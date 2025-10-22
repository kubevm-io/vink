package clients

import (
	"context"
	"encoding/json"
	"fmt"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovnv1 "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubevm.io/vink/config"
	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	spv2beta1 "github.com/spidernet-io/spiderpool/pkg/k8s/apis/spiderpool.spidernet.io/v2beta1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	clonev1alpha1 "kubevirt.io/api/clone/v1alpha1"
	virtv1 "kubevirt.io/api/core/v1"
	poolv1alpha1 "kubevirt.io/api/pool/v1alpha1"
	snapshotv1beta1 "kubevirt.io/api/snapshot/v1beta1"
	"kubevirt.io/client-go/kubecli"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func init() {
	cdiv1beta1.AddToScheme(scheme.Scheme)
	spv2beta1.AddToScheme(scheme.Scheme)
	virtv1.AddToScheme(scheme.Scheme)
	netv1.AddToScheme(scheme.Scheme)
	kubeovnv1.AddToScheme(scheme.Scheme)
	storagev1.AddToScheme(scheme.Scheme)
	v1alpha1.AddToScheme(scheme.Scheme)
	snapshotv1beta1.AddToScheme(scheme.Scheme)
	clonev1alpha1.AddToScheme(scheme.Scheme)
	poolv1alpha1.AddToScheme(scheme.Scheme)
}

var Clients = &clients{}

type clients struct {
	kubecli.KubevirtClient
	Ceph       CephInterface
	Prometheus *Prometheus

	VinkRestClient     *rest.RESTClient
	KubeOVNRestClient  *rest.RESTClient
	KubeRestClient     *rest.RESTClient
	KubeNetWorldClient *rest.RESTClient
}

func InitClients(ctx context.Context, cfg *config.Config) error {
	kubeconfig := GetK8sConfigConfigWithFile()

	kubeconfig.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(100, 200)

	vinkRestClient, err := newRestClientFromRESTConfig(kubeconfig, &v1alpha1.GroupVersion)
	if err != nil {
		return err
	}
	Clients.VinkRestClient = vinkRestClient

	kubeovnRestClient, err := newRestClientFromRESTConfig(kubeconfig, &kubeovnv1.SchemeGroupVersion)
	if err != nil {
		return err
	}
	Clients.KubeOVNRestClient = kubeovnRestClient

	kubevirtClient, err := kubecli.GetKubevirtClientFromRESTConfig(kubeconfig)
	if err != nil {
		return err
	}
	Clients.KubevirtClient = kubevirtClient

	kubeRestClient, err := newRestClientFromRESTConfig(kubeconfig, &corev1.SchemeGroupVersion)
	if err != nil {
		return err
	}
	Clients.KubeRestClient = kubeRestClient

	kubeNetworkCLient, err := newRestClientFromRESTConfig(kubeconfig, &netv1.SchemeGroupVersion)
	if err != nil {
		return err
	}
	Clients.KubeNetWorldClient = kubeNetworkCLient

	Clients.Prometheus, err = NewPrometheus(cfg.Prometheus)
	if err != nil {
		return err
	}

	// cephPassword := cfg.CephPassword
	// if len(cephPassword) == 0 {
	// 	cephSecret, err := Clients.CoreV1().Secrets(cfg.CephPasswordSecretNamespace).Get(ctx, cfg.CephPasswordSecretName, metav1.GetOptions{})
	// 	if err != nil {
	// 		return err
	// 	}
	// 	cephPassword = string(cephSecret.Data["password"])
	// }

	// Clients.Ceph, err = NewCeph(ctx, cfg.Ceph, cfg.CephUsername, cephPassword)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func newRestClientFromRESTConfig(kubeconfig *rest.Config, gv *schema.GroupVersion) (*rest.RESTClient, error) {
	shallowCopy := *kubeconfig
	if len(gv.Group) == 0 {
		shallowCopy.APIPath = "/api"
	} else {
		shallowCopy.APIPath = "/apis"
	}
	shallowCopy.GroupVersion = gv
	shallowCopy.NegotiatedSerializer = serializer.WithoutConversionCodecFactory{CodecFactory: scheme.Codecs}
	return rest.RESTClientFor(&shallowCopy)
}

func FromUnstructuredList[T any](obj *unstructured.UnstructuredList) (*T, error) {
	typedObj := new(T)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), typedObj); err != nil {
		return nil, err
	}

	return typedObj, nil
}

func FromUnstructured[T any](obj *unstructured.Unstructured) (*T, error) {
	typedObj := new(T)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), typedObj); err != nil {
		return nil, err
	}

	return typedObj, nil
}

func GetGVK(obj runtime.Object) (schema.GroupVersionKind, error) {
	gvks, _, _ := scheme.Scheme.ObjectKinds(obj)
	if len(gvks) < 1 {
		return schema.GroupVersionKind{}, fmt.Errorf("no gvk found")
	}
	return gvks[0], nil
}

func Unstructured[T runtime.Object](obj T) (*unstructured.Unstructured, error) {
	gvk, err := GetGVK(obj)
	if err != nil {
		return nil, fmt.Errorf("no gvk found")
	}
	un := &unstructured.Unstructured{}
	c, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	un.SetUnstructuredContent(c)
	un.SetAPIVersion(gvk.GroupVersion().String())
	un.SetKind(gvk.Kind)
	return un, nil
}

func UnstructuredToJSON(obj *unstructured.Unstructured) (string, error) {
	obj.SetManagedFields(nil)
	jsonBytes, err := json.Marshal(obj.Object)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func JSONToUnstructured(data string) (*unstructured.Unstructured, error) {
	obj := map[string]interface{}{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return nil, err
	}

	un := &unstructured.Unstructured{}
	un.SetUnstructuredContent(obj)
	return un, nil
}

func InterfaceToUnstructured(obj any) (*unstructured.Unstructured, error) {
	un := &unstructured.Unstructured{}
	c, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	un.SetUnstructuredContent(c)
	return un, nil
}

func InterfaceToJSON(obj any) (string, error) {
	var un *unstructured.Unstructured
	var err error
	switch payload := obj.(type) {
	case *virtv1.VirtualMachine:
		un, err = Unstructured(payload)
	case *v1alpha1.VirtualMachineSummary:
		un, err = Unstructured(payload)
	case *cdiv1beta1.DataVolume:
		un, err = Unstructured(payload)
	case *corev1.Event:
		un, err = Unstructured(payload)
	case *kubeovnv1.IPPool:
		un, err = Unstructured(payload)
	case *kubeovnv1.Vpc:
		un, err = Unstructured(payload)
	case *kubeovnv1.Subnet:
		un, err = Unstructured(payload)
	case *netv1.NetworkAttachmentDefinition:
		un, err = Unstructured(payload)
	case *corev1.Namespace:
		un, err = Unstructured(payload)
	case *corev1.Node:
		un, err = Unstructured(payload)
	case *snapshotv1beta1.VirtualMachineSnapshot:
		un, err = Unstructured(payload)
	case *snapshotv1beta1.VirtualMachineRestore:
		un, err = Unstructured(payload)
	case *clonev1alpha1.VirtualMachineClone:
		un, err = Unstructured(payload)
	case *kubeovnv1.ProviderNetwork:
		un, err = Unstructured(payload)
	case *kubeovnv1.Vlan:
		un, err = Unstructured(payload)
	case *poolv1alpha1.VirtualMachinePool:
		un, err = Unstructured(payload)
	default:
		err = fmt.Errorf("unsupported payload type %T", payload)
	}
	if err != nil {
		return "", err
	}

	jsonBytes, err := UnstructuredToJSON(un)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), err
}
