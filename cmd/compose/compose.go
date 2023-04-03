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

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/composer"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
)

func main() {
	var policiesDir, tempDir, complianceYamlPath, outputDir string
	flag.StringVar(&policiesDir, "policy-resources-dir", pkg.PathFromPkgDirectory("../out/decomposed/resources"), "path to policy resources")
	flag.StringVar(&tempDir, "temp-dir", "", "path to temp directory")
	flag.StringVar(&complianceYamlPath, "compliance-yaml", "", "path to compliance yaml")
	flag.StringVar(&outputDir, "out", pkg.PathFromPkgDirectory("../out/composed"), "path to a directory for output files")
	flag.Parse()

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}
	compliance := Compliance{}
	if err := pkg.LoadYamlFileToObject(complianceYamlPath, &compliance); err != nil {
		panic(err)
	}

	c := composer.NewComposer(policiesDir, tempDir)
	result, err := c.Compose("default", compliance, nil)
	if err != nil {
		panic(err)
	}
	yamlDataList, err := result.ToYaml()
	if err != nil {
		panic(err)
	}
	for policy, yamlData := range yamlDataList {
		resultDir := outputDir + "/composed"
		if err := os.MkdirAll(resultDir, os.ModePerm); err != nil {
			panic(err)
		}
		if err := os.WriteFile(resultDir+"/"+policy+".yaml", *yamlData, os.ModePerm); err != nil {
			panic(err)
		}
	}
}
