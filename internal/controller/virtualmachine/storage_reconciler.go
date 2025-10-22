package virtualmachine

import (
	"context"
	"fmt"

	"github.com/kubevm.io/vink/apis/annotation"
	"github.com/kubevm.io/vink/internal/controller/pkg"
	"github.com/kubevm.io/vink/pkg/log"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type StorageReconciler struct {
	Client client.Client
	Cache  cache.Cache
}

func (r *StorageReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	vm := kubevirtv1.VirtualMachine{}
	if err := r.Client.Get(ctx, request.NamespacedName, &vm); err != nil {
		if apierr.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get virtual machine: %w", err)
	}

	diskSet := make(map[string]*DiskStatus, len(vm.Spec.Template.Spec.Volumes)-1)
	for _, vol := range vm.Spec.Template.Spec.Volumes {
		if vol.DataVolume == nil {
			continue
		}

		disk := &DiskStatus{
			Name:    vol.Name,
			Mounted: false,
			Phase:   DiskPhaseUnknown,
		}
		dv := cdiv1beta1.DataVolume{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: vm.Namespace, Name: vol.DataVolume.Name}, &dv); err != nil {
			if !apierr.IsNotFound(err) {
				return ctrl.Result{}, fmt.Errorf("failed to get DataVolume: %w", err)
			}
			disk.Phase = DiskPhaseDataVolumeNotFound
		} else {
			disk.Capacity = dv.Spec.PVC.Resources.Requests.Storage().String()
			disk.StorageClassName = lo.FromPtr(dv.Spec.PVC.StorageClassName)
			if len(dv.Spec.PVC.AccessModes) > 0 {
				disk.AccessMode = string(dv.Spec.PVC.AccessModes[0])
			}
			readyConditions := lo.Filter(dv.Status.Conditions, func(condition cdiv1beta1.DataVolumeCondition, idx int) bool {
				return condition.Type == cdiv1beta1.DataVolumeReady
			})
			disk.Phase = DiskPhaseDataVolumeNotReady
			if len(readyConditions) > 0 && readyConditions[0].Status == corev1.ConditionTrue {
				disk.Phase = DiskPhaseReady
			}
		}

		diskSet[disk.Name] = disk
	}

	for _, disk := range vm.Spec.Template.Spec.Domain.Devices.Disks {
		tmp, ok := diskSet[disk.Name]
		if !ok {
			continue
		}
		if lo.FromPtr(disk.BootOrder) == 1 {
			tmp.Rootfs = true
		}
		tmp.Mounted = true
		diskSet[disk.Name] = tmp
	}

	disks := make([]*DiskStatus, 0, len(diskSet))
	for _, vol := range vm.Spec.Template.Spec.Volumes {
		if vol.DataVolume == nil {
			continue
		}
		if disk, ok := diskSet[vol.Name]; ok {
			disks = append(disks, disk)
		}
	}

	if err := pkg.PatchAnnotations(ctx, r.Client, &vm, annotation.VinkDisks.Name, disks); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

type DiskPhase string

const (
	DiskPhaseUnknown DiskPhase = "Unknown"

	DiskPhaseDataVolumeNotFound DiskPhase = "DataVolumeNotFound"

	// DiskPhaseDataVolumeReady DiskPhase = "DataVolumeReady"

	DiskPhaseDataVolumeNotReady DiskPhase = "DataVolumeNotReady"

	DiskPhaseReady DiskPhase = "Ready"
)

type DiskStatus struct {
	Name string `json:"name"`

	Capacity string `json:"capacity"`

	Phase DiskPhase `json:"phase"`

	AccessMode string `json:"accessMode"`

	StorageClassName string `json:"storageClassName"`

	Rootfs bool `json:"rootfs"`

	Mounted bool `json:"mounted"`
}

func (r *StorageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualmachine-storage").
		For(&kubevirtv1.VirtualMachine{}).
		WatchesMetadata(
			&cdiv1beta1.DataVolume{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				metadata, ok := obj.(*metav1.PartialObjectMetadata)
				if !ok || metadata.Annotations == nil {
					return nil
				}

				list := kubevirtv1.VirtualMachineList{}
				if err := r.Cache.List(ctx, &list, &client.ListOptions{Namespace: metadata.Namespace}); err != nil {
					log.Errorf("Failed to list virtual machines: %v", err)
					return nil
				}

				requests := make([]reconcile.Request, 0)
				for _, vm := range list.Items {
					for _, vol := range vm.Spec.Template.Spec.Volumes {
						if vol.DataVolume == nil {
							continue
						}
						if vol.DataVolume.Name != metadata.Name {
							continue
						}
						requests = append(requests, reconcile.Request{NamespacedName: types.NamespacedName{
							Name:      vm.Name,
							Namespace: vm.Namespace,
						}})
						break
					}
				}

				return requests
			}),
		).
		Complete(r)
}
