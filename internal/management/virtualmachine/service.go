package virtualmachine

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	vmv1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachine/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/virtualmachine/business"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/log"
	"github.com/kubevm.io/vink/pkg/watcher"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
)

func NewVirtualMachineManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) vmv1alpha1.VirtualMachineManagementServer {
	return &virtualMachineManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicVm:           dynamicx.NewClient[*types.VirtualMachine](dynamicClient),
	}
}

type virtualMachineManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicVm           *dynamicx.Client[*types.VirtualMachine]

	vmv1alpha1.UnimplementedVirtualMachineManagementServer
}

func (m *virtualMachineManagement) Watch(request *types.WatchRequest, server vmv1alpha1.VirtualMachineManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewVirtualMachineSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *virtualMachineManagement) Create(ctx context.Context, request *types.VirtualMachine) (*types.VirtualMachine, error) {
	return m.dynamicVm.Create(ctx, request)
}

func (m *virtualMachineManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.VirtualMachine, error) {
	return m.dynamicVm.Get(ctx, request.Namespace, request.Name)
}

func (m *virtualMachineManagement) List(ctx context.Context, request *types.ListRequest) (*types.VirtualMachineList, error) {
	result, err := m.dynamicVm.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.VirtualMachineList{Items: result}, nil
}

func (m *virtualMachineManagement) Update(ctx context.Context, request *types.VirtualMachine) (*types.VirtualMachine, error) {
	return m.dynamicVm.Update(ctx, request)
}

