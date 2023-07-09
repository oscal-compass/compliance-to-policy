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
	"time"

	"github.com/IBM/compliance-to-policy/pkg"
	"go.uber.org/zap"
	sigyaml "sigs.k8s.io/yaml"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typeplacementdecision "github.com/IBM/compliance-to-policy/pkg/types/placementdecision"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
	typeutils "github.com/IBM/compliance-to-policy/pkg/types/utils"
	typepolr "sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1beta1"
)

var logger *zap.Logger = pkg.GetLogger("reporter")

type Reporter struct {
	c2pParsed          typec2pcr.C2PCRParsed
	policies           []*typepolicy.Policy
	policySets         []*typepolicy.PolicySet
	placementDecisions []*typeplacementdecision.PlacementDecision
	policyReports      []*typepolr.PolicyReport
	generationType     string
}

type Reason struct {
	ClusterName     string                         `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	ComplianceState typepolicy.ComplianceState     `json:"complianceState,omitempty" yaml:"complianceState,omitempty"`
	Messages        []typepolicy.ComplianceHistory `json:"messages,omitempty" yaml:"messages,omitempty"`
}

func NewReporter(c2pParsed typec2pcr.C2PCRParsed) *Reporter {
	r := Reporter{
		c2pParsed:          c2pParsed,
		policies:           []*typepolicy.Policy{},
		policySets:         []*typepolicy.PolicySet{},
		placementDecisions: []*typeplacementdecision.PlacementDecision{},
		policyReports:      []*typepolr.PolicyReport{},
	}
	return &r
}

func (r *Reporter) SetGenerationType(generationType string) {
	r.generationType = generationType
}

func (r *Reporter) Generate(path string) (typereport.Spec, error) {
	traverseFunc := genTraverseFunc(
		func(policy typepolicy.Policy) { r.policies = append(r.policies, &policy) },
		func(policySet typepolicy.PolicySet) { r.policySets = append(r.policySets, &policySet) },
		func(placementDecision typeplacementdecision.PlacementDecision) {
			r.placementDecisions = append(r.placementDecisions, &placementDecision)
		},
	)
	if err := filepath.Walk(path, traverseFunc); err != nil {
		logger.Error(err.Error())
	}
	for _, policy := range r.policies {
		polr := ConvertToPolicyReport(*policy)
		r.policyReports = append(r.policyReports, &polr)
	}
	reportComponents := []typereport.Component{}
	for _, cdobj := range r.c2pParsed.ComponentObjects {
		policySet := typeutils.FindByNamespaceAnnotation(r.policySets, r.c2pParsed.Namespace, pkg.ANNOTATION_COMPONENT_TITLE, cdobj.ComponentTitle)
		clusterNameSets := sets.NewString()
		if policySet != nil {
			placements := []string{}
			for _, placement := range policySet.Status.Placement {
				placements = append(placements, placement.Placement)
			}
			for _, placement := range placements {
				placementDecision := typeutils.FindByNamespaceLabel(r.placementDecisions, r.c2pParsed.Namespace, "cluster.open-cluster-management.io/placement", placement)
				for _, decision := range placementDecision.Status.Decisions {
					clusterNameSets.Insert(decision.ClusterName)
				}
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
						policy := typeutils.FindByNamespaceName(r.policies, r.c2pParsed.Namespace, policyId)
						var reasons []Reason
						if r.generationType == "policy-report" {
							reasons = r.GenerateReasonsFromPolicyReports(*policy)
						} else {
							reasons = r.GenerateReasonsFromRawPolicies(*policy)
						}
						var reason string
						if statusByte, err := sigyaml.Marshal(reasons); err == nil {
							reason = string(statusByte)
						} else {
							reason = err.Error()
						}
						ruleResult := typereport.RuleResult{
							RuleId:   ruleId,
							PolicyId: policyId,
							Status:   mapToRuleStatus(policy.Status.ComplianceState),
							Reason:   reason,
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

func (r *Reporter) GenerateReasonsFromRawPolicies(policy typepolicy.Policy) []Reason {
	reasons := []Reason{}
	for _, status := range policy.Status.Status {
		clusterName := status.ClusterName
		policyPerCluster := typeutils.FindByNamespaceName(r.policies, clusterName, r.c2pParsed.Namespace+"."+policy.Name)
		if policyPerCluster == nil {
			continue
		}
		messages := []typepolicy.ComplianceHistory{}
		for _, detail := range policyPerCluster.Status.Details {
			if len(detail.History) > 0 {
				messages = append(messages, detail.History[0])
			}
		}
		reasons = append(reasons, Reason{
			ClusterName:     clusterName,
			ComplianceState: status.ComplianceState,
			Messages:        messages,
		})
	}
	return reasons

}

func (r *Reporter) GenerateReasonsFromPolicyReports(policy typepolicy.Policy) []Reason {
	reasons := []Reason{}
	for _, status := range policy.Status.Status {
		clusterName := status.ClusterName
		policyReport := findPolicyReportByNamespaceName(r.policyReports, clusterName, r.c2pParsed.Namespace+"."+policy.Name)
		clusterHistories := []typepolicy.ComplianceHistory{}
		for _, result := range policyReport.Results {
			clusterHistory := typepolicy.ComplianceHistory{
				LastTimestamp: v1.NewTime(time.Unix(result.Timestamp.Seconds, int64(result.Timestamp.Nanos))),
				EventName:     result.Properties["eventName"],
				Message:       result.Properties["details"],
			}
			clusterHistories = append(clusterHistories, clusterHistory)
		}
		reasons = append(reasons, Reason{
			ClusterName:     policyReport.Namespace,
			ComplianceState: aggregatePolicyReportSummaryToComplianceState(policyReport.Summary),
			Messages:        clusterHistories,
		})
	}
	return reasons
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

func aggregatePolicyReportSummaryToComplianceState(summary typepolr.PolicyReportSummary) typepolicy.ComplianceState {
	if summary.Pass == 0 {
		return typepolicy.NonCompliant
	} else {
		return typepolicy.Compliant
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

func genTraverseFunc(onPolicy func(typepolicy.Policy), onPolicySet func(typepolicy.PolicySet), onPlacementDesicion func(typeplacementdecision.PlacementDecision)) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
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
					onPolicy(policy)
				case "PolicySet":
					var policySet typepolicy.PolicySet
					if err := pkg.LoadYamlFileToK8sTypedObject(path, &policySet); err != nil {
						return err
					}
					onPolicySet(policySet)
				case "PlacementDecision":
					var placementDecision typeplacementdecision.PlacementDecision
					if err := pkg.LoadYamlFileToK8sTypedObject(path, &placementDecision); err != nil {
						return err
					}
					onPlacementDesicion(placementDecision)
				}
			}
		}
		return nil
	}
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
