package virtualmachine

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"strings"

// 	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
// 	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
// 	"github.com/kubevm.io/vink/apis/annotation"
// 	"github.com/kubevm.io/vink/internal/controller/pkg"
// 	"github.com/kubevm.io/vink/pkg/log"
// 	"github.com/kubevm.io/vink/pkg/utils"
// 	"github.com/samber/lo"
// 	corev1 "k8s.io/api/core/v1"
// 	apierr "k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	kubevirtv1 "kubevirt.io/api/core/v1"
// 	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/cache"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// )

// type Reconciler struct {
// 	Client client.Client
// 	Cache  cache.Cache
// }

// func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
// 	vm := kubevirtv1.VirtualMachine{}
// 	if err := r.Client.Get(ctx, request.NamespacedName, &vm); err != nil {
// 		if apierr.IsNotFound(err) {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine: %w", err)
// 	}

// 	if err := r.reconcilerNetworkStatus(ctx, &vm); err != nil {
// 		if apierr.IsConflict(err) {
// 			return ctrl.Result{Requeue: true}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to reconcile network status: %w", err)
// 	}

// 	if err := r.reconcileHostStatus(ctx, &vm); err != nil {
// 		if apierr.IsConflict(err) {
// 			return ctrl.Result{Requeue: true}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to reconcile host status: %w", err)
// 	}

// 	if err := r.reconcileDiskStatus(ctx, &vm); err != nil {
// 		if apierr.IsConflict(err) {
// 			return ctrl.Result{Requeue: true}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to reconcile disk status: %w", err)
// 	}

// 	return ctrl.Result{}, nil
// }

// func (r *Reconciler) reconcilerNetworkStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) error {
// 	ifaceStatusSet := lo.SliceToMap(vm.Spec.Template.Spec.Networks, func(network kubevirtv1.Network) (string, *NetworkInterfaceStatus) {
// 		var (
// 			networkName string
// 			defaultNet  bool
// 		)

// 		if network.Pod != nil {
// 			networkName = vm.Spec.Template.ObjectMeta.Annotations["v1.multus-cni.io/default-network"]
// 			defaultNet = true
// 		} else if network.Multus != nil {
// 			networkName = network.Multus.NetworkName
// 			defaultNet = network.Multus.Default
// 		}
// 		return network.Name, &NetworkInterfaceStatus{
// 			Name:          network.Name,
// 			MultusNetwork: networkName,
// 			Default:       defaultNet,
// 		}
// 	})

// 	ifaceStatus := make([]*NetworkInterfaceStatus, 0, len(ifaceStatusSet))
// 	for _, network := range ifaceStatusSet {
// 		ns, err := parseNetworkNamespaceName(network.MultusNetwork)
// 		if err != nil {
// 			return err
// 		}
// 		multus := netv1.NetworkAttachmentDefinition{}
// 		if err := r.Client.Get(ctx, ns, &multus); err != nil {
// 			if apierr.IsNotFound(err) {
// 				continue
// 			}
// 			return fmt.Errorf("failed to get NetworkAttachmentDefinition: %w", err)
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
// 		if err := r.Client.Get(ctx, client.ObjectKey{Name: ipsName}, &ip); err != nil {
// 			if apierr.IsNotFound(err) {
// 				continue
// 			}
// 			return fmt.Errorf("failed to get IP: %w", err)
// 		}
// 		subnet := kubeovn.Subnet{}
// 		if err := r.Client.Get(ctx, client.ObjectKey{Name: ip.Spec.Subnet}, &subnet); err != nil {
// 			return fmt.Errorf("failed to get subnet: %w", err)
// 		}
// 		ifaceStatus = append(ifaceStatus, &NetworkInterfaceStatus{
// 			Name:          network.Name,
// 			IPAddress:     ip.Spec.IPAddress,
// 			MACAddress:    ip.Spec.MacAddress,
// 			SubnetCIDR:    subnet.Spec.CIDRBlock,
// 			VPCName:       subnet.Spec.Vpc,
// 			MultusNetwork: fmt.Sprintf("%s/%s", multus.Name, multus.Namespace),
// 			Default:       network.Default,
// 		})
// 	}