func (m *virtualMachineManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicVm.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (m *virtualMachineManagement) PowerState(ctx context.Context, request *vmv1alpha1.PowerStateRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, business.PowerState(ctx, request.NamespaceName, request.PowerState)
}

func RegisterSerialConsole(router *mux.Router) {
	router.PathPrefix(business.SerialConsoleRequestPathTmpl).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		vars := mux.Vars(request)
		namespace, name := vars["namespace"], vars["name"]
		if len(namespace) == 0 || len(name) == 0 {
			log.Errorf("namespace or name is empty")
			writer.WriteHeader(http.StatusBadRequest)
			return
		}

		kubevirtRestConfig := clients.Clients.KubevirtClient.Config()

		parse, err := url.Parse(kubevirtRestConfig.Host)
		if err != nil {
			log.Errorf("Failed to parse kubevirt host: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		ws := fmt.Sprintf("wss://%s/apis/subresources.kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s/console", parse.Host, namespace, name)

		dialer := websocket.Dialer{
			HandshakeTimeout: 15 * time.Second,
			TLSClientConfig:  generateSerialConsoleTLSConfig(kubevirtRestConfig),
		}

		serverConnHeader := http.Header{}
		if len(kubevirtRestConfig.BearerToken) > 0 {
			log.Debug("Using Bearer token for serial console")
			serverConnHeader.Set("Authorization", fmt.Sprintf("Bearer %s", kubevirtRestConfig.BearerToken))
		}
		serverConn, _, err := dialer.Dial(ws, serverConnHeader)
		if err != nil {
			log.Errorf("Failed to dial server: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer serverConn.Close()

		upgrader := websocket.Upgrader{
			HandshakeTimeout: 15 * time.Second,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		clientConn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			log.Errorf("Failed to upgrade client: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer clientConn.Close()

		eg := errgroup.Group{}
		eg.Go(func() error {
			if _, err := io.Copy(clientConn.UnderlyingConn(), serverConn.UnderlyingConn()); err != nil {
				log.Errorf("Failed to copy data from server to client: %v", err)
				return err
			}
			return nil
		})
		eg.Go(func() error {
			if _, err := io.Copy(serverConn.UnderlyingConn(), clientConn.UnderlyingConn()); err != nil {
				log.Errorf("Failed to copy data from client to server: %v", err)
				return err
			}
			return nil
		})

		if err := eg.Wait(); err != nil {
			log.Errorf("Failed to copy data: %v", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}

func generateSerialConsoleTLSConfig(restConfig *rest.Config) *tls.Config {
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
		ClientAuth:         tls.NoClientCert,
	}

	if len(restConfig.CertData) == 0 || len(restConfig.KeyData) == 0 {
		return &tlsConfig
	}

	log.Debug("Using TLS client certs for serial console")
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return func(restConfig *rest.Config) (*tls.Certificate, error) {
			certBytes := restConfig.CertData
			keyBytes := restConfig.KeyData

			crt, err := tls.X509KeyPair(certBytes, keyBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to load certificate: %v", err)
			}
			leaf, err := cert.ParseCertsPEM(certBytes)
			if err != nil {
				return nil, fmt.Errorf("failed to load leaf certificate: %v", err)
			}
			crt.Leaf = leaf[0]
			return &crt, nil
		}(restConfig)
	}
	return &tlsConfig
}

// func RegisterSerialConsole(router *mux.Router, client kubecli.KubevirtClient) {
// 	router.PathPrefix(business.SerialConsoleRequestPathTmpl).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
// 		vars := mux.Vars(request)
// 		namespace, name := vars["namespace"], vars["name"]
// 		if namespace == "" || name == "" {
// 			http.Error(writer, "Missing namespace or name", http.StatusBadRequest)
// 			return
// 		}

// 		upgrader := websocket.Upgrader{
// 			HandshakeTimeout: 15 * time.Second,
// 			CheckOrigin: func(r *http.Request) bool {
// 				return true
// 			},
// 		}

// 		wsConn, err := upgrader.Upgrade(writer, request, nil)
// 		if err != nil {
// 			log.Errorf("Failed to upgrade connection: %v", err)
// 			return
// 		}
// 		defer wsConn.Close()

// 		stdinReader, stdinWriter := io.Pipe()
// 		stdoutReader, stdoutWriter := io.Pipe()
// 		resChan := make(chan error, 1)

// 		go func() {
// 			con, err := client.VirtualMachineInstance(namespace).SerialConsole(name,
// 				&kvcorev1.SerialConsoleOptions{
// 					ConnectionTimeout: 5 * time.Minute,
// 				})
// 			if err != nil {
// 				log.Errorf("Failed to connect to VMI console: %v", err)
// 				resChan <- err
// 				return
// 			}

// 			resChan <- con.Stream(kvcorev1.StreamOptions{
// 				In:  stdinReader,
// 				Out: stdoutWriter,
// 			})
// 		}()

// 		go func() {
// 			buf := make([]byte, 1024)
// 			for {
// 				n, err := stdoutReader.Read(buf)
// 				if err != nil {
// 					log.Warnf("Read from console error: %v", err)
// 					break
// 				}
// 				err = wsConn.WriteMessage(websocket.BinaryMessage, buf[:n])
// 				if err != nil {
// 					log.Warnf("Write to WebSocket error: %v", err)
// 					break
// 				}
// 			}
// 			_ = wsConn.Close()
// 		}()

// 		go func() {
// 			for {
// 				msgType, data, err := wsConn.ReadMessage()
// 				if err != nil {
// 					log.Warnf("Read from WebSocket error: %v", err)
// 					break
// 				}
// 				if msgType != websocket.BinaryMessage {
// 					log.Warnf("Unsupported WebSocket message type: %v", msgType)
// 					continue
// 				}
// 				_, err = stdinWriter.Write(data)
// 				if err != nil {
// 					log.Warnf("Write to console error: %v", err)
// 					break
// 				}
// 			}
// 			_ = wsConn.Close()
// 		}()

// 		if err := <-resChan; err != nil {
// 			log.Errorf("console stream ended with error: %v", err)
// 			_ = wsConn.WriteMessage(websocket.TextMessage, []byte("console stream error: "+err.Error()))
// 		}
// 	})
// }
