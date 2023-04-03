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

package placements

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Subject defines the resource that can be used as PlacementBinding subject
type Subject struct {
	APIGroup string `json:"apiGroup"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

// PlacementSubject defines the resource that can be used as PlacementBinding placementRef
type PlacementSubject struct {
	APIGroup string `json:"apiGroup"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

// PlacementBindingStatus defines the observed state of PlacementBinding
type PlacementBindingStatus struct{}

// PlacementBinding is the Schema for the placementbindings API
type PlacementBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	PlacementRef      PlacementSubject       `json:"placementRef"`
	Subjects          []Subject              `json:"subjects"`
	Status            PlacementBindingStatus `json:"status,omitempty"`
}

// PlacementBindingList contains a list of PlacementBinding
type PlacementBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PlacementBinding `json:"items"`
}
