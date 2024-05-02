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

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/oscal-compass/compliance-to-policy/go/cmd/ocm/result2oscal/options"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	"github.com/oscal-compass/compliance-to-policy/go/pkg/ocm"
	typec2pcr "github.com/oscal-compass/compliance-to-policy/go/pkg/types/c2pcr"
)

func New() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "result2oscal",
		Short: "Generate OSCAL Assessment Results from OCM Policy statuses",
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

func Run(options *options.Options) error {
	outputPath, c2pcrPath, policyResultsDir, tempDirPath := options.OutputPath, options.C2PCRPath, options.PolicyResultsDir, options.TempDirPath

	var c2pcrSpec typec2pcr.Spec
	if err := pkg.LoadYamlFileToObject(c2pcrPath, &c2pcrSpec); err != nil {
		panic(err)
	}

	gitUtils := pkg.NewGitUtils(pkg.NewTempDirectory(tempDirPath))
	c2pcrParser := ocm.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	if err != nil {
		panic(err)
	}

	r := ocm.NewResultToOscal(c2pcrParsed, policyResultsDir)
	arRoot, err := r.Generate()
	if err != nil {
		panic(err)
	}

	err = pkg.WriteObjToJsonFile(outputPath, arRoot)
	if err != nil {
		return err
	}

	return nil
}
