package vink

// import (
// 	"context"
// 	"fmt"

// 	cattlev1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
// 	"github.com/kubevm.io/vink/config"
// 	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
// 	corev1 "k8s.io/api/core/v1"
// 	apierr "k8s.io/apimachinery/pkg/api/errors"
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	ctrl "sigs.k8s.io/controller-runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/cache"
// 	"sigs.k8s.io/controller-runtime/pkg/client"
// 	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
// )

// type Reconciler struct {
// 	Client client.Client
// 	Cache  cache.Cache
// 	Scheme *runtime.Scheme

// 	config *config.Config
// }

// func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
// 	vink := &v1alpha1.Vink{}
// 	if err := r.Client.Get(ctx, request.NamespacedName, vink); err != nil {
// 		return ctrl.Result{}, client.IgnoreNotFound(err)
// 	}

// 	if err := r.reconcileComponents(ctx, vink); err != nil {
// 		if apierr.IsConflict(err) {
// 			return ctrl.Result{Requeue: true}, nil
// 		}
// 		return ctrl.Result{}, err
// 	}

// 	return ctrl.Result{}, nil
// }

// func (r *Reconciler) reconcileComponents(ctx context.Context, vink *v1alpha1.Vink) error {
// 	if err := r.reconcileKubeOVN(ctx, vink); err != nil {
// 		return err
// 	}
// 	if err := r.reconcileMultus(ctx, vink); err != nil {
// 		return err
// 	}
// }

// func (r *Reconciler) reconcileKubeOVN(ctx context.Context, vink *v1alpha1.Vink) error {
// 	const (
// 		kubeovnChartName = "kube-ovn"
// 		kubeovnNamespace = "kube-system"
// 	)
// 	helmChartKey := client.ObjectKey{Namespace: r.config.Namespace, Name: kubeovnChartName}

// 	helmChartConfig := &cattlev1.HelmChartConfig{}
// 	if err := r.Cache.Get(ctx, helmChartKey, helmChartConfig); err != nil {
// 		if !apierr.IsNotFound(err) {
// 			return err
// 		}
// 		helmChartConfig = NewKubeOvnHelmChartConfig()
// 		controllerutil.SetOwnerReference(vink, helmChartConfig, r.Scheme)
// 		if err := r.Client.Create(ctx, helmChartConfig); err != nil {
// 			return err
// 		}
// 	}

// 	helmChart := &cattlev1.HelmChart{}
// 	if err := r.Cache.Get(ctx, helmChartKey, helmChart); err != nil {
// 		if !apierr.IsNotFound(err) {
// 			return err
// 		}
// 		if err := r.labelKubeOvnNodes(ctx); err != nil {
// 			return err
// 		}
// 		helmChart = NewHelmChart(kubeovnNamespace, kubeovnChartName)
// 		controllerutil.SetOwnerReference(vink, helmChart, r.Scheme)
// 		if err := r.Client.Create(ctx, helmChart); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (r *Reconciler) reconcileMultus(ctx context.Context, vink *v1alpha1.Vink) error {
// 	const (
// 		multusChartName  = "multus"
// 		kubeovnNamespace = "kube-system"
// 	)
// 	helmChartKey := client.ObjectKey{Namespace: r.config.Namespace, Name: multusChartName}

// 	helmChartConfig := &cattlev1.HelmChartConfig{}
// 	if err := r.Cache.Get(ctx, helmChartKey, helmChartConfig); err != nil {
// 		if !apierr.IsNotFound(err) {
// 			return err
// 		}
// 		helmChartConfig = NewMonitoringHelmChartConfig()
// 		controllerutil.SetOwnerReference(vink, helmChartConfig, r.Scheme)
// 		if err := r.Client.Create(ctx, helmChartConfig); err != nil {
// 			return err
// 		}
// 	}

// 	helmChart := &cattlev1.HelmChart{}
// 	if err := r.Cache.Get(ctx, helmChartKey, helmChart); err != nil {
// 		if !apierr.IsNotFound(err) {
// 			return err
// 		}
// 		helmChart = NewHelmChart(kubeovnNamespace, multusChartName)
// 		controllerutil.SetOwnerReference(vink, helmChart, r.Scheme)
// 		if err := r.Client.Create(ctx, helmChart); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (r *Reconciler) labelKubeOvnNodes(ctx context.Context) error {
// 	type rule struct {
// 		selector string
// 		labels   map[string]string
// 	}

// 	rules := []rule{
// 		{
// 			selector: "beta.kubernetes.io/os=linux",
// 			labels:   map[string]string{"kubernetes.io/os": "linux"},
// 		},
// 		{
// 			selector: "node-role.kubernetes.io/control-plane",
// 			labels:   map[string]string{"kube-ovn/role": "master"},
// 		},
// 	}

// 	for _, rule := range rules {
// 		sel, err := labels.Parse(rule.selector)
// 		if err != nil {
// 			return fmt.Errorf("invalid selector %q: %w", rule.selector, err)
// 		}

// 		var nodeList corev1.NodeList
// 		if err := r.Client.List(ctx, &nodeList, &client.ListOptions{LabelSelector: sel}); err != nil {
// 			return fmt.Errorf("list nodes for selector %q: %w", rule.selector, err)
// 		}

// 		for _, node := range nodeList.Items {
// 			changed := false
// 			if node.Labels == nil {
// 				node.Labels = map[string]string{}
// 			}
// 			for k, v := range rule.labels {
// 				if node.Labels[k] != v {
// 					node.Labels[k] = v
// 					changed = true
// 				}
// 			}
// 			if changed {
// 				if err := r.Client.Update(ctx, &node); err != nil {
// 					return fmt.Errorf("update node %s: %w", node.Name, err)
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }

// func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		Named("vink").
// 		For(&v1alpha1.Vink{}).
// 		Complete(r)
// }
