package subnet

import (
	"context"

	subnet_v1alpha1 "github.com/kubevm.io/vink/apis/management/subnet/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/subnet/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewSubnetManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) subnet_v1alpha1.SubnetManagementServer {
	return &templateManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicSubnet:          dynamicx.NewClient[*types.Subnet](dynamicClient),
	}
}

type templateManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicSubnet          *dynamicx.Client[*types.Subnet]

	subnet_v1alpha1.UnsafeSubnetManagementServer
}

func (m *templateManagement) Watch(request *types.WatchRequest, server subnet_v1alpha1.SubnetManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewSubnetSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *templateManagement) Create(ctx context.Context, request *types.Subnet) (*types.Subnet, error) {
	return m.dynamicSubnet.Create(ctx, request)
}

func (m *templateManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.Subnet, error) {
	return m.dynamicSubnet.Get(ctx, request.Namespace, request.Name)
}

func (m *templateManagement) List(ctx context.Context, request *types.ListRequest) (*types.SubnetList, error) {
	result, err := m.dynamicSubnet.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.SubnetList{Items: result}, nil
}

func (m *templateManagement) Update(ctx context.Context, request *types.Subnet) (*types.Subnet, error) {
	return m.dynamicSubnet.Update(ctx, request)
}

func (m *templateManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicSubnet.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
