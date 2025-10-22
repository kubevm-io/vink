package virtualmachineclaim

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/kubevm.io/vink/pkg/k8s/apis/vink/v1alpha1"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	kubevirtv1 "kubevirt.io/api/core/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cdiStorageBindImmediateAnno = "cdi.kubevirt.io/storage.bind.immediate.requested"

	appCreatedByLabel = "app.kubernetes.io/created-by"
)

func NewVirtualMachineBuilder(tpl *v1alpha1.VirtualMachineTemplate, vmc *v1alpha1.VirtualMachineClaim, client client.Client) *VirtualMachineBuilder {
	return &VirtualMachineBuilder{
		tpl:    tpl,
		vmc:    vmc,
		client: client,
		vm: &kubevirtv1.VirtualMachine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: vmc.Namespace,
				Name:      vmc.Name,
				Labels:    map[string]string{appCreatedByLabel: v1alpha1.GroupVersion.Group},
			},
			Spec: kubevirtv1.VirtualMachineSpec{
				RunStrategy: lo.ToPtr(kubevirtv1.RunStrategyAlways),
				Template:    &kubevirtv1.VirtualMachineInstanceTemplateSpec{},
			},
		},
	}
}

type VirtualMachineBuilder struct {
	tpl *v1alpha1.VirtualMachineTemplate

	vmc *v1alpha1.VirtualMachineClaim

	vm *kubevirtv1.VirtualMachine

	client client.Client
}

func (b *VirtualMachineBuilder) Build(ctx context.Context) (*kubevirtv1.VirtualMachine, error) {
	if err := b.buildOwnerReference(); err != nil {
		return nil, err
	}
	if err := b.buildStorageTemplates(ctx); err != nil {
		return nil, err
	}
	if err := b.buildStorageVolumes(); err != nil {
		return nil, err
	}
	if err := b.buildStorageDisks(); err != nil {
		return nil, err
	}
	if err := b.buildCloudInit(); err != nil {
		return nil, err
	}
	if err := b.buildNetworks(); err != nil {
		return nil, err
	}
	if err := b.buildInterfaces(); err != nil {
		return nil, err
	}
	if err := b.buildResourceQuantity(); err != nil {
		return nil, err
	}
	return b.vm, nil
}

func (b *VirtualMachineBuilder) buildStorageTemplates(ctx context.Context) error {
	var (
		dvTemps     = make([]kubevirtv1.DataVolumeTemplateSpec, 0, len(b.tpl.Spec.Storage.DataDisks)+1)
		dvTempAnno  = map[string]string{cdiStorageBindImmediateAnno: "true"}
		imageSource = b.getImageSource()
	)

	generateDiskID := func() string {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		raw := fmt.Sprintf("%d-%d", time.Now().UnixNano(), r.Int())
		sum := md5.Sum([]byte(raw))
		return hex.EncodeToString(sum[:])[:8]
	}

	if b.tpl.Status.Os != nil {
		dv := cdiv1.DataVolume{}
		if err := b.client.Get(ctx, client.ObjectKey{Namespace: b.tpl.Status.Os.Namespace, Name: b.tpl.Status.Os.Name}, &dv); err != nil {
			return fmt.Errorf("failed to get DataVolume %q in namespace %q for VirtualMachineClaim %q: %w", b.tpl.Status.Os.Name, b.tpl.Status.Os.Namespace, b.vmc.Name, err)
		}
		if len(dv.Status.ClaimName) > 0 {
			imageSource = &cdiv1.DataVolumeSource{
				PVC: &cdiv1.DataVolumeSourcePVC{
					Namespace: b.vmc.Namespace,
					Name:      dv.Status.ClaimName,
				},
			}
		}
	}

	dvTemps = append(dvTemps, kubevirtv1.DataVolumeTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("root-%s", b.vmc.Name),
			Annotations: dvTempAnno,
		},
		Spec: cdiv1.DataVolumeSpec{
			PVC: &corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					b.tpl.Spec.Storage.RootDisk.AccessMode,
				},
				Resources: corev1.VolumeResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse(b.tpl.Spec.Storage.RootDisk.Size),
					},
				},
				StorageClassName: lo.ToPtr(b.tpl.Spec.Storage.RootDisk.StorageClass),
			},
			Source: imageSource,
		},
	})
	for _, disk := range b.tpl.Spec.Storage.DataDisks {
		dvTemps = append(dvTemps, kubevirtv1.DataVolumeTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Name:        fmt.Sprintf("%s-%s", b.vmc.Name, generateDiskID()),
				Annotations: dvTempAnno,
			},
			Spec: cdiv1.DataVolumeSpec{
				PVC: &corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						disk.AccessMode,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse(disk.Size),
						},
					},
					StorageClassName: lo.ToPtr(disk.StorageClass),
				},
				Source: &cdiv1.DataVolumeSource{
					Blank: &cdiv1.DataVolumeBlankImage{},
				},
			},
		})
	}
	b.vm.Spec.DataVolumeTemplates = dvTemps
	return nil
}

