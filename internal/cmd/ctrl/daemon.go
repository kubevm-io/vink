package ctrl

import (
	"context"
	"path"

	netv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	kubeovn "github.com/kubeovn/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/kubevm.io/vink/config"
	"github.com/kubevm.io/vink/internal/controller/virtualmachinetemplate"
	"github.com/kubevm.io/vink/internal/controller/virtualmachineclaim"
	poolv1alpha1 "kubevirt.io/api/pool/v1alpha1"

	// "github.com/kubevm.io/vink/internal/controller/template_instance"

	// "github.com/kubevm.io/vink/internal/controller"
	// "github.com/kubevm.io/vink/internal/controller/node"
	// "github.com/kubevm.io/vink/internal/controller/virtualmachine"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	uruntime "k8s.io/apimachinery/pkg/util/runtime"
	virtv1 "kubevirt.io/api/core/v1"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var scheme = runtime.NewScheme()

func init() {
	uruntime.Must(kubeovn.AddToScheme(scheme))
	uruntime.Must(v1alpha1.AddToScheme(scheme))
	uruntime.Must(corev1.AddToScheme(scheme))
	uruntime.Must(virtv1.AddToScheme(scheme))
	uruntime.Must(cdiv1beta1.AddToScheme(scheme))
	uruntime.Must(netv1.AddToScheme(scheme))
	uruntime.Must(poolv1alpha1.AddToScheme(scheme))
}

func New(cfg *config.Config) *Daemon {
	return &Daemon{config: cfg}
}

type Daemon struct {
	config *config.Config
}

func (dm *Daemon) Execute(ctx context.Context) error {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Development: true,
	})))

	mgr, err := ctrl.NewManager(clients.Clients.Config(), ctrl.Options{
		Scheme:                  scheme,
		LeaderElectionID:        "vink.kubevm.io",
		LeaderElectionNamespace: "vink",
		LeaderElection:          false,
		Metrics: server.Options{
			BindAddress: "0",
		},
		WebhookServer: webhook.NewServer(webhook.Options{
			Port: 8081,
			// CertDir: path.Dir("/Users/hetongxue/workspace/examples/operator/example/secret/"),
			CertDir: path.Dir("examples/secret/"),
		}),
	})
	if err != nil {
		return err
	}

	if err := (&virtualmachinetemplate.Reconciler{
		Client: mgr.GetClient(),
		Cache:  mgr.GetCache(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// if err := (&template_instance.Reconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	if err := (&virtualmachinetemplate.Webhook{
		Client: mgr.GetClient(),
	}).SetupWebhookWithManager(mgr); err != nil {
		return err
	}

	// if err := (&virtualmachine.HostReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.NetworkReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.StorageReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.Reconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	if err := (&virtualmachineclaim.Reconciler{
		Client: mgr.GetClient(),
		Cache:  mgr.GetCache(),
	}).SetupWithManager(mgr); err != nil {
		return err
	}

	// if err := (&virtualmachine.Webhook{
	// 	Client: mgr.GetClient(),
	// }).SetupWebhookWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&controller.DataVolumeOwnerReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.NetworkReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.OperatingSystemReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.HostReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// if err := (&virtualmachine.DiskReconciler{
	// 	Client: mgr.GetClient(),
	// 	Cache:  mgr.GetCache(),
	// }).SetupWithManager(mgr); err != nil {
	// 	return err
	// }

	// monitor, err := virtualmachine.NewCollector(mgr.GetClient(), mgr.GetCache())
	// if err != nil {
	// 	return err
	// }
	// go monitor.Collector(ctx, dm.config.MonitorInterval)

	// nodeMonitor, err := node.NewCollector(mgr.GetClient(), mgr.GetCache())
	// if err != nil {
	// 	return err
	// }
	// go nodeMonitor.Collector(ctx, dm.config.MonitorInterval)

	// storage, err := node.NewCephStorageCollector(mgr.GetClient(), mgr.GetCache())
	// if err != nil {
	// 	return err
	// }
	// go storage.Collector(ctx, dm.config.MonitorInterval)

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	return mgr.Start(ctx)
}

func (dm *Daemon) Shutdown() error {
	return nil
}
