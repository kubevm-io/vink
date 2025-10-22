package business

import (
	virtualmachinerestore_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinerestore/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewVirtualMachineRestoreSink(server virtualmachinerestore_v1alpha1.VirtualMachineRestoreManagement_WatchServer) *VirtualMachineRestoreSink {
	return &VirtualMachineRestoreSink{server: server}
}

type VirtualMachineRestoreSink struct {
	server virtualmachinerestore_v1alpha1.VirtualMachineRestoreManagement_WatchServer
}

func (s *VirtualMachineRestoreSink) OnAdd(obj *types.VirtualMachineRestore) error {
	return s.server.Send(&virtualmachinerestore_v1alpha1.WatchResponse{Added: []*types.VirtualMachineRestore{obj}})
}

func (s *VirtualMachineRestoreSink) OnUpdate(obj *types.VirtualMachineRestore) error {
	return s.server.Send(&virtualmachinerestore_v1alpha1.WatchResponse{Modified: []*types.VirtualMachineRestore{obj}})
}

func (s *VirtualMachineRestoreSink) OnDelete(obj *types.VirtualMachineRestore) error {
	return s.server.Send(&virtualmachinerestore_v1alpha1.WatchResponse{Deleted: []*types.VirtualMachineRestore{obj}})
}

func (s *VirtualMachineRestoreSink) OnReady() error {
	return s.server.Send(&virtualmachinerestore_v1alpha1.WatchResponse{})
}
