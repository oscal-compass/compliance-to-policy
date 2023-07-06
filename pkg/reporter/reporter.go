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

package reporter

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/IBM/compliance-to-policy/pkg"
	"go.uber.org/zap"

	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
)

var logger *zap.Logger = pkg.GetLogger("reporter")

type Reporter struct {
	c2pParsed  typec2pcr.C2PCRParsed
	policies   []*typepolicy.Policy
	policySets []*typepolicy.PolicySet
}

func NewReporter(c2pParsed typec2pcr.C2PCRParsed) *Reporter {
	r := Reporter{
		c2pParsed:  c2pParsed,
		policies:   []*typepolicy.Policy{},
		policySets: []*typepolicy.PolicySet{},
	}
	return &r
}

func (r *Reporter) generate(path string) (typereport.Spec, error) {
	if err := filepath.Walk(path, r.traverse); err != nil {
		logger.Error(err.Error())
	}
	reportComponents := []typereport.Component{}
	for _, cdobj := range r.c2pParsed.ComponentObjects {
		policySet := typepolicy.FindByNamespaceAnnotation(r.policySets, r.c2pParsed.Namespace, pkg.ANNOTATION_COMPONENT_TITLE, cdobj.ComponentTitle)
		if policySet != nil {
			policies := []typepolicy.Policy{}
			for _, policyName := range policySet.Spec.Policies {
				policy := typepolicy.FindByNamespaceName(r.policies, r.c2pParsed.Namespace, string(policyName))
				policies = append(policies, *policy)
			}
		}
		for _, controlImpleObj := range cdobj.ControlImpleObjects {
			controlResults := []typereport.ControlResult{}
			requiredControls := []string{}
			checkedControls := []string{}
			for _, controlObj := range controlImpleObj.ControlObjects {
				ruleResults := []typereport.RuleResult{}
				controlId := controlObj.ControlId
				for _, ruleId := range controlObj.RuleIds {
					requiredControls = append(requiredControls, controlId)
					rule, ok := oscal.FindRulesByRuleId(ruleId, cdobj.RuleObjects)
					if !ok {
						ruleResults = append(ruleResults, typereport.RuleResult{
							RuleId:   ruleId,
							PolicyId: "",
							Status:   typereport.RuleStatusUnImplemented,
						})
					} else {
						policyId := rule.PolicyId
						policy := typepolicy.FindByNamespaceName(r.policies, r.c2pParsed.Namespace, policyId)
						ruleResult := typereport.RuleResult{
							RuleId:   ruleId,
							PolicyId: policyId,
							Status:   mapToRuleStatus(policy.Status.ComplianceState),
						}
						ruleResults = append(ruleResults, ruleResult)
						checkedControls = append(checkedControls, controlId)
					}
				}
				controlResult := typereport.ControlResult{
					ControlId:        controlId,
					RuleResults:      ruleResults,
					ComplianceStatus: aggregateRuleResults(ruleResults),
				}
				controlResults = append(controlResults, controlResult)
			}
			parameters := map[string]string{}
			for _, setParam := range controlImpleObj.SetParameters {
				parameters[setParam.ParamID] = setParam.Values[0]
			}
			reportComponent := typereport.Component{
				ComponentTitle:   cdobj.ComponentTitle,
				RequiredControls: requiredControls,
				CheckedControls:  checkedControls,
				Parameters:       parameters,
				ControlResults:   controlResults,
				ComplianceStatus: aggregateControlResults(controlResults),
			}
			reportComponents = append(reportComponents, reportComponent)
		}
	}
	return typereport.Spec{
		Catalog:    r.c2pParsed.Catalog.Metadata.Title,
		Profile:    r.c2pParsed.Profile.Metadata.Title,
		Components: reportComponents,
	}, nil
}

func mapToRuleStatus(complianceState typepolicy.ComplianceState) typereport.RuleStatus {
	switch complianceState {
	case typepolicy.Compliant:
		return typereport.RuleStatusPass
	case typepolicy.NonCompliant:
		return typereport.RuleStatusFail
	case typepolicy.Pending:
		return typereport.RuleStatusFail
	default:
		return typereport.RuleStatusError
	}
}

func aggregateRuleResults(ruleResults []typereport.RuleResult) typereport.ComplianceStatus {
	countPass := 0
	countFail := 0
	countError := 0
	countUnimple := 0
	for _, ruleResult := range ruleResults {
		switch ruleResult.Status {
		case typereport.RuleStatusPass:
			countPass++
		case typereport.RuleStatusFail:
			countFail++
		case typereport.RuleStatusError:
			countError++
		case typereport.RuleStatusUnImplemented:
			countUnimple++
		}
	}
	if countPass != 0 && countPass == len(ruleResults) {
		return typereport.ComplianceStatusCompliant
	}
	return typereport.ComplianceStatusNonCompliant
}

func aggregateControlResults(controlResults []typereport.ControlResult) typereport.ComplianceStatus {
	countCompiant := 0
	countNonCompiant := 0
	for _, controlResult := range controlResults {
		switch controlResult.ComplianceStatus {
		case typereport.ComplianceStatusCompliant:
			countCompiant++
		case typereport.ComplianceStatusNonCompliant:
			countNonCompiant++
		}
	}
	if countCompiant != 0 && countCompiant == len(controlResults) {
		return typereport.ComplianceStatusCompliant
	}
	return typereport.ComplianceStatusNonCompliant
}

func (r *Reporter) traverse(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		kind, _, _, ok := parseFileName(info.Name())
		if ok {
			switch kind {
			case "Policy":
				var policy typepolicy.Policy
				if err := pkg.LoadYamlFileToK8sTypedObject(path, &policy); err != nil {
					return err
				}
				r.policies = append(r.policies, &policy)
			case "PolicySet":
				var policySet typepolicy.PolicySet
				if err := pkg.LoadYamlFileToK8sTypedObject(path, &policySet); err != nil {
					return err
				}
				r.policySets = append(r.policySets, &policySet)
			}
		}
	}
	return nil
}

func parseFileName(fname string) (kind string, namespace string, name string, ok bool) {
	splitted := strings.Split(fname, ".")
	if len(splitted) >= 4 {
		kind = splitted[0]
		namespace = splitted[1]
		name = strings.Join(splitted[2:len(splitted)-2], ".")
		ok = true
	} else {
		ok = false
	}
	return
}
