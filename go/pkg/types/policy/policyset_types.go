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

// Copyright (c) 2021 Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project

package policy

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A custom type is required since there is no way to have a kubebuilder marker
// apply to the items of a slice.

type NonEmptyString string

// PolicySetSpec describes a group of policies that are related and
// can be placed on the same managed clusters.
type PolicySetSpec struct {
	// Description of this PolicySet.
	Description string `json:"description,omitempty"`
	// Policies that are grouped together within the PolicySet.
	Policies []NonEmptyString `json:"policies"`
}

// PolicySetStatus defines the observed state of PolicySet
type PolicySetStatus struct {
	Placement     []PolicySetStatusPlacement `json:"placement,omitempty"`
	Compliant     string                     `json:"compliant,omitempty"`
	StatusMessage string                     `json:"statusMessage,omitempty"`
}

// PolicySetStatusPlacement defines a placement object for the status
type PolicySetStatusPlacement struct {
	PlacementBinding string `json:"placementBinding,omitempty"`
	Placement        string `json:"placement,omitempty"`
	PlacementRule    string `json:"placementRule,omitempty"`
}

// PolicySet is the Schema for the policysets API
type PolicySet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              PolicySetSpec   `json:"spec"`
	Status            PolicySetStatus `json:"status,omitempty"`
}

// PolicySetList contains a list of PolicySet
type PolicySetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PolicySet `json:"items"`
}
