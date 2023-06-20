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
				destDir := policyCompositionDir + "/resources"
				err := cp.Copy(sourceDir, destDir)
				if err != nil {
					return nil, err
				}
				entries, err := os.ReadDir(destDir)
				if err != nil {
					return nil, err
				}
				configPolicyDirs := []string{}
				for _, entry := range entries {
					if entry.IsDir() {
						configPolicyDirs = append(configPolicyDirs, "./resources/"+entry.Name())
					}
				}
				manifests := []pgtype.Manifest{}
				for _, path := range configPolicyDirs {
					manifest := pgtype.Manifest{
						Path: path,
					}
					manifests = append(manifests, manifest)
				}
				policyConfig := pgtype.PolicyConfig{
					Name:      policy,
					Manifests: manifests,
				}
				claim := policygenerator.PolicyGeneratorManifestClaim{
					Namespace:        namespace,
					Standards:        []string{standard.Name},
					Categories:       []string{category.Name},
					Controls:         []string{control.Name},
					Policies:         []pgtype.PolicyConfig{policyConfig},
					ClusterSelectors: clusterSelectors,
				}
				policyGeneratorPath := policyCompositionDir + "/policy-generator.yaml"
				logger.Info(fmt.Sprintf("Create policy-generator.yaml in '%s'", policyGeneratorPath))
				policyGenerator := policygenerator.GeneratePolicyGeneratorManifest(claim)
				if err := pkg.WriteObjToYamlFile(policyGeneratorPath, policyGenerator); err != nil {
					return nil, err
				}
				kustomizePath := policyCompositionDir + "/kustomization.yaml"
				if err := pkg.WriteObjToYamlFile(kustomizePath, pgtype.Kustomization{Generators: []string{"./policy-generator.yaml"}}); err != nil {
					return nil, err
				}
				logger.Info(fmt.Sprintf("Generate policy '%s' by PolicyGenerator", policyGeneratorPath))
				generatedManifests, err := policygenerator.Kustomize(policyCompositionDir)
				if err != nil {
					logger.Sugar().Error(err, "failed to run kustomize")
					return nil, err
				}
				configPolicyAbsoluteDirs := []string{}
				for _, configPolicyDir := range configPolicyDirs {
					configPolicyAbsoluteDirs = append(configPolicyAbsoluteDirs, policyCompositionDir+"/"+configPolicyDir)
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
