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
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
)

func main() {
	var c2pcrPath, tempDirPath, outputDir string
	flag.StringVar(&c2pcrPath, "c2pcr", "c2pcr.yaml", "path to c2p CR")
	flag.StringVar(&tempDirPath, "temp-dir", "", "path to temp directory")
	flag.StringVar(&outputDir, "out", ".", "path to a directory for output files")
	flag.Parse()

	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	var c2pcrSpec typec2pcr.Spec
	if err := pkg.LoadYamlFileToObject(c2pcrPath, &c2pcrSpec); err != nil {
		panic(err)
	}

	gitUtils := composer.NewGitUtils(composer.NewTempDirectory(tempDirPath))
	c2pcrParser := composer.NewC2PCRParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	if err != nil {
		panic(err)
	}

	composer := composer.NewComposerV2ByTempDirectory(c2pcrParsed.PolicyResoureDir, composer.NewTempDirectory(tempDirPath))
	if err := composer.ComposeByC2PParsed(c2pcrParsed); err != nil {
		panic(err)
	}
	policySet, err := composer.GeneratePolicySet()
	if err != nil {
		panic(err)
	}
	policySetYaml, err := (*policySet).AsYaml()
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile(outputDir+"/manifests.yaml", policySetYaml, os.ModePerm); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}
	if err := composer.CopyAllTo(outputDir); err != nil {
		panic(err)
	}
}
