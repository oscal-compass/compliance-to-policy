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

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/kyverno"
	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	kyverno.PolicyResourceIndex
	DestPath string `json:"destPath,omitempty"`
}
type Summary struct {
	ResourcesHavingContext []string `json:"resourcesHavingContext,omitempty"`
}
type Result struct {
	PolicyResourceIndice []policyResourceIndex `json:"policyResourceIndice,omitempty"`
	Summary              Summary               `json:"summary,omitempty"`
}

func Run(options *Options) error {
	srcDir, destDir := options.SourceDir, options.DestinationDir

	if _, err := pkg.MakeDir(destDir); err != nil {
		logger.Error(fmt.Sprintf("Failed to create a destination directory %s", destDir))
		return err
	}

	fl := kyverno.NewFileLoader()

	err := fl.LoadFromDirectory(srcDir)
	if err != nil {
		return err
	}

	inverseMap := map[string][]*policyResourceIndex{}
	policyResourceIndice := []policyResourceIndex{}
	for _, pri := range fl.GetPolicyResourceIndice() {
		policyResourceIndice = append(policyResourceIndice, policyResourceIndex{
			PolicyResourceIndex: pri,
		})
	}
	for idx, pri := range policyResourceIndice {
		_, found := inverseMap[pri.PolicyResourceIndex.Name]
		if found {
			inverseMap[pri.PolicyResourceIndex.Name] = append(inverseMap[pri.PolicyResourceIndex.Name], &policyResourceIndice[idx])
		} else {
			inverseMap[pri.PolicyResourceIndex.Name] = []*policyResourceIndex{&policyResourceIndice[idx]}
		}
	}
	for name, pris := range inverseMap {
		if len(pris) > 1 {
			logger.Warn(fmt.Sprintf("There are duplicate policies for %s", name))
			for _, pri := range pris {
				logger.Warn(fmt.Sprintf("  - %s", pri.PolicyResourceIndex.SrcPath))
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
			pris[idx].DestPath = targetDir + "/" + pri.PolicyResourceIndex.Name + ".yaml"
			if err := cp.Copy(pri.PolicyResourceIndex.SrcPath, pris[idx].DestPath); err != nil {
				logger.Error(fmt.Sprintf("Failed to copy %s", pri.PolicyResourceIndex.SrcPath))
				return err
			}
		}
	}

	resourcesHavingContext := []string{}
	for name, pris := range inverseMap {
		for _, pri := range pris {
			if pri.PolicyResourceIndex.HasContext {
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
