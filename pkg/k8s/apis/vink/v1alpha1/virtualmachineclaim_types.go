package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// VirtualMachineClaimPhasePending indicates that the VirtualMachineClaim has not been processed yet.
	VirtualMachineClaimPhasePending = "Pending"

	// VirtualMachineClaimPhaseProvisioning indicates that the VirtualMachineClaim is being processed.
	VirtualMachineClaimPhaseProvisioning = "Provisioning"

	// VirtualMachineClaimPhaseCompleted indicates that the VirtualMachineClaim has been successfully processed.
	VirtualMachineClaimPhaseCompleted = "Completed"

	// VirtualMachineClaimPhaseFailed indicates that the VirtualMachineClaim has failed to be processed.
	VirtualMachineClaimPhaseFailed = "Failed"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=vink,path="virtualmachineclaims",scope="Namespaced",shortName="vmc",singular="virtualmachineclaim"
// +kubebuilder:printcolumn:name="Template",type=string,JSONPath=".spec.templateRef.name",description="Referenced template name"
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=".status.phase",description="Phase"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

type VirtualMachineClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualMachineClaimSpec `json:"spec"`

	Status VirtualMachineClaimStatus `json:"status,omitempty"`
}

type VirtualMachineClaimSpec struct {
	// +kubebuilder:validation:Required
	TemplateRef *VirtualMachineTemplateRef `json:"templateRef"`

	// +optional
	IPMACAssignments []IPMACAssignment `json:"ipMacAssignments,omitempty"`

	// +optional
	// Replicas *int `json:"replicas,omitempty"`

	// Parameters to override defaults in the template.
	// +optional
	Parameters map[string]string `json:"parameters,omitempty"`
}

type VirtualMachineTemplateRef struct {
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

type IPMACAssignment struct {
	// +kubebuilder:validation:Required
	Subnet string `json:"subnet"`

	// +kubebuilder:validation:Required
	IP string `json:"ip"`

	// +optional
	MAC string `json:"mac,omitempty"`
}

type VirtualMachineClaimStatus struct {
	// +kubebuilder:validation:Enum:=Pending;Provisioning;Completed;Failed
	// +kubebuilder:default:=Pending
	Phase string `json:"phase,omitempty"`

	// Reason provides any error or status message.
	// +optional
	Reason string `json:"reason,omitempty"`

	// VirtualMachineRef is the reference to the generated VM object.
	// +optional
	VirtualMachineRef *corev1.ObjectReference `json:"virtualMachineRef,omitempty"`

	// +optional
	IPPoolAllocations []VirtualMachineClaimStatusIPPool `json:"ipPoolAllocations,omitempty"`
}

type VirtualMachineClaimStatusIPPool struct {
	// +kubebuilder:validation:Required
	Subnet string `json:"subnet"`

	// +kubebuilder:validation:Required
	IPPool string `json:"ipPool"`
}

//+kubebuilder:object:root=true

type VirtualMachineClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineClaim{}, &VirtualMachineClaimList{})
}
