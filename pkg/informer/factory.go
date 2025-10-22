package informer

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	poolv1alpha1 "kubevirt.io/api/pool/v1alpha1"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/clients/gvr"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"github.com/kubevm.io/vink/pkg/log"
	k8sv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	clonev1alpha1 "kubevirt.io/api/clone/v1alpha1"
	kubev1 "kubevirt.io/api/core/v1"
	snapshotv1beta1 "kubevirt.io/api/snapshot/v1beta1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

var unexpectedObjectError = errors.New("unexpected object")

type newSharedInformer func() cache.SharedIndexInformer

type KubeInformerFactory interface {
	// Starts any informers that have not been started yet
	// This function is thread safe and idempotent
	Start(stopCh <-chan struct{})

	// Waits for all informers to sync
	WaitForCacheSync(stopCh <-chan struct{})

	VirtualMachineInstances() cache.SharedIndexInformer

	// VirtualMachine handles the VMIs that are stopped or not running
	VirtualMachine() cache.SharedIndexInformer

	VirtualMachinePool() cache.SharedIndexInformer

	DataVolume() cache.SharedIndexInformer

	VirtualMachineSnapshot() cache.SharedIndexInformer

	VirtualMachineRestore() cache.SharedIndexInformer

	VirtualMachineClone() cache.SharedIndexInformer

	Node() cache.SharedIndexInformer

	Multus() cache.SharedIndexInformer

	Subnet() cache.SharedIndexInformer

	IPPool() cache.SharedIndexInformer

	VPC() cache.SharedIndexInformer

	ProviderNetwork() cache.SharedIndexInformer

	VLAN() cache.SharedIndexInformer

	Event() cache.SharedIndexInformer

	Namespace() cache.SharedIndexInformer

	Template() cache.SharedIndexInformer

	TemplateInstance() cache.SharedIndexInformer

	K8SInformerFactory() informers.SharedInformerFactory

	InformerForGVR(gvr schema.GroupVersionResource) (cache.SharedIndexInformer, bool)

	Informers() map[schema.GroupVersionResource]cache.SharedIndexInformer
}

type kubeInformerFactory struct {
	lock          sync.Mutex
	defaultResync time.Duration

	informers        map[schema.GroupVersionResource]cache.SharedIndexInformer
	startedInformers map[schema.GroupVersionResource]bool
	k8sInformers     informers.SharedInformerFactory
}

func NewKubeInformerFactory() KubeInformerFactory {
	return &kubeInformerFactory{
		// Resulting resync period will be between 12 and 24 hours, like the default for k8s
		defaultResync:    resyncPeriod(12 * time.Hour),
		informers:        make(map[schema.GroupVersionResource]cache.SharedIndexInformer),
		startedInformers: make(map[schema.GroupVersionResource]bool),
		k8sInformers:     informers.NewSharedInformerFactoryWithOptions(clients.Clients.KubevirtClient, 0),
	}
}

// Start can be called from multiple controllers in different go routines safely.
// Only informers that have not started are triggered by this function.
// Multiple calls to this function are idempotent.
func (f *kubeInformerFactory) Start(stopCh <-chan struct{}) {
	f.lock.Lock()
	defer f.lock.Unlock()

	for name, informer := range f.informers {
		if f.startedInformers[name] {
			// skip informers that have already started.
			log.Debugf("SKIPPING informer %T", name)
			continue
		}
		log.Infof("STARTING informer %T", name)
		go informer.Run(stopCh)
		f.startedInformers[name] = true
	}
	f.k8sInformers.Start(stopCh)
}

func (f *kubeInformerFactory) WaitForCacheSync(stopCh <-chan struct{}) {
	syncs := []cache.InformerSynced{}

	f.lock.Lock()
	for name, informer := range f.informers {
		log.Infof("Waiting for cache sync of informer %T", name)
		syncs = append(syncs, informer.HasSynced)
	}
	f.lock.Unlock()

	cache.WaitForCacheSync(stopCh, syncs...)
}

func GetVMIInformerIndexers() cache.Indexers {
	return cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		"node": func(obj interface{}) (strings []string, e error) {
			return []string{obj.(*kubev1.VirtualMachineInstance).Status.NodeName}, nil
		},
		"dv": func(obj interface{}) ([]string, error) {
			vmi, ok := obj.(*kubev1.VirtualMachineInstance)
			if !ok {
				return nil, unexpectedObjectError
			}
			var dvs []string
			for _, vol := range vmi.Spec.Volumes {
				if vol.DataVolume != nil {
					dvs = append(dvs, fmt.Sprintf("%s/%s", vmi.Namespace, vol.DataVolume.Name))
				}
			}
			return dvs, nil
		},
		"pvc": func(obj interface{}) ([]string, error) {
			vmi, ok := obj.(*kubev1.VirtualMachineInstance)
			if !ok {
				return nil, unexpectedObjectError
			}
			var pvcs []string
			for _, vol := range vmi.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil {
					pvcs = append(pvcs, fmt.Sprintf("%s/%s", vmi.Namespace, vol.PersistentVolumeClaim.ClaimName))
				}
			}
			return pvcs, nil
		},
	}
}

func (f *kubeInformerFactory) VirtualMachineInstances() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubev1.VirtualMachineInstance{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.RestClient(), "virtualmachineinstances", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubev1.VirtualMachineInstance{}, f.defaultResync, GetVMIInformerIndexers())
	})
}

