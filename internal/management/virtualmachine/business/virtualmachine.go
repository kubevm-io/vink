package business

import (
	"context"

	vmv1alpha1 "github.com/kubevm.io/vink/apis/management/virtualmachine/v1alpha1"
	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/pkg/clients"
	virtv1 "kubevirt.io/api/core/v1"
)

const (
	SerialConsoleRequestPathTmpl = "/apis/vink.io/v1alpha1/namespaces/{namespace}/virtualmachines/{name}/console"
)

func NewVirtualMachineSink(server vmv1alpha1.VirtualMachineManagement_WatchServer) *VirtualMachineSink {
	return &VirtualMachineSink{server: server}
}

type VirtualMachineSink struct {
	server vmv1alpha1.VirtualMachineManagement_WatchServer
}

func (s *VirtualMachineSink) OnAdd(obj *types.VirtualMachine) error {
	return s.server.Send(&vmv1alpha1.WatchResponse{Added: []*types.VirtualMachine{obj}})
}

func (s *VirtualMachineSink) OnUpdate(obj *types.VirtualMachine) error {
	return s.server.Send(&vmv1alpha1.WatchResponse{Modified: []*types.VirtualMachine{obj}})
}

func (s *VirtualMachineSink) OnDelete(obj *types.VirtualMachine) error {
	return s.server.Send(&vmv1alpha1.WatchResponse{Deleted: []*types.VirtualMachine{obj}})
}

func (s *VirtualMachineSink) OnReady() error {
	return s.server.Send(&vmv1alpha1.WatchResponse{})
}

func PowerState(ctx context.Context, namespaceName *types.NamespaceName, powerState vmv1alpha1.PowerStateRequest_PowerState) error {
	cli := clients.Clients.VirtualMachine(namespaceName.Namespace)

	switch powerState {
	case vmv1alpha1.PowerStateRequest_ON:
		return cli.Start(ctx, namespaceName.Name, &virtv1.StartOptions{})
	case vmv1alpha1.PowerStateRequest_OFF:
		return cli.Stop(ctx, namespaceName.Name, &virtv1.StopOptions{})
	case vmv1alpha1.PowerStateRequest_REBOOT:
		return cli.Restart(ctx, namespaceName.Name, &virtv1.RestartOptions{})
	case vmv1alpha1.PowerStateRequest_FORCE_OFF:
		return cli.ForceStop(ctx, namespaceName.Name, &virtv1.StopOptions{})
	case vmv1alpha1.PowerStateRequest_FORCE_REBOOT:
		return cli.ForceRestart(ctx, namespaceName.Name, &virtv1.RestartOptions{})
	}

	return nil
}
