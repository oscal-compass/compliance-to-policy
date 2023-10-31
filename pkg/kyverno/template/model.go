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

package template

type Subject struct {
	// Name
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	// UUID
	UUID string `json:"uuid,omitempty" yaml:"uuid,omitempty"`
	// Result
	Result string `json:"result,omitempty" yaml:"result,omitempty"`
	// Reason
	Reason string `json:"reason,omitempty" yaml:"reason,omitempty"`
}

type RuleResult struct {
	// Rule ID
	RuleId string `json:"ruleId,omitempty" yaml:"ruleId,omitempty"`
	// Subjects
	Subjects []Subject `json:"subjects,omitempty" yaml:"subjects,omitempty"`
}

type ControlResult struct {
	// Control ID
	ControlId string `json:"controlId,omitempty" yaml:"controlId,omitempty"`
	// Results per rule
	RuleResults []RuleResult `json:"ruleResults,omitempty" yaml:"ruleResults,omitempty"`
}

type Component struct {
	// Component title in component-definition
	ComponentTitle string `json:"componentTitle,omitempty" yaml:"componentTitle,omitempty"`
	// Results per control
	ControlResults []ControlResult `json:"controlResults,omitempty" yaml:"controlResults,omitempty"`
}

type TemplateValue struct {
	CatalogTitle string
	Components   []Component
}
