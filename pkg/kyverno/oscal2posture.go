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

package kyverno

import (
	"bytes"
	"embed"
	"html/template"
	"os"

	"github.com/IBM/compliance-to-policy/pkg"
	"go.uber.org/zap"

	tp "github.com/IBM/compliance-to-policy/pkg/kyverno/template"
	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typear "github.com/IBM/compliance-to-policy/pkg/types/oscal/assessmentresults"
	typecd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
)

//go:embed template/*.md
var embeddedResources embed.FS

type Oscal2Posture struct {
	logger       *zap.Logger
	c2pParsed    typec2pcr.C2PCRParsed
	templateFile *string
}

type TemplateValues struct {
	CatalogTitle     string
	Components       typecd.ComponentDefinition
	AssessmentResult typear.AssessmentResults
}

func NewOscal2Posture(c2pParsed typec2pcr.C2PCRParsed, templateFile *string) *Oscal2Posture {
	return &Oscal2Posture{
		logger:       pkg.GetLogger("kyverno/oscal2posture"),
		c2pParsed:    c2pParsed,
		templateFile: templateFile,
	}
}

func (r *Oscal2Posture) findSubjects(ruleId string) []typear.Subject {
	subjects := []typear.Subject{}
	ars := r.c2pParsed.AssessmentResults
	for _, ar := range ars.AssessmentResults.Results {
		for _, ob := range ar.Observations {
			prop, found := oscal.FindProp("assessment-rule-id", ob.Props)
			if found && prop.Value == ruleId {
				subjects = append(subjects, ob.Subjects...)
			}
		}
	}
	return subjects
}

func (r *Oscal2Posture) toTemplateValue() tp.TemplateValue {
	templateValue := tp.TemplateValue{
		CatalogTitle: r.c2pParsed.Catalog.Catalog.Metadata.Title,
		Components:   []tp.Component{},
	}
	for _, componentObject := range r.c2pParsed.ComponentObjects {
		if componentObject.ComponentType == "validation" {
			continue
		}
		component := tp.Component{
			ComponentTitle: componentObject.ComponentTitle,
			ControlResults: []tp.ControlResult{},
		}
		for _, cio := range componentObject.ControlImpleObjects {
			for _, co := range cio.ControlObjects {
				controlResult := tp.ControlResult{
					ControlId:   co.GetControlId(),
					RuleResults: []tp.RuleResult{},
				}
				for _, ruleId := range co.RuleIds {
					subjects := []tp.Subject{}
					rawSubjects := r.findSubjects(ruleId)
					for _, rawSubject := range rawSubjects {
						var result, reason string
						resultProp, resultFound := oscal.FindProp("result", rawSubject.Props)
						reasonProp, reasonFound := oscal.FindProp("reason", rawSubject.Props)

						if resultFound {
							result = resultProp.Value
							if reasonFound {
								reason = reasonProp.Value
							}
						} else {
							result = "Error"
							reason = "No results found."
						}
						subject := tp.Subject{
							Title:  rawSubject.Title,
							UUID:   rawSubject.SubjectUUID,
							Result: result,
							Reason: reason,
						}
						subjects = append(subjects, subject)
					}
					controlResult.RuleResults = append(controlResult.RuleResults, tp.RuleResult{
						RuleId:   ruleId,
						Subjects: subjects,
					})
				}
				component.ControlResults = append(component.ControlResults, controlResult)
			}
		}
		templateValue.Components = append(templateValue.Components, component)
	}
	return templateValue
}

func (r *Oscal2Posture) Generate() ([]byte, error) {
	var templateData []byte
	var err error
	if r.templateFile == nil {
		templateData, err = embeddedResources.ReadFile("template/template.md")
	} else {
		templateData, err = os.ReadFile(*r.templateFile)
	}
	if err != nil {
		return nil, err
	}
	templateString := string(templateData)
	tmpl, err := template.New("report.md").Parse(templateString)
	if err != nil {
		return nil, err
	}
	templateValue := r.toTemplateValue()
	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, templateValue)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
