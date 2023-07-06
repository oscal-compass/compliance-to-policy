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
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/oscal"
	policygenerator "github.com/IBM/compliance-to-policy/pkg/policygenerator"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	pgtype "github.com/IBM/compliance-to-policy/pkg/types/policygenerator"
	cp "github.com/otiai10/copy"
	"go.uber.org/zap"
	"sigs.k8s.io/kustomize/api/resmap"
	typekustomize "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

type ComposerV2 struct {
	policiesDir string
	tempDir     pkg.TempDirectory
}

func NewComposerV2(policiesDir string, tempDir string) *ComposerV2 {
	return NewComposerV2ByTempDirectory(policiesDir, pkg.NewTempDirectory(tempDir))
}

func NewComposerV2ByTempDirectory(policiesDir string, tempDir pkg.TempDirectory) *ComposerV2 {
	return &ComposerV2{
		policiesDir: policiesDir,
		tempDir:     tempDir,
	}
}

func (c *ComposerV2) GetPoliciesDir() string {
	return c.policiesDir
}

func (c *ComposerV2) ComposeByC2PParsed(c2pParsed typec2pcr.C2PCRParsed) error {
	return c.Compose(c2pParsed.Namespace, c2pParsed.ComponentObjects, c2pParsed.ClusterSelectors)
}

