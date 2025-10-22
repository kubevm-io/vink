package template

import (
	"context"

	template_v1alpha1 "github.com/kubevm.io/vink/apis/management/template/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/template/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewTemplateManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) template_v1alpha1.TemplateManagementServer {
	return &templateManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicTpl:          dynamicx.NewClient[*types.Template](dynamicClient),
	}
}

type templateManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicTpl          *dynamicx.Client[*types.Template]

	template_v1alpha1.UnsafeTemplateManagementServer
}

func (m *templateManagement) Watch(request *types.WatchRequest, server template_v1alpha1.TemplateManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewTemplateSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *templateManagement) Create(ctx context.Context, request *types.Template) (*types.Template, error) {
	return m.dynamicTpl.Create(ctx, request)
}

func (m *templateManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.Template, error) {
	return m.dynamicTpl.Get(ctx, request.Namespace, request.Name)
}

func (m *templateManagement) List(ctx context.Context, request *types.ListRequest) (*types.TemplateList, error) {
	result, err := m.dynamicTpl.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.TemplateList{Items: result}, nil
}

func (m *templateManagement) Update(ctx context.Context, request *types.Template) (*types.Template, error) {
	return m.dynamicTpl.Update(ctx, request)
}

func (m *templateManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicTpl.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
