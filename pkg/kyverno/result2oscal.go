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
	"fmt"
	"strings"
	"time"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typear "github.com/IBM/compliance-to-policy/pkg/types/oscal/assessmentresults"
	typeoscalcommon "github.com/IBM/compliance-to-policy/pkg/types/oscal/common"
	kyvernov1 "github.com/kyverno/kyverno/api/kyverno/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
	typepolr "sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1beta1"
)

type ResultToOscal struct {
	logger                  *zap.Logger
	c2pParsed               typec2pcr.C2PCRParsed
	policyReportList        *typepolr.PolicyReportList
	clusterPolicyReportList *typepolr.ClusterPolicyReportList
	policyList              *kyvernov1.PolicyList
	clusterPolicyList       *kyvernov1.ClusterPolicyList
}

type PolicyReportContainer struct {
	PolicyReports        []*typepolr.PolicyReport
	ClusterPolicyReports []*typepolr.ClusterPolicyReport
}

type PolicyResourceIndexContainer struct {
	PolicyResourceIndex PolicyResourceIndex
	ControlIds          []string
}

func NewResultToOscal(c2pParsed typec2pcr.C2PCRParsed) *ResultToOscal {
	r := ResultToOscal{
		logger:                  pkg.GetLogger("kyverno/reporter"),
		c2pParsed:               c2pParsed,
		policyReportList:        &typepolr.PolicyReportList{},
		clusterPolicyReportList: &typepolr.ClusterPolicyReportList{},
		policyList:              &kyvernov1.PolicyList{},
		clusterPolicyList:       &kyvernov1.ClusterPolicyList{},
	}
	return &r
}

func (r *ResultToOscal) aggregateComponentObjects() (policyResourceIndice []PolicyResourceIndex, controlIds []string) {
	controlIdSets := sets.NewString()
	for _, componentObject := range r.c2pParsed.ComponentObjects {
		if componentObject.ComponentType == "validation" {
			continue
		}
		for _, ruleObject := range componentObject.RuleObjects {
			sourceDir := fmt.Sprintf("%s/%s", r.c2pParsed.PolicyResoureDir, ruleObject.RuleId)
			fl := NewFileLoader()
			if err := fl.LoadFromDirectory(sourceDir); err != nil {
				r.logger.Error(fmt.Sprintf("Failed to load %s", sourceDir))
				continue
			}
			policyResourceIndice = append(policyResourceIndice, fl.GetPolicyResourceIndice()...)
		}
		for _, cio := range componentObject.ControlImpleObjects {
			for _, cos := range cio.ControlObjects {
				controlIdSets = controlIdSets.Insert(cos.GetControlId())
			}
		}
	}
	controlIds = controlIdSets.List()
	return
}

func (r *ResultToOscal) findControls(ruleId string) []oscal.ControlObject {
	controls := []oscal.ControlObject{}
	for _, componentObject := range r.c2pParsed.ComponentObjects {
		if componentObject.ComponentType == "validation" {
			continue
		}
		for _, cio := range componentObject.ControlImpleObjects {
			for _, cos := range cio.ControlObjects {
				for _, _ruleId := range cos.RuleIds {
					if ruleId == _ruleId {
						controls = append(controls, cos)
					}
				}
			}
		}
	}
	return controls
}

func (r *ResultToOscal) retrievePolicyReportResults(name string) []*typepolr.PolicyReportResult {
	prrs := []*typepolr.PolicyReportResult{}
	for _, polr := range r.policyReportList.Items {
		for _, result := range polr.Results {
			policy := result.Policy
			if policy == name {
				prrs = append(prrs, result)
			}
		}
	}
	return prrs
}

func (r *ResultToOscal) loadData(path string, out interface{}) error {
	if err := pkg.LoadYamlFileToK8sTypedObject(r.c2pParsed.PolicyResultsDir+path, &out); err != nil {
		return err
	}
	return nil
}

func makeProp(name string, value string) typeoscalcommon.Prop {
	return typeoscalcommon.Prop{
		Name:  name,
		Value: value,
	}
}

func (r *ResultToOscal) GenerateAssessmentResults() (*typear.AssessmentResults, error) {
	var polList kyvernov1.PolicyList
	if err := r.loadData("/policies.kyverno.io.yaml", &polList); err != nil {
		return nil, err
	}

	var cpolList kyvernov1.ClusterPolicyList
	if err := r.loadData("/clusterpolicies.kyverno.io.yaml", &cpolList); err != nil {
		return nil, err
	}

	var polrList typepolr.PolicyReportList
	if err := r.loadData("/policyreports.wgpolicyk8s.io.yaml", &polrList); err != nil {
		return nil, err
	}
	r.policyReportList = &polrList

	var cpolrList typepolr.ClusterPolicyReportList
	if err := r.loadData("/clusterpolicyreports.wgpolicyk8s.io.yaml", &cpolrList); err != nil {
		return nil, err
	}

	observations := []typear.Observation{}
	pris, controlIds := r.aggregateComponentObjects()

	for _, pri := range pris {
		name := pri.Name
		prrs := r.retrievePolicyReportResults(name)
		props := []typeoscalcommon.Prop{}
		props = append(props, makeProp("assessment-rule-id", name))
		controls := r.findControls(name)
		controlIds := sets.NewString()
		for _, control := range controls {
			controlIds = controlIds.Insert(control.GetControlId())
		}
		props = append(props, makeProp("controls", strings.Join(controlIds.List(), ",")))
		observation := typear.Observation{
			UUID:        oscal.GenerateUUID(),
			Description: fmt.Sprintf("Observation of rule %s", pri.Name),
			Methods:     []string{"TEST-AUTOMATED"},
			Props:       props,
			Subjects:    []typear.Subject{},
		}
		for _, prr := range prrs {
			props := []typeoscalcommon.Prop{}
			props = append(props, makeProp("result", string(prr.Result)))
			props = append(props, makeProp("reason", prr.Description))
			for _, resource := range prr.Subjects {
				gvknsn := fmt.Sprintf("ApiVersion: %s, Kind: %s, Namespace: %s, Name: %s", resource.APIVersion, resource.Kind, resource.Namespace, resource.Name)
				subject := typear.Subject{
					SubjectUUID: string(resource.UID),
					Title:       gvknsn,
					Type:        "resource",
					Props:       props,
				}
				observation.Subjects = append(observation.Subjects, subject)
			}
		}
		observations = append(observations, observation)
	}

	metadata := typear.Metadata{
		Title:        "OSCAL Assessment Results",
		LastModified: time.Now(),
		Version:      "0.0.1",
		OscalVersion: "1.0.4",
	}
	importAp := typear.ImportAp{
		Href: "http://...",
	}
	ar := typear.AssessmentResults{
		UUID:     oscal.GenerateUUID(),
		Metadata: metadata,
		ImportAp: importAp,
		Results:  []typear.Result{},
	}

	scs := []typear.SelectControlById{}
	for _, controlId := range controlIds {
		scs = append(scs, typear.SelectControlById{
			ControlID: controlId,
		})
	}
	controlSelection := typear.ControlSelection{
		IncludeControls: scs,
	}
	result := typear.Result{
		UUID:        oscal.GenerateUUID(),
		Title:       "Assessment Results by Kyverno Policy",
		Description: "Assessment Results by Kyverno Policy...",
		Start:       time.Now(),
		ReviewedControls: []typear.ReviewedControl{{
			ControlSelections: []typear.ControlSelection{controlSelection},
		}},
		Observations: observations,
	}

	ar.Results = append(ar.Results, result)

	return &ar, nil
}
