package business

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/kubevm.io/vink/apis/types"
	"github.com/kubevm.io/vink/internal/controller/pkg"
	"github.com/kubevm.io/vink/pkg/clients"
	"github.com/kubevm.io/vink/pkg/log"
)

func List(ctx context.Context, cephCli clients.CephInterface, nodeName string) ([]*types.StorageDevice, error) {
	osds, err := cephCli.ListOsds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list osds: %w", err)
	}

	osdSet := make(map[string][]*clients.Osd)
	for _, osd := range osds {
		if len(nodeName) > 0 && !strings.Contains(osd.OsdMetadata.Hostname, nodeName) {
			continue
		}
		osdSet[osd.OsdMetadata.Hostname] = append(osdSet[osd.OsdMetadata.Hostname], osd)
	}

	result := make([]*types.StorageDevice, 0, len(osdSet))

	for nodeName, osds := range osdSet {
		sdev := types.StorageDevice{
			Spec: &types.StorageDeviceSpec{NodeName: nodeName},
			Status: &types.StorageDeviceStatus{
				BlockDevices: make([]*types.BlockDeviceStatus, 0, len(osds)),
			},
		}

		for _, osd := range osds {
			bds := types.BlockDeviceStatus{
				Osd:                  int32(osd.OsdMap.Osd),
				Up:                   osd.OsdMap.Up == 1,
				BluestoreBdevDevNode: osd.OsdMetadata.BluestoreBdevDevNode,
				BluestoreBdevType:    osd.OsdMetadata.BluestoreBdevType,
			}
			if total, err := queryPrometheusForStorageTotal(ctx, nodeName, osd); err != nil {
				log.Errorf("failed to get total storage for OSD %d: %v", osd.OsdMap.Osd, err)
			} else {
				bds.Total = float32(total)
			}
			if usage, err := queryPrometheusForStorageUsage(ctx, nodeName, osd); err != nil {
				log.Errorf("failed to get usage storage for OSD %d: %v", osd.OsdMap.Osd, err)
			} else {
				bds.Usage = float32(usage)
			}
			sdev.Status.BlockDevices = append(sdev.Status.BlockDevices, &bds)
		}
		result = append(result, &sdev)
	}

	return result, nil
}

func queryPrometheusForStorageTotal(ctx context.Context, nodeName string, osd *clients.Osd) (float64, error) {
	return pkg.QueryPrometheus(ctx, fmt.Sprintf(`ceph_osd_stat_bytes{ceph_daemon="osd.%s"} * on (pod, namespace) group_left(node) kube_pod_info{node="%s"}`, strconv.Itoa(osd.OsdMap.Osd), nodeName))
}

func queryPrometheusForStorageUsage(ctx context.Context, nodeName string, osd *clients.Osd) (float64, error) {
	return pkg.QueryPrometheus(ctx, fmt.Sprintf(`ceph_osd_stat_bytes_used{ceph_daemon="osd.%v"} * on (pod, namespace) group_left(node) kube_pod_info{node="%s"}`, strconv.Itoa(osd.OsdMap.Osd), nodeName))
}
