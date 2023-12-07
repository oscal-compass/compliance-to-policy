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

package main

import (
	"fmt"
	"os"

	"github.com/IBM/compliance-to-policy/cmd/c2pcli/cmd"
	"github.com/spf13/cobra"
)

var (
	version = "none"
	commit  = "none"
	date    = "unknown"
)

func newVersionSubCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "version",
		Short: "Display version",
		RunE: func(cmd *cobra.Command, args []string) error {
			message := fmt.Sprintf("version: %s, commit: %s, date: %s", version, commit, date)
			fmt.Fprintln(os.Stdout, message)
			return nil
		},
	}
	return command
}

func main() {
	command := cmd.New()
	command.AddCommand(newVersionSubCommand())
	err := command.Execute()
	if err != nil {
		os.Exit(1)
	}
}
