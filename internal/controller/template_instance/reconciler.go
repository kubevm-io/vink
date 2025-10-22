package template_instance

import (
	"context"
	"fmt"
	"reflect"

	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"github.com/samber/lo"
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
	tplInstance := &v1alpha1.TemplateInstance{}
	if err := r.Client.Get(ctx, request.NamespacedName, tplInstance); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	vm := kubevirtv1.VirtualMachine{}
	if err := r.Client.Get(ctx, request.NamespacedName, &vm); err != nil && !apierr.IsNotFound(err) {
		return ctrl.Result{}, fmt.Errorf("failed to get VirtualMachine %q in namespace %q for TemplateInstance %q: %w", request.Name, request.Namespace, tplInstance.Name, err)
	} else if err == nil {
		if err := r.reconcileStatusForUnownedVM(ctx, tplInstance, &vm); err != nil {
			if apierr.IsConflict(err) {
				return ctrl.Result{Requeue: true}, nil
			}
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	var tpl v1alpha1.VirtualMachineTemplate
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: tplInstance.Namespace, Name: tplInstance.Spec.Template}, &tpl); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to get Template %q in namespace %q for TemplateInstance %q: %w", tplInstance.Spec.Template, tplInstance.Namespace, tplInstance.Name, err)
	}

	vmCfg, err := r.buildVirtualMachineFromTemplate(ctx, &tpl, tplInstance)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to build VirtualMachine from Template %q: %w", tpl.Name, err)
	}

	// data, _ := json.MarshalIndent(vmCfg, "", "  ")
	// fmt.Println(string(data))

	if err := r.Client.Create(ctx, vmCfg); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create VirtualMachine %q: %w", vmCfg.Name, err)
	}

	statusCopy := tplInstance.Status.DeepCopy()
	statusCopy.Applied = true
	statusCopy.VirtualMachine = vmCfg.Name
	statusCopy.Reason = ""
	if reflect.DeepEqual(&tplInstance.Status, statusCopy) {
		return ctrl.Result{}, nil
	}

	tplInstance.Status = lo.FromPtr(statusCopy)
	if err := r.Client.Status().Update(ctx, tplInstance); err != nil {
		if apierr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to update TemplateInstance %q: %w", tplInstance.Name, err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) reconcileStatusForUnownedVM(ctx context.Context, tplInstance *v1alpha1.TemplateInstance, vm *kubevirtv1.VirtualMachine) error {
	statusCopy := tplInstance.Status.DeepCopy()
	if vm.Labels == nil || vm.Labels[appCreatedByLabel] != v1alpha1.GroupVersion.Group {
		statusCopy.Applied = false
		statusCopy.VirtualMachine = vm.Name
		statusCopy.Reason = fmt.Sprintf("VirtualMachine %q is not owned by TemplateInstance controller", vm.Name)
	}
	if reflect.DeepEqual(&tplInstance.Status, statusCopy) {
		return nil
	}

	tplInstance.Status = lo.FromPtr(statusCopy)
	if err := r.Client.Status().Update(ctx, tplInstance); err != nil {
		return fmt.Errorf("failed to update TemplateInstance %q: %w", tplInstance.Name, err)
	}
	return nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("templateinstance").
		For(&v1alpha1.TemplateInstance{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
