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

	cmdparse "github.com/IBM/compliance-to-policy/cmd/parse/modules"
	"github.com/IBM/compliance-to-policy/pkg/decomposer"
	cp "github.com/otiai10/copy"
	"go.uber.org/zap"
)

func main() {
	var policyCollectionDir, outputDir string
	flag.StringVar(&policyCollectionDir, "policy-collection-dir", ".", "The root directory path to policy-collection repository.")
	flag.StringVar(&outputDir, "out", "./out", "output")
	flag.Parse()

	logger, _ := zap.NewDevelopment()

	parsedResultsDir := outputDir + "/parsed"
	if err := os.MkdirAll(parsedResultsDir, os.ModePerm); err != nil {
		panic(err)
	}
	decomposedResultsDir := outputDir + "/decomposed"
	if err := os.MkdirAll(decomposedResultsDir, os.ModePerm); err != nil {
		panic(err)
	}

	parsedOutput := cmdparse.Parse(logger, policyCollectionDir, parsedResultsDir)
	f, err := os.Open(parsedOutput.ResourcesCsvPath)
	if err != nil {
		panic(err)
	}

	decomposer, err := decomposer.NewDecomposer(f, decomposedResultsDir)
	if err != nil {
		panic(err)
	}
	_, err = decomposer.Decompose()
	if err != nil {
		panic(err)
	}

	if err := cp.Copy(parsedOutput.SourcesDir, decomposedResultsDir+"/_sources"); err != nil {
		panic(err)
	}
}
