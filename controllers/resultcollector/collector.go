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

package resultcollector

import (
	"context"
	"fmt"
	"strings"

	c2pv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/kcpclient"
	wgpolicyk8sv1alpha2 "github.com/IBM/compliance-to-policy/controllers/wgpolicyk8s.io/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type ClusterPolicyValidationResult struct {
	policyValidationResults []PolicyValidationResult
	workspace               utils.Workspace
	location                string
}

type ClusterPolicyReport struct {
	policyReport wgpolicyk8sv1alpha2.PolicyReport
	workspace    utils.Workspace
	location     string
}

func (r *ResultCollectorReconciler) collect(
	ctx context.Context,
	cr c2pv1alpha1.ResultCollector,
	workspaces []utils.Workspace,
	policyValidationRequests []c2pv1alpha1.PolicyValidationRequest,
) error {
	clusterPolicyValidationResults, locations, err := gatherCheckResults(ctx, *r.Cfg, cr, workspaces, policyValidationRequests)
	if err != nil {
		logger.Error(err, "Failed to gather checkResults")
		return err
	}
	clusterPolicyReports, err := r.generateReportsPerCluster(ctx, cr, clusterPolicyValidationResults)
	if err != nil {
		logger.Error(err, "Failed to generate PolicyReports per cluster")
		return err
	}
	return r.generateSummaryReport(ctx, cr, clusterPolicyReports, locations)
}

func gatherCheckResults(ctx context.Context, cfg rest.Config, cr c2pv1alpha1.ResultCollector, workspaces []utils.Workspace, policyValidationRequests []c2pv1alpha1.PolicyValidationRequest) ([]ClusterPolicyValidationResult, []string, error) {
	clusterPolicyValidationResults := []ClusterPolicyValidationResult{}
	locations := []string{}
	sts, err := utils.GetSts(ctx, cfg, cr.Spec.ComplianceDeployment.Target.Workspace, cr.Name)
	if err != nil {
		logger.Error(err, "Failed to get sts")
		return nil, locations, err
	}
	for _, workspace := range workspaces {
		logger.V(3).Info(fmt.Sprintf("\nworkspace: %s", workspace))
		wsName := workspace.Name
		location := ""
		for _, destination := range sts.Destinations {
			if destination.SyncTargetName == workspace.SyncTargetName {
				location = destination.LocationName
			}
		}
		if location == "" {
			continue
		}
		locations = append(locations, location)
		kcpClient, err := kcpclient.NewKcpClient(cfg, wsName)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to create KcpClient for workspace '%s'", workspace))
			return nil, locations, err
		}
		policyValidationResults, err := validate(ctx, policyValidationRequests, kcpClient)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to validate for workspace '%s'", workspace))
			return nil, locations, err
		}
		clusterPolicyValidationResult := ClusterPolicyValidationResult{
			policyValidationResults: policyValidationResults,
			workspace:               workspace,
			location:                location,
		}
		clusterPolicyValidationResults = append(clusterPolicyValidationResults, clusterPolicyValidationResult)
	}
	return clusterPolicyValidationResults, locations, nil
}

