package virtualmachinerestore

import (
	"context"

	virtualmachinerestore_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinerestore/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/virtualmachinerestore/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewVirtualMachineRestoreManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) virtualmachinerestore_v1alpha1.VirtualMachineRestoreManagementServer {
	return &snapshotManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicRestore:        dynamicx.NewClient[*types.VirtualMachineRestore](dynamicClient),
	}
}

type snapshotManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicRestore        *dynamicx.Client[*types.VirtualMachineRestore]

	virtualmachinerestore_v1alpha1.UnsafeVirtualMachineRestoreManagementServer
}

func (m *snapshotManagement) Watch(request *types.WatchRequest, server virtualmachinerestore_v1alpha1.VirtualMachineRestoreManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewVirtualMachineRestoreSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *snapshotManagement) Create(ctx context.Context, request *types.VirtualMachineRestore) (*types.VirtualMachineRestore, error) {
	return m.dynamicRestore.Create(ctx, request)
}

func (m *snapshotManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.VirtualMachineRestore, error) {
	return m.dynamicRestore.Get(ctx, request.Namespace, request.Name)
}

func (m *snapshotManagement) List(ctx context.Context, request *types.ListRequest) (*types.VirtualMachineRestoreList, error) {
	result, err := m.dynamicRestore.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.VirtualMachineRestoreList{Items: result}, nil
}

func (m *snapshotManagement) Update(ctx context.Context, request *types.VirtualMachineRestore) (*types.VirtualMachineRestore, error) {
	return m.dynamicRestore.Update(ctx, request)
}

func (m *snapshotManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicRestore.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
