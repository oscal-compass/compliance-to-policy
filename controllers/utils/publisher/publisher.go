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

package publisher

import (
	"fmt"
	"os"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/composer"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/gitrepo"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/go-logr/logr"
	cp "github.com/otiai10/copy"
	ctrl "sigs.k8s.io/controller-runtime"
)

var logger logr.Logger = ctrl.Log.WithName("publisher")

func PublishPolicyCollection(
	compDeploy compliancetopolicycontrollerv1alpha1.ComplianceDeployment,
	composer *composer.Composer,
	gitRepo gitrepo.GitRepo,
	path string,
) error {
	destDir := gitRepo.GetDirectory() + path
	if err := cp.Copy(composer.GetPoliciesDir(), destDir); err != nil {
		return err
	}
	if err := gitRepo.Commit(".", "update policy collection"); err != nil {
		return err
	}
	return gitRepo.Push()
}

func Publish(
	namespace string,
	tempDir string,
	compDeploy compliancetopolicycontrollerv1alpha1.ComplianceDeployment,
	composer *composer.Composer,
	gitRepo gitrepo.GitRepo,
	path string,
) error {
	crComposit, err := utils.MakeControlReference(tempDir, compDeploy)
	cr := crComposit.ControlReference
	if err != nil {
		return err
	}
	intCompliance := utils.ConvertComplianceToIntCompliance(cr.Spec.Compliance)

	composedResult, err := composer.Compose(namespace, intCompliance, *compDeploy.Spec.ClusterGroups[0].MatchLabels)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to compose %v", intCompliance))
		return err
	}
	outputDir := gitRepo.GetDirectory() + path
	if err := composedResult.AddGeneratedPolicyManifest(); err != nil {
		return err
	}
	if err := composedResult.CopyTo(outputDir + "/raw-policies"); err != nil {
		return err
	}

	yamlDataList, err := composedResult.ToYaml()
	if err != nil {
		return err
	}
	for policy, yamlData := range yamlDataList {
		resultDir := outputDir + "/deliverable-policies"
		if err := os.MkdirAll(resultDir, os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(resultDir+"/"+policy+".yaml", *yamlData, os.ModePerm); err != nil {
			return err
		}
	}

	outputOscalDir, err := pkg.MakeDir(outputDir + "/oscal")
	if err != nil {
		return err
	}
	if err := pkg.WriteObjToJsonFile(outputOscalDir+"/catalog.json", crComposit.Catalog); err != nil {
		return err
	}
	if err := pkg.WriteObjToJsonFile(outputOscalDir+"/profile.json", crComposit.Profile); err != nil {
		return err
	}
	if err := pkg.WriteObjToJsonFile(outputOscalDir+"/component-definition.json", crComposit.ComponentDefinition); err != nil {
		return err
	}
	if err := pkg.WriteObjToYamlFile(outputDir+"/cr.yaml", compDeploy); err != nil {
		return err
	}
	if err := composedResult.WriteSelectedPoliciesToYamlFile(outputDir + "/selected-policies-in-component-definitions.yaml"); err != nil {
		return err
	}
	if err := gitRepo.Commit(".", "update"); err != nil {
		return err
	}
	return gitRepo.Push()
}
