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

// Copyright (c) 2020 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package policy

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// RemediationAction describes weather to enforce or inform
type RemediationAction string

const (
	// Enforce is an remediationAction to make changes
	Enforce RemediationAction = "Enforce"

	// Inform is an remediationAction to only inform
	Inform RemediationAction = "Inform"
)

// PolicyTemplate template for custom security policy
type PolicyTemplate struct {
	// A Kubernetes object defining the policy to apply to a managed cluster
	ObjectDefinition runtime.RawExtension `json:"objectDefinition"`

	// Additional PolicyDependencies that only apply to this template
	ExtraDependencies []PolicyDependency `json:"extraDependencies,omitempty"`

	// Ignore this template's Pending status when calculating the overall Policy status
	IgnorePending bool `json:"ignorePending,omitempty"`
}

// ComplianceState shows the state of enforcement
type ComplianceState string

const (
	// Compliant is a ComplianceState
	Compliant ComplianceState = "Compliant"

	// NonCompliant is a ComplianceState
	NonCompliant ComplianceState = "NonCompliant"

	// Pending is a ComplianceState
	Pending ComplianceState = "Pending"
)

// Each PolicyDependency defines an object reference which must be in a certain compliance
// state before the policy should be created.
type PolicyDependency struct {
	metav1.TypeMeta `json:",inline"`

	// The name of the object to be checked
	Name string `json:"name"`

	// The namespace of the object to be checked (optional)
	Namespace string `json:"namespace,omitempty"`

	// The ComplianceState (at path .status.compliant) required before the policy should be created
	Compliance ComplianceState `json:"compliance"`
}

// PolicySpec defines the desired state of Policy
type PolicySpec struct {
	// This provides the ability to enable and disable your policies.
	Disabled bool `json:"disabled"`

	// If set to true (default), all the policy's labels and annotations will be copied to the replicated policy.
	// If set to false, only the policy framework specific policy labels and annotations will be copied to the
	// replicated policy.
	CopyPolicyMetadata *bool `json:"copyPolicyMetadata,omitempty"`

	// This value (Enforce or Inform) will override the remediationAction on each template
	RemediationAction RemediationAction `json:"remediationAction,omitempty"`

	// Used to create one or more policies to apply to a managed cluster
	PolicyTemplates []*PolicyTemplate `json:"policy-templates"`

	// PolicyDependencies that apply to each template in this Policy
	Dependencies []PolicyDependency `json:"dependencies,omitempty"`
}

// PlacementDecision defines the decision made by controller
type PlacementDecision struct {
	ClusterName      string `json:"clusterName,omitempty"`
	ClusterNamespace string `json:"clusterNamespace,omitempty"`
}

// Placement defines the placement results
type Placement struct {
	PlacementBinding string              `json:"placementBinding,omitempty"`
	PlacementRule    string              `json:"placementRule,omitempty"`
	Placement        string              `json:"placement,omitempty"`
	Decisions        []PlacementDecision `json:"decisions,omitempty"`
	PolicySet        string              `json:"policySet,omitempty"`
}

// CompliancePerClusterStatus defines compliance per cluster status
type CompliancePerClusterStatus struct {
	ComplianceState  ComplianceState `json:"compliant,omitempty"`
	ClusterName      string          `json:"clustername,omitempty"`
	ClusterNamespace string          `json:"clusternamespace,omitempty"`
}

// DetailsPerTemplate defines compliance details and history
type DetailsPerTemplate struct {
	TemplateMeta    metav1.ObjectMeta   `json:"templateMeta,omitempty"`
	ComplianceState ComplianceState     `json:"compliant,omitempty"`
	History         []ComplianceHistory `json:"history,omitempty"`
}

// ComplianceHistory defines compliance details history
type ComplianceHistory struct {
	LastTimestamp metav1.Time `json:"lastTimestamp,omitempty" protobuf:"bytes,7,opt,name=lastTimestamp"`
	Message       string      `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	EventName     string      `json:"eventName,omitempty"`
}

// PolicyStatus defines the observed state of Policy
type PolicyStatus struct {
	Placement []*Placement                  `json:"placement,omitempty"` // used by root policy
	Status    []*CompliancePerClusterStatus `json:"status,omitempty"`    // used by root policy

	// +kubebuilder:validation:Enum=Compliant;Pending;NonCompliant
	ComplianceState ComplianceState       `json:"compliant,omitempty"` // used by replicated policy
	Details         []*DetailsPerTemplate `json:"details,omitempty"`   // used by replicated policy
}

// Policy is the Schema for the policies API
type Policy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PolicySpec   `json:"spec"`
	Status PolicyStatus `json:"status,omitempty"`
}

// PolicyList contains a list of Policy
type PolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Policy `json:"items"`
}