func (r *ResultCollectorReconciler) generateReportsPerCluster(ctx context.Context, cr c2pv1alpha1.ResultCollector, clusterPolicyValidationResults []ClusterPolicyValidationResult) ([]ClusterPolicyReport, error) {
	clusterPolicyReports := []ClusterPolicyReport{}
	for idx, clusterPolicyValidationResult := range clusterPolicyValidationResults {
		policyResults := []*wgpolicyk8sv1alpha2.PolicyReportResult{}
		policySummary := wgpolicyk8sv1alpha2.PolicyReportSummary{}
		messages := []string{}
		policyReport := wgpolicyk8sv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("%s-per-cluster-%d", cr.Name, idx),
				Namespace: cr.Namespace,
				Annotations: map[string]string{
					"workspaceName": clusterPolicyValidationResult.workspace.Name,
					"locationName":  clusterPolicyValidationResult.location,
				},
			},
			Summary: wgpolicyk8sv1alpha2.PolicyReportSummary{},
		}
		for _, policyValidationResult := range clusterPolicyValidationResult.policyValidationResults {
			for _, checkResult := range policyValidationResult.checkPolicyResults {
				policyResult := wgpolicyk8sv1alpha2.PolicyReportResult{}
				policyResult.Policy = policyValidationResult.policyId
				policyResult.Rule = checkResult.checkPolicy.Name
				subjects := []*corev1.ObjectReference{}
				pass := true
				errored := false
				for _, testResult := range checkResult.testResults {
					pass = testResult.pass && pass
					if testResult.error != nil {
						errored = true
					}
					subject := &corev1.ObjectReference{
						Kind:       testResult.objectDefinition.GetKind(),
						APIVersion: testResult.objectDefinition.GetAPIVersion(),
						Namespace:  testResult.objectDefinition.GetNamespace(),
						Name:       testResult.objectDefinition.GetName(),
					}
					subjects = append(subjects, subject)
					messages = append(messages, testResult.message)
				}
				if errored {
					policyResult.Result = "error"
					policySummary.Error++
				} else if pass {
					policyResult.Result = "pass"
					policySummary.Pass++
				} else {
					policyResult.Result = "fail"
					policySummary.Fail++
				}
				policyResult.Subjects = subjects
				policyResult.Description = strings.Join(messages, "\n")
				policyResults = append(policyResults, &policyResult)
			}
		}
		policyReport.Results = policyResults
		policyReport.Summary = policySummary
		clusterPolicyReports = append(clusterPolicyReports, ClusterPolicyReport{policyReport: policyReport, location: clusterPolicyValidationResult.location, workspace: clusterPolicyValidationResult.workspace})
		if err := utils.CreateOrUpdate(ctx, r.Client, &policyReport, &wgpolicyk8sv1alpha2.PolicyReport{}); err != nil {
			logger.Error(err, fmt.Sprintf("Failed to create or update PolicyReport '%s' for workspace '%s'", policyReport.Name, clusterPolicyValidationResult.workspace.Name))
			return nil, err
		}
	}
	return clusterPolicyReports, nil
}

