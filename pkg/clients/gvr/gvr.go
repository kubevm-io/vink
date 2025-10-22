package gvr

import (
	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clonev1alpha1 "kubevirt.io/api/clone/v1alpha1"
	virtv1 "kubevirt.io/api/core/v1"
	poolv1alpha1 "kubevirt.io/api/pool/v1alpha1"
	snapshotv1beta1 "kubevirt.io/api/snapshot/v1beta1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
)

func From[T any](o T) schema.GroupVersionResource {
	switch any(o).(type) {
	case cdiv1beta1.DataVolume, *cdiv1beta1.DataVolume:
		return schema.GroupVersionResource{
			Group:    cdiv1beta1.SchemeGroupVersion.Group,
			Version:  cdiv1beta1.SchemeGroupVersion.Version,
			Resource: "datavolumes",
		}
	case corev1.PersistentVolumeClaim, *corev1.PersistentVolumeClaim:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "persistentvolumeclaims",
		}
	case corev1.ConfigMap, *corev1.ConfigMap:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "configmaps",
		}
	case storagev1.StorageClass, *storagev1.StorageClass:
		return schema.GroupVersionResource{
			Group:    storagev1.SchemeGroupVersion.Group,
			Version:  storagev1.SchemeGroupVersion.Version,
			Resource: "storageclasses",
		}
	case corev1.Secret, *corev1.Secret:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "secrets",
		}
	case corev1.Node, *corev1.Node:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "nodes",
		}
	case virtv1.VirtualMachine, *virtv1.VirtualMachine:
		return schema.GroupVersionResource{
			Group:    virtv1.SchemeGroupVersion.Group,
			Version:  virtv1.SchemeGroupVersion.Version,
			Resource: "virtualmachines",
		}
	case poolv1alpha1.VirtualMachinePool, *poolv1alpha1.VirtualMachinePool:
		return schema.GroupVersionResource{
			Group:    poolv1alpha1.SchemeGroupVersion.Group,
			Version:  poolv1alpha1.SchemeGroupVersion.Version,
			Resource: "virtualmachinepools",
		}
	case virtv1.VirtualMachineInstance, *virtv1.VirtualMachineInstance:
		return schema.GroupVersionResource{
			Group:    virtv1.SchemeGroupVersion.Group,
			Version:  virtv1.SchemeGroupVersion.Version,
			Resource: "virtualmachineinstances",
		}
	case corev1.Namespace, *corev1.Namespace:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "namespaces",
		}
	case netv1.NetworkAttachmentDefinition, *netv1.NetworkAttachmentDefinition:
		return schema.GroupVersionResource{
			Group:    netv1.SchemeGroupVersion.Group,
			Version:  netv1.SchemeGroupVersion.Version,
			Resource: "network-attachment-definitions",
		}
	case kubeovn.Subnet, *kubeovn.Subnet:
		return kubeovn.SchemeGroupVersion.WithResource("subnets")
	case kubeovn.Vpc, *kubeovn.Vpc:
		return kubeovn.SchemeGroupVersion.WithResource("vpcs")
	case kubeovn.IPPool, *kubeovn.IPPool:
		return kubeovn.SchemeGroupVersion.WithResource("ippools")
	case kubeovn.IP, *kubeovn.IP:
		return kubeovn.SchemeGroupVersion.WithResource("ips")
	case v1alpha1.VirtualMachineSummary, *v1alpha1.VirtualMachineSummary:
		return v1alpha1.VirtualMachineSummaryGVR
	case corev1.Event, *corev1.Event:
		return schema.GroupVersionResource{
			Group:    corev1.SchemeGroupVersion.Group,
			Version:  corev1.SchemeGroupVersion.Version,
			Resource: "events",
		}
	case snapshotv1beta1.VirtualMachineSnapshot, *snapshotv1beta1.VirtualMachineSnapshot:
		return schema.GroupVersionResource{
			Group:    snapshotv1beta1.SchemeGroupVersion.Group,
			Version:  snapshotv1beta1.SchemeGroupVersion.Version,
			Resource: "virtualmachinesnapshots",
		}
	case snapshotv1beta1.VirtualMachineRestore, *snapshotv1beta1.VirtualMachineRestore:
		return schema.GroupVersionResource{
			Group:    snapshotv1beta1.SchemeGroupVersion.Group,
			Version:  snapshotv1beta1.SchemeGroupVersion.Version,
			Resource: "virtualmachinerestores",
		}
	case clonev1alpha1.VirtualMachineClone, *clonev1alpha1.VirtualMachineClone:
		return schema.GroupVersionResource{
			Group:    clonev1alpha1.SchemeGroupVersion.Group,
			Version:  clonev1alpha1.SchemeGroupVersion.Version,
			Resource: "virtualmachineclones",
		}
	case kubeovn.Vlan, *kubeovn.Vlan:
		return kubeovn.SchemeGroupVersion.WithResource("vlans")
	case kubeovn.ProviderNetwork, *kubeovn.ProviderNetwork:
		return kubeovn.SchemeGroupVersion.WithResource("provider-networks")
	case v1alpha1.VirtualMachineTemplate, *v1alpha1.VirtualMachineTemplate, types.Template, *types.Template:
		return v1alpha1.SchemeGroupVersion.WithResource("virtualmachinetemplates")
	}

	return schema.GroupVersionResource{}
}

func ResolveGVR(rt types.ResourceType) schema.GroupVersionResource {
	switch rt {
	case types.ResourceType_VIRTUAL_MACHINE:
		return From(virtv1.VirtualMachine{})
	case types.ResourceType_VIRTUAL_MACHINE_INSTANCE:
		return From(virtv1.VirtualMachineInstance{})
	case types.ResourceType_DATA_VOLUME:
		return From(cdiv1beta1.DataVolume{})
	case types.ResourceType_NODE:
		return From(corev1.Node{})
	case types.ResourceType_NAMESPACE:
		return From(corev1.Namespace{})
	case types.ResourceType_MULTUS:
		return From(netv1.NetworkAttachmentDefinition{})
	case types.ResourceType_SUBNET:
		return From(kubeovn.Subnet{})
	case types.ResourceType_VPC:
		return From(kubeovn.Vpc{})
	case types.ResourceType_IPPOOL:
		return From(kubeovn.IPPool{})
	case types.ResourceType_STORAGE_CLASS:
		return From(storagev1.StorageClass{})
	case types.ResourceType_IPS:
		return From(kubeovn.IP{})
	case types.ResourceType_VIRTUAL_MACHINE_SUMMARY:
		return From(v1alpha1.VirtualMachineSummary{})
	case types.ResourceType_EVENT:
		return From(corev1.Event{})
	case types.ResourceType_VIRTUAL_MACHINE_SNAPSHOT:
		return From(snapshotv1beta1.VirtualMachineSnapshot{})
	case types.ResourceType_VIRTUAL_MACHINE_RESTORE:
		return From(snapshotv1beta1.VirtualMachineRestore{})
	case types.ResourceType_VIRTUAL_MACHINE_CLONE:
		return From(clonev1alpha1.VirtualMachineClone{})
	case types.ResourceType_PROVIDER_NETWORK:
		return From(kubeovn.ProviderNetwork{})
	case types.ResourceType_VLAN:
		return From(kubeovn.Vlan{})
	case types.ResourceType_VIRTUAL_MACHINE_POOL:
		return From(poolv1alpha1.VirtualMachinePool{})
	}

	return schema.GroupVersionResource{}
}
