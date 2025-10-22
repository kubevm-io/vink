package sc

import (
	"context"

	sc_v1alpha1 "github.com/kubevm.io/vink/apis/management/sc/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/sc/business"
	"github.com/kubevm.io/vink/pkg/dynamicx"
	"github.com/kubevm.io/vink/pkg/informer"
	"github.com/kubevm.io/vink/pkg/watcher"
	"google.golang.org/protobuf/types/known/emptypb"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func NewStorageClassManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface) sc_v1alpha1.StorageClassManagementServer {
	return &templateManagement{
		kubeInformerFactory: kubeInformerFactory,
		dynamicSc:          dynamicx.NewClient[*types.StorageClass](dynamicClient),
	}
}

type templateManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	dynamicSc          *dynamicx.Client[*types.StorageClass]

	sc_v1alpha1.UnsafeStorageClassManagementServer
}

func (m *templateManagement) Watch(request *types.WatchRequest, server sc_v1alpha1.StorageClassManagement_WatchServer) error {
	return watcher.Watch(server.Context(), m.kubeInformerFactory, business.NewStorageClassSink(server), &k8stypes.NamespacedName{Namespace: request.Namespace, Name: request.Name})
}

func (m *templateManagement) Create(ctx context.Context, request *types.StorageClass) (*types.StorageClass, error) {
	return m.dynamicSc.Create(ctx, request)
}

func (m *templateManagement) Get(ctx context.Context, request *types.NamespaceName) (*types.StorageClass, error) {
	return m.dynamicSc.Get(ctx, request.Namespace, request.Name)
}

func (m *templateManagement) List(ctx context.Context, request *types.ListRequest) (*types.StorageClassList, error) {
	result, err := m.dynamicSc.List(ctx, request.Namespace)
	if err != nil {
		return nil, err
	}
	return &types.StorageClassList{Items: result}, nil
}

func (m *templateManagement) Update(ctx context.Context, request *types.StorageClass) (*types.StorageClass, error) {
	return m.dynamicSc.Update(ctx, request)
}

func (m *templateManagement) Delete(ctx context.Context, request *types.NamespaceName) (*emptypb.Empty, error) {
	if err := m.dynamicSc.Delete(ctx, request.Namespace, request.Name); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
