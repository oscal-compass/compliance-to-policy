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

package ocm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/IBM/compliance-to-policy/pkg"
	sigyaml "sigs.k8s.io/yaml"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typear "github.com/IBM/compliance-to-policy/pkg/types/oscal/assessmentresults"
	typeoscalcommon "github.com/IBM/compliance-to-policy/pkg/types/oscal/common"
	typeplacementdecision "github.com/IBM/compliance-to-policy/pkg/types/placementdecision"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
	typeutils "github.com/IBM/compliance-to-policy/pkg/types/utils"
)

type ResultToOscal struct {
	c2pParsed          typec2pcr.C2PCRParsed
	policies           []*typepolicy.Policy
	policySets         []*typepolicy.PolicySet
	placementDecisions []*typeplacementdecision.PlacementDecision
}

type Reason struct {
	ClusterName     string                         `json:"clusterName,omitempty" yaml:"clusterName,omitempty"`
	ComplianceState typepolicy.ComplianceState     `json:"complianceState,omitempty" yaml:"complianceState,omitempty"`
	Messages        []typepolicy.ComplianceHistory `json:"messages,omitempty" yaml:"messages,omitempty"`
}

type GenerationType string

const (
	GenerationTypeRaw          GenerationType = "raw"
	GenerationTypePolicyReport GenerationType = "policy-report"
)

func NewResultToOscal(c2pParsed typec2pcr.C2PCRParsed) *ResultToOscal {
	r := ResultToOscal{
		c2pParsed:          c2pParsed,
		policies:           []*typepolicy.Policy{},
		policySets:         []*typepolicy.PolicySet{},
		placementDecisions: []*typeplacementdecision.PlacementDecision{},
	}
	return &r
}

func (r *ResultToOscal) Generate() (*typear.AssessmentResultsRoot, error) {
	traverseFunc := genTraverseFunc(
		func(policy typepolicy.Policy) { r.policies = append(r.policies, &policy) },
		func(policySet typepolicy.PolicySet) { r.policySets = append(r.policySets, &policySet) },
		func(placementDecision typeplacementdecision.PlacementDecision) {
			r.placementDecisions = append(r.placementDecisions, &placementDecision)
		},
	)
	if err := filepath.Walk(r.c2pParsed.PolicyResultsDir, traverseFunc); err != nil {
		logger.Error(err.Error())
	}

	inventories := []typear.InventoryItem{}
	clusternameIndex := map[string]bool{}
	for _, policy := range r.policies {
		if policy.Namespace == r.c2pParsed.Namespace {
			for _, s := range policy.Status.Status {
				_, exist := clusternameIndex[s.ClusterName]
				if !exist {
					clusternameIndex[s.ClusterName] = true
					item := typear.InventoryItem{
						UUID: oscal.GenerateUUID(),
						Props: []typeoscalcommon.Prop{{
							Name:  "cluster-name",
							Value: s.ClusterName,
						}},
					}
					inventories = append(inventories, item)
				}
			}
		}
	}
	observations := []typear.Observation{}
	for _, cdobj := range r.c2pParsed.ComponentObjects {
		policySets := typeutils.FilterByAnnotation(r.policySets, pkg.ANNOTATION_COMPONENT_TITLE, cdobj.ComponentTitle)
		clusterNameSets := sets.NewString()
		var policySet *typepolicy.PolicySet
		if len(policySets) > 0 {
			policySet = policySets[0]
		}
		if policySet != nil {
			placements := []string{}
			for _, placement := range policySet.Status.Placement {
				placements = append(placements, placement.Placement)
			}
			for _, placement := range placements {
				placementDecision := typeutils.FindByNamespaceLabel(r.placementDecisions, policySet.Namespace, "cluster.open-cluster-management.io/placement", placement)
				for _, decision := range placementDecision.Status.Decisions {
					clusterNameSets.Insert(decision.ClusterName)
				}
			}
		}
		for _, controlImpleObj := range cdobj.ControlImpleObjects {
			requiredControls := sets.NewString()
			checkedControls := sets.NewString()
			for _, controlObj := range controlImpleObj.ControlObjects {
				ruleResults := []typereport.RuleResult{}
				controlId := controlObj.ControlId
				for _, ruleId := range controlObj.RuleIds {
					requiredControls.Insert(controlId)
					rule, ok := oscal.FindRulesByRuleId(ruleId, cdobj.RuleObjects)
					if !ok {
						ruleResults = append(ruleResults, typereport.RuleResult{
							RuleId:   ruleId,
							PolicyId: "",
							Status:   typereport.RuleStatusUnImplemented,
						})
					} else {
						policyId := rule.PolicyId
						var policy *typepolicy.Policy
						if policySet != nil {
							policy = typeutils.FindByNamespaceName(r.policies, policySet.Namespace, policyId)
						}
						var ruleStatus typereport.RuleStatus
						subjects := []typear.Subject{}
						if policy != nil {
							reasons := r.GenerateReasonsFromRawPolicies(*policy)
							ruleStatus = mapToRuleStatus(policy.Status.ComplianceState)
							for _, reason := range reasons {
								clusterName := "N/A"
								inventoryUuid := ""
								for _, inventory := range inventories {
									prop, ok := oscal.FindProp("cluster-name", inventory.Props)
									if ok && prop.Value == reason.ClusterName {
										clusterName = prop.Value
										inventoryUuid = inventory.UUID
										break
									}
								}
								if inventoryUuid != "" {
									var message string
									if messageByte, err := sigyaml.Marshal(reason.Messages); err == nil {
										message = string(messageByte)
									} else {
										message = err.Error()
									}
									subject := typear.Subject{
										SubjectUUID: inventoryUuid,
										Type:        "resource",
										Title:       "Cluster Name: " + clusterName,
										Props: []typeoscalcommon.Prop{{
											Name:  "result",
											Value: string(mapToRuleStatus(reason.ComplianceState)),
										}, {
											Name:  "reason",
											Value: message,
										}},
									}
									subjects = append(subjects, subject)
								}
							}
						} else {
							ruleStatus = typereport.RuleStatusError
						}
						observation := typear.Observation{
							UUID:        oscal.GenerateUUID(),
							Description: fmt.Sprintf("Observation of policy %s", policyId),
							Methods:     []string{"TEST-AUTOMATED"},
							Props: []typeoscalcommon.Prop{{
								Name:  "assessment-rule-id",
								Value: ruleId,
							}, {
								Name:  "policy-id",
								Value: policyId,
							}, {
								Name:  "control-id",
								Value: controlId,
							}, {
								Name:  "result",
								Value: string(ruleStatus),
							}},
							Subjects: subjects,
						}
						observations = append(observations, observation)
						checkedControls.Insert(controlId)
					}
				}
			}
		}
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
	result := typear.Result{
		UUID:        oscal.GenerateUUID(),
		Title:       "Assessment Results by OCM",
		Description: "Assessment Results by OCM...",
		Start:       time.Now(),
		LocalDefinitions: typear.LocalDefinitions{
			InventoryItems: inventories,
		},
		Observations: observations,
	}
	ar.Results = append(ar.Results, result)
	arRoot := typear.AssessmentResultsRoot{AssessmentResults: ar}

	return &arRoot, nil
}

func (r *ResultToOscal) GenerateReasonsFromRawPolicies(policy typepolicy.Policy) []Reason {
	reasons := []Reason{}
	for _, status := range policy.Status.Status {
		clusterName := status.ClusterName
		policyPerCluster := typeutils.FindByNamespaceName(r.policies, clusterName, policy.Namespace+"."+policy.Name)
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