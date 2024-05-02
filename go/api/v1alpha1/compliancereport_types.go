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
	wgpolicyk8sv1alpha2 "github.com/oscal-compass/compliance-to-policy/go/controllers/wgpolicyk8s.io/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=Compliant;NonCompliant
type CompliancePolicyResult string

type ComplianceReportCluster struct {
	Name    string                           `json:"name,omitempty"`
	Result  wgpolicyk8sv1alpha2.PolicyResult `json:"result,omitempty"`
	Message string                           `json:"message,omitempty"`
}

// ComplianceReportSpec defines the desired state of ComplianceReport
type ComplianceReportResult struct {
	Source   string                           `json:"source,omitempty"`
	Category string                           `json:"category,omitempty"`
	Control  string                           `json:"control,omitempty"`
	Policies []string                         `json:"policies,omitempty"`
	Result   wgpolicyk8sv1alpha2.PolicyResult `json:"result,omitempty"`
	Clusters []ComplianceReportCluster        `json:"clusters,omitempty"`
}

type ComplianceReportSummary struct {
	Standard             string                 `json:"standard,omitempty"`
	Policy               string                 `json:"policy,omitempty"`
	Category             string                 `json:"category,omitempty"`
	Control              string                 `json:"control,omitempty"`
	Result               CompliancePolicyResult `json:"result,omitempty"`
	CompliantClusters    string                 `json:"compliantClusters,omitempty"`
	NonCompliantClusters string                 `json:"nonCompliantClusters,omitempty"`
	TargetClusters       string                 `json:"targetClusters,omitempty"`
	Message              string                 `json:"massage,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:printcolumn:name="Result",type=string,JSONPath=`.summary.result`
//+kubebuilder:printcolumn:name="Compliant_Clusters",type=string,JSONPath=`.summary.compliantClusters`
//+kubebuilder:printcolumn:name="NonCompliant_Clusters",type=string,JSONPath=`.summary.nonCompliantClusters`
//+kubebuilder:printcolumn:name="Target_Clusters",type=string,JSONPath=`.summary.targetClusters`
//+kubebuilder:printcolumn:name="Standard",type=string,JSONPath=`.summary.standard`
//+kubebuilder:printcolumn:name="Control",type=string,JSONPath=`.summary.control`
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// ComplianceReport is the Schema for the controlreferences API
type ComplianceReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Results []ComplianceReportResult `json:"results,omitempty"`
	Summary ComplianceReportSummary  `json:"summary,omitempty"`
}

//+kubebuilder:object:root=true

// ComplianceReportList contains a list of ComplianceReport
type ComplianceReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComplianceReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComplianceReport{}, &ComplianceReportList{})
}
