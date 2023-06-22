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
	"fmt"

	"github.com/IBM/compliance-to-policy/pkg/oscal"
	"github.com/IBM/compliance-to-policy/pkg/types/c2pcr"
	typesoscal "github.com/IBM/compliance-to-policy/pkg/types/oscal"
	typecd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
)

type C2PCRParser struct {
	gitUtils GitUtils
	parsed   C2PCRParsed
}

type C2PCRParsed struct {
	namespace           string
	PolicyResoureDir    string
	catalog             typesoscal.CatalogRoot
	profile             typesoscal.ProfileRoot
	componentDefinition typecd.ComponentDefinitionRoot
	componentObjects    []oscal.ComponentObject
	clusterSelectors    map[string]string
}

func NewC2PCRParser(gitUtils GitUtils) C2PCRParser {
	return C2PCRParser{
		gitUtils: gitUtils,
	}
}

func (p *C2PCRParser) Parse(c2pcrSpec c2pcr.Spec) (C2PCRParsed, error) {
	parsed := C2PCRParsed{}
	parsed.namespace = c2pcrSpec.Target.Namespace
	parsed.clusterSelectors = *c2pcrSpec.ClusterGroups[0].MatchLabels
	cloneDir, path, err := p.gitUtils.GitClone(c2pcrSpec.PolicyResources.Url)
	if err != nil {
		logger.Sugar().Error(err, fmt.Sprintf("Failed to load policy resources %v", c2pcrSpec.PolicyResources.Url))
		return parsed, err
	}
	parsed.PolicyResoureDir = cloneDir + "/" + path

	logger.Info(fmt.Sprintf("Component-definition is loaded from %s", c2pcrSpec.Compliance.ComponentDefinition.Url))
	if err := p.gitUtils.loadFromGit(c2pcrSpec.Compliance.ComponentDefinition.Url, &parsed.componentDefinition); err != nil {
		logger.Sugar().Error(err, "Failed to load component-definition")
		return parsed, err
	}

	logger.Info(fmt.Sprintf("Catalog is loaded from %s", c2pcrSpec.Compliance.Catalog.Url))
	if err := p.gitUtils.loadFromWeb(c2pcrSpec.Compliance.Catalog.Url, &parsed.catalog); err != nil {
		logger.Sugar().Error(err, "Failed to load catalog")
		return parsed, err
	}

	logger.Info(fmt.Sprintf("Profile is loaded from %s", c2pcrSpec.Compliance.Profile.Url))
	if err := p.gitUtils.loadFromWeb(c2pcrSpec.Compliance.Profile.Url, &parsed.profile); err != nil {
		logger.Sugar().Error(err, "Failed to load profile")
		return parsed, err
	}

	parsed.componentObjects = oscal.ParseComponentDefinition(parsed.componentDefinition)

	return parsed, err
}
