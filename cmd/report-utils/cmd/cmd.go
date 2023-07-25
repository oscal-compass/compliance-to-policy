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

	"github.com/IBM/compliance-to-policy/cmd/report-utils/options"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/reporter"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
)

func New() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "report-utils",
		Short: "Utilities for reporting",
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

	var report typereport.ComplianceReport
	err := pkg.LoadYamlFileToK8sTypedObject(options.ComplianceReportFile, &report)
	if err != nil {
		panic(err)
	}

	md := reporter.NewMarkdown()
	generated, err := md.Generate(options.MdTemplateFile, report)
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stdout, string(generated))

	return nil
}