// 	if vm.Annotations == nil {
// 		vm.Annotations = make(map[string]string)
// 	}

// 	var (
// 		ifaceValue     = vm.Annotations[annotation.VinkNetworks.Name]
// 		oldIfaceStatus = make([]*NetworkInterfaceStatus, 0)
// 	)
// 	if len(ifaceValue) > 0 {
// 		if err := json.Unmarshal([]byte(ifaceValue), &oldIfaceStatus); err != nil {
// 			log.Warnf("Failed to unmarshal VirtualMachine Networks info from annotation %q: %v. Skipping this annotation", annotation.VinkNetworks.Name, err)
// 		}
// 	}

// 	if len(ifaceStatus) == 0 && len(oldIfaceStatus) == 0 {
// 		return nil
// 	}

// 	if err := pkg.PatchAnnotations(ctx, r.Client, vm, annotation.VinkNetworks.Name, ifaceStatus); err != nil {
// 		return fmt.Errorf("failed to patch virtual machine networks: %w", err)
// 	}

// 	return nil
// }

// func (r *Reconciler) reconcileHostStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) error {
// 	var vmi kubevirtv1.VirtualMachineInstance
// 	err := r.Client.Get(ctx, client.ObjectKeyFromObject(vm), &vmi)
// 	if err != nil && !apierr.IsNotFound(err) {
// 		return fmt.Errorf("failed to get virtual machine instance: %w", err)
// 	}
// 	if apierr.IsNotFound(err) || len(vmi.Status.NodeName) == 0 {
// 		if err := pkg.PatchAnnotations(ctx, r.Client, vm, annotation.VinkHost.Name, struct{}{}); err != nil {
// 			return fmt.Errorf("failed to patch virtual machine host: %w", err)
// 		}
// 		return nil
// 	}

// 	if vm.Annotations == nil {
// 		vm.Annotations = make(map[string]string)
// 	}

// 	oldHostStatus := HostStatus{}
// 	if err := json.Unmarshal([]byte(vm.Annotations[annotation.VinkHost.Name]), &oldHostStatus); err != nil {
// 		log.Warnf("Failed to unmarshal virtual machine Host info from annotation %q: %v. Skipping this annotation", annotation.VinkHost.Name, err)
// 	}

// 	node := corev1.Node{}
// 	if err := r.Client.Get(ctx, types.NamespacedName{Name: vmi.Status.NodeName}, &node); err != nil {
// 		if apierr.IsNotFound(err) {
// 			return nil
// 		}
// 		return fmt.Errorf("failed to get Node: %w", err)
// 	}

// 	newIPs := make([]string, 0)
// 	for _, addr := range node.Status.Addresses {
// 		if addr.Type == corev1.NodeInternalIP || addr.Type == corev1.NodeExternalIP && len(addr.Address) > 0 {
// 			newIPs = append(newIPs, addr.Address)
// 		}
// 	}

// 	newHostStatus := HostStatus{Name: node.Name, Ips: newIPs}
// 	if oldHostStatus.Name == newHostStatus.Name && utils.CompareArrays(newIPs, oldHostStatus.Ips) {
// 		return nil
// 	}

// 	if err := pkg.PatchAnnotations(ctx, r.Client, vm, annotation.VinkHost.Name, &newHostStatus); err != nil {
// 		return fmt.Errorf("failed to patch virtual machine host: %w", err)
// 	}

// 	return nil
// }

// func (r *Reconciler) reconcileDiskStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) error {
// 	diskSet := make(map[string]*DiskStatus, len(vm.Spec.Template.Spec.Volumes)-1)
// 	for _, disk := range vm.Spec.Template.Spec.Volumes {
// 		if disk.DataVolume == nil {
// 			continue
// 		}
// 		dv := cdiv1beta1.DataVolume{}
// 		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: vm.Namespace, Name: disk.DataVolume.Name}, &dv); err != nil {
// 			if apierr.IsNotFound(err) {
// 				continue
// 			}
// 			return fmt.Errorf("failed to get DataVolume: %w", err)
// 		}
// 		disk := &DiskStatus{
// 			Name:             disk.Name,
// 			Capacity:         dv.Spec.PVC.Resources.Requests.Storage().String(),
// 			StorageClassName: lo.FromPtr(dv.Spec.PVC.StorageClassName),
// 			Mounted:          false,
// 			Ready:            false,
// 		}
// 		if len(dv.Spec.PVC.AccessModes) > 0 {
// 			disk.AccessMode = string(dv.Spec.PVC.AccessModes[0])
// 		}
// 		readyConditions := lo.Filter(dv.Status.Conditions, func(condition cdiv1beta1.DataVolumeCondition, idx int) bool {
// 			return condition.Type == cdiv1beta1.DataVolumeReady
// 		})
// 		if len(readyConditions) > 0 && readyConditions[0].Status == corev1.ConditionTrue {
// 			disk.Ready = true
// 		}
// 		diskSet[disk.Name] = disk
// 	}

