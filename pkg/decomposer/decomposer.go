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

package decomposer

import (
	"io"
	"os"
	"path/filepath"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/tables/resources"
	. "github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	"github.com/IBM/compliance-to-policy/pkg/types/policycomposition"
)

type Decomposer struct {
	resourceTable *resources.Table
	outputDirPath string
}

func NewDecomposer(resourceTableFile io.Reader, outputDirPath string) (*Decomposer, error) {
	if err := os.MkdirAll(outputDirPath, os.ModePerm); err != nil {
		return nil, err
	}
	return &Decomposer{
		resourceTable: resources.FromCsv(resourceTableFile),
		outputDirPath: outputDirPath,
	}, nil
}

func (d *Decomposer) Decompose() (string, error) {
	policyResourcesDir, err := pkg.MakeDir(d.outputDirPath + "/resources")
	if err != nil {
		return policyResourcesDir, err
	}
	groupedByPolicy := d.resourceTable.GroupBy("policy")
	filenameCreator := pkg.NewFilenameCreator("", &pkg.FilenameCreatorOption{
		UnlabelToZero: true,
	})
	for policy, table := range groupedByPolicy {
		policyFilename := filenameCreator.Get(policy)
		policyDir, err := pkg.MakeDir(policyResourcesDir + "/" + policyFilename)
		if err != nil {
			return policyResourcesDir, err
		}
		if err := pkg.CopyFile(table.List()[0].PolicyDir+"/kustomization.yaml", policyDir+"/kustomization.yaml"); err != nil {
			return policyResourcesDir, err
		}
		if err := pkg.CopyFile(table.List()[0].PolicyDir+"/policy-generator.yaml", policyDir+"/policy-generator.yaml"); err != nil {
			return policyResourcesDir, err
		}
		policyResourcesDir := policyDir
		groupedByPolicyByCompliance := GroupByCompliance(table)
		for _, table := range groupedByPolicyByCompliance {
			groupedByConfigPolicy := table.GroupBy("config-policy")
			for configPolicy, table := range groupedByConfigPolicy {
				configPolicyDir, err := pkg.MakeDir(policyResourcesDir + "/" + configPolicy)
				if err != nil {
					return policyResourcesDir, err
				}
				for _, row := range table.List() {
					fname := filepath.Base(row.Source)
					if err := pkg.CopyFile(row.Source, configPolicyDir+"/"+fname); err != nil {
						return policyResourcesDir, err
					}
				}
			}
		}
	}
	return policyResourcesDir, nil
}

func mapToHierarchy(x0 map[string]map[string]map[string][]string, key1 string, key2 string, key3 string, value []string) {
	x1, ok := x0[key1]
	if !ok {
		x1 = map[string]map[string][]string{}
		x0[key1] = x1
	}
	x2, ok := x1[key2]
	if !ok {
		x2 = map[string][]string{}
		x1[key2] = x2
	}
	_, ok = x2[key3]
	if !ok {
		x3 := value
		x2[key3] = x3
	}
}

func GroupByComplianceInHierarchy(resourceTable *resources.Table) []Compliance {
	groupedByPolicyByCompliance := GroupByCompliance(resourceTable)

	standards := map[string]map[string]map[string][]string{}
	for compliance, table := range groupedByPolicyByCompliance {
		standard := compliance.Standard
		category := compliance.Category
		control := compliance.Control

		policies := []string{}
		for policy := range table.GroupBy("policy") {
			policies = append(policies, policy)
		}
		mapToHierarchy(standards, standard, category, control, policies)
	}
	compliances := []Compliance{}
	for x1 := range standards {
		categories := []Category{}
		for x2 := range standards[x1] {
			controls := []Control{}
			for x3 := range standards[x1][x2] {
				control := Control{
					Name:        x3,
					ControlRefs: standards[x1][x2][x3],
				}
				controls = append(controls, control)
			}
			category := Category{
				Name:     x2,
				Controls: controls,
			}
			categories = append(categories, category)
		}
		standard := Standard{
			Name:       x1,
			Categories: categories,
		}
		compliance := Compliance{
			Standard: standard,
		}
		compliances = append(compliances, compliance)
	}
	return compliances
}

func (d *Decomposer) GroupByCompliance() error {
	compliancesDir, err := pkg.MakeDir(d.outputDirPath + "/compliances")
	if err != nil {
		return err
	}
	complianceLists := GroupByComplianceInHierarchy(d.resourceTable)
	for _, complianceList := range complianceLists {
		path := compliancesDir + "/" + complianceList.Standard.Name + ".yaml"
		_ = pkg.WriteObjToYamlFile(path, complianceList)
	}
	return nil
}

func GroupByCompliance(table *resources.Table) map[policycomposition.Compliance]*resources.Table {
	groupedByPolicyByCompliance := map[policycomposition.Compliance]*resources.Table{}
	groupedByStandard := table.GroupBy("standard")
	for standard, table := range groupedByStandard {
		groupedByCategory := table.GroupBy("category")
		for category, table := range groupedByCategory {
			groupedByControl := table.GroupBy("control")
			for control, table := range groupedByControl {
				compliance := policycomposition.Compliance{
					Standard: standard,
					Category: category,
					Control:  control,
				}
				groupedByPolicyByCompliance[compliance] = table
			}
		}
	}
	return groupedByPolicyByCompliance
}
