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
	C2PCRPath   string
	TempDirPath string
	OutputPath  string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.C2PCRPath, "config", "c", "", "path to c2p-config.yaml")
	fs.StringVar(&o.TempDirPath, "temp-dir", "", "path to temp directory")
	fs.StringVarP(&o.OutputPath, "out", "o", "./assessment-results.json", "path to output OSCAL Assessment Results")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() error {
	if o.C2PCRPath == "" {
		return errors.New("-c or --config <c2p-config.yaml> is required")
	}
	return nil
}
