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
	"strings"

	policygenerator "github.com/IBM/compliance-to-policy/pkg/policygenerator"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	pgtype "github.com/IBM/compliance-to-policy/pkg/types/policygenerator"
	cp "github.com/otiai10/copy"
	typekustomize "sigs.k8s.io/kustomize/api/types"

	"github.com/IBM/compliance-to-policy/pkg"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/sets"
)

var logger *zap.Logger = pkg.GetLogger("composer")

type Composer struct {
	policiesDir string
	tempDir     pkg.TempDirectory
}

func NewComposer(policiesDir string, tempDir string) *Composer {
	return NewComposerByTempDirectory(policiesDir, pkg.NewTempDirectory(tempDir))
}

func NewComposerByTempDirectory(policiesDir string, tempDir pkg.TempDirectory) *Composer {
	return &Composer{
		policiesDir: policiesDir,
		tempDir:     tempDir,
	}
}

func (c *Composer) GetPoliciesDir() string {
	return c.policiesDir
}

type ControlDirectory struct {
	Path      string
	ControlId string
	Policies  []string
}

type ControlPolicy struct {
	Path                    string
	PolicyId                string
	ConfigPolicies          []string
	Kustomization           File
	PolicyGeneratorManifest File
}

type ControlConfigPolicy struct {
	Path       string
	SourcePath string
}
type File struct {
	Path       string
	SourcePath string
}

func (c *Composer) Compose(namespace string, compliance Compliance, clusterSelectors map[string]string) (*ComposedResult, error) {

	if clusterSelectors == nil {
		clusterSelectors = map[string]string{"env": "dev"}
	}
	policyCompositions := []PolicyComposition{}

	result := ComposedResult{}

	count := 0
	standard := compliance.Standard
	for _, category := range standard.Categories {
		for _, control := range category.Controls {
			for _, policy := range control.ControlRefs {
				logger.Info(fmt.Sprintf("Start generating policy '%s'", policy))

				sourceDir := fmt.Sprintf("%s/%s", c.policiesDir, policy)
				policyCompositionDir := fmt.Sprintf("%s/%s", c.tempDir.GetTempDir(), policy)
				err := cp.Copy(sourceDir, policyCompositionDir)
				if err != nil {
					return nil, err
				}
				policyGeneratorManifestPath := policyCompositionDir + "/policy-generator.yaml"
				var policyGeneratorManifest pgtype.PolicyGenerator
				if err := pkg.LoadYamlFileToObject(sourceDir+"/policy-generator.yaml", &policyGeneratorManifest); err != nil {
					return nil, err
				}
				policyGeneratorManifest.PolicyDefaults.Namespace = namespace
				policyGeneratorManifest.PolicyDefaults.PolicyOptions.Standards = []string{standard.Name}
				policyGeneratorManifest.PolicyDefaults.PolicyOptions.Categories = []string{category.Name}
				policyGeneratorManifest.PolicyDefaults.PolicyOptions.Controls = []string{control.Name}
				policyGeneratorManifest.PolicyDefaults.PolicyOptions.Placement.ClusterSelectors = clusterSelectors
				if err := pkg.WriteObjToYamlFileByGoYaml(policyGeneratorManifestPath, policyGeneratorManifest); err != nil {
					return nil, err
				}
				logger.Info(fmt.Sprintf("Generate policy '%s' by PolicyGenerator", policyGeneratorManifestPath))
				generatedManifests, err := policygenerator.Kustomize(policyCompositionDir)
				if err != nil {
					logger.Sugar().Error(err, "failed to run kustomize")
					return nil, err
				}
				entries, err := os.ReadDir(policyCompositionDir)
				if err != nil {
					return nil, err
				}
				configPolicyAbsoluteDirs := []string{}
				for _, entry := range entries {
					if entry.IsDir() {
						configPolicyAbsoluteDirs = append(configPolicyAbsoluteDirs, policyCompositionDir+"/"+entry.Name())
					}
				}
				policyComposition := PolicyComposition{
					Id:                      policy,
					ControlId:               control.Name,
					PolicyCompositionDir:    policyCompositionDir,
					configPolicyDirs:        configPolicyAbsoluteDirs,
					composedManifests:       &generatedManifests,
					policyGeneratorManifest: policyGeneratorManifest,
				}
				policyCompositions = append(policyCompositions, policyComposition)
				logger.Info(fmt.Sprintf("Finish generating policy '%s'", policy))
				count = count + 1
			}
		}
	}
	result.policyCompositions = policyCompositions
	result.internalCompliance = compliance
	result.namespace = namespace
	result.clusterSelectors = clusterSelectors

	policySetsGeneratorManifest := generatePolicySetsGeneratorManifest(&result)
	if err := pkg.WriteObjToYamlFileByGoYaml(c.tempDir.GetTempDir()+"/policy-generator.yaml", policySetsGeneratorManifest); err != nil {
		return nil, err
	}
	kustomize := typekustomize.Kustomization{Generators: []string{"./policy-generator.yaml"}}
	if err := pkg.WriteObjToYamlFile(c.tempDir.GetTempDir()+"/kustomization.yaml", kustomize); err != nil {
		return nil, err
	}

	generatedManifests, err := policygenerator.Kustomize(c.tempDir.GetTempDir())
	if err != nil {
		logger.Sugar().Error(err, "failed to run kustomize")
		return nil, err
	}
	result.composedManifests = &generatedManifests

	logger.Info("")
	logger.Info(fmt.Sprintf("%d policies are created", count))

	return &result, nil
}

