package business

import (
	subnet_v1alpha1 "github.com/kubevm.io/vink/apis/management/subnet/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewSubnetSink(server subnet_v1alpha1.SubnetManagement_WatchServer) *SubnetSink {
	return &SubnetSink{server: server}
}

type SubnetSink struct {
	server subnet_v1alpha1.SubnetManagement_WatchServer
}

func (s *SubnetSink) OnAdd(obj *types.Subnet) error {
	return s.server.Send(&subnet_v1alpha1.WatchResponse{Added: []*types.Subnet{obj}})
}

func (s *SubnetSink) OnUpdate(obj *types.Subnet) error {
	return s.server.Send(&subnet_v1alpha1.WatchResponse{Modified: []*types.Subnet{obj}})
}

func (s *SubnetSink) OnDelete(obj *types.Subnet) error {
	return s.server.Send(&subnet_v1alpha1.WatchResponse{Deleted: []*types.Subnet{obj}})
}

func (s *SubnetSink) OnReady() error {
	return s.server.Send(&subnet_v1alpha1.WatchResponse{})
}
