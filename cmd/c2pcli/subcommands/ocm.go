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

package subcommands

import (
	"github.com/spf13/cobra"

	"github.com/IBM/compliance-to-policy/cmd/c2pcli/options"
	oscal2policycmd "github.com/IBM/compliance-to-policy/cmd/ocm/oscal2policy/cmd"
	result2oscalcmd "github.com/IBM/compliance-to-policy/cmd/ocm/result2oscal/cmd"
	oscal2posturecmd "github.com/IBM/compliance-to-policy/cmd/pvpcommon/oscal2posture/cmd"
	"github.com/IBM/compliance-to-policy/pkg"
)

func NewOcmSubCommand() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "ocm",
		Short: "C2P CLI OCM plugin",
	}

	opts.AddFlags(command.Flags())

	command.AddCommand(oscal2policycmd.New())
	command.AddCommand(result2oscalcmd.New())
	command.AddCommand(oscal2posturecmd.New(pkg.GetLogger("ocm/oscal2posture")))

	return command
}
