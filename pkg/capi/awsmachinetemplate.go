package capi

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AWSMachineTemplateSpec defines the desired state of AWSMachineTemplate
type AWSMachineTemplateSpec struct {
	Template AWSMachineTemplateResource `json:"template"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=awsmachinetemplates,scope=Namespaced,categories=cluster-api,shortName=awsmt
// +kubebuilder:storageversion

// AWSMachineTemplate is the Schema for the awsmachinetemplates API
type AWSMachineTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AWSMachineTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// AWSMachineTemplateList contains a list of AWSMachineTemplate.
type AWSMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AWSMachineTemplate `json:"items"`
}
