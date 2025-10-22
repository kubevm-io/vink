package node

import (
	"context"

	node_v1alpha1 "github.com/kubevm.io/vink/apis/management/node/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/node/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewNodeManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) node_v1alpha1.NodeManagementServer {
	return &nodeManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicTpl:          dynamicx.NewClient[*types.Node](dynamicClient),
	}
}

type nodeManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicTpl          *dynamicx.Client[*types.Node]

	node_v1alpha1.NodeManagementServer
}

func (m *nodeManagement) Watch(request *types.WatchRequest, server node_v1alpha1.NodeManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewNodeSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *nodeManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.Node, error) {
	return m.dynamicTpl.Get(ctx, request.Namespace, request.Name)
}

func (m *nodeManagement) List(ctx context.Context, request *types.ListRequest) (*types.NodeList, error) {
	result, err := m.dynamicTpl.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.NodeList{Items: result}, nil
}

func (m *nodeManagement) Update(ctx context.Context, request *types.Node) (*types.Node, error) {
	return m.dynamicTpl.Update(ctx, request)
}

func (m *nodeManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicTpl.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
