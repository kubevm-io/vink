package virtualmachinetemplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	Client client.Client
	Cache  cache.Cache
}

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	tpl := &v1alpha1.VirtualMachineTemplate{}
	if err := r.Client.Get(ctx, request.NamespacedName, tpl); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileDataVolume(ctx, tpl); err != nil {
		if apierr.IsConflict(err) {
			return ctrl.Result{Requeue: true}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to reconcile DataVolume: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) ensureOsValid(ctx context.Context, tpl *v1alpha1.VirtualMachineTemplate) error {
	if tpl.Status.Os == nil {
		return nil
	}
	dv := cdiv1.DataVolume{}
	key := client.ObjectKey{Namespace: tpl.Status.Os.Namespace, Name: tpl.Status.Os.Name}
	if err := r.Client.Get(ctx, key, &dv); err != nil {
		if apierr.IsNotFound(err) {
			tpl.Status.Os = nil
			return r.Client.Status().Update(ctx, tpl)
		}
		return err
	}
	if dv.Spec.Source.HTTP != nil && dv.Spec.Source.HTTP.URL != tpl.Spec.General.Source.Http.Url {
		if err := r.Client.Delete(ctx, &dv); err != nil && !apierr.IsNotFound(err) {
			return err
		}
		tpl.Status.Os = nil
		return r.Client.Status().Update(ctx, tpl)
	}
	if dv.Labels == nil {
		dv.Labels = make(map[string]string)
	}
	if dv.Labels["vink.kubevm.io/os.name"] != tpl.Spec.General.Os.Name || dv.Labels["vink.kubevm.io/os.version"] != tpl.Spec.General.Os.Version {
		dv.Labels["vink.kubevm.io/os.name"] = tpl.Spec.General.Os.Name
		dv.Labels["vink.kubevm.io/os.version"] = tpl.Spec.General.Os.Version
		return r.Client.Update(ctx, &dv)
	}
	return nil
}

func (r *Reconciler) reconcileDataVolume(ctx context.Context, tpl *v1alpha1.VirtualMachineTemplate) error {
	if err := r.ensureOsValid(ctx, tpl); err != nil {
		return err
	}
	if tpl.Status.Os != nil {
		return nil
	}

	dv, err := r.createOs(ctx, tpl)
	if err != nil {
		return fmt.Errorf("failed to create DataVolume for VirtualMachineTemplate %q: %w", tpl.Name, err)
	}

	tpl.Status.Os = &v1alpha1.DataVolumeRef{
		Name:      dv.Name,
		Namespace: dv.Namespace,
	}
	if err := r.Client.Status().Update(ctx, tpl); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) createOs(ctx context.Context, tpl *v1alpha1.VirtualMachineTemplate) (*cdiv1.DataVolume, error) {
	dv := cdiv1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", strings.ToLower(tpl.Spec.General.Os.Name)),
			Namespace:    tpl.Namespace,
			Annotations: map[string]string{
				"cdi.kubevirt.io/storage.bind.immediate.requested": "true",
			},
			Labels: map[string]string{
				"vink.kubevm.io/os.name":    tpl.Spec.General.Os.Name,
				"vink.kubevm.io/os.version": tpl.Spec.General.Os.Version,
			},
			OwnerReferences: []metav1.OwnerReference{
				newOwnerReferenceFor(tpl),
			},
		},
		Spec: cdiv1.DataVolumeSpec{
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					tpl.Spec.Storage.RootDisk.AccessMode,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("10Gi"),
					},
				},
				StorageClassName: lo.ToPtr("ceph-block"),
			},
			Source: buildImageSource(tpl.Spec.General.Source),
		},
	}
	return &dv, r.Client.Create(ctx, &dv)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("virtualmachinetemplate").
		For(&v1alpha1.VirtualMachineTemplate{}).
		Complete(r)
}

func newOwnerReferenceFor(tpl *v1alpha1.VirtualMachineTemplate) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion:         tpl.APIVersion,
		Kind:               tpl.Kind,
		Name:               tpl.Name,
		UID:                tpl.UID,
		Controller:         lo.ToPtr(true),
		BlockOwnerDeletion: lo.ToPtr(true),
	}
}

func buildImageSource(source *v1alpha1.ImageSource) *cdiv1.DataVolumeSource {
	switch {
	case source.Builtin != nil:
		return nil
	case source.Http != nil:
		return &cdiv1.DataVolumeSource{HTTP: &cdiv1.DataVolumeSourceHTTP{URL: source.Http.Url}}
	case source.S3 != nil:
		return &cdiv1.DataVolumeSource{S3: &cdiv1.DataVolumeSourceS3{URL: source.S3.Url}}
	case source.Registry != nil:
		return &cdiv1.DataVolumeSource{Registry: &cdiv1.DataVolumeSourceRegistry{URL: &source.Registry.Url}}
	case source.Pvc != nil:
		return &cdiv1.DataVolumeSource{PVC: &cdiv1.DataVolumeSourcePVC{Name: source.Pvc.Name}}
	case source.DataVolume != nil:
		return nil
	default:
		return nil
	}
}
