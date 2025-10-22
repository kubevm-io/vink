package virtualmachineclaim

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"github.com/kubevm.io/vink/pkg/log"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"

	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Client client.Client
	Cache  cache.Cache
}

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	vmc := &v1alpha1.VirtualMachineClaim{}
	if err := r.Client.Get(ctx, request.NamespacedName, vmc); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var tpl v1alpha1.VirtualMachineTemplate
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: vmc.Spec.TemplateRef.Namespace, Name: vmc.Spec.TemplateRef.Name}, &tpl); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Template %q in namespace %q for VirtualMachineClaim %q: %w", vmc.Spec.TemplateRef.Name, vmc.Namespace, vmc.Name, err)
	}

	statusCopy := vmc.Status.DeepCopy()
	statusCopy.Reason = ""
	statusCopy.Phase = v1alpha1.VirtualMachineClaimPhaseCompleted
	if err := r.reconcileVirtualMachine(ctx, &tpl, vmc); err != nil {
		if apierr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		statusCopy.Reason = err.Error()
		statusCopy.Phase = v1alpha1.VirtualMachineClaimPhaseFailed
		log.Errorf("Failed to reconcile VirtualMachine for VirtualMachineClaim %q: %v", vmc.Name, err)
	} else {
		statusCopy.VirtualMachineRef = toObjectRefFromVirtualMachine(vmc)
	}
	if reflect.DeepEqual(&vmc.Status, statusCopy) {
		return ctrl.Result{}, nil
	}

	vmc.Status = lo.FromPtr(statusCopy)
	if err := r.Client.Status().Update(ctx, vmc); err != nil {
		if apierr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update VirtualMachineClaim %q: %w", vmc.Name, err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileVirtualMachine(ctx context.Context, tpl *v1alpha1.VirtualMachineTemplate, vmc *v1alpha1.VirtualMachineClaim) error {
	vm := kubevirtv1.VirtualMachine{}
	if err := r.Client.Get(ctx, client.ObjectKeyFromObject(vmc), &vm); err != nil && !apierr.IsNotFound(err) {
		return fmt.Errorf("failed to get VirtualMachine %q in namespace %q for VirtualMachineClaim %q: %w", vmc.Name, vmc.Namespace, vmc.Name, err)
	} else if err == nil {
		return r.reconcileStatusForUnownedVM(ctx, vmc, &vm)
	}

	vmbuilder := NewVirtualMachineBuilder(tpl, vmc, r.Client)
	vmcfg, err := vmbuilder.Build(ctx)
	if err != nil {
		return fmt.Errorf("failed to build VirtualMachine from VirtualMachineClaim %q: %w", vmc.Name, err)
	}

	if err := r.Client.Create(ctx, vmcfg); err != nil {
		return fmt.Errorf("failed to create VirtualMachine %q: %w", vmcfg.Name, err)
	}
	return nil
}

func (r *Reconciler) reconcileStatusForUnownedVM(ctx context.Context, vmc *v1alpha1.VirtualMachineClaim, vm *kubevirtv1.VirtualMachine) error {
	statusCopy := vmc.Status.DeepCopy()
	if vm.Labels == nil || vm.Labels[appCreatedByLabel] != v1alpha1.GroupVersion.Group {
		statusCopy.Phase = v1alpha1.VirtualMachineClaimPhaseFailed
		statusCopy.VirtualMachineRef = toObjectRefFromVirtualMachine(vmc)
		statusCopy.Reason = fmt.Sprintf("VirtualMachine %q is not owned by VirtualMachineClaim controller", vm.Name)
	}
	if reflect.DeepEqual(&vmc.Status, statusCopy) {
		return nil
	}

	vmc.Status = lo.FromPtr(statusCopy)
	if err := r.Client.Status().Update(ctx, vmc); err != nil {
		return fmt.Errorf("failed to update VirtualMachineClaim %q: %w", vmc.Name, err)
	}
	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualmachineclaim").
		For(&v1alpha1.VirtualMachineClaim{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}

func toObjectRefFromVirtualMachine(vmc *v1alpha1.VirtualMachineClaim) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:       "VirtualMachine",
		Namespace:  vmc.Namespace,
		Name:       vmc.Name,
		APIVersion: vmc.APIVersion,
	}
}