func (b *VirtualMachineBuilder) getImageSource() *cdiv1.DataVolumeSource {
	switch {
	case b.tpl.Spec.General.Source.Builtin != nil:
		return nil
	case b.tpl.Spec.General.Source.Http != nil:
		return &cdiv1.DataVolumeSource{HTTP: &cdiv1.DataVolumeSourceHTTP{URL: b.tpl.Spec.General.Source.Http.Url}}
	case b.tpl.Spec.General.Source.S3 != nil:
		return &cdiv1.DataVolumeSource{S3: &cdiv1.DataVolumeSourceS3{URL: b.tpl.Spec.General.Source.S3.Url}}
	case b.tpl.Spec.General.Source.Registry != nil:
		return &cdiv1.DataVolumeSource{Registry: &cdiv1.DataVolumeSourceRegistry{URL: &b.tpl.Spec.General.Source.Registry.Url}}
	case b.tpl.Spec.General.Source.Pvc != nil:
		return &cdiv1.DataVolumeSource{PVC: &cdiv1.DataVolumeSourcePVC{Name: b.tpl.Spec.General.Source.Pvc.Name}}
	case b.tpl.Spec.General.Source.DataVolume != nil:
		return nil
	default:
		return nil
	}
}

func (b *VirtualMachineBuilder) buildStorageVolumes() error {
	volumes := make([]kubevirtv1.Volume, 0, len(b.vm.Spec.DataVolumeTemplates))
	for _, tpl := range b.vm.Spec.DataVolumeTemplates {
		volumes = append(volumes, kubevirtv1.Volume{
			Name: tpl.Name,
			VolumeSource: kubevirtv1.VolumeSource{
				DataVolume: &kubevirtv1.DataVolumeSource{
					Name: tpl.Name,
				},
			},
		})
	}
	b.vm.Spec.Template.Spec.Volumes = volumes
	return nil
}

func (b *VirtualMachineBuilder) buildStorageDisks() error {
	disks := make([]kubevirtv1.Disk, 0, len(b.vm.Spec.Template.Spec.Volumes))
	for _, volume := range b.vm.Spec.Template.Spec.Volumes {
		var bootOrder *uint
		if strings.HasPrefix(volume.Name, "root-") {
			bootOrder = lo.ToPtr[uint](1)
		}
		disks = append(disks, kubevirtv1.Disk{
			BootOrder: bootOrder,
			Name:      volume.Name,
			DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{
					Bus: "virtio",
				},
			},
		})
	}
	b.vm.Spec.Template.Spec.Domain.Devices.Disks = disks
	return nil
}

