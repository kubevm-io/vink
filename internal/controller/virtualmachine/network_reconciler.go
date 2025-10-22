package virtualmachine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubevm.io/vink/apis/annotation"
	"github.com/kubevm.io/vink/internal/controller/pkg"
	"github.com/kubevm.io/vink/pkg/log"
	"github.com/samber/lo"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type NetworkReconciler struct {
	Client client.Client
	Cache  cache.Cache
}

func (r *NetworkReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	vm := kubevirtv1.VirtualMachine{}
	if err := r.Client.Get(ctx, request.NamespacedName, &vm); err != nil {
		if apierr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine: %w", err)
	}

	ifaceStatusSet := lo.SliceToMap(vm.Spec.Template.Spec.Networks, func(network kubevirtv1.Network) (string, *NetworkInterfaceStatus) {
		var (
			networkName string
			defaultNet  bool
		)

		if network.Pod != nil {
			networkName = vm.Spec.Template.ObjectMeta.Annotations["v1.multus-cni.io/default-network"]
			defaultNet = true
		} else if network.Multus != nil {
			networkName = network.Multus.NetworkName
			defaultNet = network.Multus.Default
		}
		return network.Name, &NetworkInterfaceStatus{
			Name:          network.Name,
			MultusNetwork: networkName,
			Default:       defaultNet,
		}
	})

	ifaceStatus := make([]*NetworkInterfaceStatus, 0, len(ifaceStatusSet))
	for _, v := range vm.Spec.Template.Spec.Networks {
		network := ifaceStatusSet[v.Name]
		if network == nil {
			continue
		}
		ns, err := parseNetworkNamespaceName(network.MultusNetwork)
		if err != nil {
			return ctrl.Result{}, err
		}
		multus := netv1.NetworkAttachmentDefinition{}
		if err := r.Client.Get(ctx, ns, &multus); err != nil {
			if apierr.IsNotFound(err) {
				continue
			}
			return ctrl.Result{}, fmt.Errorf("failed to get NetworkAttachmentDefinition: %w", err)
		}
		config := KubeOvnConfig{}
		if err := json.Unmarshal([]byte(multus.Spec.Config), &config); err != nil {
			log.Warnf("Failed to unmarshal NetworkAttachmentDefinition config: %v", err)
			continue
		}
		if config.Type != "kube-ovn" {
			continue
		}
		ipsName := fmt.Sprintf("%s.%s", vm.Name, vm.Namespace)
		if len(config.Provider) > 0 && config.Provider != "ovn" {
			ipsName = fmt.Sprintf("%s.%s.%s", vm.Name, vm.Namespace, config.Provider)
		}
		ip := kubeovn.IP{}
		if err := r.Client.Get(ctx, client.ObjectKey{Name: ipsName}, &ip); err != nil {
			if apierr.IsNotFound(err) {
				continue
			}
			return ctrl.Result{}, fmt.Errorf("failed to get IP: %w", err)
		}
		if ip.Spec.PodType != "VirtualMachine" || ip.Spec.PodName != vm.Name || ip.Spec.Namespace != vm.Namespace {
			continue
		}

		subnet := kubeovn.Subnet{}
		if err := r.Client.Get(ctx, client.ObjectKey{Name: ip.Spec.Subnet}, &subnet); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to get subnet: %w", err)
		}
		ifaceStatus = append(ifaceStatus, &NetworkInterfaceStatus{
			Name:          network.Name,
			IPAddress:     ip.Spec.IPAddress,
			MACAddress:    ip.Spec.MacAddress,
			Subnet:        subnet.Name,
			SubnetCIDR:    subnet.Spec.CIDRBlock,
			VPCName:       subnet.Spec.Vpc,
			MultusNetwork: fmt.Sprintf("%s/%s", multus.Name, multus.Namespace),
			Default:       network.Default,
		})
	}

	if vm.Annotations == nil {
		vm.Annotations = make(map[string]string)
	}

	var (
		ifaceValue     = vm.Annotations[annotation.VinkNetworks.Name]
		oldIfaceStatus = make([]*NetworkInterfaceStatus, 0)
	)
	if len(ifaceValue) > 0 {
		if err := json.Unmarshal([]byte(ifaceValue), &oldIfaceStatus); err != nil {
			log.Warnf("Failed to unmarshal VirtualMachine Networks info from annotation %q: %v. Skipping this annotation", annotation.VinkNetworks.Name, err)
		}
	}

	if len(ifaceStatus) == 0 && len(oldIfaceStatus) == 0 {
		return ctrl.Result{}, nil
	}

	if err := pkg.PatchAnnotations(ctx, r.Client, &vm, annotation.VinkNetworks.Name, ifaceStatus); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to patch virtual machine networks: %w", err)
	}

	return ctrl.Result{}, nil
}

type NetworkInterfaceStatus struct {
	Name string `json:"name"`

	IPAddress string `json:"ipAddress"`

	MACAddress string `json:"macAddress"`

	Subnet string `json:"subnet"`

	SubnetCIDR string `json:"subnetCidr"`

	VPCName string `json:"vpcName"`

	MultusNetwork string `json:"multusNetwork"`

	Default bool `json:"default"`
}

type KubeOvnConfig struct {
	Type string `json:"type"`

	Provider string `json:"provider"`
}

func parseNetworkNamespaceName(input string) (types.NamespacedName, error) {
	parts := strings.SplitN(input, "/", 2)
	if len(parts) == 2 {
		return types.NamespacedName{Namespace: parts[0], Name: parts[1]}, nil
	}
	return types.NamespacedName{}, fmt.Errorf("invalid network namespace name: %s", input)
}

func (r *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualmachine-network").
		For(&kubevirtv1.VirtualMachine{}).
		Watches(
			&kubeovn.IP{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				ip, ok := obj.(*kubeovn.IP)
				if !ok || ip.Spec.PodType != "VirtualMachine" {
					return nil
				}
				return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: ip.Spec.Namespace, Name: ip.Spec.PodName}}}
			}),
		).
		Complete(r)
}
