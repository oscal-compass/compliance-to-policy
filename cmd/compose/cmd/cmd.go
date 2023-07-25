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
	"strings"

	"github.com/spf13/cobra"

	"github.com/IBM/compliance-to-policy/cmd/compose/options"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/c2pcr"
	"github.com/IBM/compliance-to-policy/pkg/composer"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
)

func New() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "compose",
		Short: "Compose deliverable OCM Policies from OSCAL",
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
		panic(err)
	}

	var c2pcrSpec typec2pcr.Spec
	if err := pkg.LoadYamlFileToObject(options.C2PCRPath, &c2pcrSpec); err != nil {
		panic(err)
	}

	gitUtils := pkg.NewGitUtils(pkg.NewTempDirectory(options.TempDirPath))
	c2pcrParser := c2pcr.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	if err != nil {
		panic(err)
	}

	tmpdir := pkg.NewTempDirectory(options.TempDirPath)
	composer := composer.NewComposerByTempDirectory(c2pcrParsed.PolicyResoureDir, tmpdir)
	if err := composer.ComposeByC2PParsed(c2pcrParsed); err != nil {
		panic(err)
	}
	policySet, err := composer.GeneratePolicySet()
	if err != nil {
		panic(err)
	}

	for _, resource := range (*policySet).Resources() {
		name := resource.GetName()
		kind := resource.GetKind()
		namespace := resource.GetNamespace()
		yamlByte, err := resource.AsYAML()
		if err != nil {
			panic(err)
		}
		fnamesTokens := []string{kind, namespace, name}
		fname := strings.Join(fnamesTokens, ".") + ".yaml"
		if err := os.WriteFile(options.OutputDir+"/"+fname, yamlByte, os.ModePerm); err != nil {
			panic(err)
		}
	}

	if options.OutputDirForPolicyGenerator != "" {
		if err := composer.CopyAllTo(options.OutputDirForPolicyGenerator); err != nil {
			panic(err)
		}
	}

	return nil
}
