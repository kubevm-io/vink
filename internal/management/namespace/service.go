package namespace

import (
	"context"

	namespace_v1alpha1 "github.com/kubevm.io/vink/apis/management/namespace/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"k8s.io/client-go/dynamic"
)

func NewNamespaceManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) namespace_v1alpha1.NamespaceManagementServer {
	return &nodeManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicNamespace:    dynamicx.NewClient[*types.Namespace](dynamicClient),
	}
}

type nodeManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicNamespace    *dynamicx.Client[*types.Namespace]

	namespace_v1alpha1.NamespaceManagementServer
}

func (m *nodeManagement) List(ctx context.Context, request *types.ListRequest) (*types.NamespaceList, error) {
	result, err := m.dynamicNamespace.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.NamespaceList{Items: result}, nil
}
