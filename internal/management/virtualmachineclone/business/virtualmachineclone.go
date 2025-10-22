package business

import (
	virtualmachineclone_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachineclone/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewVirtualMachineCloneSink(server virtualmachineclone_v1alpha1.VirtualMachineCloneManagement_WatchServer) *VirtualMachineCloneSink {
	return &VirtualMachineCloneSink{server: server}
}

type VirtualMachineCloneSink struct {
	server virtualmachineclone_v1alpha1.VirtualMachineCloneManagement_WatchServer
}

func (s *VirtualMachineCloneSink) OnAdd(obj *types.VirtualMachineClone) error {
	return s.server.Send(&virtualmachineclone_v1alpha1.WatchResponse{Added: []*types.VirtualMachineClone{obj}})
}

func (s *VirtualMachineCloneSink) OnUpdate(obj *types.VirtualMachineClone) error {
	return s.server.Send(&virtualmachineclone_v1alpha1.WatchResponse{Modified: []*types.VirtualMachineClone{obj}})
}

func (s *VirtualMachineCloneSink) OnDelete(obj *types.VirtualMachineClone) error {
	return s.server.Send(&virtualmachineclone_v1alpha1.WatchResponse{Deleted: []*types.VirtualMachineClone{obj}})
}

func (s *VirtualMachineCloneSink) OnReady() error {
	return s.server.Send(&virtualmachineclone_v1alpha1.WatchResponse{})
}
