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
	"fmt"
	"os"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/types/configurationpolicy"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	"github.com/IBM/compliance-to-policy/pkg/types/placements"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/resource"
)

func TestCompose(t *testing.T) {

	var policyDir = pkg.PathFromPkgDirectory("./composer/testdata/policies")
	var tempDir = pkg.PathFromPkgDirectory("./composer/_test")
	var expectedDir = pkg.PathFromPkgDirectory("./composer/testdata/expected")

	complianceYaml := pkg.PathFromPkgDirectory("./composer/testdata/compliance.yaml")
	if err := os.MkdirAll(tempDir, os.ModePerm); err != nil {
		panic(err)
	}

	c := NewComposer(policyDir, tempDir)

	compliance := Compliance{}
	if err := pkg.LoadYamlFileToObject(complianceYaml, &compliance); err != nil {
		panic(err)
	}
	namespace := "default"
	result, err := c.Compose(namespace, compliance, nil)
	if err != nil {
		panic(err)
	}
	yamlDataList, err := result.ToYaml()
	assert.NoError(t, err, "Should not happen")
	for policy, yamlData := range yamlDataList {
		resultDir := tempDir + "/composed-policies"
		if err := os.MkdirAll(resultDir, os.ModePerm); err != nil {
			panic(err)
		}
		if err := os.WriteFile(resultDir+"/"+policy+".yaml", *yamlData, os.ModePerm); err != nil {
			logger.Sugar().Error(err, fmt.Sprintf("failed to write composed policy for %s", policy))
			panic(err)
		}
	}

	for _, policyComposition := range result.policyCompositions {
		policy := policyComposition.Id
		composedManifests := policyComposition.composedManifests
		expectedManifests, err := pkg.LoadYaml(expectedDir + "/composed-policies/" + policy + ".yaml")
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
				assert.Equal(t, namespace, actualPolicy.GetNamespace())
				assert.Equal(t, len(expectedPolicy.Spec.PolicyTemplates), len(actualPolicy.Spec.PolicyTemplates))
			case "PlacementRule":
				var expectedPR, actualPR placements.PlacementRule
				err := toTypedObject(expectedManifest, &expectedPR, resource, &actualPR)
				assert.NoError(t, err, "Should not happen")
				assert.Equal(t, namespace, actualPR.GetNamespace())
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
				assert.Equal(t, namespace, actualPB.GetNamespace())
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

	configPoliciesByPolicy, err := result.ToConfigPoliciesByPolicy()
	assert.NoError(t, err, "Should not happen")

	configPolicyDir := c.tempDir.GetTempDir() + "/config-policy"
	for policy, configPolicies := range configPoliciesByPolicy {
		for _, configPolicy := range configPolicies {
			err := pkg.MakeDirAndWriteObjToYamlFile(fmt.Sprintf("%s/%s", configPolicyDir, policy), configPolicy.Name+".yaml", configPolicy)
			assert.NoError(t, err, "Should not happen")
			var expectedConfigPolicy configurationpolicy.ConfigurationPolicy
			err = pkg.LoadYamlFileToK8sTypedObject(fmt.Sprintf("%s/composed-config-policies/%s/%s.yaml", expectedDir, policy, configPolicy.Name), &expectedConfigPolicy)
			assert.NoError(t, err, "Should not happen")
			expectedObjectTemplates := expectedConfigPolicy.Spec.ObjectTemplates
			objectTemplates := configPolicy.Spec.ObjectTemplates
			assert.Equal(t, len(expectedObjectTemplates), len(objectTemplates))
			for _, expectedObjectTemplate := range expectedObjectTemplates {
				var expectedUnst unstructured.Unstructured
				err := pkg.LoadByteToK8sTypedObject(expectedObjectTemplate.ObjectDefinition.Raw, &expectedUnst)
				assert.NoError(t, err, "Should not happen")
				expectedKind := expectedUnst.GetKind()
				found := false
				for _, objectTemplate := range objectTemplates {
					var unst unstructured.Unstructured
					err := pkg.LoadByteToK8sTypedObject(objectTemplate.ObjectDefinition.Raw, &unst)
					assert.NoError(t, err, "Should not happen")
					if unst.GetKind() == expectedKind {
						found = true
						break
					}
				}
				assert.Equal(t, true, found)
			}
		}
	}
}

func findManifestByKind(kind string, manifests []*unstructured.Unstructured) (*unstructured.Unstructured, bool) {
	for _, manifest := range manifests {
		if manifest.GetKind() == kind {
			return manifest, true
		}
	}
	return nil, false
}

func toTypedObject(expectedManifest *unstructured.Unstructured, expected interface{}, actualResource *resource.Resource, actual interface{}) error {
	if err := pkg.ToK8sTypedObject(expectedManifest, expected); err != nil {
		return err
	}
	yamlData, err := actualResource.AsYAML()
	if err != nil {
		return err
	}
	if err := utilyaml.Unmarshal(yamlData, actual); err != nil {
		return err
	}
	return nil
}

func findFromMatchExpressions(key string, matchExpressions []v1.LabelSelectorRequirement) (*v1.LabelSelectorRequirement, bool) {
	for _, me := range matchExpressions {
		if me.Key == key {
			return &me, true
		}
	}
	return nil, false
}
