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

package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/tables"
	"github.com/IBM/compliance-to-policy/pkg/tables/resources"
	"github.com/IBM/compliance-to-policy/pkg/types/configurationpolicy"
	"github.com/IBM/compliance-to-policy/pkg/types/policycomposition"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

var logger *zap.Logger = pkg.GetLogger("parser")

type Collector struct {
	outputDir     string
	table         *tables.Table
	resourceTable *resources.Table
	erroredTable  *resources.Table
}

func NewCollector(outputDir string) *Collector {
	return &Collector{
		outputDir:     outputDir,
		table:         &tables.Table{},
		resourceTable: &resources.Table{},
		erroredTable:  &resources.Table{},
	}
}

func (c *Collector) GetTable() *tables.Table {
	return c.table
}

func (c *Collector) GetResourceTable() *resources.Table {
	return c.resourceTable
}

func (c *Collector) GetErroredTable() *resources.Table {
	return c.erroredTable
}

func (c *Collector) TraversalFunc(target string) func(path string, info os.FileInfo, err error) error {
	outputTargetDir := c.outputDir + "/" + target
	_ = os.MkdirAll(outputTargetDir, os.ModePerm)
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if err := c.ParseFile(target, outputTargetDir, path, info, err); err != nil {
				logger.Sugar().Infof("Ignore parsing %s due to %s", path, err.Error())
			}
		}
		return nil
	}
}

func (c *Collector) parseFile(target string, outputDir string, path string, filename string, reader io.Reader) error {
	policies, placementBindings, plaementRules, err := loadAndUnmarshal(reader)
	if err != nil {
		return err
	}
	if len(policies) == 0 {
		return fmt.Errorf("No policies are found for %s.", target)
	}
	placementDir := outputDir + "/placements"
	if err := os.MkdirAll(placementDir, os.ModePerm); err != nil {
		return err
	}
	for _, pb := range placementBindings {
		if err := pkg.WriteObjToYamlFile(placementDir+"/"+pb.GetName()+".yaml", pb.Object); err != nil {
			logger.Sugar().Errorf("%v", err)
			return err
		}
	}
	for _, pr := range plaementRules {
		if err := pkg.WriteObjToYamlFile(placementDir+"/"+pr.GetName()+".yaml", pr.Object); err != nil {
			logger.Sugar().Errorf("%v", err)
			return err
		}
	}
	policiesDir := outputDir + "/policies"
	annotations := policies[0].Annotations
	for _, policy := range policies {
		policyDir := policiesDir + "/" + policy.Name
		standard := policy.Annotations["policy.open-cluster-management.io/standards"]
		category := policy.Annotations["policy.open-cluster-management.io/categories"]
		control := policy.Annotations["policy.open-cluster-management.io/controls"]
		for _, policyTemplate := range policy.Spec.PolicyTemplates {
			raw := policyTemplate.ObjectDefinition.Raw
			var configPolicy configurationpolicy.ConfigurationPolicy
			err := utilyaml.Unmarshal(raw, &configPolicy)
			if err != nil {
				logger.Sugar().Errorf("%v", err)
				c.erroredTable.Add(resources.Row{
					Name:         "",
					Kind:         "",
					ApiVersion:   "",
					Policy:       policy.Name,
					ConfigPolicy: configPolicy.Name,
					Standard:     standard,
					Category:     category,
					Control:      control,
					Source:       "",
					PolicyDir:    policyDir,
				})
				continue
			}
			configPoliciesDir := policyDir + "/config-policies"
			configPolicyDir := configPoliciesDir + "/" + configPolicy.Name
			if err := os.MkdirAll(configPolicyDir, os.ModePerm); err != nil {
				return err
			}
			filenameCreator := pkg.NewFilenameCreator(".yaml", nil)
			if configPolicy.Spec.ObjectTemplates == nil {
				c.erroredTable.Add(resources.Row{
					Name:         "",
					Kind:         "",
					ApiVersion:   "",
					Policy:       policy.Name,
					ConfigPolicy: configPolicy.Name,
					Standard:     standard,
					Category:     category,
					Control:      control,
					Source:       "",
					PolicyDir:    policyDir,
				})
				continue
			}
			for _, objectTemplate := range configPolicy.Spec.ObjectTemplates {
				raw := objectTemplate.ObjectDefinition.Raw
				var unst unstructured.Unstructured
				err := utilyaml.Unmarshal(raw, &unst)
				if err != nil {
					logger.Sugar().Errorf("%v", err)
					c.erroredTable.Add(resources.Row{
						Name:         "",
						Kind:         "",
						ApiVersion:   "",
						Policy:       policy.Name,
						ConfigPolicy: configPolicy.Name,
						Standard:     standard,
						Category:     category,
						Control:      control,
						Source:       "",
						PolicyDir:    policyDir,
					})
					continue
				}
				kind := unst.GetKind()
				name := unst.GetName()
				if name == "" {
					name = "noname"
				}
				fnameFmt := "%s.%s"
				fname := fmt.Sprintf(fnameFmt, kind, name)
				fname = filenameCreator.Get(fname)
				outputResourceFile := configPolicyDir + "/" + fname
				if err := pkg.WriteObjToYamlFile(outputResourceFile, unst.Object); err != nil {
					logger.Sugar().Errorf("%v", err)
					c.erroredTable.Add(resources.Row{
						Name:         name,
						Kind:         kind,
						ApiVersion:   unst.GetAPIVersion(),
						Policy:       policy.Name,
						ConfigPolicy: configPolicy.Name,
						Standard:     standard,
						Category:     category,
						Control:      control,
						Source:       outputResourceFile,
						PolicyDir:    policyDir,
					})
					continue
				}
				c.resourceTable.Add(resources.Row{
					Name:         name,
					Kind:         kind,
					ApiVersion:   unst.GetAPIVersion(),
					Policy:       policy.Name,
					ConfigPolicy: configPolicy.Name,
					Standard:     standard,
					Category:     category,
					Control:      control,
					Source:       outputResourceFile,
					PolicyDir:    policyDir,
				})
			}
		}
	}
	c.table.Add(tables.Row{
		Name:        filename,
		Group:       target,
		Standard:    annotations["policy.open-cluster-management.io/standards"],
		Category:    annotations["policy.open-cluster-management.io/categories"],
		Control:     annotations["policy.open-cluster-management.io/controls"],
		Source:      path,
		Destination: outputDir,
	})
	return nil
}

