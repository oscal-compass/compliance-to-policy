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

package placementdecision

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PlacementDecision indicates a decision from a placement
// PlacementDecision should has a label cluster.open-cluster-management.io/placement={placement name}
// to reference a certain placement.
//
// If a placement has spec.numberOfClusters specified, the total number of decisions contained in
// status.decisions of PlacementDecisions should always be NumberOfClusters; otherwise, the total
// number of decisions should be the number of ManagedClusters which match the placement requirements.
//
// Some of the decisions might be empty when there are no enough ManagedClusters meet the placement
// requirements.
type PlacementDecision struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Status represents the current status of the PlacementDecision
	Status PlacementDecisionStatus `json:"status,omitempty"`
}

// The placementDecsion label name holding the placement name
const (
	PlacementLabel string = "cluster.open-cluster-management.io/placement"
)

// PlacementDecisionStatus represents the current status of the PlacementDecision.
type PlacementDecisionStatus struct {
	// Decisions is a slice of decisions according to a placement
	// The number of decisions should not be larger than 100
	Decisions []ClusterDecision `json:"decisions"`
}

// ClusterDecision represents a decision from a placement
// An empty ClusterDecision indicates it is not scheduled yet.
type ClusterDecision struct {
	// ClusterName is the name of the ManagedCluster. If it is not empty, its value should be unique cross all
	// placement decisions for the Placement.
	ClusterName string `json:"clusterName"`

	// Reason represents the reason why the ManagedCluster is selected.
	Reason string `json:"reason"`
}

// ClusterDecisionList is a collection of PlacementDecision.
type PlacementDecisionList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is a list of PlacementDecision.
	Items []PlacementDecision `json:"items"`
}
