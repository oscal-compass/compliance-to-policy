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
	"fmt"

	"github.com/IBM/compliance-to-policy/pkg"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	cp "github.com/otiai10/copy"
	"go.uber.org/zap"
)

type Oscal2Policy struct {
	policiesDir string
	tempDir     pkg.TempDirectory
	logger      *zap.Logger
}

func NewOscal2Policy(policiesDir string, tempDir pkg.TempDirectory) *Oscal2Policy {
	return &Oscal2Policy{
		policiesDir: policiesDir,
		tempDir:     tempDir,
		logger:      pkg.GetLogger("kyverno/composer"),
	}
}

func (c *Oscal2Policy) Generate(c2pParsed typec2pcr.C2PCRParsed) error {
	for _, componentObject := range c2pParsed.ComponentObjects {
		if componentObject.ComponentType == "validation" {
			continue
		}
		for _, ruleObject := range componentObject.RuleObjects {
			sourceDir := fmt.Sprintf("%s/%s", c.policiesDir, ruleObject.RuleId)
			destDir := fmt.Sprintf("%s/%s", c.tempDir.GetTempDir(), ruleObject.RuleId)
			err := cp.Copy(sourceDir, destDir)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Oscal2Policy) CopyAllTo(destDir string) error {
	if _, err := pkg.MakeDir(destDir); err != nil {
		return err
	}
	if err := cp.Copy(c.tempDir.GetTempDir(), destDir); err != nil {
		return err
	}
	return nil
}
