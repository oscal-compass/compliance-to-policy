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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/IBM/compliance-to-policy/cmd/pvpcommon/oscal2posture/options"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/kyverno"
	"github.com/IBM/compliance-to-policy/pkg/pvpcommon"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
)

func New(logger *zap.Logger) *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "oscal2posture",
		Short: "Generate Compliance Posture from OSCAL artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}
			return Run(opts, logger)
		},
	}

	opts.AddFlags(command.Flags())

	return command
}

func Run(options *options.Options, logger *zap.Logger) error {
	output, c2pcrPath, tempDirPath := options.Out, options.C2PCRPath, options.TempDirPath

	var c2pcrSpec typec2pcr.Spec
	if err := pkg.LoadYamlFileToObject(c2pcrPath, &c2pcrSpec); err != nil {
		panic(err)
	}

	gitUtils := pkg.NewGitUtils(pkg.NewTempDirectory(tempDirPath))
	c2pcrParser := kyverno.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	if err != nil {
		return err
	}

	arRoot, err := c2pcrParser.LoadAssessmentResults(options.AssessmentResults)
	if err != nil {
		return err
	}

	r := pvpcommon.NewOscal2Posture(c2pcrParsed, arRoot, nil, logger)
	data, err := r.Generate()
	if err != nil {
		return err
	}

	if output == "-" {
		fmt.Fprintln(os.Stdout, string(data))
	} else {
		return os.WriteFile(output, data, os.ModePerm)
	}

	return nil
}
