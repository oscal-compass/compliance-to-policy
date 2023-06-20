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
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/parser"
	"github.com/IBM/compliance-to-policy/pkg/tables/resources"
	"github.com/IBM/compliance-to-policy/pkg/types/policycomposition"
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
	err := indexer(collector)
	if err != nil {
		panic(err)
	}
	err = appendCompliance(collector)
	if err != nil {
		panic(err)
	}
	o := &Outputs{}
	o.SourcesDir, err = createPolicySourcesDir(collector)
	if err != nil {
		panic(err)
	}
	o.PolicyCsvPath, o.ResourcesCsvPath = parser.WriteToCSVs(collector, outputDir)
	return o
}

func createPolicySourcesDir(c *parser.Collector) (string, error) {
	sourcesDir, err := pkg.MakeDir(c.GetOutputDir() + "/_sources")
	if err != nil {
		return sourcesDir, err
	}
	if err := createPolicySources(sourcesDir, c.GetResourceTable()); err != nil {
		return sourcesDir, err
	}

	return sourcesDir, createPolicySources(sourcesDir, c.GetErroredTable())
}

func createPolicySources(sourcesDir string, resourcesTable *resources.Table) error {
	filenameCreator := pkg.NewFilenameCreator("", &pkg.FilenameCreatorOption{
		UnlabelToZero: true,
	})
	groupedByPolicy := resourcesTable.GroupBy("policy")
	for policy, table := range groupedByPolicy {
		policyFilename := filenameCreator.Get(policy)
		sourcesPolicyDir, err := pkg.MakeDir(sourcesDir + "/" + policyFilename)
		if err != nil {
			return err
		}
		if err := pkg.CopyFile(table.List()[0].PolicyDir+"/../../policy.yaml", sourcesPolicyDir+"/policy.yaml"); err != nil {
			return err
		}
	}
	return nil
}

func appendCompliance(c *parser.Collector) error {
	type mapKey struct {
		standard string
		category string
		control  string
	}
	groupedByPolicy := c.GetResourceTable().GroupBy("policy")
	for _, table := range groupedByPolicy {
		groupedByPolicyByCompliance := map[mapKey]*resources.Table{}
		groupedByStandard := table.GroupBy("standard")
		for standard, table := range groupedByStandard {
			groupedByCategory := table.GroupBy("category")
			for category, table := range groupedByCategory {
				groupedByControl := table.GroupBy("control")
				for control, table := range groupedByControl {
					mapKey := mapKey{
						standard: standard,
						category: category,
						control:  control,
					}
					groupedByPolicyByCompliance[mapKey] = table
				}
			}
		}
		compliances := []policycomposition.Compliance{}
		policyDir := table.List()[0].PolicyDir
		for mapKey := range groupedByPolicyByCompliance {
			compliance := policycomposition.Compliance{
				Standard: mapKey.standard,
				Category: mapKey.category,
				Control:  mapKey.control,
			}
			compliances = append(compliances, compliance)
		}
		yamlData, err := yaml.Marshal(compliances)
		if err != nil {
			return err
		}
		if err := os.WriteFile(policyDir+"/compliance.yaml", yamlData, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func indexer(c *parser.Collector) error {
	resourcesDir := c.GetOutputDir() + "/resources"
	groupedByApiVersion := c.GetResourceTable().GroupBy("api-version")
	filenameCreator := pkg.NewFilenameCreator(".yaml", nil)
	for apiVersion, table := range groupedByApiVersion {
		apiVersionDir := resourcesDir + "/" + apiVersion
		if err := os.MkdirAll(apiVersionDir, os.ModePerm); err != nil {
			return err
		}
		groupedByKind := table.GroupBy("kind")
		for kind, table := range groupedByKind {
			kindDir := apiVersionDir + "/" + kind
			if err := os.MkdirAll(kindDir, os.ModePerm); err != nil {
				return err
			}
			for _, row := range table.List() {
				name := row.Name
				if name == "" {
					name = "noname"
				}
				fnameFmt := "%s/%s.%s"
				fname := fmt.Sprintf(fnameFmt, kindDir, row.Policy, name)
				fname = filenameCreator.Get(fname)
				if err := pkg.CopyFile(row.Source, fname); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
