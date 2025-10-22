package business

import (
	virtualmachinesnapshot_v1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachinesnapshot/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
)

func NewVirtualMachineSnapshotSink(server virtualmachinesnapshot_v1alpha1.VirtualMachineSnapshotManagement_WatchServer) *VirtualMachineSnapshotSink {
	return &VirtualMachineSnapshotSink{server: server}
}

type VirtualMachineSnapshotSink struct {
	server virtualmachinesnapshot_v1alpha1.VirtualMachineSnapshotManagement_WatchServer
}

func (s *VirtualMachineSnapshotSink) OnAdd(obj *types.VirtualMachineSnapshot) error {
	return s.server.Send(&virtualmachinesnapshot_v1alpha1.WatchResponse{Added: []*types.VirtualMachineSnapshot{obj}})
}

func (s *VirtualMachineSnapshotSink) OnUpdate(obj *types.VirtualMachineSnapshot) error {
	return s.server.Send(&virtualmachinesnapshot_v1alpha1.WatchResponse{Modified: []*types.VirtualMachineSnapshot{obj}})
}

func (s *VirtualMachineSnapshotSink) OnDelete(obj *types.VirtualMachineSnapshot) error {
	return s.server.Send(&virtualmachinesnapshot_v1alpha1.WatchResponse{Deleted: []*types.VirtualMachineSnapshot{obj}})
}

func (s *VirtualMachineSnapshotSink) OnReady() error {
	return s.server.Send(&virtualmachinesnapshot_v1alpha1.WatchResponse{})
}
