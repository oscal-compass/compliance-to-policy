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

// ClusterSelector defines the target clusters to which the policy composition is distributed.
type ClusterSelectors struct {
	// 'matchLabels' is a map of {key,value} pairs matching objects by label.
	MatchLabels *map[string]string `json:"matchLabels,omitempty"`
}

type ComplianceDeploymentResourceRef struct {
	Url string `json:"url,omitempty"`
}

type ComplianceDeploymentCompliance struct {
	// Name of compliance
	Name string `json:"name,omitempty"`
	// Reference to OSCAL Catalog json
	Catalog ComplianceDeploymentResourceRef `json:"catalog,omitempty"`
	// Reference to OSCAL Profile json
	Profile ComplianceDeploymentResourceRef `json:"profile,omitempty"`
	// Reference to OSCAL Component Definition json
	ComponentDefinition ComplianceDeploymentResourceRef `json:"componentDefinition,omitempty"`
}

type ComplianceDeploymentClusterGroup struct {
	// Name
	Name string `json:"name,omitempty"`
	// Cluster selector
	MatchLabels *map[string]string `json:"matchLabels,omitempty"`
}

type ComplianceDeploymentBinding struct {
	// Compliance name
	Compliance string `json:"compliance,omitempty"`
	// The list of cluster group to be bound to the compliance
	ClusterGroups []string `json:"clusterGroups,omitempty"`
}

type ComplianceDeploymentTarget struct {
	// Namespace for generated policies to be placed in Hub
	Namespace string `json:"namespace,omitempty"`
	// KCP workspace
	Workspace string `json:"workspace,omitempty"`
}

// ComplianceDeploymentSpec defines the desired state of ComplianceDeployment
type ComplianceDeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Compliance      ComplianceDeploymentCompliance     `json:"compliance,omitempty"`
	PolicyResources ComplianceDeploymentResourceRef    `json:"policyResources,omitempty"`
	ClusterGroups   []ComplianceDeploymentClusterGroup `json:"clusterGroups,omitempty"`
	Binding         ComplianceDeploymentBinding        `json:"binding,omitempty"`
	Target          ComplianceDeploymentTarget         `json:"target,omitempty"`
}

// ComplianceDeploymentStatus defines the observed state of ComplianceDeployment
type ComplianceDeploymentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ComplianceDeployment is the Schema for the compliancedeployments API
type ComplianceDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComplianceDeploymentSpec   `json:"spec,omitempty"`
	Status ComplianceDeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ComplianceDeploymentList contains a list of ComplianceDeployment
type ComplianceDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplianceDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComplianceDeployment{}, &ComplianceDeploymentList{})
}
