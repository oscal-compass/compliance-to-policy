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

package composer

import (
	"os"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/c2pcr"
	typec2pcr "github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	"github.com/stretchr/testify/assert"
)

func TestComposer(t *testing.T) {
	policyDir := pkg.PathFromPkgDirectory("./testdata/policies")
	catalogPath := pkg.PathFromPkgDirectory("./testdata/oscal/catalog.json")
	profilePath := pkg.PathFromPkgDirectory("./testdata/oscal/profile.json")
	cdPath := pkg.PathFromPkgDirectory("./testdata/oscal/component-definition.json")
	// expectedDir := pkg.PathFromPkgDirectory("./composer/testdata/expected/c2pcr-parser-composed-policies")

	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err := os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := pkg.NewTempDirectory(tempDirPath)

	gitUtils := pkg.NewGitUtils(tempDir)

	c2pcrSpec := typec2pcr.Spec{
		Compliance: typec2pcr.Compliance{
			Name: "Test Compliance",
			Catalog: typec2pcr.ResourceRef{
				Url: catalogPath,
			},
			Profile: typec2pcr.ResourceRef{
				Url: profilePath,
			},
			ComponentDefinition: typec2pcr.ResourceRef{
				Url: cdPath,
			},
		},
		PolicyResources: typec2pcr.ResourceRef{
			Url: policyDir,
		},
		PolicyRersults: typec2pcr.ResourceRef{
			Url: "/1/2/3",
		},
		ClusterGroups: []typec2pcr.ClusterGroup{{
			Name:        "test-group",
			MatchLabels: &map[string]string{"environment": "test"},
		}},
		Binding: typec2pcr.Binding{
			Compliance:    "Test Compliance",
			ClusterGroups: []string{"test-group"},
		},
		Target: typec2pcr.Target{
			Namespace: "test",
		},
	}
	c2pcrParser := c2pcr.NewParser(gitUtils)
	c2pcrParsed, err := c2pcrParser.Parse(c2pcrSpec)
	assert.NoError(t, err, "Should not happen")

	composer := NewComposerByTempDirectory(c2pcrParsed.PolicyResoureDir, tempDir)
	err = composer.Compose(c2pcrParsed.Namespace, c2pcrParsed.ComponentObjects, c2pcrParsed.ClusterSelectors)
	assert.NoError(t, err, "Should not happen")
}