func (b *VirtualMachineBuilder) buildCloudInit() error {
	cloudInitVolume := kubevirtv1.Volume{
		Name:         "cloud-init",
		VolumeSource: kubevirtv1.VolumeSource{},
	}
	cloudInitDisk := kubevirtv1.Disk{
		Name: "cloud-init",
		DiskDevice: kubevirtv1.DiskDevice{
			Disk: &kubevirtv1.DiskTarget{
				Bus: "virtio",
			},
		},
	}

	var init = b.tpl.Spec.Initialization
	if init == nil || init.CloudInit == nil || len(init.CloudInit.UserDataBase64) == 0 || len(init.CloudInit.UserData) == 0 {
		defaultInit, err := b.getDefaultCloudInit()
		if err != nil {
			return err
		}
		encoded := base64.StdEncoding.EncodeToString([]byte(defaultInit))
		cloudInitVolume.VolumeSource.CloudInitNoCloud = &kubevirtv1.CloudInitNoCloudSource{
			UserDataBase64: encoded,
		}
		b.vm.Spec.Template.Spec.Volumes = append(b.vm.Spec.Template.Spec.Volumes, cloudInitVolume)
		b.vm.Spec.Template.Spec.Domain.Devices.Disks = append(b.vm.Spec.Template.Spec.Domain.Devices.Disks, cloudInitDisk)
		return nil
	}

	if len(init.CloudInit.UserDataBase64) > 0 {
		cloudInitVolume.VolumeSource.CloudInitNoCloud = &kubevirtv1.CloudInitNoCloudSource{
			UserDataBase64: init.CloudInit.UserDataBase64,
		}
	} else {
		cloudInitVolume.VolumeSource.CloudInitNoCloud = &kubevirtv1.CloudInitNoCloudSource{
			UserData: init.CloudInit.UserData,
		}
	}
	b.vm.Spec.Template.Spec.Volumes = append(b.vm.Spec.Template.Spec.Volumes, cloudInitVolume)
	b.vm.Spec.Template.Spec.Domain.Devices.Disks = append(b.vm.Spec.Template.Spec.Domain.Devices.Disks, cloudInitDisk)

	return nil
}

func (b *VirtualMachineBuilder) buildNetworks() error {
	var (
		infacs      = b.tpl.Spec.Network.Interfaces
		networks    = make([]kubevirtv1.Network, 0, len(infacs))
		annotations = make(map[string]string)
	)

	ipMacMap := lo.SliceToMap(b.vmc.Spec.IPMACAssignments, func(item v1alpha1.IPMACAssignment) (string, v1alpha1.IPMACAssignment) {
		return item.Subnet, item
	})

	multusToDashQualifiedName := func(s string) string {
		return strings.ReplaceAll(s, "/", "-")
	}

	multusToQualifiedName := func(s string) string {
		parts := strings.SplitN(s, "/", 2)
		if len(parts) != 2 {
			return s
		}
		return parts[1] + "." + parts[0]
	}

	ovnAnnotationKey := func(nad, suffix string) string {
		return fmt.Sprintf("%s.ovn.kubernetes.io/%s", multusToQualifiedName(nad), suffix)
	}

	for idx, iface := range infacs {
		networks = append(networks, kubevirtv1.Network{
			Name: multusToDashQualifiedName(iface.Nad),
			NetworkSource: kubevirtv1.NetworkSource{
				Multus: &kubevirtv1.MultusNetwork{
					NetworkName: iface.Nad,
					Default:     idx == 0,
				},
			},
		})
		annotations[ovnAnnotationKey(iface.Nad, "logical_switch")] = iface.Subnet

		if ippool, ok := lo.Find(b.vmc.Status.IPPoolAllocations, func(item v1alpha1.VirtualMachineClaimStatusIPPool) bool {
			return item.Subnet == iface.Subnet
		}); ok {
			annotations[ovnAnnotationKey(iface.Nad, "ip_pool")] = ippool.IPPool
			break
		}

		if len(iface.IpPool) > 0 {
			annotations[ovnAnnotationKey(iface.Nad, "ip_pool")] = iface.IpPool
		}
		if ipMac, ok := ipMacMap[iface.Subnet]; ok {
			annotations[ovnAnnotationKey(iface.Nad, "ip_address")] = ipMac.IP
			annotations[ovnAnnotationKey(iface.Nad, "mac_address")] = ipMac.MAC
			delete(ipMacMap, iface.Subnet)
		}
	}
	b.vm.Spec.Template.Spec.Networks = networks
	b.vm.Spec.Template.ObjectMeta.Annotations = annotations
	return nil
}

