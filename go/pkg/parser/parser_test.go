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

package parser

import (
	"embed"
	"os"
	"testing"

	"github.com/oscal-compass/compliance-to-policy/go/pkg"
)

//go:embed testdata/*
var static embed.FS

func Test(t *testing.T) {
	outputDir := pkg.ChdirFromPkgDirectory("./parser") + "/_test"
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	f, _ := static.Open("testdata/policy.yaml")
	c := NewCollector(outputDir)
	if err := c.parseFile("test", outputDir+"/test", "", "test_policy.yaml", f); err != nil {
		panic(err)
	}

	_, _ = WriteToCSVs(c, outputDir)
}
