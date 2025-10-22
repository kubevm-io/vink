package nad

import (
	"context"

	nad_v1alpha1 "github.com/kubevm.io/vink/apis/management/nad/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/nad/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewNetworkAttachmentDefinitionManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) nad_v1alpha1.NetworkAttachmentDefinitionManagementServer {
	return &templateManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicNad:          dynamicx.NewClient[*types.NetworkAttachmentDefinition](dynamicClient),
	}
}

type templateManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicNad          *dynamicx.Client[*types.NetworkAttachmentDefinition]

	nad_v1alpha1.UnsafeNetworkAttachmentDefinitionManagementServer
}

func (m *templateManagement) Watch(request *types.WatchRequest, server nad_v1alpha1.NetworkAttachmentDefinitionManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewNetworkAttachmentDefinitionSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *templateManagement) Create(ctx context.Context, request *types.NetworkAttachmentDefinition) (*types.NetworkAttachmentDefinition, error) {
	return m.dynamicNad.Create(ctx, request)
}

func (m *templateManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.NetworkAttachmentDefinition, error) {
	return m.dynamicNad.Get(ctx, request.Namespace, request.Name)
}

func (m *templateManagement) List(ctx context.Context, request *types.ListRequest) (*types.NetworkAttachmentDefinitionList, error) {
	result, err := m.dynamicNad.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.NetworkAttachmentDefinitionList{Items: result}, nil
}

func (m *templateManagement) Update(ctx context.Context, request *types.NetworkAttachmentDefinition) (*types.NetworkAttachmentDefinition, error) {
	return m.dynamicNad.Update(ctx, request)
}

func (m *templateManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicNad.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
