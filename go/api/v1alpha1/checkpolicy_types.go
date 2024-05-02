/*
Copyright (c) 2020 Red Hat, Inc.
Copyright Contributors to the Open Cluster Management project

Modifications copyright (C) 2023 IBM Corporation
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// ResultCollectorStatus defines the observed state of CheckPolicy
type CheckPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// ComplianceType describes whether we must or must not have a given resource
// +kubebuilder:validation:Enum=MustHave;Musthave;musthave;MustOnlyHave;Mustonlyhave;mustonlyhave;MustNotHave;Mustnothave;mustnothave
type ComplianceType string

const (
	// MustNotHave is an enforcement state to exclude a resource
	MustNotHave ComplianceType = "Mustnothave"

	// MustHave is an enforcement state to include a resource
	MustHave ComplianceType = "Musthave"

	// MustOnlyHave is an enforcement state to exclusively include a resource
	MustOnlyHave ComplianceType = "Mustonlyhave"
)

type CheckPolicyObjectTemplate struct {
	ComplianceType ComplianceType `json:"complianceType,omitempty"`
	// ObjectDefinition defines required fields for the object
	// +kubebuilder:pruning:PreserveUnknownFields
	ObjectDefinition runtime.RawExtension `json:"objectDefinition,omitempty"`
}

// CheckPolicySpec defines the desired state of CheckPolicy
type CheckPolicySpec struct {
	// 'object-templates' is arrays of objects to be checked. It's refered to OCM Configuration Policy.
	ObjectTemplates []CheckPolicyObjectTemplate `json:"object-templates,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CheckPolicy is the Schema for the controlreferences API
type CheckPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CheckPolicySpec   `json:"spec,omitempty"`
	Status CheckPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CheckPolicyList contains a list of CheckPolicy
type CheckPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CheckPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CheckPolicy{}, &CheckPolicyList{})
}
