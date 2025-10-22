package virtualmachineclone

import (
	"context"

	virtualmachineclone_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachineclone/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/virtualmachineclone/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewVirtualMachineCloneManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) virtualmachineclone_v1alpha1.VirtualMachineCloneManagementServer {
	return &cloneManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicClone:        dynamicx.NewClient[*types.VirtualMachineClone](dynamicClient),
	}
}

type cloneManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicClone        *dynamicx.Client[*types.VirtualMachineClone]

	virtualmachineclone_v1alpha1.UnsafeVirtualMachineCloneManagementServer
}

func (m *cloneManagement) Watch(request *types.WatchRequest, server virtualmachineclone_v1alpha1.VirtualMachineCloneManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewVirtualMachineCloneSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *cloneManagement) Create(ctx context.Context, request *types.VirtualMachineClone) (*types.VirtualMachineClone, error) {
	return m.dynamicClone.Create(ctx, request)
}

func (m *cloneManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.VirtualMachineClone, error) {
	return m.dynamicClone.Get(ctx, request.Namespace, request.Name)
}

func (m *cloneManagement) List(ctx context.Context, request *types.ListRequest) (*types.VirtualMachineCloneList, error) {
	result, err := m.dynamicClone.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.VirtualMachineCloneList{Items: result}, nil
}

func (m *cloneManagement) Update(ctx context.Context, request *types.VirtualMachineClone) (*types.VirtualMachineClone, error) {
	return m.dynamicClone.Update(ctx, request)
}

func (m *cloneManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicClone.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
