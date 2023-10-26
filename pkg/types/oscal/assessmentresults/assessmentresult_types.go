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

package assessmentresults

import (
	"time"

	"github.com/IBM/compliance-to-policy/pkg/types/oscal/common"
)

type Metadata struct {
	Title        string    `json:"title"`
	LastModified time.Time `json:"last-modified"`
	Version      string    `json:"version"`
	OscalVersion string    `json:"oscal-version"`
}

type ImportAp struct {
	Href    string `json:"href"`
	Remarks string `json:"remarks,omitempty"`
}

type InventoryItem struct {
	UUID        string        `json:"uuid"`
	Description string        `json:"description"`
	Props       []common.Prop `json:"props"`
}

type LocalDefinitions struct {
	InventoryItems []InventoryItem `json:"inventory-items"`
}

type SelectControlById struct {
	ControlID    string   `json:"control-id"`
	StatementIds []string `json:"statement-ids,omitempty"`
}

type ControlSelection struct {
	IncludeControls []SelectControlById `json:"include-controls"`
}

type ControlObjectiveSelection struct {
	Props []common.Prop `json:"props,omitempty"`
	Links []common.Link `json:"links,omitempty"`
}

type ReviewedControl struct {
	Description                string                      `json:"description,omitempty"`
	Props                      []common.Prop               `json:"props,omitempty"`
	Links                      []common.Link               `json:"links,omitempty"`
	ControlSelections          []ControlSelection          `json:"control-selections"`
	ControlObjectiveSelections []ControlObjectiveSelection `json:"control-objective-selections,omitempty"`
}

type Subject struct {
	SubjectUUID string        `json:"subject-uuid"`
	Type        string        `json:"type"`
	Title       string        `json:"title"`
	Props       []common.Prop `json:"props"`
}

type Actor struct {
	Type      string        `json:"type,omitempty"`
	ActorUUID string        `json:"actor-uuid"`
	RoleId    string        `json:"role-id,omitempty"`
	Props     []common.Prop `json:"props,omitempty"`
	Links     []common.Link `json:"links,omitempty"`
}

type Origin struct {
	Actors Actor `json:"actors"`
}

type Observation struct {
	UUID             string                    `json:"uuid"`
	Title            string                    `json:"title,omitempty"`
	Description      string                    `json:"description"`
	Props            []common.Prop             `json:"props,omitempty"`
	Links            []common.Link             `json:"links,omitempty"`
	Methods          []string                  `json:"methods"`
	Types            []string                  `json:"types,omitempty"`
	Origins          []Origin                  `json:"origins,omitempty"`
	Subjects         []Subject                 `json:"subjects,omitempty"`
	RelevantEvidence []common.RelevantEvidence `json:"relevant-evidence,omitempty"`
	Collected        time.Time                 `json:"collected"`
	Expires          time.Time                 `json:"expires,omitempty"`
	Remarks          string                    `json:"markup-multiline,omitempty"`
}

type Result struct {
	UUID             string            `json:"uuid"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Start            time.Time         `json:"start"`
	Props            []interface{}     `json:"props,omitempty"`
	LocalDefinitions LocalDefinitions  `json:"local-definitions,omitempty"`
	ReviewedControls []ReviewedControl `json:"reviewed-controls,omitempty"`
	Observations     []Observation     `json:"observations,omitempty"`
}

type AssessmentResults struct {
	UUID     string   `json:"uuid"`
	Metadata Metadata `json:"metadata"`
	ImportAp ImportAp `json:"import-ap"`
	Results  []Result `json:"results"`
}

type AssessmentResultsRoot struct {
	AssessmentResults AssessmentResults `json:"assessment-results"`
}
