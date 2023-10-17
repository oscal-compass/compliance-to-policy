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

package kyverno

import (
	"errors"

	"github.com/spf13/pflag"
)

type Options struct {
	SourceDir      string
	DestinationDir string
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.SourceDir, "src", "", "path to a directory of Kyverno policy collection")
	fs.StringVar(&o.DestinationDir, "dest", "", "path to a directory for output retrieved Kyverno policies")
}

func (o *Options) Complete() error {
	return nil
}

func (o *Options) Validate() error {
	if o.SourceDir == "" {
		return errors.New("--src is required")
	}
	if o.DestinationDir == "" {
		return errors.New("--dest is required")
	}
	return nil
}
