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
	"bytes"
	"embed"
	"os"
	"text/template"

	typereport "github.com/IBM/compliance-to-policy/pkg/types/report"
)

//go:embed *.md
var embeddedResources embed.FS

type Markdown struct{}

func NewMarkdown() Markdown {
	return Markdown{}
}

func (m *Markdown) Generate(templateFile string, report typereport.ComplianceReport) ([]byte, error) {
	var templateData []byte
	var err error
	if templateFile == "" {
		templateData, err = embeddedResources.ReadFile("template.md")
	} else {
		templateData, err = os.ReadFile(templateFile)
	}
	if err != nil {
		return nil, err
	}
	templateString := string(templateData)
	tmpl, err := template.New("report.md").Parse(templateString)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(buffer, report.Spec)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