func (c *Collector) ParseFile(target string, outputTargetDir string, path string, info os.FileInfo, _err error) error {
	fname := info.Name()
	outputDir := outputTargetDir + "/" + fname
	outputDir = strings.ReplaceAll(outputDir, ".yaml", "")
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return err
	}
	if err := pkg.CopyFile(path, outputDir+"/policy.yaml"); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("%s, %s", path, fname))

	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := c.parseFile(target, outputDir, path, filepath.Base(f.Name()), f); err != nil {
		logger.Sugar().Infof("Ignore %s and cleanup the output directory %s due to %s", target, outputDir, err.Error())
		if err := os.RemoveAll(outputDir); err != nil {
			logger.Sugar().Errorf("%v", err)
		}
	}
	return nil
}

func (c *Collector) CreatePolicySourcesDir() (string, error) {
	sourcesDir, err := pkg.MakeDir(c.outputDir + "/_sources")
	if err != nil {
		return sourcesDir, err
	}
	if err := createPolicySources(sourcesDir, c.resourceTable); err != nil {
		return sourcesDir, err
	}
	return sourcesDir, createPolicySources(sourcesDir, c.erroredTable)
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

func (c *Collector) Indexer() error {
	resourcesDir := c.outputDir + "/resources"
	groupedByApiVersion := c.resourceTable.GroupBy("api-version")
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

func (c *Collector) AppendCompliance() error {
	type mapKey struct {
		standard string
		category string
		control  string
	}
	groupedByPolicy := c.resourceTable.GroupBy("policy")
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
