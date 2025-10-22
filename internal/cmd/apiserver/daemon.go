package apiserver

import (
	"context"
	"fmt"

	"github.com/kubevm.io/vink/config"
	"github.com/kubevm.io/vink/internal/management"

	grpcwebproxy "github.com/kubevm.io/vink/internal/pkg/grpc-web-proxy"
	"github.com/kubevm.io/vink/internal/pkg/servers"

	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/log"
)

func New(cfg *config.Config) *Daemon {
	return &Daemon{config: cfg}
}

type Daemon struct {
	informerFactory informer.KubeInformerFactory

	grpcServer servers.Server
	httpServer servers.Server

	config *config.Config
}

func (dm *Daemon) Execute(ctx context.Context) error {
	dm.informerFactory = informer.NewKubeInformerFactory()
	_ = dm.informerFactory.VirtualMachine()
	_ = dm.informerFactory.VirtualMachineInstances()
	_ = dm.informerFactory.DataVolume()
	_ = dm.informerFactory.VirtualMachineSnapshot()
	_ = dm.informerFactory.VirtualMachineRestore()
	_ = dm.informerFactory.VirtualMachineClone()
	_ = dm.informerFactory.VirtualMachinePool()
	_ = dm.informerFactory.Multus()
	_ = dm.informerFactory.Subnet()
	_ = dm.informerFactory.IPPool()
	_ = dm.informerFactory.VPC()
	_ = dm.informerFactory.VLAN()
	_ = dm.informerFactory.ProviderNetwork()
	_ = dm.informerFactory.Event()
	_ = dm.informerFactory.Namespace()
	_ = dm.informerFactory.Node()
	// _ = dm.informerFactory.Template()
	// _ = dm.informerFactory.TemplateInstance()

	dm.informerFactory.Start(ctx.Done())
	dm.informerFactory.WaitForCacheSync(ctx.Done())

	register, err := management.RegisterGRPCRoutes(dm.informerFactory)
	if err != nil {
		return err
	}

	httpAddress := fmt.Sprintf(":%v", dm.config.APIServerHTTP)
	grpcAddress := fmt.Sprintf(":%v", dm.config.APIServerGRPC)

	dm.grpcServer = servers.NewGRPCServer(grpcAddress, register)
	log.Infof("Starting gRPC server at: %s", grpcAddress)

	errCh := make(chan error)
	go func() {
		if err := dm.grpcServer.Run(); err != nil {
			errCh <- err
		}
	}()

	httpRegister, err := management.RegisterHTTPRoutes()
	if err != nil {
		return err
	}
	dm.httpServer = servers.NewHTTPServer("apiserver", httpAddress, httpRegister)
	log.Infof("Starting http server at: %s", httpAddress)
	go func() {
		if err := dm.httpServer.Run(); err != nil {
			errCh <- err
		}
	}()

	grpcweb := grpcwebproxy.NewDetaultProxy(dm.config)

	go func() {
		if err := grpcweb.Run(ctx); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (dm *Daemon) Shutdown() error {
	if err := dm.grpcServer.Stop(); err != nil {
		log.Errorf("Failed to stop grpc server: %v", err)
	}
	return dm.httpServer.Stop()
}
