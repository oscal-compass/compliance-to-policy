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

package ocm

import (
	"fmt"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/oscal"
	"github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
)

type C2PCRParser struct {
	gitUtils pkg.GitUtils
}

func NewParser(gitUtils pkg.GitUtils) C2PCRParser {
	return C2PCRParser{
		gitUtils: gitUtils,
	}
}

func (p *C2PCRParser) Parse(c2pcrSpec c2pcr.Spec) (c2pcr.C2PCRParsed, error) {
	var err error
	parsed := c2pcr.C2PCRParsed{}
	parsed.Namespace = c2pcrSpec.Target.Namespace
	parsed.ClusterSelectors = *c2pcrSpec.ClusterGroups[0].MatchLabels
	parsed.PolicyResoureDir, err = p.loadResourceFromUrl(c2pcrSpec.PolicyResources.Url)
	if err != nil {
		return parsed, err
	}

	logger.Info(fmt.Sprintf("Component-definition is loaded from %s", c2pcrSpec.Compliance.ComponentDefinition.Url))
	if err := p.gitUtils.LoadFromGit(c2pcrSpec.Compliance.ComponentDefinition.Url, &parsed.ComponentDefinition); err != nil {
		logger.Sugar().Error(err, "Failed to load component-definition")
		return parsed, err
	}

	if c2pcrSpec.Compliance.Catalog.Url != "" {
		logger.Info(fmt.Sprintf("Catalog is loaded from %s", c2pcrSpec.Compliance.Catalog.Url))
		if err := p.gitUtils.LoadFromWeb(c2pcrSpec.Compliance.Catalog.Url, &parsed.Catalog); err != nil {
			logger.Sugar().Error(err, "Failed to load catalog")
			return parsed, err
		}
	}

	if c2pcrSpec.Compliance.Profile.Url != "" {
		logger.Info(fmt.Sprintf("Profile is loaded from %s", c2pcrSpec.Compliance.Profile.Url))
		if err := p.gitUtils.LoadFromWeb(c2pcrSpec.Compliance.Profile.Url, &parsed.Profile); err != nil {
			logger.Sugar().Error(err, "Failed to load profile")
			return parsed, err
		}
	}

	parsed.ComponentObjects = oscal.ParseComponentDefinition(parsed.ComponentDefinition)

	return parsed, err
}

func (p *C2PCRParser) loadResourceFromUrl(url string) (dirpath string, err error) {
	cloneDir, path, err := p.gitUtils.GitClone(url)
	if err != nil {
		logger.Sugar().Error(err, fmt.Sprintf("Failed to load %v", url))
		return dirpath, err
	}
	return cloneDir + "/" + path, nil
}
