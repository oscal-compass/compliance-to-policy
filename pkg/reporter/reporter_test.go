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
	"fmt"
	"os"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/c2pcr"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
	"github.com/stretchr/testify/assert"
)

func TestReporter(t *testing.T) {

	policyDir := pkg.PathFromPkgDirectory("./testdata/policies")
	policyResultsDir := pkg.PathFromPkgDirectory("./testdata/policy-results")
	catalogPath := pkg.PathFromPkgDirectory("./testdata/oscal/reporter-test/catalog.json")
	profilePath := pkg.PathFromPkgDirectory("./testdata/oscal/reporter-test/profile.json")
	cdPath := pkg.PathFromPkgDirectory("./testdata/oscal/reporter-test/component-definition.json")

	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := pkg.NewTempDirectory(tempDirPath)

	gitUtils := pkg.NewGitUtils(tempDir)

	c2pcrSpec := typec2pcr.Spec{
		Compliance: typec2pcr.Compliance{
			Name: "Test Compliance",
			Catalog: typec2pcr.ResourceRef{
				Url: catalogPath,
			},
			Profile: typec2pcr.ResourceRef{
				Url: profilePath,
			},
			ComponentDefinition: typec2pcr.ResourceRef{
				Url: cdPath,
			},
		},
		PolicyResources: typec2pcr.ResourceRef{
			Url: policyDir,
		},
		PolicyRersults: typec2pcr.ResourceRef{
			Url: policyResultsDir,
		},
		ClusterGroups: []typec2pcr.ClusterGroup{{
			Name:        "test-group",
			MatchLabels: &map[string]string{"environment": "test"},
		}},
		Binding: typec2pcr.Binding{
			Compliance:    "Test Compliance",
			ClusterGroups: []string{"test-group"},
		},
		Target: typec2pcr.Target{
			Namespace: "c2p",
		},
	}
	c2pcrParser := c2pcr.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	assert.NoError(t, err, "Should not happen")

	reporter := NewReporter(c2pcrParsed)
	report, err := reporter.Generate()
	assert.NoError(t, err, "Should not happen")

	err = pkg.WriteObjToYamlFile(tempDir.GetTempDir()+"/compliance-report.yaml", report)
	assert.NoError(t, err, "Should not happen")

	var expected typereport.ComplianceReport
	err = pkg.LoadYamlFileToK8sTypedObject(pkg.PathFromPkgDirectory("./testdata/reports/compliance-report.yaml"), &expected)

	// Timestamp is currently set by Now(). Since the timestamp should be always different from expected one, reset creationTimestamp of expected one to actual one.
	expected.CreationTimestamp = report.CreationTimestamp

	assert.NoError(t, err, "Should not happen")
	assert.Equal(t, expected, report)

	reporter.SetGenerationType("policy-report")
	reportFromPolicyReports, err := reporter.Generate()
	assert.NoError(t, err, "Should not happen")

	// Timestamp is currently set by Now(). Since the timestamp should be always different from expected one, reset creationTimestamp of expected one to actual one.
	reportFromPolicyReports.CreationTimestamp = report.CreationTimestamp

	assert.Equal(t, report, reportFromPolicyReports)

	for _, policyReport := range reporter.policyReports {
		fname := fmt.Sprintf("policy-report.%s.%s.yaml", policyReport.Namespace, policyReport.Name)
		err := pkg.WriteObjToYamlFile(tempDir.GetTempDir()+"/"+fname, policyReport)
		assert.NoError(t, err, "Should not happen")
	}

}
