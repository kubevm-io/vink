package business

import (
	sc_v1alpha1 "github.com/kubevm.io/vink/apis/management/sc/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewStorageClassSink(server sc_v1alpha1.StorageClassManagement_WatchServer) *StorageClassSink {
	return &StorageClassSink{server: server}
}

type StorageClassSink struct {
	server sc_v1alpha1.StorageClassManagement_WatchServer
}

func (s *StorageClassSink) OnAdd(obj *types.StorageClass) error {
	return s.server.Send(&sc_v1alpha1.WatchResponse{Added: []*types.StorageClass{obj}})
}

func (s *StorageClassSink) OnUpdate(obj *types.StorageClass) error {
	return s.server.Send(&sc_v1alpha1.WatchResponse{Modified: []*types.StorageClass{obj}})
}

func (s *StorageClassSink) OnDelete(obj *types.StorageClass) error {
	return s.server.Send(&sc_v1alpha1.WatchResponse{Deleted: []*types.StorageClass{obj}})
}

func (s *StorageClassSink) OnReady() error {
	return s.server.Send(&sc_v1alpha1.WatchResponse{})
}
