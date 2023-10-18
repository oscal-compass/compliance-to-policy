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

package kyverno

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/IBM/compliance-to-policy/pkg"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var logger *zap.Logger = pkg.GetLogger("cmd/tools/kyverno")

func New() *cobra.Command {
	opts := NewOptions()

	command := &cobra.Command{
		Use:   "kyverno-policy-resource",
		Short: "Retrieve policies from Kyverno policy collection and create policy resource directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}
			return Run(opts)
		},
	}

	opts.AddFlags(command.Flags())

	return command
}

type policyResourceIndex struct {
	Kind       string `json:"kind,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
	Name       string `json:"name,omitempty"`
	SrcPath    string `json:"srcPath,omitempty"`
	DestPath   string `json:"destPath,omitempty"`
	HasContext bool   `json:"hasContext,omitempty"`
}

type Summary struct {
	ResourcesHavingContext []string `json:"resourcesHavingContext,omitempty"`
}
type Result struct {
	PolicyResourceIndice []policyResourceIndex `json:"policyResourceIndice,omitempty"`
	Summary              Summary               `json:"summary,omitempty"`
}

func mapLoadedObject(unstObj *unstructured.Unstructured, path string) *policyResourceIndex {
	kind, apiVersion, name := unstObj.GetKind(), unstObj.GetAPIVersion(), unstObj.GetName()
	logger.Info(fmt.Sprintf("load yaml %s: %s/%s/%s", path, kind, apiVersion, name))
	return &policyResourceIndex{
		ApiVersion: unstObj.GetAPIVersion(),
		Kind:       unstObj.GetKind(),
		Name:       name,
		SrcPath:    path,
	}
}

func filterByGVKN(pri *policyResourceIndex, unstObj *unstructured.Unstructured) *policyResourceIndex {
	if !(pri.ApiVersion == "kyverno.io/v1" && (pri.Kind == "ClusterPolicy" || pri.Kind == "Policy")) {
		return nil
	}
	_, found, err := unstructured.NestedMap(unstObj.Object, "status")
	if err != nil {
		logger.Info(fmt.Sprintf("  ignore %s due to something error when getting 'status' field", pri.Name))
		return nil
	} else if err == nil && found {
		logger.Info(fmt.Sprintf("  ignore %s since 'status' field is found", pri.Name))
		return nil
	} else {
		return pri
	}
}

func filterByAnnotation(pri *policyResourceIndex, unstObj *unstructured.Unstructured) *policyResourceIndex {
	if pri != nil {
		annotations := unstObj.GetAnnotations()
		_, found := annotations["policies.kyverno.io/title"]
		if !found {
			logger.Info(fmt.Sprintf("  ignore %s due to missing 'policies.kyverno.io/title' annotation", pri.Name))
			return nil
		}
	}
	return pri
}

func addFlag(pri *policyResourceIndex, unstObj *unstructured.Unstructured) *policyResourceIndex {
	if pri != nil {
		rules, found, err := unstructured.NestedSlice(unstObj.Object, "spec", "rules")
		if err == nil && found {
			for _, rule := range rules {
				rule, ok := rule.(map[string]interface{})
				if !ok {
					logger.Warn("Failed to cast")
				} else {
					_, found1, err1 := unstructured.NestedSlice(rule, "context")
					if err1 == nil && found1 {
						pri.HasContext = true
						return pri
					}
				}
			}
		}
	}
	return pri
}

func Run(options *Options) error {
	srcDir, destDir := options.SourceDir, options.DestinationDir

	if _, err := pkg.MakeDir(destDir); err != nil {
		logger.Error(fmt.Sprintf("Failed to create a destination directory %s", destDir))
		return err
	}

	policyResourceIndice := []policyResourceIndex{}

	re := regexp.MustCompile(`^[\.*]`)
	callback := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(fmt.Sprintf("Failed on %s: %v", path, err.Error()))
		}
		if info.IsDir() && re.MatchString(info.Name()) {
			return filepath.SkipDir
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			unstObjs, err := pkg.LoadYaml(path)
			if err == nil {
				for _, unstObj := range unstObjs {
					pri := mapLoadedObject(unstObj, path)
					pri = filterByGVKN(pri, unstObj)
					pri = filterByAnnotation(pri, unstObj)
					pri = addFlag(pri, unstObj)
					if pri != nil {
						policyResourceIndice = append(policyResourceIndice, *pri)
					}
				}
			} else {
				logger.Warn(fmt.Sprintf("%s is not k8s object: %v", path, err.Error()))
			}
		}
		return nil
	}

	err := filepath.Walk(srcDir, callback)
	if err != nil {
		return err
	}

	inverseMap := map[string][]*policyResourceIndex{}
	for idx, pri := range policyResourceIndice {
		_, found := inverseMap[pri.Name]
		if found {
			inverseMap[pri.Name] = append(inverseMap[pri.Name], &policyResourceIndice[idx])
		} else {
			inverseMap[pri.Name] = []*policyResourceIndex{&policyResourceIndice[idx]}
		}
	}
	for name, pris := range inverseMap {
		if len(pris) > 1 {
			logger.Warn(fmt.Sprintf("There are duplicate policies for %s", name))
			for _, pri := range pris {
				logger.Warn(fmt.Sprintf("  - %s", pri.SrcPath))
			}
		}
	}

	fnameCreator := pkg.NewFilenameCreator("", &pkg.FilenameCreatorOption{UnlabelToZero: true})
	for name, pris := range inverseMap {
		for idx, pri := range pris {
			numberedName := fnameCreator.Get(name)
			targetDir, err := pkg.MakeDir(destDir + "/" + numberedName)
			if err != nil {
				return err
			}
			pris[idx].DestPath = targetDir + "/" + pri.Name + ".yaml"
			if err := cp.Copy(pri.SrcPath, pris[idx].DestPath); err != nil {
				logger.Error(fmt.Sprintf("Failed to copy %s", pri.SrcPath))
				return err
			}
		}
	}

	resourcesHavingContext := []string{}
	for name, pris := range inverseMap {
		for _, pri := range pris {
			if pri.HasContext {
				resourcesHavingContext = append(resourcesHavingContext, name)
			}
		}
	}

	result := Result{
		PolicyResourceIndice: policyResourceIndice,
		Summary: Summary{
			ResourcesHavingContext: resourcesHavingContext,
		},
	}

	return pkg.WriteObjToJsonFile(destDir+"/result.json", result)
}
