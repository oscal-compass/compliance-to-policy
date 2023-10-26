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
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type PolicyResourceIndex struct {
	Kind       string `json:"kind,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	SrcPath    string `json:"srcPath,omitempty"`
	HasContext bool   `json:"hasContext,omitempty"`
}

type FileLoader struct {
	logger               *zap.Logger
	policyResourceIndice []PolicyResourceIndex
}

func NewFileLoader() *FileLoader {
	return &FileLoader{
		logger:               pkg.GetLogger("kyverno/fileloader"),
		policyResourceIndice: []PolicyResourceIndex{},
	}
}

func (fl *FileLoader) GetPolicyResourceIndice() []PolicyResourceIndex {
	return fl.policyResourceIndice
}

func (fl *FileLoader) LoadFromDirectory(dir string) error {

	re := regexp.MustCompile(`^[\.*]`)
	callback := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fl.logger.Error(fmt.Sprintf("Failed on %s: %v", path, err.Error()))
		}
		if info.IsDir() && re.MatchString(info.Name()) {
			return filepath.SkipDir
		}
		if !info.IsDir() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			unstObjs, err := pkg.LoadYaml(path)
			if err == nil {
				for _, unstObj := range unstObjs {
					pri := fl.mapLoadedObject(unstObj, path)
					pri = fl.filterByGVKN(pri, unstObj)
					pri = fl.filterByAnnotation(pri, unstObj)
					pri = fl.addFlag(pri, unstObj)
					if pri != nil {
						fl.policyResourceIndice = append(fl.policyResourceIndice, *pri)
					}
				}
			} else {
				fl.logger.Warn(fmt.Sprintf("%s is not k8s object: %v", path, err.Error()))
			}
		}
		return nil
	}

	err := filepath.Walk(dir, callback)
	if err != nil {
		return err
	}

	return nil
}

func (fl *FileLoader) mapLoadedObject(unstObj *unstructured.Unstructured, path string) *PolicyResourceIndex {
	kind, apiVersion, name := unstObj.GetKind(), unstObj.GetAPIVersion(), unstObj.GetName()
	fl.logger.Info(fmt.Sprintf("load yaml %s: %s/%s/%s", path, kind, apiVersion, name))
	return &PolicyResourceIndex{
		ApiVersion: unstObj.GetAPIVersion(),
		Kind:       unstObj.GetKind(),
		Name:       name,
		Namespace:  unstObj.GetNamespace(),
		SrcPath:    path,
	}
}

func (fl *FileLoader) filterByGVKN(pri *PolicyResourceIndex, unstObj *unstructured.Unstructured) *PolicyResourceIndex {
	if !(pri.ApiVersion == "kyverno.io/v1" && (pri.Kind == "ClusterPolicy" || pri.Kind == "Policy")) {
		return nil
	}
	_, found, err := unstructured.NestedMap(unstObj.Object, "status")
	if err != nil {
		fl.logger.Info(fmt.Sprintf("  ignore %s due to something error when getting 'status' field", pri.Name))
		return nil
	} else if err == nil && found {
		fl.logger.Info(fmt.Sprintf("  ignore %s since 'status' field is found", pri.Name))
		return nil
	} else {
		return pri
	}
}

func (fl *FileLoader) filterByAnnotation(pri *PolicyResourceIndex, unstObj *unstructured.Unstructured) *PolicyResourceIndex {
	if pri != nil {
		annotations := unstObj.GetAnnotations()
		_, found := annotations["policies.kyverno.io/title"]
		if !found {
			fl.logger.Info(fmt.Sprintf("  ignore %s due to missing 'policies.kyverno.io/title' annotation", pri.Name))
			return nil
		}
	}
	return pri
}

func (fl *FileLoader) addFlag(pri *PolicyResourceIndex, unstObj *unstructured.Unstructured) *PolicyResourceIndex {
	if pri != nil {
		rules, found, err := unstructured.NestedSlice(unstObj.Object, "spec", "rules")
		if err == nil && found {
			for _, rule := range rules {
				rule, ok := rule.(map[string]interface{})
				if !ok {
					fl.logger.Warn("Failed to cast")
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
