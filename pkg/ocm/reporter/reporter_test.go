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
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/c2pcr"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typear "github.com/IBM/compliance-to-policy/pkg/types/oscal/assessmentresults"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestReporter(t *testing.T) {

	policyDir := pkg.PathFromPkgDirectory("./testdata/ocm/policies")
	policyResultsDir := pkg.PathFromPkgDirectory("./testdata/ocm/policy-results")
	catalogPath := pkg.PathFromPkgDirectory("./testdata/ocm/catalog.json")
	profilePath := pkg.PathFromPkgDirectory("./testdata/ocm/profile.json")
	cdPath := pkg.PathFromPkgDirectory("./testdata/ocm/component-definition.json")

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
	arRoot, err := reporter.Generate()
	assert.NoError(t, err, "Should not happen")

	err = pkg.WriteObjToJsonFile(tempDir.GetTempDir()+"/assessment-results.json", arRoot)
	assert.NoError(t, err, "Should not happen")

	var expected typear.AssessmentResultsRoot
	err = pkg.LoadYamlFileToK8sTypedObject(pkg.PathFromPkgDirectory("./testdata/ocm/assessment-results.json"), &expected)

	assert.NoError(t, err, "Should not happen")
	diff := cmp.Diff(expected, *arRoot,
		cmpopts.IgnoreFields(typear.AssessmentResults{}, "UUID"),
		cmpopts.IgnoreFields(typear.Metadata{}, "LastModified"),
		cmpopts.IgnoreFields(typear.Result{}, "UUID", "Start"),
		cmpopts.IgnoreFields(typear.InventoryItem{}, "UUID"),
		cmpopts.IgnoreFields(typear.Subject{}, "SubjectUUID"),
		cmpopts.IgnoreFields(typear.Observation{}, "UUID"),
	)
	assert.Equal(t, diff, "", "assessment-result matched")
}