func (b *VirtualMachineBuilder) buildInterfaces() error {
	interfaces := make([]kubevirtv1.Interface, 0, len(b.vm.Spec.Template.Spec.Networks))
	for _, iface := range b.vm.Spec.Template.Spec.Networks {
		interfaces = append(interfaces, kubevirtv1.Interface{
			Name: iface.Name,
			InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{
				Bridge: &kubevirtv1.InterfaceBridge{},
			},
		})
	}
	b.vm.Spec.Template.Spec.Domain.Devices.Interfaces = interfaces
	return nil
}

func (b *VirtualMachineBuilder) buildResourceQuantity() error {
	var (
		guestMemQty = resource.MustParse(b.tpl.Spec.Compute.Memory.Size)

		requestMemMi = float64(guestMemQty.Value()) / (1024 * 1024) / b.tpl.Spec.Compute.Memory.OvercommitRatio
		limitMemMi   = float64(guestMemQty.Value()) / (1024 * 1024)

		requestMemQty = resource.MustParse(fmt.Sprintf("%.0fMi", requestMemMi))
		requestCPUQty = resource.MustParse(fmt.Sprintf("%dm", int(float64(b.tpl.Spec.Compute.Cpu.Cores)*1000/b.tpl.Spec.Compute.Cpu.OvercommitRatio)))
		limitMemQty   = resource.MustParse(fmt.Sprintf("%.0fMi", limitMemMi))
		limitCPUQty   = resource.MustParse(fmt.Sprintf("%dm", b.tpl.Spec.Compute.Cpu.Cores*1000))
	)

	b.vm.Spec.Template.Spec.Domain.CPU = &kubevirtv1.CPU{
		Cores:   uint32(b.tpl.Spec.Compute.Cpu.Cores),
		Threads: uint32(b.tpl.Spec.Compute.Cpu.Threads),
	}
	b.vm.Spec.Template.Spec.Domain.Resources = kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: requestMemQty,
			corev1.ResourceCPU:    requestCPUQty,
		},
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: limitMemQty,
			corev1.ResourceCPU:    limitCPUQty,
		},
	}
	b.vm.Spec.Template.Spec.Domain.Memory = &kubevirtv1.Memory{
		Guest: &guestMemQty,
	}
	return nil
}

func (b *VirtualMachineBuilder) buildOwnerReference() error {
	b.vm.OwnerReferences = []metav1.OwnerReference{{
		APIVersion:         b.vmc.APIVersion,
		Kind:               b.vmc.Kind,
		Name:               b.vmc.Name,
		UID:                b.vmc.UID,
		Controller:         lo.ToPtr(true),
		BlockOwnerDeletion: lo.ToPtr(true),
	}}
	return nil
}

func (b *VirtualMachineBuilder) getDefaultCloudInit() (string, error) {
	getChpasswdList := func() string {
		var bf strings.Builder
		for _, user := range b.tpl.Spec.General.Users {
			bf.WriteString(fmt.Sprintf("%s:%s\n", user.Name, user.Password))
		}
		return bf.String()
	}

	getCloudInitUsers := func() []map[string]any {
		var result []map[string]any
		for _, user := range b.tpl.Spec.General.Users {
			entry := map[string]any{
				"name":                user.Name,
				"lock_passwd":         false,
				"ssh-authorized-keys": user.SshKey,
			}
			result = append(result, entry)
		}

		return result
	}

	cfg := map[string]any{
		"ssh_pwauth":   b.tpl.Spec.Access.Ssh.Enabled,
		"disable_root": false,
		"users":        getCloudInitUsers(),
		"chpasswd": map[string]any{
			"list":   getChpasswdList(),
			"expire": false,
		},
		"manage_resolv_conf": true,
		"resolv_conf": map[string]any{
			"nameservers": []string{"8.8.8.8", "1.1.1.1"},
		},
		"runcmd": []string{
			"dhclient -r && dhclient",
		},
	}
	if b.tpl.Spec.Access.Ssh.Enabled {
		cfg["runcmd"] = append(cfg["runcmd"].([]string), `sed -i '/^#\?PermitRootLogin/s/.*/PermitRootLogin yes/' /etc/ssh/sshd_config`)
		cfg["runcmd"] = append(cfg["runcmd"].([]string), "systemctl restart sshd")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", err
	}

	return "#cloud-config\n" + string(data), nil
}