func generatePolicySetsGeneratorManifest(cr *ComposedResult) pgtype.PolicyGenerator {
	policyDefaults := pgtype.PolicyDefaults{
		Namespace: cr.namespace,
		PolicyOptions: pgtype.PolicyOptions{
			Placement: pgtype.PlacementConfig{
				ClusterSelectors: cr.clusterSelectors,
			},
		},
	}
	policyConfigMap := map[string]pgtype.PolicyConfig{}
	policyListPerControl := map[string][]string{}
	for _, policyComposition := range cr.policyCompositions {
		policyName := policyComposition.Id
		controlId := policyComposition.ControlId
		policyList, ok := policyListPerControl[controlId]
		if ok {
			policyListPerControl[controlId] = append(policyList, policyName)
		} else {
			policyListPerControl[controlId] = []string{policyName}
		}
		policyConfig, ok := policyConfigMap[policyName]
		policyGeneratorManifest := policyComposition.policyGeneratorManifest
		if ok {
			policyConfig.Standards = appendUnique(policyConfig.Standards, policyGeneratorManifest.PolicyDefaults.Standards...)
			policyConfig.Categories = appendUnique(policyConfig.Categories, policyGeneratorManifest.PolicyDefaults.Categories...)
			policyConfig.Controls = appendUnique(policyConfig.Controls, policyGeneratorManifest.PolicyDefaults.Controls...)
			policyConfigMap[policyName] = policyConfig
		} else {
			policyConfig := policyGeneratorManifest.Policies[0]
			policyConfig.Standards = policyGeneratorManifest.PolicyDefaults.Standards
			policyConfig.Categories = policyGeneratorManifest.PolicyDefaults.Categories
			policyConfig.Controls = policyGeneratorManifest.PolicyDefaults.Controls
			for idx, manifest := range policyConfig.Manifests {
				policyConfig.Manifests[idx].Path = strings.Replace(manifest.Path, "./", fmt.Sprintf("./%s/", policyName), 1)
			}
			policyConfigMap[policyName] = policyConfig
		}
	}
	policyConfigs := []pgtype.PolicyConfig{}
	for _, policyConfig := range policyConfigMap {
		policyConfigs = append(policyConfigs, policyConfig)
	}
	policySetGeneratorManifest := policygenerator.BuildPolicyGeneratorManifest("policy-set", policyDefaults, policyConfigs)
	policySetGeneratorManifest.PlacementBindingDefaults.Name = "policy-set"
	policySetGeneratorManifest.PolicySets = []pgtype.PolicySetConfig{}
	for controlId, policyList := range policyListPerControl {
		policySetConfig := pgtype.PolicySetConfig{
			Name:     controlId,
			Policies: policyList,
		}
		policySetGeneratorManifest.PolicySets = append(policySetGeneratorManifest.PolicySets, policySetConfig)
	}
	return policySetGeneratorManifest
}

func appendUnique(slice []string, elems ...string) []string {
	a := append(slice, elems...)
	return sets.List[string](sets.New[string](a...))
}

func (c *Composer) CopyAllTo(destDir string) error {
	if _, err := pkg.MakeDir(destDir); err != nil {
		return err
	}
	if err := cp.Copy(c.tempDir.GetTempDir(), destDir); err != nil {
		return err
	}
	return nil
}
