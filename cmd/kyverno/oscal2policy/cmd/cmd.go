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
	"os"

	"github.com/spf13/cobra"

	"github.com/IBM/compliance-to-policy/cmd/kyverno/oscal2policy/options"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/c2pcr"
	"github.com/IBM/compliance-to-policy/pkg/kyverno"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
)

func New() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "oscal2policy",
		Short: "Compose deliverable Kyverno policies from OSCAL",
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
	if err := os.MkdirAll(options.OutputDir, os.ModePerm); err != nil {
		return err
	}

	var c2pcrSpec typec2pcr.Spec
	if err := pkg.LoadYamlFileToObject(options.C2PCRPath, &c2pcrSpec); err != nil {
		return err
	}

	gitUtils := pkg.NewGitUtils(pkg.NewTempDirectory(options.TempDirPath))
	c2pcrParser := c2pcr.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	if err != nil {
		return err
	}

	tmpdir := pkg.NewTempDirectory(options.TempDirPath)
	composer := kyverno.NewOscal2Policy(c2pcrParsed.PolicyResoureDir, tmpdir)
	if err := composer.Generate(c2pcrParsed); err != nil {
		return err
	}

	if options.OutputDir != "" {
		if err := composer.CopyAllTo(options.OutputDir); err != nil {
			return err
		}
	}
	return nil
}
