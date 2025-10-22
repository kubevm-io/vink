package business

import (
	template_v1alpha1 "github.com/kubevm.io/vink/apis/management/template/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewTemplateSink(server template_v1alpha1.TemplateManagement_WatchServer) *TemplateSink {
	return &TemplateSink{server: server}
}

type TemplateSink struct {
	server template_v1alpha1.TemplateManagement_WatchServer
}

func (s *TemplateSink) OnAdd(obj *types.Template) error {
	return s.server.Send(&template_v1alpha1.WatchResponse{Added: []*types.Template{obj}})
}

func (s *TemplateSink) OnUpdate(obj *types.Template) error {
	return s.server.Send(&template_v1alpha1.WatchResponse{Modified: []*types.Template{obj}})
}

func (s *TemplateSink) OnDelete(obj *types.Template) error {
	return s.server.Send(&template_v1alpha1.WatchResponse{Deleted: []*types.Template{obj}})
}

func (s *TemplateSink) OnReady() error {
	return s.server.Send(&template_v1alpha1.WatchResponse{})
}
