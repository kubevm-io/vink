package business

import (
	node_v1alpha1 "github.com/kubevm.io/vink/apis/management/node/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewNodeSink(server node_v1alpha1.NodeManagement_WatchServer) *NodeSink {
	return &NodeSink{server: server}
}

type NodeSink struct {
	server node_v1alpha1.NodeManagement_WatchServer
}

func (s *NodeSink) OnAdd(obj *types.Node) error {
	return s.server.Send(&node_v1alpha1.WatchResponse{Added: []*types.Node{obj}})
}

func (s *NodeSink) OnUpdate(obj *types.Node) error {
	return s.server.Send(&node_v1alpha1.WatchResponse{Modified: []*types.Node{obj}})
}

func (s *NodeSink) OnDelete(obj *types.Node) error {
	return s.server.Send(&node_v1alpha1.WatchResponse{Deleted: []*types.Node{obj}})
}

func (s *NodeSink) OnReady() error {
	return s.server.Send(&node_v1alpha1.WatchResponse{})
}