func (f *kubeInformerFactory) Multus() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(netv1.NetworkAttachmentDefinition{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeNetWorldClient, "network-attachment-definitions", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &netv1.NetworkAttachmentDefinition{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) Subnet() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubeovn.Subnet{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeOVNRestClient, "subnets", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubeovn.Subnet{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) IPPool() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubeovn.IPPool{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeOVNRestClient, "ippools", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubeovn.IPPool{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VPC() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubeovn.Vpc{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeOVNRestClient, "vpcs", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubeovn.Vpc{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VLAN() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubeovn.Vlan{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeOVNRestClient, "vlans", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubeovn.Vlan{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) ProviderNetwork() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubeovn.ProviderNetwork{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeOVNRestClient, "provider-networks", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubeovn.ProviderNetwork{}, f.defaultResync, cache.Indexers{})
	})
}

func GetVirtualMachineInformerIndexers() cache.Indexers {
	return cache.Indexers{
		cache.NamespaceIndex: cache.MetaNamespaceIndexFunc,
		"dv": func(obj interface{}) ([]string, error) {
			vm, ok := obj.(*kubev1.VirtualMachine)
			if !ok {
				return nil, unexpectedObjectError
			}
			var dvs []string
			for _, vol := range vm.Spec.Template.Spec.Volumes {
				if vol.DataVolume != nil {
					dvs = append(dvs, fmt.Sprintf("%s/%s", vm.Namespace, vol.DataVolume.Name))
				}
			}
			return dvs, nil
		},
		"pvc": func(obj interface{}) ([]string, error) {
			vm, ok := obj.(*kubev1.VirtualMachine)
			if !ok {
				return nil, unexpectedObjectError
			}
			var pvcs []string
			for _, vol := range vm.Spec.Template.Spec.Volumes {
				if vol.PersistentVolumeClaim != nil {
					pvcs = append(pvcs, fmt.Sprintf("%s/%s", vm.Namespace, vol.PersistentVolumeClaim.ClaimName))
				}
			}
			return pvcs, nil
		},
	}
}

func (f *kubeInformerFactory) VirtualMachine() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(kubev1.VirtualMachine{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.RestClient(), "virtualmachines", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &kubev1.VirtualMachine{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) DataVolume() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(cdiv1.DataVolume{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.CdiClient().CdiV1beta1().RESTClient(), "datavolumes", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &cdiv1.DataVolume{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VirtualMachineSnapshot() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(snapshotv1beta1.VirtualMachineSnapshot{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.GeneratedKubeVirtClient().SnapshotV1beta1().RESTClient(), "virtualmachinesnapshots", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &snapshotv1beta1.VirtualMachineSnapshot{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VirtualMachineRestore() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(snapshotv1beta1.VirtualMachineRestore{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.GeneratedKubeVirtClient().SnapshotV1beta1().RESTClient(), "virtualmachinerestores", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &snapshotv1beta1.VirtualMachineRestore{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VirtualMachineClone() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(clonev1alpha1.VirtualMachineClone{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.GeneratedKubeVirtClient().CloneV1alpha1().RESTClient(), "virtualmachineclones", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &clonev1alpha1.VirtualMachineClone{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) VirtualMachinePool() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(poolv1alpha1.VirtualMachinePool{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubevirtClient.GeneratedKubeVirtClient().PoolV1alpha1().RESTClient(), "virtualmachinepools", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &poolv1alpha1.VirtualMachinePool{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) Node() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(k8sv1.Node{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeRestClient, "nodes", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &k8sv1.Node{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) Event() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(k8sv1.Event{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeRestClient, "events", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &k8sv1.Event{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) Namespace() cache.SharedIndexInformer {
	return f.getInformer(gvr.From(k8sv1.Namespace{}), func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.KubeRestClient, "namespaces", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &k8sv1.Namespace{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) Template() cache.SharedIndexInformer {
	gvr, _ := dynamicx.ResolveGVRAndGVK(v1alpha1.VirtualMachineTemplate{})
	return f.getInformer(gvr, func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.VinkRestClient, "virtualmachinetemplates", k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &v1alpha1.VirtualMachineTemplate{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) TemplateInstance() cache.SharedIndexInformer {
	gvr, _ := dynamicx.ResolveGVRAndGVK(v1alpha1.TemplateInstance{})
	return f.getInformer(gvr, func() cache.SharedIndexInformer {
		lw := cache.NewListWatchFromClient(clients.Clients.VinkRestClient, gvr.Resource, k8sv1.NamespaceAll, fields.Everything())
		return cache.NewSharedIndexInformer(lw, &v1alpha1.TemplateInstance{}, f.defaultResync, cache.Indexers{})
	})
}

func (f *kubeInformerFactory) K8SInformerFactory() informers.SharedInformerFactory {
	return f.k8sInformers
}

func (f *kubeInformerFactory) InformerForGVR(gvr schema.GroupVersionResource) (cache.SharedIndexInformer, bool) {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, ok := f.informers[gvr]
	return informer, ok
}

func (f *kubeInformerFactory) Informers() map[schema.GroupVersionResource]cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informers := make(map[schema.GroupVersionResource]cache.SharedIndexInformer, len(f.informers))
	for gvr, informer := range f.informers {
		informers[gvr] = informer
	}
	return informers
}

// internal function used to retrieve an already created informer
// or create a new informer if one does not already exist.
// Thread safe
func (f *kubeInformerFactory) getInformer(key schema.GroupVersionResource, newFunc newSharedInformer) cache.SharedIndexInformer {
	f.lock.Lock()
	defer f.lock.Unlock()

	informer, exists := f.informers[key]
	if exists {
		return informer
	}
	informer = newFunc()
	f.informers[key] = informer

	return informer
}

// resyncPeriod computes the time interval a shared informer waits before resyncing with the api server
func resyncPeriod(minResyncPeriod time.Duration) time.Duration {
	// #nosec no need for better randomness
	factor := rand.Float64() + 1
	return time.Duration(float64(minResyncPeriod.Nanoseconds()) * factor)
}
