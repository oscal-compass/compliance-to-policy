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

	"github.com/oscal-compass/compliance-to-policy/go/pkg/parser"
)

var TARGETS = []string{
	// "AC-Access-Control",
	// "AU-Audit-and-Accountability",
	// "CA-Security-Assessment-and-Authorization",
	"CM-Configuration-Management",
	// "SC-System-and-Communications-Protection",
	// "SI-System-and-Information-Integrity",
}

func main() {
	var policyPath, outputDir string
	flag.StringVar(&policyPath, "policy-path", ".", "path to ocm-policy")
	flag.StringVar(&outputDir, "out", "./out", "output")
	flag.Parse()

	if err := os.Mkdir(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	collector := parser.NewCollector(outputDir)
	in, err := os.Open(policyPath)
	if err != nil {
		panic(err)
	}
	info, err := in.Stat()
	if err := collector.ParseFile("xxx", outputDir, policyPath, info, err); err != nil {
		panic(err)
	}
	parser.WriteToCSVs(collector, outputDir)
}
