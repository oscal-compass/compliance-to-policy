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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PolicyReportRef struct {
	// Policy ID
	PolicyId string `json:"policyId,omitempty"`
	// PolicyReport name
	Name string `json:"name,omitempty"`
	// Namespace
	Namespace string `json:"namespace,omitempty"`
}

// ControlReferenceSpec defines the desired state of ResultCollector
type ResultCollectorSpec struct {
	// Resource name of CompliacneDeployment
	ComplianceDeployment string `json:"complianceDeployment,omitempty"`
	// Resource name of ControlReference
	ControlReference string `json:"controlReference,omitempty"`
	// List of PolicyReport to be collected
	PolicyReports []PolicyReportRef `json:"policyReports,omitempty"`
	// List of ClusterPolicyReport to be collected
	ClusterPolicyReports []PolicyReportRef `json:"clusterPolicyReports,omitempty"`
	//+kubebuilder:validation:Pattern=`^(?:(?:(?:[0-9]+(?:.[0-9])?)(?:h|m|s|(?:ms)|(?:us)|(?:ns)))|never)+$`
	// Interval to watch
	Interval string `json:"interval,omitempty"`
}

func parseInterval(interval string) (time.Duration, error) {
	if interval == "" {
		return 0, nil
	}

	parsedInterval, err := time.ParseDuration(interval)
	if err != nil {
		return 0, err
	}

	return parsedInterval, nil
}

func (c ResultCollectorSpec) parseInterval(interval string) (time.Duration, error) {
	return parseInterval(interval)
}

func (c ResultCollectorSpec) GetInterval() (time.Duration, error) {
	return c.parseInterval(c.Interval)
}

// ResultCollectorStatus defines the observed state of ResultCollector
type ResultCollectorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ResultCollector is the Schema for the controlreferences API
type ResultCollector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResultCollectorSpec   `json:"spec,omitempty"`
	Status ResultCollectorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResultCollectorList contains a list of ResultCollector
type ResultCollectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResultCollector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResultCollector{}, &ResultCollectorList{})
}