func (c *ComposerV2) Compose(namespace string, componentObjects []oscal.ComponentObject, clusterSelectors map[string]string) error {

	if clusterSelectors == nil {
		clusterSelectors = map[string]string{"env": "dev"}
	}

	logger.Info("Start composing policySets")
	parameters := map[string]string{}
	policyConfigMap := map[string]pgtype.PolicyConfig{}
	policySets := []pgtype.PolicySetConfig{}
	policySetPatches := []typekustomize.Patch{}
	for _, componentObject := range componentObjects {
		logger := logger.With(zap.Namespace(fmt.Sprintf("component %s", componentObject.ComponentTitle)))
		logger.Info("Start generating policy")
		for _, ruleObject := range componentObject.RuleObjects {
			sourceDir := fmt.Sprintf("%s/%s", c.policiesDir, ruleObject.PolicyId)
			destDir := fmt.Sprintf("%s/%s", c.tempDir.GetTempDir(), ruleObject.PolicyId)
			err := cp.Copy(sourceDir, destDir)
			if err != nil {
				return err
			}
		}

		for idx, controlImpleObject := range componentObject.ControlImpleObjects {
			policyListPerControlImple := []string{}
			for _, param := range controlImpleObject.SetParameters {
				parameters[param.ParamID] = param.Values[0]
			}
			for _, controlObject := range controlImpleObject.ControlObjects {
				for _, ruleId := range controlObject.RuleIds {
					ruleObject, ok := oscal.FindRulesByRuleId(ruleId, componentObject.RuleObjects)
					if ok {
						policyId := ruleObject.PolicyId
						destDir := fmt.Sprintf("%s/%s", c.tempDir.GetTempDir(), policyId)
						policyGeneratorManifestPath := destDir + "/policy-generator.yaml"
						var policyGeneratorManifest pgtype.PolicyGenerator
						if err := pkg.LoadYamlFileToObject(policyGeneratorManifestPath, &policyGeneratorManifest); err != nil {
							return err
						}
						policyGeneratorManifest.PolicyDefaults.Namespace = namespace
						policyGeneratorManifest.PolicyDefaults.PolicyOptions.Standards = []string{""}
						policyGeneratorManifest.PolicyDefaults.PolicyOptions.Categories = []string{""}
						policyGeneratorManifest.PolicyDefaults.PolicyOptions.Controls = []string{controlObject.ControlId}
						policyGeneratorManifest.PolicyDefaults.PolicyOptions.Placement.ClusterSelectors = clusterSelectors
						if err := pkg.WriteObjToYamlFileByGoYaml(policyGeneratorManifestPath, policyGeneratorManifest); err != nil {
							return err
						}
						// For policySet
						policyListPerControlImple = appendUnique(policyListPerControlImple, policyId)
						policyConfig, ok := policyConfigMap[policyId]
						if ok {
							policyConfig.Standards = appendUnique(policyConfig.Standards, policyGeneratorManifest.PolicyDefaults.Standards...)
							policyConfig.Categories = appendUnique(policyConfig.Categories, policyGeneratorManifest.PolicyDefaults.Categories...)
							policyConfig.Controls = appendUnique(policyConfig.Controls, policyGeneratorManifest.PolicyDefaults.Controls...)
							policyConfigMap[policyId] = policyConfig
						} else {
							policyConfig := policyGeneratorManifest.Policies[0]
							policyConfig.Standards = policyGeneratorManifest.PolicyDefaults.Standards
							policyConfig.Categories = policyGeneratorManifest.PolicyDefaults.Categories
							policyConfig.Controls = policyGeneratorManifest.PolicyDefaults.Controls
							for idx, manifest := range policyConfig.Manifests {
								policyConfig.Manifests[idx].Path = strings.Replace(manifest.Path, "./", fmt.Sprintf("./%s/", policyId), 1)
							}
							policyConfigMap[policyId] = policyConfig
						}
					}
				}
			}
			suffix := ""
			if idx > 0 {
				suffix = fmt.Sprintf("-%d", idx)
			}
			policySetConfig := pgtype.PolicySetConfig{
				Name:     toDNSCompliant(componentObject.ComponentTitle + suffix),
				Policies: policyListPerControlImple,
			}
			policySets = append(policySets, policySetConfig)
			policySetPatch := typekustomize.Patch{
				Target: &typekustomize.Selector{
					ResId: resid.FromString(fmt.Sprintf("PolicySet../%s.", policySetConfig.Name)),
				},
				Patch: fmt.Sprintf(`[{"op": "replace", "path": "/metadata/annotations/%s", "value": "%s"}]`, pkg.ANNOTATION_COMPONENT_TITLE, componentObject.ComponentTitle),
			}
			policySetPatches = append(policySetPatches, policySetPatch)
		}
	}

	policyDefaults := pgtype.PolicyDefaults{
		Namespace: namespace,
		PolicyOptions: pgtype.PolicyOptions{
			Placement: pgtype.PlacementConfig{
				LabelSelector: clusterSelectors,
			},
		},
		ConfigurationPolicyOptions: pgtype.ConfigurationPolicyOptions{
			NamespaceSelector: pgtype.NamespaceSelector{
				Exclude: []string{"kube-system", "open-cluster-management", "open-cluster-management-agent", "open-cluster-management-agent-addon"},
				Include: []string{"*"},
			},
		},
	}
	policyConfigs := []pgtype.PolicyConfig{}
	for _, policyConfig := range policyConfigMap {
		policyConfigs = append(policyConfigs, policyConfig)
	}
	policySetGeneratorManifest := policygenerator.BuildPolicyGeneratorManifest("policy-set", policyDefaults, policyConfigs)
	policySetGeneratorManifest.PlacementBindingDefaults.Name = "policy-set"
	policySetGeneratorManifest.PolicySets = policySets
	policySetGeneratorManifest.PolicySetDefaults = pgtype.PolicySetDefaults{
		PolicySetOptions: pgtype.PolicySetOptions{
			Placement: policyDefaults.Placement,
		},
	}

	if err := pkg.WriteObjToYamlFileByGoYaml(c.tempDir.GetTempDir()+"/policy-generator.yaml", policySetGeneratorManifest); err != nil {
		return err
	}

	logger.Info("Create configmapt for templatized parameters")
	parametersConfigmap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c2p-parameters",
			Namespace: "c2p",
		},
		Data: parameters,
	}
	if err := pkg.WriteObjToYamlFile(c.tempDir.GetTempDir()+"/parameters.yaml", parametersConfigmap); err != nil {
		return err
	}

	kustomize := typekustomize.Kustomization{
		Generators: []string{"./policy-generator.yaml"},
		Resources:  []string{"./parameters.yaml"},
		Patches:    policySetPatches,
	}
	if err := pkg.WriteObjToYamlFile(c.tempDir.GetTempDir()+"/kustomization.yaml", kustomize); err != nil {
		return err
	}
	logger.Info("")

	return nil
}

func (c *ComposerV2) CopyAllTo(destDir string) error {
	if _, err := pkg.MakeDir(destDir); err != nil {
		return err
	}
	if err := cp.Copy(c.tempDir.GetTempDir(), destDir); err != nil {
		return err
	}
	return nil
}

func (c *ComposerV2) GeneratePolicySet() (*resmap.ResMap, error) {
	generatedManifests, err := policygenerator.Kustomize(c.tempDir.GetTempDir())
	if err != nil {
		logger.Sugar().Error(err, "failed to run kustomize")
		return nil, err
	}
	return &generatedManifests, nil
}

func toDNSCompliant(name string) string {
	var result string
	result = strings.ToLower(name)
	result = strings.ReplaceAll(result, " ", "-")
	return result
}
