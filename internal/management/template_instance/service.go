package templateinstance

import (
	"context"

	template_instance_v1alpha1 "github.com/kubevm.io/vink/apis/management/template_instance/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/template_instance/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/client-go/dynamic"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

func NewTemplateInstanceManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) template_instance_v1alpha1.TemplateInstanceManagementServer {
	return &templateinstanceManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicTpli:         dynamicx.NewClient[*types.TemplateInstance](dynamicClient),
	}
}

type templateinstanceManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicTpli         *dynamicx.Client[*types.TemplateInstance]

	template_instance_v1alpha1.UnsafeTemplateInstanceManagementServer
}

func (m *templateinstanceManagement) Watch(request *types.WatchRequest, server template_instance_v1alpha1.TemplateInstanceManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewTemplateInstanceSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *templateinstanceManagement) Create(ctx context.Context, request *types.TemplateInstance) (*types.TemplateInstance, error) {
	return m.dynamicTpli.Create(ctx, request)
}

func (m *templateinstanceManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.TemplateInstance, error) {
	return m.dynamicTpli.Get(ctx, request.Namespace, request.Name)
}

func (m *templateinstanceManagement) List(ctx context.Context, request *types.ListRequest) (*types.TemplateInstanceList, error) {
	result, err := m.dynamicTpli.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.TemplateInstanceList{Items: result}, nil
}

func (m *templateinstanceManagement) Update(ctx context.Context, request *types.TemplateInstance) (*types.TemplateInstance, error) {
	return m.dynamicTpli.Update(ctx, request)
}

func (m *templateinstanceManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicTpli.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
