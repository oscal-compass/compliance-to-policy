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

package composer

import (
	"os"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	"github.com/IBM/compliance-to-policy/pkg/types/placements"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	"github.com/stretchr/testify/assert"
)

func TestC2PCRParser(t *testing.T) {
	policyDir := pkg.PathFromPkgDirectory("./composer/testdata/policies")
	catalogPath := pkg.PathFromPkgDirectory("./composer/testdata/oscal/catalog.json")
	profilePath := pkg.PathFromPkgDirectory("./composer/testdata/oscal/profile.json")
	cdPath := pkg.PathFromPkgDirectory("./composer/testdata/oscal/component-definition.json")
	expectedDir := pkg.PathFromPkgDirectory("./composer/testdata/expected/c2pcr-parser-composed-policies")

	tempDirPath := pkg.PathFromPkgDirectory("./composer/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := NewTempDirectory(tempDirPath)

	gitUtils := NewGitUtils(tempDir)

	c2pcrSpec := typec2pcr.Spec{
		Compliance: typec2pcr.Compliance{
			Name: "Test Compliance",
			Catalog: typec2pcr.ResourceRef{
				Url: "local://" + catalogPath,
			},
			Profile: typec2pcr.ResourceRef{
				Url: "local://" + profilePath,
			},
			ComponentDefinition: typec2pcr.ResourceRef{
				Url: "local://" + cdPath,
			},
		},
		PolicyResources: typec2pcr.ResourceRef{
			Url: "local://" + policyDir,
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
			Namespace: "test",
		},
	}
	c2pcrParser := NewC2PCRParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	assert.NoError(t, err, "Should not happen")

	composer := NewComposerByTempDirectory(c2pcrParsed.PolicyResoureDir, tempDir)
	result, err := composer.ComposeByC2PCRParsed(c2pcrParsed)
	assert.NoError(t, err, "Should not happen")

	yamlDataList, err := result.ToYaml()
	assert.NoError(t, err, "Should not happen")
	for policy, yamlData := range yamlDataList {
		resultDir := tempDir.getTempDir() + "/composed-policies"
		err := os.MkdirAll(resultDir, os.ModePerm)
		assert.NoError(t, err, "Should not happen")
		err = os.WriteFile(resultDir+"/"+policy+".yaml", *yamlData, os.ModePerm)
		assert.NoError(t, err, "Should not happen")
	}

	for _, policyComposition := range result.policyCompositions {
		policy := policyComposition.Id
		composedManifests := policyComposition.composedManifests
		expectedManifests, err := pkg.LoadYaml(expectedDir + "/" + policy + ".yaml")
		if err != nil {
			panic(err)
		}
		for _, resource := range (*composedManifests).Resources() {
			kind := resource.GetKind()
			expectedManifest, ok := findManifestByKind(kind, expectedManifests)
			assert.Equal(t, true, ok, "The resource should exist.")
			switch kind {
			case "Policy":
				expectedPolicy := typepolicy.Policy{}
				actualPolicy := typepolicy.Policy{}
				err := toTypedObject(expectedManifest, &expectedPolicy, resource, &actualPolicy)
				assert.NoError(t, err, "Should not happen")
				assert.Equal(t, c2pcrSpec.Target.Namespace, actualPolicy.GetNamespace())
				assert.Equal(t, len(expectedPolicy.Spec.PolicyTemplates), len(actualPolicy.Spec.PolicyTemplates))
			case "PlacementRule":
				var expectedPR, actualPR placements.PlacementRule
				err := toTypedObject(expectedManifest, &expectedPR, resource, &actualPR)
				assert.NoError(t, err, "Should not happen")
				assert.Equal(t, c2pcrSpec.Target.Namespace, actualPR.GetNamespace())
				assert.Equal(t, len(expectedPR.Spec.ClusterSelector.MatchExpressions), len(actualPR.Spec.ClusterSelector.MatchExpressions))
				for _, me := range expectedPR.Spec.ClusterSelector.MatchExpressions {
					actualMe, ok := findFromMatchExpressions(me.Key, actualPR.Spec.ClusterSelector.MatchExpressions)
					assert.Equal(t, true, ok, "The resource should exist.")
					assert.Equal(t, me.Operator, actualMe.Operator)
					assert.Equal(t, me.Values, actualMe.Values)
				}
			case "PlacementBinding":
				var expectedPB, actualPB placements.PlacementBinding
				err := toTypedObject(expectedManifest, &expectedPB, resource, &actualPB)
				assert.NoError(t, err, "Should not happen")
				assert.Equal(t, c2pcrSpec.Target.Namespace, actualPB.GetNamespace())
				assert.Equal(t, expectedPB.PlacementRef.Name, actualPB.PlacementRef.Name)
				assert.Equal(t, len(expectedPB.Subjects), len(actualPB.Subjects))
				for _, subject := range expectedPB.Subjects {
					kind := subject.Kind
					apiGroup := subject.APIGroup
					name := subject.Name
					found := false
					for _, actualSubject := range actualPB.Subjects {
						if actualSubject.Kind == kind && actualSubject.APIGroup == apiGroup && actualSubject.Name == name {
							found = true
							break
						}
					}
					assert.Equal(t, true, found, "Subscjects are not same as expected one.")
				}
			}
		}
	}
}
