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
	ComplianceReportFile string
	MdTemplateFile       string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ComplianceReportFile, "compliance-report", "", "path to a compliance report")
	fs.StringVar(&o.MdTemplateFile, "template", "", "path to a template of markdown (default: pre-defined template)")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() error {
	if o.ComplianceReportFile == "" {
		return errors.New("--compliance-report is required")
	}
	return nil
}
