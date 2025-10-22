package virtualmachine

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Webhook struct {
	Client client.Client
}

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (webhook *Webhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&kubevirtv1.VirtualMachine{}).
		WithDefaulter(webhook).
		// WithValidator(webhook).
		Complete()
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (webhook *Webhook) Default(ctx context.Context, obj runtime.Object) error {
	vm, ok := obj.(*kubevirtv1.VirtualMachine)
	if !ok {
		return fmt.Errorf("object is not a template")
	}

	fmt.Println(vm)

	return nil

	// if len(vm.Labels[LabelTpl]) == 0 {
	// 	return nil
	// }

	// return fmt.Errorf("this is a test")
}

// // ValidateCreate implements webhook.Validator so a webhook will be registered for the type
// func (webhook *Webhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
// 	return nil, nil
// }

// // ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
// func (webhook *Webhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
// 	return nil, nil
// }

// // ValidateDelete implements webhook.Validator so a webhook will be registered for the type
// func (webhook *Webhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
// 	return nil, nil
// }

// func (webhook *Webhook) validate(ctx context.Context, obj runtime.Object) error {
// 	vm, ok := obj.(*kubevirtv1.VirtualMachine)
// 	if !ok {
// 		return fmt.Errorf("object is not a template")
// 	}

// 	// if len(vm.Labels[LabelTpl]) == 0 {
// 	// 	return nil
// 	// }
// }
