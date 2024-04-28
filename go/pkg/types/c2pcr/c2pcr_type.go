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

package c2pcr

type ClusterSelectors struct {
	// 'matchLabels' is a map of {key,value} pairs matching objects by label.
	MatchLabels *map[string]string `json:"matchLabels,omitempty"`
}

type ResourceRef struct {
	Url string `json:"url,omitempty"`
}

type Compliance struct {
	// Name of compliance
	Name string `json:"name,omitempty"`
	// Reference to OSCAL Catalog json
	Catalog ResourceRef `json:"catalog,omitempty"`
	// Reference to OSCAL Profile json
	Profile ResourceRef `json:"profile,omitempty"`
	// Reference to OSCAL Component Definition json
	ComponentDefinition ResourceRef `json:"componentDefinition,omitempty"`
	// Reference to OSCAL Assessment Results json
	AssessmentResults ResourceRef `json:"assessmentResults,omitempty"`
}

type ClusterGroup struct {
	// Name
	Name string `json:"name,omitempty"`
	// Cluster selector
	MatchLabels *map[string]string `json:"matchLabels,omitempty"`
}

type Binding struct {
	// Compliance name
	Compliance string `json:"compliance,omitempty"`
	// The list of cluster group to be bound to the compliance
	ClusterGroups []string `json:"clusterGroups,omitempty"`
}

type Target struct {
	// Namespace for generated policies to be placed in Hub
	Namespace string `json:"namespace,omitempty"`
}

// C2P CR Spec defines the desired state of C2P CR
type Spec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Compliance      Compliance     `json:"compliance,omitempty"`
	PolicyResources ResourceRef    `json:"policyResources,omitempty"`
	PolicyRersults  ResourceRef    `json:"policyResults,omitempty"`
	ClusterGroups   []ClusterGroup `json:"clusterGroups,omitempty"`
	Binding         Binding        `json:"binding,omitempty"`
	Target          Target         `json:"target,omitempty"`
}
