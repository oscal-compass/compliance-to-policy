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

package componentdefinition

import "time"

type ComponentDefinitionRoot struct {
	ComponentDefinition `json:"component-definition"`
}

type ComponentDefinition struct {
	UUID       string      `json:"uuid"`
	Metadata   Metadata    `json:"metadata"`
	Components []Component `json:"components"`
}

type Metadata struct {
	Title        string    `json:"title"`
	LastModified time.Time `json:"last-modified"`
	Version      string    `json:"version"`
	OscalVersion string    `json:"oscal-version"`
}

type Component struct {
	UUID                   string                  `json:"uuid"`
	Type                   string                  `json:"type"`
	Title                  string                  `json:"title"`
	Description            string                  `json:"description"`
	Props                  []Prop                  `json:"props"`
	ControlImplementations []ControlImplementation `json:"control-implementations"`
}

type ControlImplementation struct {
	UUID                    string                   `json:"uuid"`
	Source                  string                   `json:"source"`
	Description             string                   `json:"description"`
	Props                   []Prop                   `json:"props"`
	SetParameters           []SetParameter           `json:"set-parameters"`
	ImplementedRequirements []ImplementedRequirement `json:"implemented-requirements"`
}

type ImplementedRequirement struct {
	UUID        string      `json:"uuid"`
	ControlID   string      `json:"control-id"`
	Description string      `json:"description"`
	Props       []Prop      `json:"props"`
	Statements  []Statement `json:"statements,omitempty"`
}

type Statement struct {
	StatementId string `json:"statement-id"`
	UUID        string `json:"uuid"`
	Description string `json:"description,omitempty"`
	Props       []Prop `json:"props,omitempty"`
}

type Prop struct {
	Name    string `json:"name"`
	Ns      string `json:"ns"`
	Value   string `json:"value"`
	Class   string `json:"class,omitempty"`
	Remarks string `json:"remarks"`
}

type SetParameter struct {
	ParamID string   `json:"param-id"`
	Values  []string `json:"values"`
}
