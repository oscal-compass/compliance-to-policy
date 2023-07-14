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

package options

import (
	"errors"

	"github.com/spf13/pflag"
)

type Options struct {
	C2PCRPath                   string
	TempDirPath                 string
	OutputDir                   string
	OutputDirForPolicyGenerator string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.C2PCRPath, "c2pcr", "", "path to c2p CR")
	fs.StringVar(&o.TempDirPath, "temp-dir", "", "path to temp directory")
	fs.StringVar(&o.OutputDir, "out", ".", "path to a directory for output manifest files of generated OCM Policy manifests")
	fs.StringVar(&o.OutputDirForPolicyGenerator, "out-for-policy-generator", "", "path to a directory for output files for policy generator to generate OCM Policy manifests (default: system temporary directory or directory specified by --temp-dir)")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() error {
	if o.C2PCRPath == "" {
		return errors.New("--c2pcr is required")
	}
	return nil
}
