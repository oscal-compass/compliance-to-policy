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

	"go.uber.org/zap"

	"github.com/IBM/compliance-to-policy/cmd/parse/modules"
)

var TARGETS = []string{
	"AC-Access-Control",
	"AU-Audit-and-Accountability",
	"CA-Security-Assessment-and-Authorization",
	"CM-Configuration-Management",
	"SC-System-and-Communications-Protection",
	"SI-System-and-Information-Integrity",
}

func main() {
	var policyCollectionDir, outputDir string
	flag.StringVar(&policyCollectionDir, "policy-collection-dir", ".", "The root directory path to policy-collection repository.")
	flag.StringVar(&outputDir, "out", "./out", "output")
	flag.Parse()

	logger, _ := zap.NewDevelopment()

	if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	_ = modules.Parse(logger, policyCollectionDir, outputDir)
}
