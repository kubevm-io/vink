package vink

import (
	cattlev1 "github.com/k3s-io/helm-controller/pkg/apis/helm.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	VinkNamespace = "vink"

	
)

func NewHelmChart(namespace, name string) *cattlev1.HelmChart {
	return &cattlev1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartSpec{
			Repo:            "https://kubevm-io.github.io/helm-charts",
			TargetNamespace: namespace,
			Chart:           name,
			CreateNamespace: true,
			// ValuesContent: ,
		},
	}
}

func NewKubeOvnHelmChartConfig() *cattlev1.HelmChartConfig {
	return &cattlev1.HelmChartConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-ovn",
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartConfigSpec{
			ValuesContent: `
masterNodes: 198.19.249.103
masterNodesLabel: node-role.kubernetes.io/control-plane=true
cni:
  configPriority: "00"
central:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
ovn:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
controller:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
cni:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
pinger:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
monitor:
  resources:
    requests:
      cpu: 10m
      memory: 64Mi
`,
		},
	}
}

func NewKubeVirtHelmChartConfig() *cattlev1.HelmChartConfig {
	return &cattlev1.HelmChartConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-virt",
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartConfigSpec{
			ValuesContent: `
kubevirt:
  configuration:
    developerConfiguration:
      useEmulation: true
`,
		},
	}
}

func NewMonitoringHelmChartConfig() *cattlev1.HelmChartConfig {
	return &cattlev1.HelmChartConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "monitoring",
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartConfigSpec{
			ValuesContent: `
alertmanager:
  enabled: false

kubeStateMetrics:
  enabled: true

prometheus:
  prometheusSpec:
    serviceMonitorSelectorNilUsesHelmValues: false

grafana:
  grafana.ini:
    auth.anonymous:
      enabled: true
      org_role: Viewer
    http:
      enable_cors: true
      allow_from_origin: "*"
    server:
      root_url: /grafana
      serve_from_sub_path: true
    security:
      allow_embedding: true

prometheus-node-exporter:
  prometheus:
    monitor:
      attachMetadata:
        node: true
      relabelings:
        - action: replace
          sourceLabels: [__meta_kubernetes_node_name]
          targetLabel: nodename
`,
		},
	}
}

