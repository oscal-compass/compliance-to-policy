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

package main

import (
	"flag"
	"os"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/composer"
	"github.com/IBM/compliance-to-policy/controllers/utils/gitrepo"
	"github.com/IBM/compliance-to-policy/controllers/utils/publisher"
	"github.com/IBM/compliance-to-policy/pkg"
)

func main() {
	var policiesDir, tempDir, cdFile, crFile, outputPath string
	var publishPolicyCollection bool
	flag.StringVar(&policiesDir, "policy-collection-dir", pkg.PathFromPkgDirectory("../out/decomposed/policies"), "path to policy collection")
	flag.StringVar(&tempDir, "temp-dir", "", "path to temp directory")
	flag.StringVar(&cdFile, "cd", "", "path to component-definition.json")
	flag.StringVar(&crFile, "cr", "", "path to compliance-deployment.yaml")
	flag.BoolVar(&publishPolicyCollection, "publish-policy-collection", false, "")
	flag.StringVar(&outputPath, "path", "/published", "")
	flag.Parse()

	_, err := pkg.MakeDir(tempDir)
	if err != nil {
		panic(err)
	}

	composer := composer.NewComposer(policiesDir, tempDir)

	var compDeploy compliancetopolicycontrollerv1alpha1.ComplianceDeployment
	if err := pkg.LoadYamlFileToObject(crFile, &compDeploy); err != nil {
		panic(err)
	}

	username := os.Getenv("username")
	token := os.Getenv("token")
	url := os.Getenv("url")
	gitRepo, err := gitrepo.NewGitRepoWithAuth(tempDir, url, username, token)
	if err != nil {
		panic(err)
	}

	if publishPolicyCollection {
		if err := publisher.PublishPolicyCollection(compDeploy, composer, gitRepo, outputPath); err != nil {
			panic(err)
		}
	} else {
		if err := publisher.Publish("namespace", tempDir, compDeploy, composer, gitRepo, outputPath); err != nil {
			panic(err)
		}
	}
}
