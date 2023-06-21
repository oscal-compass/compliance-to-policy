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

	policygenerator "github.com/IBM/compliance-to-policy/pkg/policygenerator"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	pgtype "github.com/IBM/compliance-to-policy/pkg/types/policygenerator"
	cp "github.com/otiai10/copy"

	"github.com/IBM/compliance-to-policy/pkg"
	"go.uber.org/zap"
)

var logger *zap.Logger = pkg.GetLogger("composer")

type Composer struct {
	policiesDir string
	tempDir     string
}

func NewComposer(policiesDir string, tempDir string) *Composer {
	dir, err := os.MkdirTemp(tempDir, "tmp-")
	if err != nil {
		panic(err)
	}
	return &Composer{
		policiesDir: policiesDir,
		tempDir:     dir,
	}
}

func (c *Composer) GetPoliciesDir() string {
	return c.policiesDir
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
			controlDir := c.tempDir + "/" + control.Name
			for _, policy := range control.ControlRefs {
				logger.Info(fmt.Sprintf("Start generating policy '%s'", policy))

				sourceDir := fmt.Sprintf("%s/%s", c.policiesDir, policy)
				policyCompositionDir := fmt.Sprintf("%s/%s", controlDir, policy)
				err := cp.Copy(sourceDir, policyCompositionDir)
				if err != nil {
					return nil, err
				}
				policyGeneratorManifestPath := policyCompositionDir + "/policy-generator.yaml"
				var policyGeneratorManifest pgtype.PolicyGenerator
				if err := pkg.LoadYamlFileToObject(policyGeneratorManifestPath, &policyGeneratorManifest); err != nil {
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
					Id:                   policy,
					PolicyCompositionDir: policyCompositionDir,
					configPolicyDirs:     configPolicyAbsoluteDirs,
					composedManifests:    &generatedManifests,
				}
				policyCompositions = append(policyCompositions, policyComposition)
				logger.Info(fmt.Sprintf("Finish generating policy '%s'", policy))
				count = count + 1
			}
		}
	}
	result.policyCompositions = policyCompositions
	result.internalCompliance = compliance

	logger.Info("")
	logger.Info(fmt.Sprintf("%d policies are created", count))

	return &result, nil
}

func (c *Composer) CopyPoliciesDirTo(destDir string) error {
	if _, err := pkg.MakeDir(destDir); err != nil {
		return err
	}
	if err := cp.Copy(c.policiesDir, destDir); err != nil {
		return err
	}
	return nil
}
