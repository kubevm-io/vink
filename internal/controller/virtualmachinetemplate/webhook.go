package virtualmachinetemplate

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const (
	defaultStorageClass = "ceph-block"

	defaultNetworkType = "bridge"

	defaultNetworkNad = "vink/default"

	defaultNetworkSubnet = "vink"
)

var (
	networkField *field.Path = field.NewPath("spec").Child("network").Child("interfaces").Child("type")

	userField *field.Path = field.NewPath("spec").Child("general").Child("user")

	osField *field.Path = field.NewPath("spec").Child("general").Child("os")

	sourceField *field.Path = field.NewPath("spec").Child("general").Child("source")

	storageField *field.Path = field.NewPath("spec").Child("storage")
)

var allowedInterTypes = map[string]struct{}{"bridge": {}, "sriov": {}, "masquerade": {}}

type Webhook struct {
	Client client.Client
}

func (webhook *Webhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&v1alpha1.VirtualMachineTemplate{}).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type.
func (webhook *Webhook) Default(ctx context.Context, obj runtime.Object) error {
	return webhook.mutate(ctx, obj)
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (webhook *Webhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, webhook.validate(ctx, obj)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (webhook *Webhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	return nil, webhook.validate(ctx, newObj)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (webhook *Webhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (webhook *Webhook) mutate(ctx context.Context, obj runtime.Object) error {
	tpl, ok := obj.(*v1alpha1.VirtualMachineTemplate)
	if !ok {
		return fmt.Errorf("object is not a template")
	}

	var allErrs field.ErrorList

	if err := webhook.mutateOs(ctx, tpl); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := webhook.mutateStorage(ctx, tpl); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := webhook.mutateNetwork(ctx, tpl); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := webhook.mutateAccess(ctx, tpl); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return errors.NewInvalid(schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: "VirtualMachineTemplate"}, tpl.Name, allErrs)
}

func (webhook *Webhook) mutateOs(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) *field.Error {
	if tpl.Spec.General.Source.Builtin == nil {
		return nil
	}

	if tpl.Spec.General.Os == nil {
		tpl.Spec.General.Os = &v1alpha1.OperatingSystemSpec{}
	}
	tpl.Spec.General.Os.Name = tpl.Spec.General.Source.Builtin.Distribution
	tpl.Spec.General.Os.Version = tpl.Spec.General.Source.Builtin.Version

	return nil
}

func (webhook *Webhook) mutateStorage(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) *field.Error {
	if len(tpl.Spec.Storage.RootDisk.StorageClass) == 0 {
		tpl.Spec.Storage.RootDisk.StorageClass = defaultStorageClass
	}

	for idx, disk := range tpl.Spec.Storage.DataDisks {
		if len(disk.StorageClass) == 0 {
			tpl.Spec.Storage.DataDisks[idx].StorageClass = defaultStorageClass
		}
	}

	return nil
}

func (webhook *Webhook) mutateNetwork(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) *field.Error {
	if tpl.Spec.Network == nil {
		tpl.Spec.Network = &v1alpha1.NetworkSpec{}
	}

	if len(tpl.Spec.Network.Interfaces) == 0 {
		tpl.Spec.Network.Interfaces = []v1alpha1.NetworkInterface{
			{
				Nad:    defaultNetworkNad,
				Subnet: defaultNetworkSubnet,
				Type:   defaultNetworkType,
			},
		}
	}

	return nil
}

func (webhook *Webhook) mutateAccess(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) *field.Error {
	if tpl.Spec.Access == nil {
		tpl.Spec.Access = &v1alpha1.AccessSpec{}
	}
	if tpl.Spec.Access.Ssh == nil {
		tpl.Spec.Access.Ssh = &v1alpha1.SshAccessSpec{Enabled: true}
	}
	if tpl.Spec.Access.Ssh.Port == 0 {
		tpl.Spec.Access.Ssh.Port = 22
	}
	if tpl.Spec.Access.Console == nil {
		tpl.Spec.Access.Console = &v1alpha1.ConsoleAccessSpec{
			Serial: true,
			Vnc:    true,
		}
	}
	return nil
}

func (webhook *Webhook) validate(ctx context.Context, obj runtime.Object) error {
	tpl, ok := obj.(*v1alpha1.VirtualMachineTemplate)
	if !ok {
		return fmt.Errorf("object is not a template")
	}

	var allErrs field.ErrorList

	if errs := webhook.validateSourceExclusive(ctx, tpl); errs != nil {
		allErrs = append(allErrs, errs...)
	}
	if errs := webhook.validateOs(ctx, tpl); errs != nil {
		allErrs = append(allErrs, errs...)
	}
	if errs := webhook.validateNetwork(ctx, tpl); errs != nil {
		allErrs = append(allErrs, errs...)
	}
	if errs := webhook.validateUsers(ctx, tpl); errs != nil {
		allErrs = append(allErrs, errs...)
	}
	if errs := webhook.validateStorage(ctx, tpl); errs != nil {
		allErrs = append(allErrs, errs...)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(schema.GroupKind{Group: v1alpha1.GroupVersion.Group, Kind: "VirtualMachineTemplate"}, tpl.Name, allErrs)
}

func (webhook *Webhook) validateNetwork(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) (errs field.ErrorList) {
	var (
		network       = tpl.Spec.Network
		duplicatedNad = map[string]int{}
	)

	if network == nil {
		return
	}

	for _, iface := range network.Interfaces {
		duplicatedNad[iface.Nad]++
		if _, ok := allowedInterTypes[iface.Type]; !ok {
			errs = append(errs, field.Invalid(networkField, iface.Type, "invalid interface type"))
		}
	}
	for name, count := range duplicatedNad {
		if count > 1 {
			errs = append(errs, field.Invalid(networkField, name, "duplicated nad"))
		}
	}
	for _, iface := range network.Interfaces {
		if len(iface.Subnet) == 0 {
			errs = append(errs, field.Invalid(networkField, iface.Subnet, "subnet cannot be empty"))
		}
	}
	return
}

func (webhook *Webhook) validateUsers(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) (errs field.ErrorList) {
	var users = tpl.Spec.General.Users
	for _, user := range users {
		var (
			passwordCount = 0
			sshKeyCount   = 0
		)
		if len(user.Password) > 0 {
			passwordCount++
		}
		if len(user.PasswordBase64) > 0 {
			passwordCount++
		}
		if len(user.PasswordSecretRef) > 0 {
			passwordCount++
		}
		if len(user.SshKey) > 0 {
			sshKeyCount++
		}
		if len(user.SshKeyBase64) > 0 {
			sshKeyCount++
		}
		if passwordCount > 1 {
			errs = append(errs, field.Invalid(userField, user, "only one of password, passwordBase64, passwordSecretRef can be set"))
		}
		if sshKeyCount > 1 {
			errs = append(errs, field.Invalid(userField, user, "only one of sshKey, sshKeyBase64, sshKeySecretRef can be set"))
		}
	}
	return
}

func (webhook *Webhook) validateOs(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) (errs field.ErrorList) {
	var (
		os     = tpl.Spec.General.Os
		source = tpl.Spec.General.Source
	)

	if source.Builtin == nil && (os == nil || len(os.Name) == 0) {
		errs = append(errs, field.Invalid(osField, os, "os is required"))
	}

	if source.Builtin != nil && os != nil {
		if os.Name != source.Builtin.Distribution && os.Version != source.Builtin.Version {
			errs = append(errs, field.Invalid(osField, os, "os name and version must be the same as the builtin image"))
		}
	}
	return
}

func (webhook *Webhook) validateStorage(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) (errs field.ErrorList) {
	var (
		storage = tpl.Spec.Storage
	)

	if len(storage.RootDisk.Size) == 0 {
		errs = append(errs, field.Invalid(storageField.Child("rootDisk").Child("size"), storage.RootDisk.Size, "root disk size must be specified"))
	} else if !isValidDiskSize(storage.RootDisk.Size) {
		errs = append(errs, field.Invalid(storageField.Child("rootDisk").Child("size"), storage.RootDisk.Size, "root disk size must be Gi or Mi"))
	}
	for idx, disk := range storage.DataDisks {
		if len(disk.Size) == 0 {
			errs = append(errs, field.Invalid(storageField.Child("dataDisks").Index(idx).Child("size"), disk.Size, "data disk size must be specified"))
		} else if !isValidDiskSize(disk.Size) {
			errs = append(errs, field.Invalid(storageField.Child("dataDisks").Index(idx).Child("size"), disk.Size, "data disk size must be Gi or Mi"))
		}
	}

	return
}

func (webhook *Webhook) validateSourceExclusive(_ context.Context, tpl *v1alpha1.VirtualMachineTemplate) (errs field.ErrorList) {
	var (
		count  = 0
		source = tpl.Spec.General.Source
	)
	if source.Builtin != nil {
		count++
	}
	if source.Registry != nil {
		count++
	}
	if source.Http != nil {
		count++
	}
	if source.S3 != nil {
		count++
	}
	if source.Pvc != nil {
		count++
	}
	if source.DataVolume != nil {
		count++
	}
	if count != 1 {
		errs = append(errs, field.Invalid(sourceField, source, fmt.Sprintf("exactly one of [builtin, registry, http, s3, pvc, dataVolume] must be set, but got %d", count)))
	}
	return
}

func isValidDiskSize(size string) bool {
	size = strings.TrimSpace(size)
	if size == "" {
		return false
	}

	match, _ := regexp.MatchString(`^\d+(Gi|Mi)$`, size)
	return match
}
