package virtualmachinesnapshot

import (
	"context"

	virtualmachinesnapshot_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinesnapshot/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/virtualmachinesnapshot/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewVirtualMachineSnapshotManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) virtualmachinesnapshot_v1alpha1.VirtualMachineSnapshotManagementServer {
	return &snapshotManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicSnapshot:        dynamicx.NewClient[*types.VirtualMachineSnapshot](dynamicClient),
	}
}

type snapshotManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicSnapshot        *dynamicx.Client[*types.VirtualMachineSnapshot]

	virtualmachinesnapshot_v1alpha1.UnsafeVirtualMachineSnapshotManagementServer
}

func (m *snapshotManagement) Watch(request *types.WatchRequest, server virtualmachinesnapshot_v1alpha1.VirtualMachineSnapshotManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewVirtualMachineSnapshotSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *snapshotManagement) Create(ctx context.Context, request *types.VirtualMachineSnapshot) (*types.VirtualMachineSnapshot, error) {
	return m.dynamicSnapshot.Create(ctx, request)
}

func (m *snapshotManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.VirtualMachineSnapshot, error) {
	return m.dynamicSnapshot.Get(ctx, request.Namespace, request.Name)
}

func (m *snapshotManagement) List(ctx context.Context, request *types.ListRequest) (*types.VirtualMachineSnapshotList, error) {
	result, err := m.dynamicSnapshot.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.VirtualMachineSnapshotList{Items: result}, nil
}

func (m *snapshotManagement) Update(ctx context.Context, request *types.VirtualMachineSnapshot) (*types.VirtualMachineSnapshot, error) {
	return m.dynamicSnapshot.Update(ctx, request)
}

func (m *snapshotManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicSnapshot.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
