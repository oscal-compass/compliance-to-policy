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

package report

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RuleStatus is a status of rule result
type RuleStatus string

const (
	// If test passed
	RuleStatusPass RuleStatus = "pass"

	// If test failed
	RuleStatusFail RuleStatus = "fail"

	// If test ended with error
	RuleStatusError RuleStatus = "error"

	// If rule doesn't have any implementation
	RuleStatusUnImplemented RuleStatus = "unimplemented"
)

// RuleStatus is a status of rule result
type ComplianceStatus string

const (
	ComplianceStatusCompliant    ComplianceStatus = "Compliant"
	ComplianceStatusNonCompliant ComplianceStatus = "NonCompliant"
)

type RuleResult struct {
	// Rule ID
	RuleId string `json:"ruleId,omitempty" yaml:"ruleId,omitempty"`
	// Policy ID
	PolicyId string `json:"policyId,omitempty" yaml:"policyId,omitempty"`
	// Status
	Status RuleStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// Reason
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

type ControlResult struct {
	// Control ID
	ControlId string `json:"controlId,omitempty" yaml:"controlId,omitempty"`
	// Compliance status
	ComplianceStatus ComplianceStatus `json:"complianceStatus,omitempty" yaml:"complianceStatus,omitempty"`
	// Results per rule
	RuleResults []RuleResult `json:"ruleResults,omitempty" yaml:"ruleResults,omitempty"`
}

type Component struct {
	// Component title in component-definition
	ComponentTitle string `json:"componentTitle,omitempty" yaml:"componentTitle,omitempty"`
	// Compliance status
	ComplianceStatus ComplianceStatus `json:"complianceStatus,omitempty" yaml:"complianceStatus,omitempty"`
	// Required controls
	RequiredControls []string `json:"requiredControls,omitempty" yaml:"requiredControls,omitempty"`
	// Checked controls
	CheckedControls []string `json:"checkedControls,omitempty" yaml:"checkedControls,omitempty"`
	// Used parameters
	Parameters map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// Results per control
	ControlResults []ControlResult `json:"controlResults,omitempty" yaml:"controlResults,omitempty"`
}

type Spec struct {
	Catalog    string      `json:"catalog,omitempty" yaml:"catalog,omitempty"`
	Profile    string      `json:"profile,omitempty" yaml:"profile,omitempty"`
	Components []Component `json:"components,omitempty" yaml:"components,omitempty"`
}

type ComplianceReport struct {
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec              Spec `json:"spec" yaml:"spec"`
}
