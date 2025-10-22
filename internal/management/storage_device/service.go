package storage_device

import (
	"context"

	storage_device_v1alpha1 "github.com/kubevm.io/vink/apis/management/storage_device/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/management/storage_device/business"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/informer"
	"k8s.io/client-go/dynamic"
)

func NewStorageDeviceManagement(kubeInformerFactory informer.KubeInformerFactory, dynamicClient dynamic.Interface, cephCli clients.CephInterface) storage_device_v1alpha1.StorageDeviceManagementServer {
	return &storageDeviceManagement{
		kubeInformerFactory: kubeInformerFactory,
		cephCli:             cephCli,
	}
}

type storageDeviceManagement struct {
	kubeInformerFactory informer.KubeInformerFactory
	cephCli             clients.CephInterface

	storage_device_v1alpha1.StorageDeviceManagementServer
}

func (m *storageDeviceManagement) List(ctx context.Context, request *types.ListRequest) (*types.StorageDeviceList, error) {
	result, err := business.List(ctx, m.cephCli, request.Name)
	if err != nil {
		return nil, err
	}
	return &types.StorageDeviceList{Items: result}, nil
}
