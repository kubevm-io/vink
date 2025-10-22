package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=vink,path="vinks",scope="Cluster",shortName="vink",singular="vink"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type Vink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VinkSpec `json:"spec"`

	Status VinkStatus `json:"status,omitempty"`
}

type VinkSpec struct {
	// +kubebuilder:validation:optional
	Multus *VinkComponent `json:"multus,omitempty"`

	// +kubebuilder:validation:optional
	KubeOVN *VinkComponent `json:"kubeovn,omitempty"`

	// +kubebuilder:validation:optional
	RookCeph *VinkComponent `json:"rookceph,omitempty"`

	// +kubebuilder:validation:optional
	RookCephCluster *VinkComponent `json:"rookcephcluster,omitempty"`

	// +kubebuilder:validation:optional
	Snapshot *VinkComponent `json:"snapshot,omitempty"`

	// +kubebuilder:validation:optional
	Monitoring *VinkComponent `json:"monitoring,omitempty"`

	// +kubebuilder:validation:optional
	KubeVirt *VinkComponent `json:"kubevirt,omitempty"`

	// +kubebuilder:validation:optional
	Cdi *VinkComponent `json:"cdi,omitempty"`
}

type VinkComponent struct {
	// +kubebuilder:validation:optional
	HelmChart *HelmChartRef `json:"helmChart"`

	// +kubebuilder:validation:optional
	HelmValues *HelmValuesRef `json:"helmValues,omitempty"`
}

type HelmChartRef struct {
	// +kubebuilder:validation:optional
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type HelmValuesRef struct {
	// +kubebuilder:validation:optional
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type VinkStatus struct {}

//+kubebuilder:object:root=true

type VinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vink `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Vink{}, &VinkList{})
}
