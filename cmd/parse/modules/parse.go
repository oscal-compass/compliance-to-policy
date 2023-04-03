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

package modules

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/IBM/compliance-to-policy/pkg/parser"
)

var TARGETS = []string{
	"AC-Access-Control",
	"AU-Audit-and-Accountability",
	"CA-Security-Assessment-and-Authorization",
	"CM-Configuration-Management",
	"SC-System-and-Communications-Protection",
	"SI-System-and-Information-Integrity",
}

type Outputs struct {
	SourcesDir       string
	PolicyCsvPath    string
	ResourcesCsvPath string
}

func Parse(logger *zap.Logger, policyCollectionDir string, outputDir string) *Outputs {

	collector := parser.NewCollector(outputDir)

	for _, target := range TARGETS {
		d := fmt.Sprintf("%s/community/%s", policyCollectionDir, target)
		if err := filepath.Walk(d, collector.TraversalFunc(target)); err != nil {
			logger.Error(err.Error())
		}
	}
	err := collector.Indexer()
	if err != nil {
		panic(err)
	}
	err = collector.AppendCompliance()
	if err != nil {
		panic(err)
	}
	o := &Outputs{}
	o.SourcesDir, err = collector.CreatePolicySourcesDir()
	if err != nil {
		panic(err)
	}
	o.PolicyCsvPath, o.ResourcesCsvPath = parser.WriteToCSVs(collector, outputDir)
	return o
}
