/*
Copyright 2023 IBM Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ControlReferenceTarget struct {
	Namespace string `json:"namespace,omitempty"`
}

// ControlReferenceSpec defines the desired state of ControlReference
type ControlReferenceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Internal representation of Compliance
	Compliance Compliance `json:"compliance,omitempty"`
	// Namespace to deploy generated resources
	Target ControlReferenceTarget `json:"target,omitempty"`

	PolicyResources ComplianceDeploymentResourceRef `json:"policyResources,omitempty"`

	Summary map[string]string `json:"summary,omitempty"`
}

// ControlReferenceStatus defines the observed state of ControlReference
type ControlReferenceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ControlReference is the Schema for the controlreferences API
type ControlReference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlReferenceSpec   `json:"spec,omitempty"`
	Status ControlReferenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ControlReferenceList contains a list of ControlReference
type ControlReferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ControlReference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ControlReference{}, &ControlReferenceList{})
}