func NewRookCephClusterHelmChartConfig() *cattlev1.HelmChartConfig {
	return &cattlev1.HelmChartConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rook-ceph-cluster",
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartConfigSpec{
			ValuesContent: `
toolbox:
  enabled: true
  resources:
    requests:
      cpu: "0"
      memory: "0"

cephClusterSpec:
  crashCollector:
    disable: false
  mon:
    count: 1
    allowMultiplePerNode: true
  mgr:
    count: 1
    allowMultiplePerNode: true
  resources:
    mgr:
      limits:
        memory: 3Gi
      requests:
        cpu: "0"
        memory: "0"
    mon:
      requests:
        cpu: "0"
        memory: "0"
    osd:
      requests:
        cpu: "0"
        memory: "0"
    prepareosd:
      requests:
        cpu: "0"
        memory: "0"
    mgr-sidecar:
      requests:
        cpu: "0"
        memory: "0"
    crashcollector:
      requests:
        cpu: "0"
        memory: "0"
    logcollector:
      requests:
        cpu: "0"
        memory: "0"
    cleanup:
      requests:
        cpu: "0"
        memory: "0"
    exporter:
      requests:
        cpu: "0"
        memory: "0"

cephFileSystems:
  - name: ceph-filesystem
    spec:
      metadataPool:
        replicated:
          size: 1
      dataPools:
        - failureDomain: host
          replicated:
            size: 1
          name: data0
      metadataServer:
        activeCount: 1
        activeStandby: true
        resources:
          limits:
            memory: "0"
          requests:
            cpu: "0"
            memory: "0"
        priorityClassName: system-cluster-critical
    storageClass:
      enabled: true
      name: ceph-filesystem
      isDefault: false
      pool: data0
      reclaimPolicy: Delete
      allowVolumeExpansion: true
      volumeBindingMode: Immediate
      annotations: {}
      labels: {}
      mountOptions: []
      parameters:
        csi.storage.k8s.io/provisioner-secret-name: rook-csi-cephfs-provisioner
        csi.storage.k8s.io/provisioner-secret-namespace: rook-ceph
        csi.storage.k8s.io/controller-expand-secret-name: rook-csi-cephfs-provisioner
        csi.storage.k8s.io/controller-expand-secret-namespace: rook-ceph
        csi.storage.k8s.io/node-stage-secret-name: rook-csi-cephfs-node
        csi.storage.k8s.io/node-stage-secret-namespace: rook-ceph
        csi.storage.k8s.io/fstype: ext4

cephObjectStores:
  - name: ceph-objectstore
    spec:
      metadataPool:
        failureDomain: host
        replicated:
          size: 1
      dataPool:
        failureDomain: host
        erasureCoded:
          dataChunks: 2
          codingChunks: 1
        parameters:
          bulk: "true"
      preservePoolsOnDelete: true
      gateway:
        port: 80
        resources:
          limits:
            memory: "0"
          requests:
            cpu: "0"
            memory: "0"
        instances: 1
        priorityClassName: system-cluster-critical
    storageClass:
      enabled: false
      name: ceph-bucket
      reclaimPolicy: Delete
      volumeBindingMode: Immediate
      annotations: {}
      labels: {}
      parameters:
        region: us-east-1
    ingress:
      enabled: false

cephBlockPools:
  - name: ceph-blockpool
    spec:
      failureDomain: host
      replicated:
        size: 1
    storageClass:
      enabled: true
      name: ceph-block
      isDefault: true
      reclaimPolicy: Delete
      allowVolumeExpansion: true
      volumeBindingMode: Immediate
      annotations: {}
      labels: {}
      mountOptions: []
      allowedTopologies: []
      parameters:
        imageFormat: "2"
        imageFeatures: layering
        csi.storage.k8s.io/provisioner-secret-name: rook-csi-rbd-provisioner
        csi.storage.k8s.io/provisioner-secret-namespace: rook-ceph
        csi.storage.k8s.io/controller-expand-secret-name: rook-csi-rbd-provisioner
        csi.storage.k8s.io/controller-expand-secret-namespace: rook-ceph
        csi.storage.k8s.io/node-stage-secret-name: rook-csi-rbd-node
        csi.storage.k8s.io/node-stage-secret-namespace: rook-ceph
        csi.storage.k8s.io/fstype: ext4

monitoring:
  enabled: true
`,
		},
	}
}

func NewRookCephHelmChartConfig() *cattlev1.HelmChartConfig {
	return &cattlev1.HelmChartConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rook-ceph",
			Namespace: "vink",
		},
		Spec: cattlev1.HelmChartConfigSpec{
			ValuesContent: `
resources:
  requests:
    cpu: "0"
    memory: "0"

csi:
  enableRbdDriver: true
  enableCephfsDriver: true
  provisionerReplicas: 1

  csiRBDProvisionerResource: |
    - name: csi-provisioner
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-resizer
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-attacher
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-snapshotter
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-rbdplugin
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-omap-generator
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: liveness-prometheus
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"

  csiRBDPluginResource: |
    - name: driver-registrar
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-rbdplugin
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: liveness-prometheus
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"

  csiCephFSProvisionerResource: |
    - name: csi-provisioner
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-resizer
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-attacher
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-snapshotter
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-cephfsplugin
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: liveness-prometheus
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"

  csiCephFSPluginResource: |
    - name: driver-registrar
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-cephfsplugin
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: liveness-prometheus
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"

  csiNFSProvisionerResource: |
    - name: csi-provisioner
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-nfsplugin
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-attacher
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"

  csiNFSPluginResource: |
    - name: driver-registrar
      resource:
        requests:
          cpu: "0"
          memory: "0"
        limits:
          memory: "0"
    - name: csi-nfsplugin
      resource:
        requests:
          cpu: "0"
`,
		},
	}
}