// 	for _, disk := range vm.Spec.Template.Spec.Domain.Devices.Disks {
// 		tmp, ok := diskSet[disk.Name]
// 		if !ok {
// 			continue
// 		}
// 		if lo.FromPtr(disk.BootOrder) == 1 {
// 			tmp.Rootfs = true
// 		}
// 		tmp.Mounted = true
// 		diskSet[disk.Name] = tmp
// 	}

// 	if err := pkg.PatchAnnotations(ctx, r.Client, vm, annotation.VinkDisks.Name, lo.Values(diskSet)); err != nil {
// 		return err
// 	}

// 	return nil
// }

// type NetworkInterfaceStatus2 struct {
// 	Name string `json:"name"`

// 	IPAddress string `json:"ipAddress"`

// 	MACAddress string `json:"macAddress"`

// 	SubnetCIDR string `json:"subnetCidr"`

// 	VPCName string `json:"vpcName"`

// 	MultusNetwork string `json:"multusNetwork"`

// 	Default bool `json:"default"`
// }

// type KubeOvnConfig2 struct {
// 	Type string `json:"type"`

// 	Provider string `json:"provider"`
// }

// type HostStatus2 struct {
// 	Name string   `json:"name"`
// 	Ips  []string `json:"ips"`
// }

// type DiskStatus struct {
// 	Name             string `json:"name"`
// 	Capacity         string `json:"capacity"`
// 	Ready            bool   `json:"ready"`
// 	AccessMode       string `json:"accessMode"`
// 	StorageClassName string `json:"storageClassName"`
// 	Rootfs           bool   `json:"rootfs"`
// 	Mounted          bool   `json:"mounted"`
// }

// func parseNetworkNamespaceName2(input string) (types.NamespacedName, error) {
// 	parts := strings.SplitN(input, "/", 2)
// 	if len(parts) == 2 {
// 		return types.NamespacedName{Namespace: parts[0], Name: parts[1]}, nil
// 	}
// 	return types.NamespacedName{}, fmt.Errorf("invalid network namespace name: %s", input)
// }

// func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		Named("virtualmachine").
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
// 		WatchesMetadata(
// 			&kubevirtv1.VirtualMachineInstance{},
// 			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
// 				metadata, ok := obj.(*metav1.PartialObjectMetadata)
// 				if !ok {
// 					return nil
// 				}
// 				return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: metadata.Namespace, Name: metadata.Name}}}
// 			}),
// 		).
// 		WatchesMetadata(
// 			&cdiv1beta1.DataVolume{},
// 			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
// 				metadata, ok := obj.(*metav1.PartialObjectMetadata)
// 				if !ok || metadata.Annotations == nil {
// 					return nil
// 				}
// 				ownerString, ok := metadata.Annotations[annotation.VinkDatavolumeOwner.Name]
// 				if !ok || len(ownerString) == 0 {
// 					return nil
// 				}
// 				owners := make([]string, 0)
// 				if err := json.Unmarshal([]byte(ownerString), &owners); err != nil {
// 					log.Errorf("Failed to unmarshal DataVolume owners: %v", err)
// 					return nil
// 				}
// 				requests := make([]reconcile.Request, 0, len(owners))
// 				for _, owner := range owners {
// 					requests = append(requests, reconcile.Request{NamespacedName: client.ObjectKey{Namespace: metadata.Namespace, Name: owner}})
// 				}
// 				return requests
// 			}),
// 		).
// 		Complete(r)
// }
