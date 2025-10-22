package dynamicx

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
	snapshotv1beta1 "kubevirt.io/api/snapshot/v1beta1"
)

func ResolveGVRAndGVK[T any](o T) (schema.GroupVersionResource, schema.GroupVersionKind) {
	switch any(o).(type) {
	case virtv1.VirtualMachine, *virtv1.VirtualMachine, types.VirtualMachine, *types.VirtualMachine:
		return virtv1.GroupVersion.WithResource("virtualmachines"), virtv1.GroupVersion.WithKind("VirtualMachine")
	case v1alpha1.VirtualMachineTemplate, *v1alpha1.VirtualMachineTemplate, types.Template, *types.Template:
		return v1alpha1.SchemeGroupVersion.WithResource("virtualmachinetemplates"), v1alpha1.SchemeGroupVersion.WithKind("VirtualMachineTemplate")
	case v1alpha1.TemplateInstance, *v1alpha1.TemplateInstance, types.TemplateInstance, *types.TemplateInstance:
		return v1alpha1.SchemeGroupVersion.WithResource("templateinstances"), v1alpha1.SchemeGroupVersion.WithKind("TemplateInstance")
	case corev1.Node, *corev1.Node, types.Node, *types.Node:
		return corev1.SchemeGroupVersion.WithResource("nodes"), corev1.SchemeGroupVersion.WithKind("Node")
	case corev1.Namespace, *corev1.Namespace, types.Namespace, *types.Namespace:
		return corev1.SchemeGroupVersion.WithResource("namespaces"), corev1.SchemeGroupVersion.WithKind("Namespace")
	case kubeovn.Subnet, *kubeovn.Subnet, types.Subnet, *types.Subnet:
		return kubeovn.SchemeGroupVersion.WithResource("subnets"), kubeovn.SchemeGroupVersion.WithKind("Subnet")
	case netv1.NetworkAttachmentDefinition, *netv1.NetworkAttachmentDefinition, types.NetworkAttachmentDefinition, *types.NetworkAttachmentDefinition:
		return netv1.SchemeGroupVersion.WithResource("network-attachment-definitions"), netv1.SchemeGroupVersion.WithKind("NetworkAttachmentDefinition")
	case storagev1.StorageClass, *storagev1.StorageClass, types.StorageClass, *types.StorageClass:
		return storagev1.SchemeGroupVersion.WithResource("storageclasses"), storagev1.SchemeGroupVersion.WithKind("StorageClass")
	case clonev1alpha1.VirtualMachineClone, *clonev1alpha1.VirtualMachineClone, types.VirtualMachineClone, *types.VirtualMachineClone:
		return clonev1alpha1.SchemeGroupVersion.WithResource("virtualmachineclones"), clonev1alpha1.SchemeGroupVersion.WithKind("VirtualMachineClone")
	case snapshotv1beta1.VirtualMachineSnapshot, *snapshotv1beta1.VirtualMachineSnapshot, types.VirtualMachineSnapshot, *types.VirtualMachineSnapshot:
		return snapshotv1beta1.SchemeGroupVersion.WithResource("virtualmachinesnapshots"), snapshotv1beta1.SchemeGroupVersion.WithKind("VirtualMachineSnapshot")
	case snapshotv1beta1.VirtualMachineRestore, *snapshotv1beta1.VirtualMachineRestore, types.VirtualMachineRestore, *types.VirtualMachineRestore:
		return snapshotv1beta1.SchemeGroupVersion.WithResource("virtualmachinerestores"), snapshotv1beta1.SchemeGroupVersion.WithKind("VirtualMachineRestore")
	}

	return schema.GroupVersionResource{}, schema.GroupVersionKind{}
}
