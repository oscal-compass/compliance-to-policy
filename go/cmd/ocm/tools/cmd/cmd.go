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

	"github.com/oscal-compass/compliance-to-policy/go/cmd/c2pcli/options"
	oscal2posturecmd "github.com/oscal-compass/compliance-to-policy/go/cmd/pvpcommon/oscal2posture/cmd"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
)

func New() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "tools",
		Short: "Tools",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Complete(); err != nil {
				return err
			}

			if err := opts.Validate(); err != nil {
				return err
			}
			return nil
		},
	}

	opts.AddFlags(command.Flags())

	command.AddCommand(oscal2posturecmd.New(pkg.GetLogger("ocm/oscal2posture")))

	return command
}
