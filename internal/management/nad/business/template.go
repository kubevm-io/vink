package business

import (
	nad_v1alpha1 "github.com/kubevm.io/vink/apis/management/nad/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewNetworkAttachmentDefinitionSink(server nad_v1alpha1.NetworkAttachmentDefinitionManagement_WatchServer) *NetworkAttachmentDefinitionSink {
	return &NetworkAttachmentDefinitionSink{server: server}
}

type NetworkAttachmentDefinitionSink struct {
	server nad_v1alpha1.NetworkAttachmentDefinitionManagement_WatchServer
}

func (s *NetworkAttachmentDefinitionSink) OnAdd(obj *types.NetworkAttachmentDefinition) error {
	return s.server.Send(&nad_v1alpha1.WatchResponse{Added: []*types.NetworkAttachmentDefinition{obj}})
}

func (s *NetworkAttachmentDefinitionSink) OnUpdate(obj *types.NetworkAttachmentDefinition) error {
	return s.server.Send(&nad_v1alpha1.WatchResponse{Modified: []*types.NetworkAttachmentDefinition{obj}})
}

func (s *NetworkAttachmentDefinitionSink) OnDelete(obj *types.NetworkAttachmentDefinition) error {
	return s.server.Send(&nad_v1alpha1.WatchResponse{Deleted: []*types.NetworkAttachmentDefinition{obj}})
}

func (s *NetworkAttachmentDefinitionSink) OnReady() error {
	return s.server.Send(&nad_v1alpha1.WatchResponse{})
}
