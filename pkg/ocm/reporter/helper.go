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
	typeconfigpolicy "github.com/IBM/compliance-to-policy/pkg/types/configurationpolicy"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	typepolr "sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1beta1"
)

// - pass: the policy requirements are met
// - fail: the policy requirements are not met
// - warn: the policy requirements are not met and the policy is not scored
// - error: the policy could not be evaluated
// - skip: the policy was not selected based on user inputs or applicability
type PolicyResult string

const (
	// the policy requirements are met
	PolicyResultPass PolicyResult = "pass"

	// the policy requirements are not met
	PolicyResultFail PolicyResult = "fail"

	// the policy requirements are not met and the policy is not scored
	PolicyResultWarn PolicyResult = "warn"

	// the policy could not be evaluated
	PolicyResultError PolicyResult = "error"

	// the policy was not selected based on user inputs or applicability
	PolicyResultSkip PolicyResult = "skip"
)

func mapToPolicyResult(complianceState typepolicy.ComplianceState) typepolr.PolicyResult {
	var result PolicyResult
	switch complianceState {
	case typepolicy.Compliant:
		result = PolicyResultPass
	case typepolicy.NonCompliant:
		result = PolicyResultFail
	case typepolicy.Pending:
		result = PolicyResultError
	default:
		result = PolicyResultError
	}
	return typepolr.PolicyResult(result)
}

// Severity : low, medium, high, or critical
// PolicyResultSeverity has one of the following values:
//   - critical
//   - high
//   - low
//   - medium
//   - info
func mapToSeverity(severity typeconfigpolicy.Severity) typepolr.PolicyResultSeverity {
	switch severity {
	case "low":
		return "low"
	case "medium":
		return "medium"
	case "high":
		return "high"
	case "critical":
		return "critical"
	default:
		return "info"
	}
}

func mapToTimestamp(details typepolicy.DetailsPerTemplate) metav1.Timestamp {
	if len(details.History) > 0 {
		return *details.History[0].LastTimestamp.ProtoTime()
	}
	return metav1.Timestamp{
		Seconds: 0,
		Nanos:   0,
	}
}

func mapToProps(details typepolicy.DetailsPerTemplate) map[string]string {
	props := map[string]string{}
	if len(details.History) > 0 {
		props["details"] = details.History[0].Message
		props["eventName"] = details.History[0].EventName
		props["lastTimestamp"] = details.History[0].LastTimestamp.DeepCopy().String()
	}
	return props
}

func findConfigPolicyStatus(policy typepolicy.Policy, configPolicy typeconfigpolicy.ConfigurationPolicy) typepolicy.DetailsPerTemplate {
	for _, detail := range policy.Status.Details {
		if detail.TemplateMeta.Name == configPolicy.GetName() {
			return *detail
		}
	}
	return typepolicy.DetailsPerTemplate{}
}

func summary(policyReport typepolr.PolicyReport) typepolr.PolicyReportSummary {
	reportSummary := typepolr.PolicyReportSummary{}
	for _, result := range policyReport.Results {
		pr := PolicyResult(result.Result)
		switch pr {
		case PolicyResultPass:
			reportSummary.Pass++
		case PolicyResultFail:
			reportSummary.Fail++
		case PolicyResultWarn:
			reportSummary.Warn++
		case PolicyResultError:
			reportSummary.Error++
		case PolicyResultSkip:
			reportSummary.Skip++
		default:
			reportSummary.Skip++
		}
	}
	return reportSummary
}

func findPolicyReportByNamespaceName(policyReports []*typepolr.PolicyReport, namespace string, name string) *typepolr.PolicyReport {
	for _, policyReport := range policyReports {
		if policyReport.Namespace == namespace && policyReport.Name == name {
			return policyReport
		}
	}
	return nil
}
