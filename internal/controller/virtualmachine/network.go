package virtualmachine

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
// 	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
// 	"github.com/kubevm.io/vink/apis/annotation"
// 	apitypes "github.com/kubevm.io/vink/apis/types"
// 	"github.com/kubevm.io/vink/internal/controller/pkg"
// 	"github.com/kubevm.io/vink/pkg/log"
// 	"github.com/samber/lo"
// 	apierr "k8s.io/apimachinery/pkg/api/errors"
// 	kubevirtv1 "kubevirt.io/api/core/v1"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/cache"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// )

// type NetworkReconciler struct {
// 	Client client.Client
// 	Cache  cache.Cache
// }

// // type KubeOvnConfig struct {
// // 	Type     string `json:"type"`
// // 	Provider string `json:"provider"`
// // }

// func (reconciler *NetworkReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
// 	vm := kubevirtv1.VirtualMachine{}
// 	if err := reconciler.Client.Get(ctx, request.NamespacedName, &vm); err != nil {
// 		if apierr.IsNotFound(err) {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine: %w", err)
// 	}

// 	networkMap := lo.SliceToMap(vm.Spec.Template.Spec.Networks, func(network kubevirtv1.Network) (string, *apitypes.VirtualMachineNetwork) {
// 		var (
// 			net         string
// 			networkName string
// 			defaultNet  bool
// 		)
// 		if network.Pod != nil {
// 			net = "pod"
// 			networkName = vm.Spec.Template.ObjectMeta.Annotations["v1.multus-cni.io/default-network"]
// 			defaultNet = true
// 		} else if network.Multus != nil {
// 			net = "multus"
// 			networkName = network.Multus.NetworkName
// 			defaultNet = network.Multus.Default
// 		}
// 		return network.Name, &apitypes.VirtualMachineNetwork{
// 			Name:    network.Name,
// 			Network: net,
// 			Multus:  networkName,
// 			Default: defaultNet,
// 		}
// 	})
// 	for _, inter := range vm.Spec.Template.Spec.Domain.Devices.Interfaces {
// 		network, ok := networkMap[inter.Name]
// 		if !ok {
// 			continue
// 		}
// 		switch {
// 		case inter.Bridge != nil:
// 			network.Interface = "bridge"
// 		case inter.Masquerade != nil:
// 			network.Interface = "masquerade"
// 		case inter.SRIOV != nil:
// 			network.Interface = "sriov"
// 		case inter.DeprecatedMacvtap != nil:
// 			network.Interface = "macvtap"
// 		case inter.DeprecatedPasst != nil:
// 			network.Interface = "passt"
// 		default:
// 			continue
// 		}
// 		networkMap[inter.Name] = network
// 	}

// 	newNetworks := make([]*apitypes.VirtualMachineNetwork, 0, len(networkMap))
// 	for _, network := range networkMap {
// 		ns, err := parseNetworkNamespaceName(network.Multus)
// 		if err != nil {
// 			return ctrl.Result{}, err
// 		}
// 		multus := netv1.NetworkAttachmentDefinition{}
// 		if err := reconciler.Client.Get(ctx, ns, &multus); err != nil {
// 			if apierr.IsNotFound(err) {
// 				continue
// 			}
// 			return ctrl.Result{}, fmt.Errorf("failed to get NetworkAttachmentDefinition: %w", err)
// 		}
// 		config := KubeOvnConfig{}
// 		if err := json.Unmarshal([]byte(multus.Spec.Config), &config); err != nil {
// 			log.Warnf("Failed to unmarshal NetworkAttachmentDefinition config: %v", err)
// 			continue
// 		}
// 		if config.Type != "kube-ovn" {
// 			continue
// 		}
// 		ipsName := fmt.Sprintf("%s.%s.%s", vm.Name, vm.Namespace, config.Provider)
// 		ip := kubeovn.IP{}
// 		if err := reconciler.Client.Get(ctx, client.ObjectKey{Name: ipsName}, &ip); err != nil {
// 			if apierr.IsNotFound(err) {
// 				continue
// 			}
// 			return ctrl.Result{}, fmt.Errorf("failed to get IP: %w", err)
// 		}
// 		subnet := kubeovn.Subnet{}
// 		if err := reconciler.Client.Get(ctx, client.ObjectKey{Name: ip.Spec.Subnet}, &subnet); err != nil {
// 			return ctrl.Result{}, fmt.Errorf("failed to get subnet: %w", err)
// 		}
// 		newNetworks = append(newNetworks, &apitypes.VirtualMachineNetwork{
// 			Name:      network.Name,
// 			Network:   network.Network,
// 			Interface: network.Interface,
// 			IpsName:   ip.Name,
// 			Ip:        ip.Spec.IPAddress,
// 			Subnet:    subnet.Name,
// 			Mac:       ip.Spec.MacAddress,
// 			Vpc:       subnet.Spec.Vpc,
// 			Multus:    network.Multus,
// 			Default:   network.Default,
// 		})
// 	}

// 	if vm.Annotations == nil {
// 		vm.Annotations = make(map[string]string)
// 	}

// 	networks := make([]*apitypes.VirtualMachineNetwork, 0)
// 	if err := json.Unmarshal([]byte(vm.Annotations[annotation.VinkNetworks.Name]), &networks); err != nil {
// 		log.Warnf("Failed to unmarshal VirtualMachine Networks info from annotation %q: %v. Skipping this annotation", annotation.VinkNetworks.Name, err)
// 	}

// 	if len(networks) == 0 && len(newNetworks) == 0 {
// 		return ctrl.Result{}, nil
// 	}

// 	if err := pkg.PatchAnnotations(ctx, reconciler.Client, &vm, annotation.VinkNetworks.Name, newNetworks); err != nil {
// 		return ctrl.Result{}, fmt.Errorf("failed to patch virtual machine networks: %w", err)
// 	}

// 	return ctrl.Result{}, nil
// }

// func (reconciler *NetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		Named("virtualmachine_network").
// 		For(&kubevirtv1.VirtualMachine{}).
// 		Watches(
// 			&kubeovn.IP{},
// 			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
// 				ip, ok := obj.(*kubeovn.IP)
// 				if !ok || ip.Spec.PodType != "VirtualMachine" {
// 					return nil
// 				}
// 				return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: ip.Spec.Namespace, Name: ip.Spec.PodName}}}
// 			}),
// 		).
// 		Complete(reconciler)
// }

// // func parseNetworkNamespaceName(input string) (types.NamespacedName, error) {
// // 	parts := strings.SplitN(input, "/", 2)
// // 	if len(parts) == 2 {
// // 		return types.NamespacedName{Namespace: parts[0], Name: parts[1]}, nil
// // 	}
// // 	return types.NamespacedName{}, fmt.Errorf("invalid network namespace name: %s", input)
// // }
