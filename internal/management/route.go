package management

import (
	"github.com/gorilla/mux"
	nad_v1alpha1 "github.com/kubevm.io/vink/apis/management/nad/v1alpha1"
	clone_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachineclone/v1alpha1"
	snapshot_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinesnapshot/v1alpha1"
	restore_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinerestore/v1alpha1"
	subnet_v1alpha1 "github.com/kubevm.io/vink/apis/management/subnet/v1alpha1"
	sc_v1alpha1 "github.com/kubevm.io/vink/apis/management/sc/v1alpha1"
	namespace_v1alpha1 "github.com/kubevm.io/vink/apis/management/namespace/v1alpha1"
	node_v1alpha1 "github.com/kubevm.io/vink/apis/management/node/v1alpha1"
	resource_v1alpha1 "github.com/kubevm.io/vink/apis/management/resource/v1alpha1"
	storage_device_v1alpha1 "github.com/kubevm.io/vink/apis/management/storage_device/v1alpha1"
	template_v1alpha1 "github.com/kubevm.io/vink/apis/management/template/v1alpha1"
	template_instance_v1alpha1 "github.com/kubevm.io/vink/apis/management/template_instance/v1alpha1"
	vmv1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachine/v1alpha1"
	"github.com/kubevm.io/vink/internal/management/nad"
	"github.com/kubevm.io/vink/internal/management/subnet"
	"github.com/kubevm.io/vink/internal/management/sc"
	"github.com/kubevm.io/vink/internal/management/namespace"
	"github.com/kubevm.io/vink/internal/management/node"
	"github.com/kubevm.io/vink/internal/management/virtualmachinerestore"
	"github.com/kubevm.io/vink/internal/management/virtualmachinesnapshot"
	"github.com/kubevm.io/vink/internal/management/virtualmachineclone"
	"github.com/kubevm.io/vink/internal/management/resource"
	"github.com/kubevm.io/vink/internal/management/storage_device"
	"github.com/kubevm.io/vink/internal/management/template"
	templateinstance "github.com/kubevm.io/vink/internal/management/template_instance"
	"github.com/kubevm.io/vink/internal/management/virtualmachine"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/informer"
	"google.golang.org/grpc/reflection"
)

func RegisterGRPCRoutes(kubeInformerFactory informer.KubeInformerFactory) (func(s reflection.GRPCServer), error) {
	return func(s reflection.GRPCServer) {
		resource_v1alpha1.RegisterResourceWatchManagementServer(s, resource.NewResourceWatchManagement(kubeInformerFactory))
		resource_v1alpha1.RegisterResourceManagementServer(s, resource.NewResourceManagement())
		vmv1alpha1.RegisterVirtualMachineManagementServer(s, virtualmachine.NewVirtualMachineManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		template_v1alpha1.RegisterTemplateManagementServer(s, template.NewTemplateManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		template_instance_v1alpha1.RegisterTemplateInstanceManagementServer(s, templateinstance.NewTemplateInstanceManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		node_v1alpha1.RegisterNodeManagementServer(s, node.NewNodeManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		storage_device_v1alpha1.RegisterStorageDeviceManagementServer(s, storage_device.NewStorageDeviceManagement(kubeInformerFactory, clients.Clients.DynamicClient(), clients.Clients.Ceph))
		namespace_v1alpha1.RegisterNamespaceManagementServer(s, namespace.NewNamespaceManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		nad_v1alpha1.RegisterNetworkAttachmentDefinitionManagementServer(s, nad.NewNetworkAttachmentDefinitionManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		subnet_v1alpha1.RegisterSubnetManagementServer(s, subnet.NewSubnetManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		sc_v1alpha1.RegisterStorageClassManagementServer(s, sc.NewStorageClassManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		clone_v1alpha1.RegisterVirtualMachineCloneManagementServer(s, virtualmachineclone.NewVirtualMachineCloneManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		snapshot_v1alpha1.RegisterVirtualMachineSnapshotManagementServer(s, virtualmachinesnapshot.NewVirtualMachineSnapshotManagement(kubeInformerFactory, clients.Clients.DynamicClient()))
		restore_v1alpha1.RegisterVirtualMachineRestoreManagementServer(s, virtualmachinerestore.NewVirtualMachineRestoreManagement(kubeInformerFactory, clients.Clients.DynamicClient()))

		reflection.Register(s)
	}, nil
}

func RegisterHTTPRoutes() (func(r *mux.Router), error) {
	return func(router *mux.Router) {
		virtualmachine.RegisterSerialConsole(router)
	}, nil
}
