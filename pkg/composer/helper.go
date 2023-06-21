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
	"sort"
	"strings"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/types/configurationpolicy"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	typespolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	pgtype "github.com/IBM/compliance-to-policy/pkg/types/policygenerator"
	cp "github.com/otiai10/copy"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
)

type ComposedResult struct {
	policyCompositions []PolicyComposition
	internalCompliance Compliance
	namespace          string
	clusterSelectors   map[string]string
}

type PolicyComposition struct {
	Id                      string
	ControlId               string
	PolicyCompositionDir    string
	configPolicyDirs        []string
	composedManifests       *resmap.ResMap
	policyGeneratorManifest pgtype.PolicyGenerator
}

// map of policy to yaml bytes
func (cr *ComposedResult) ToYaml() (map[string]*[]byte, error) {
	policies := map[string]*[]byte{}
	for _, pc := range cr.policyCompositions {
		yamlData, err := (*pc.composedManifests).AsYaml()
		if err != nil {
			return nil, err
		}
		policies[pc.Id] = &yamlData
	}
	return policies, nil
}

func (cr *ComposedResult) ToResourcesByPolicy() map[string][]*resource.Resource {
	resources := map[string][]*resource.Resource{}
	for _, pc := range cr.policyCompositions {
		resources[pc.Id] = (*pc.composedManifests).Resources()
	}
	return resources
}

func (cr *ComposedResult) ToConfigPoliciesByPolicy() (map[string][]configurationpolicy.ConfigurationPolicy, error) {
	configPolicyResources := map[string][]configurationpolicy.ConfigurationPolicy{}
	resourcesByPolicy := cr.ToResourcesByPolicy()
	for policyId, resources := range resourcesByPolicy {
		for _, resource := range resources {
			kind := resource.GetKind()
			switch kind {
			case "Policy":
				var policy typespolicy.Policy
				yamlData, err := resource.AsYAML()
				if err != nil {
					logger.Sugar().Error(err, fmt.Sprintf("Failed to convert Policy '%s' to yaml", policyId))
					return nil, err
				}
				if err := utilyaml.Unmarshal(yamlData, &policy); err != nil {
					logger.Sugar().Error(err, fmt.Sprintf("Failed to unmarshal Policy '%s' to yaml", policyId))
					return nil, err
				}
				configPolicies := []configurationpolicy.ConfigurationPolicy{}
				for idx, policyTemplate := range policy.Spec.PolicyTemplates {
					raw := policyTemplate.ObjectDefinition.Raw
					var configPolicy configurationpolicy.ConfigurationPolicy
					if err := utilyaml.Unmarshal(raw, &configPolicy); err != nil {
						logger.Sugar().Error(err, fmt.Sprintf("Failed to unmarshal ConfigPolicy '%d/%d' in Policy '%s' to yaml", idx, len(policy.Spec.PolicyTemplates), policyId))
						return nil, err
					}
					labels := configPolicy.GetLabels()
					if labels != nil {
						labels["policy-id"] = policyId
					} else {
						labels = map[string]string{"policy-id": policyId}
					}
					configPolicy.SetLabels(labels)
					configPolicies = append(configPolicies, configPolicy)
				}
				configPolicyResources[policyId] = configPolicies
			}
		}
	}
	return configPolicyResources, nil
}

func (cr *ComposedResult) ToPrimitiveResourcesByPolicy() (map[string][]unstructured.Unstructured, error) {
	resourcesByPolicy := map[string][]unstructured.Unstructured{}
	for _, pc := range cr.policyCompositions {
		resources := []unstructured.Unstructured{}
		for _, configpolicyDir := range pc.configPolicyDirs {
			entries, err := os.ReadDir(configpolicyDir)
			if err != nil {
				return nil, err
			}
			for _, entry := range entries {
				var unst *unstructured.Unstructured
				if err := pkg.LoadYamlFileToObject(configpolicyDir+"/"+entry.Name(), &unst); err != nil {
					return nil, err
				}
				resources = append(resources, *unst)
			}
		}
		resourcesByPolicy[pc.Id] = resources
	}
	return resourcesByPolicy, nil
}

func (cr *ComposedResult) ToCheckPoliciesByPolicy() (map[string][]unstructured.Unstructured, error) {
	resourcesByPolicy := map[string][]unstructured.Unstructured{}
	for _, pc := range cr.policyCompositions {
		resources := []unstructured.Unstructured{}
		for _, configpolicyDir := range pc.configPolicyDirs {
			path := configpolicyDir + ".yaml"
			if _, err := os.Stat(path); err != nil {
				continue
			}
			var unst *unstructured.Unstructured
			if err := pkg.LoadYamlFileToObject(path, &unst); err != nil {
				return nil, err
			}
			resources = append(resources, *unst)
		}
		resourcesByPolicy[pc.Id] = resources
	}
	return resourcesByPolicy, nil
}

func (cr *ComposedResult) AddGeneratedPolicyManifest() error {
	for _, pc := range cr.policyCompositions {
		yamlData, err := (*pc.composedManifests).AsYaml()
		if err != nil {
			return err
		}
		if err := os.WriteFile(pc.PolicyCompositionDir+"/generated-ocm-policy.yaml", yamlData, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func (cr *ComposedResult) WriteSelectedPoliciesToYamlFile(path string) error {
	type Control struct {
		Id       string   `json:"id,omitempty"`
		Policies []string `json:"policies,omitempty"`
	}
	type Summary struct {
		Controls []Control `json:"controls,omitempty"`
	}
	controls := []Control{}
	for _, category := range cr.internalCompliance.Standard.Categories {
		for _, control := range category.Controls {
			sort.Strings(control.ControlRefs)
			controls = append(controls, Control{
				Id:       control.Name,
				Policies: control.ControlRefs,
			})
		}
	}
	sort.Slice(controls, func(i, j int) bool {
		return strings.Compare(controls[i].Id, controls[j].Id) < 0
	})
	summary := Summary{
		Controls: controls,
	}
	return pkg.WriteObjToYamlFile(path, summary)
}

func (cr *ComposedResult) CopyTo(dir string) error {
	if _, err := pkg.MakeDir(dir); err != nil {
		return err
	}
	for _, pc := range cr.policyCompositions {
		destDir, err := pkg.MakeDir(dir + "/" + pc.Id)
		if err != nil {
			return err
		}
		if err := cp.Copy(pc.PolicyCompositionDir, destDir); err != nil {
			return err
		}
	}
	return nil
}
