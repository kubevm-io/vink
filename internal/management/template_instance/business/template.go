package business

import (
	template_instance_v1alpha1 "github.com/kubevm.io/vink/apis/management/template_instance/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewTemplateInstanceSink(server template_instance_v1alpha1.TemplateInstanceManagement_WatchServer) *TemplateInstanceSink {
	return &TemplateInstanceSink{server: server}
}

type TemplateInstanceSink struct {
	server template_instance_v1alpha1.TemplateInstanceManagement_WatchServer
}

func (s *TemplateInstanceSink) OnAdd(obj *types.TemplateInstance) error {
	return s.server.Send(&template_instance_v1alpha1.WatchResponse{Added: []*types.TemplateInstance{obj}})
}

func (s *TemplateInstanceSink) OnUpdate(obj *types.TemplateInstance) error {
	return s.server.Send(&template_instance_v1alpha1.WatchResponse{Modified: []*types.TemplateInstance{obj}})
}

func (s *TemplateInstanceSink) OnDelete(obj *types.TemplateInstance) error {
	return s.server.Send(&template_instance_v1alpha1.WatchResponse{Deleted: []*types.TemplateInstance{obj}})
}

func (s *TemplateInstanceSink) OnReady() error {
	return s.server.Send(&template_instance_v1alpha1.WatchResponse{})
}
