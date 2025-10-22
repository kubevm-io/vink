package virtualmachine

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"

// 	"github.com/kubevm.io/vink/apis/annotation"
// 	apitypes "github.com/kubevm.io/vink/apis/types"
// 	"github.com/kubevm.io/vink/internal/controller/pkg"
// 	"github.com/kubevm.io/vink/pkg/log"
// 	"github.com/kubevm.io/vink/pkg/utils"
// 	corev1 "k8s.io/api/core/v1"
// 	apierr "k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	kubevirtv1 "kubevirt.io/api/core/v1"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/cache"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/handler"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"
// )

// type HostReconciler struct {
// 	Client client.Client
// 	Cache  cache.Cache
// }

// func (reconciler *HostReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
// 	var vm kubevirtv1.VirtualMachine
// 	if err := reconciler.Client.Get(ctx, request.NamespacedName, &vm); err != nil {
// 		if apierr.IsNotFound(err) {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine: %w", err)
// 	}

// 	var vmi kubevirtv1.VirtualMachineInstance
// 	err := reconciler.Client.Get(ctx, request.NamespacedName, &vmi)
// 	if err != nil && !apierr.IsNotFound(err) {
// 		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine instance: %w", err)
// 	}
// 	if apierr.IsNotFound(err) || len(vmi.Status.NodeName) == 0 {
// 		if err := pkg.PatchAnnotations(ctx, reconciler.Client, &vm, annotation.VinkHost.Name, struct{}{}); err != nil {
// 			return ctrl.Result{}, fmt.Errorf("failed to patch virtual machine host: %w", err)
// 		}
// 		return ctrl.Result{}, nil
// 	}

// 	if vm.Annotations == nil {
// 		vm.Annotations = make(map[string]string)
// 	}
// 	oldHost := apitypes.VirtualMachineHost{}
// 	if err := json.Unmarshal([]byte(vm.Annotations[annotation.VinkHost.Name]), &oldHost); err != nil {
// 		log.Warnf("Failed to unmarshal virtual machine Host info from annotation %q: %v. Skipping this annotation", annotation.VinkHost.Name, err)
// 	}

// 	node := corev1.Node{}
// 	if err := reconciler.Client.Get(ctx, types.NamespacedName{Name: vmi.Status.NodeName}, &node); err != nil {
// 		if apierr.IsNotFound(err) {
// 			return ctrl.Result{}, nil
// 		}
// 		return ctrl.Result{}, fmt.Errorf("failed to get Node: %w", err)
// 	}

// 	newIPs := make([]string, 0)
// 	for _, addr := range node.Status.Addresses {
// 		if addr.Type == corev1.NodeInternalIP || addr.Type == corev1.NodeExternalIP && len(addr.Address) > 0 {
// 			newIPs = append(newIPs, addr.Address)
// 		}
// 	}

// 	newHost := apitypes.VirtualMachineHost{Name: node.Name, Ips: newIPs}
// 	if oldHost.Name == newHost.Name && utils.CompareArrays(newIPs, oldHost.Ips) {
// 		return ctrl.Result{}, nil
// 	}

// 	if err := pkg.PatchAnnotations(ctx, reconciler.Client, &vm, annotation.VinkHost.Name, &newHost); err != nil {
// 		return ctrl.Result{}, fmt.Errorf("failed to patch virtual machine host: %w", err)
// 	}

// 	return ctrl.Result{}, nil
// }

// func (reconciler *HostReconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		Named("virtualmachine_host").
// 		For(&kubevirtv1.VirtualMachine{}).
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
// 		Complete(reconciler)
// }
