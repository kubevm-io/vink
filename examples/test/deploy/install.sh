#!/bin/bash

set -euo pipefail
set -x

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"


echo "Installing Kube-OVN"
helm repo add kubeovn https://kubeovn.github.io/kube-ovn --force-update
helm repo update kubeovn

kubectl label node -l beta.kubernetes.io/os=linux kubernetes.io/os=linux --overwrite
kubectl label node -l node-role.kubernetes.io/control-plane kube-ovn/role=master --overwrite

MASTER_NODE_IPS=$(kubectl get nodes -l node-role.kubernetes.io/control-plane -o=jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}' | tr ' ' ',')

helm upgrade --install --create-namespace kube-ovn kubeovn/kube-ovn \
    --namespace kube-system \
    --set MASTER_NODES=${MASTER_NODE_IPS} \
    -f ${DIR}/kubeovn/values.yaml \
    --wait \
    --timeout 600s \
    --debug

echo "Installing Multus CNI"
kubectl apply -n kube-system -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml

echo "Installing Rook Ceph"
helm repo add rook-release https://charts.rook.io/release --force-update
helm repo update rook-release

helm upgrade --install --create-namespace rook-ceph rook-release/rook-ceph \
    --namespace rook-ceph \
    -f ${DIR}/rook-ceph/values.yaml \
    --wait \
    --timeout 600s \
    --debug

helm upgrade --install --create-namespace rook-ceph-cluster rook-release/rook-ceph-cluster \
    --namespace rook-ceph \
    -f ${DIR}/rook-ceph-cluster/values.yaml \
    --wait \
    --timeout 600s \
    --debug

echo "Installing Monitor"
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts --force-update
helm repo update prometheus-community
helm upgrade --install --create-namespace --namespace monitoring monitoring \
    prometheus-community/kube-prometheus-stack \
    -f ${DIR}/monitoring/values.yaml \
    --wait \
    --timeout 600s \
    --debug

echo "Installing Snapshotter"
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotclasses.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshotcontents.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/client/config/crd/snapshot.storage.k8s.io_volumesnapshots.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/rbac-snapshot-controller.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/snapshot-controller/setup-snapshot-controller.yaml

kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/csi-snapshotter/rbac-csi-snapshotter.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/csi-snapshotter/rbac-external-provisioner.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes-csi/external-snapshotter/master/deploy/kubernetes/csi-snapshotter/setup-csi-snapshotter.yaml

echo "Installing KubeVirt"
helm upgrade --install --create-namespace --namespace kubevirt kubevirt \
    oci://registry-1.docker.io/hejianmin/kubevirt --version 0.0.1 \
    -f ${DIR}/kubevirt/values.yaml \
    --wait \
    --timeout 600s \
    --debug

echo "Installing CDI"
helm upgrade --install --create-namespace --namespace cdi cdi \
    oci://registry-1.docker.io/hejianmin/cdi --version 0.0.1 \
    --wait \
    --timeout 600s \
    --debug

echo "Installing Vink"
kubectl apply -f vink/namespace.yaml
kubectl apply -f vink/
