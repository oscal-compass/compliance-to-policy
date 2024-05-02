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

	"github.com/oscal-compass/compliance-to-policy/go/cmd/c2pcli/options"
	oscal2policycmd "github.com/oscal-compass/compliance-to-policy/go/cmd/kyverno/oscal2policy/cmd"
	result2oscalcmd "github.com/oscal-compass/compliance-to-policy/go/cmd/kyverno/result2oscal/cmd"
	toolscmd "github.com/oscal-compass/compliance-to-policy/go/cmd/kyverno/tools/cmd"
)

func NewKyvernoSubCommand() *cobra.Command {
	opts := options.NewOptions()

	command := &cobra.Command{
		Use:   "kyverno",
		Short: "C2P CLI Kyverno plugin",
	}

	opts.AddFlags(command.Flags())

	command.AddCommand(oscal2policycmd.New())
	command.AddCommand(result2oscalcmd.New())
	command.AddCommand(toolscmd.New())

	return command
}