func (r *ResultCollectorReconciler) generateSummaryReport(ctx context.Context, cr c2pv1alpha1.ResultCollector, clusterPolicyReports []ClusterPolicyReport, locations []string) error {
	type ClusterPolicyReportResult struct {
		policyReportName string
		reportResult     wgpolicyk8sv1alpha2.PolicyReportResult
		workspace        utils.Workspace
		location         string
	}
	clusterPolicyReportResultsPerPolicy := map[string][]ClusterPolicyReportResult{}
	for _, clusterPolicyReport := range clusterPolicyReports {
		for _, reportResult := range clusterPolicyReport.policyReport.Results {
			clusterPolicyReportResult := ClusterPolicyReportResult{
				policyReportName: clusterPolicyReport.policyReport.Name,
				reportResult:     *reportResult,
				workspace:        clusterPolicyReport.workspace,
				location:         clusterPolicyReport.location,
			}
			_, ok := clusterPolicyReportResultsPerPolicy[reportResult.Policy]
			if !ok {
				clusterPolicyReportResultsPerPolicy[reportResult.Policy] = []ClusterPolicyReportResult{clusterPolicyReportResult}
			} else {
				clusterPolicyReportResultsPerPolicy[reportResult.Policy] = append(clusterPolicyReportResultsPerPolicy[reportResult.Policy], clusterPolicyReportResult)
			}
		}
	}
	type clusterExtend struct {
		complianceReportCluster c2pv1alpha1.ComplianceReportCluster
		policy                  string
	}
	contains := func(results []wgpolicyk8sv1alpha2.PolicyResult, value string) bool {
		for _, result := range results {
			if string(result) == value {
				return true
			}
		}
		return false
	}
	mergeResults := func(results []wgpolicyk8sv1alpha2.PolicyResult) wgpolicyk8sv1alpha2.PolicyResult {
		if contains(results, "fail") {
			return "fail"
		} else if contains(results, "error") {
			return "error"
		} else {
			return "pass"
		}
	}
	totalControls := []string{}
	complianceReportResults := []c2pv1alpha1.ComplianceReportResult{}
	for _, category := range cr.Spec.Compliance.Standard.Categories {
		for _, control := range category.Controls {
			totalControls = append(totalControls, control.Name)
			clusters := []clusterExtend{}
			for _, policy := range control.ControlRefs {
				clusterPolicyReportResults, ok := clusterPolicyReportResultsPerPolicy[policy]
				if ok {
					for _, clusterPolicyReportResult := range clusterPolicyReportResults {
						result := clusterPolicyReportResult.reportResult.Result
						location := clusterPolicyReportResult.location
						message := fmt.Sprintf("The detail information is available in %s.\n%s", clusterPolicyReportResult.policyReportName, clusterPolicyReportResult.reportResult.Description)
						cluster := c2pv1alpha1.ComplianceReportCluster{Name: location, Result: result, Message: message}
						clusters = append(clusters, clusterExtend{complianceReportCluster: cluster, policy: policy})
					}
				}
			}
			perCluster := map[string][]clusterExtend{}
			for _, cluster := range clusters {
				_, ok := perCluster[cluster.complianceReportCluster.Name]
				if !ok {
					perCluster[cluster.complianceReportCluster.Name] = []clusterExtend{cluster}
				} else {
					perCluster[cluster.complianceReportCluster.Name] = append(perCluster[cluster.complianceReportCluster.Name], cluster)
				}
			}
			clustersPerControl := []c2pv1alpha1.ComplianceReportCluster{}
			for name, clusters := range perCluster {
				messages := []string{}
				results := []wgpolicyk8sv1alpha2.PolicyResult{}
				for _, cluster := range clusters {
					messages = append(messages, cluster.complianceReportCluster.Message)
					results = append(results, cluster.complianceReportCluster.Result)
				}
				cluster := c2pv1alpha1.ComplianceReportCluster{Name: name, Result: mergeResults(results), Message: strings.Join(messages, "\n")}
				clustersPerControl = append(clustersPerControl, cluster)
			}

			policyReportResult := c2pv1alpha1.ComplianceReportResult{
				Control:  control.Name,
				Policies: control.ControlRefs,
				Source:   "C2P",
				Category: category.Name,
				Clusters: clustersPerControl,
			}
			complianceReportResults = append(complianceReportResults, policyReportResult)
		}
	}
	clusterList := []string{}
	for _, clusterPolicyReport := range clusterPolicyReports {
		clusterList = append(clusterList, clusterPolicyReport.location)
	}
	findByCluster := func(clusters []c2pv1alpha1.ComplianceReportCluster, cluster string) *c2pv1alpha1.ComplianceReportCluster {
		for _, _cluster := range clusters {
			if cluster == _cluster.Name {
				return &_cluster
			}
		}
		return nil
	}
	compliantClusters := []string{}
	nonCompliantClusters := []string{}
	for _, cluster := range clusterList {
		results := []wgpolicyk8sv1alpha2.PolicyResult{}
		complianceReportResultsPerClusters := []c2pv1alpha1.ComplianceReportCluster{}
		for _, complianceReportResult := range complianceReportResults {
			complianceReportCluster := findByCluster(complianceReportResult.Clusters, cluster)
			complianceReportResultsPerClusters = append(complianceReportResultsPerClusters, *complianceReportCluster)
			results = append(results, complianceReportCluster.Result)
		}
		result := mergeResults(results)
		if result == "pass" {
			compliantClusters = append(compliantClusters, cluster)
		} else {
			nonCompliantClusters = append(nonCompliantClusters, cluster)
		}
	}
	var result c2pv1alpha1.CompliancePolicyResult
	if len(nonCompliantClusters) == 0 {
		result = "Compliant"
	} else {
		result = "NonCompliant"
	}
	summary := c2pv1alpha1.ComplianceReportSummary{
		Standard:             cr.Spec.Compliance.Standard.Name,
		Control:              strings.Join(totalControls, ","),
		CompliantClusters:    strings.Join(compliantClusters, ","),
		NonCompliantClusters: strings.Join(nonCompliantClusters, ","),
		TargetClusters:       strings.Join(locations, ","),
		Result:               result,
	}
	complianceReport := c2pv1alpha1.ComplianceReport{
		ObjectMeta: v1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Results: complianceReportResults,
		Summary: summary,
	}
	if err := utils.CreateOrUpdate(ctx, r.Client, &complianceReport, &c2pv1alpha1.ComplianceReport{}); err != nil {
		logger.Error(err, fmt.Sprintf("Failed to create or update ComplianceReport '%s'", complianceReport.Name))
		return err
	}
	return nil
}
