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

package reporter

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/IBM/compliance-to-policy/pkg"
	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
)

func TestMarkdown(t *testing.T) {
	tmplateData, err := embeddedResources.ReadFile("template.md")
	assert.NoError(t, err, "Should not happen")

	var report typereport.ComplianceReport
	err = pkg.LoadYamlFileToK8sTypedObject(pkg.PathFromPkgDirectory("./testdata/reports/compliance-report.yaml"), &report)
	assert.NoError(t, err, "Should not happen")

	markdown := NewMarkdown()
	generated, err := markdown.Generate(string(tmplateData), report)
	assert.NoError(t, err, "Should not happen")

	err = os.WriteFile(pkg.PathFromPkgDirectory("./testdata/reports/compliance-report.md"), generated, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
}
