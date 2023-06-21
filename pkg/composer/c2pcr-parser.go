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
	"github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	typesoscal "github.com/IBM/compliance-to-policy/pkg/types/oscal"
	typecd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
)

type C2PCRParser struct {
	gitUtils GitUtils
	tempDir  TempDirectory
}

func NewC2PCRParser(gitUtils GitUtils, tempDir TempDirectory) C2PCRParser {
	return C2PCRParser{
		gitUtils: gitUtils,
		tempDir:  tempDir,
	}
}

func (p *C2PCRParser) parse(c2pcrSpec c2pcr.Spec) (*ComposedResult, error) {
	composer, err := p.createComposer(c2pcrSpec)
	if err != nil {
		return nil, err
	}
	internalCompliance, err := p.toInternalCompliance(c2pcrSpec)
	if err != nil {
		return nil, err
	}
	return composer.Compose(c2pcrSpec.Target.Namespace, internalCompliance, *c2pcrSpec.ClusterGroups[0].MatchLabels)
}

func (p *C2PCRParser) createComposer(c2pcrSpec c2pcr.Spec) (*Composer, error) {
	cloneDir, path, err := p.gitUtils.GitClone(c2pcrSpec.PolicyResources.Url)
	if err != nil {
		logger.Sugar().Error(err, fmt.Sprintf("Failed to load policy resources %v", c2pcrSpec.PolicyResources.Url))
		return nil, err
	}
	return NewComposerByTempDirectory(cloneDir+"/"+path, p.tempDir), nil
}

func (p *C2PCRParser) toInternalCompliance(c2pcrSpec c2pcr.Spec) (internalcompliance.Compliance, error) {
	var internalCompliance internalcompliance.Compliance

	logger.Info(fmt.Sprintf("Component-definition is loaded from %s", c2pcrSpec.Compliance.ComponentDefinition.Url))
	var cdobj typecd.ComponentDefinitionRoot
	if err := p.gitUtils.loadFromGit(c2pcrSpec.Compliance.ComponentDefinition.Url, &cdobj); err != nil {
		logger.Sugar().Error(err, "Failed to load component-definition")
		return internalCompliance, err
	}

	logger.Info(fmt.Sprintf("Catalog is loaded from %s", c2pcrSpec.Compliance.Catalog.Url))
	var catalogObj typesoscal.CatalogRoot
	if err := p.gitUtils.loadFromWeb(c2pcrSpec.Compliance.Catalog.Url, &catalogObj); err != nil {
		logger.Sugar().Error(err, "Failed to load catalog")
		return internalCompliance, err
	}

	logger.Info(fmt.Sprintf("Profile is loaded from %s", c2pcrSpec.Compliance.Profile.Url))
	var profileObj typesoscal.ProfileRoot
	if err := p.gitUtils.loadFromWeb(c2pcrSpec.Compliance.Profile.Url, &profileObj); err != nil {
		logger.Sugar().Error(err, "Failed to load profile")
		return internalCompliance, err
	}

	profiledCd := oscal.IntersectProfileWithCD(cdobj.ComponentDefinition, profileObj.Profile)
	internalCompliance = makeInternalCompliance(catalogObj.Catalog, profileObj.Profile, profiledCd)

	return internalCompliance, nil
}

func listRules(props []typecd.Prop) []typecd.Prop {
	newProps := []typecd.Prop{}
	for _, prop := range props {
		if prop.Name == "Rule_Id" {
			newProps = append(newProps, prop)
		}
	}
	return newProps
}

func findRuleSetByRuleId(ruleSetMap map[string]*RuleSet, ruleId string) (*RuleSet, bool) {
	for _, ruleSet := range ruleSetMap {
		if ruleSet.ruleId == ruleId {
			return ruleSet, true
		}
	}
	return nil, false
}

type RuleSet struct {
	ruleId          string
	ruleDescription string
	policyId        string
}

func makeInternalCompliance(catalog typesoscal.Catalog, profile typesoscal.Profile, cd typecd.ComponentDefinition) internalcompliance.Compliance {

	unknownGroup := typesoscal.Group{ID: "unknown", Title: "unknown", Class: "unknown"}

	intCompliances := []internalcompliance.Compliance{}
	for _, component := range cd.Components {
		ruleSetMap := map[string]*RuleSet{}
		for _, prop := range component.Props {
			ruleSetId := prop.Remarks
			ruleSet, ok := ruleSetMap[ruleSetId]
			if !ok {
				ruleSet = &RuleSet{}
				ruleSetMap[ruleSetId] = ruleSet
			}
			switch prop.Name {
			case "Rule_Id":
				ruleSet.ruleId = prop.Value
			case "Rule_Description":
				ruleSet.ruleDescription = prop.Value
			case "Policy_Id":
				ruleSet.policyId = prop.Value
			}
		}
		for _, controlImpl := range component.ControlImplementations {
			controlsPerGroup := map[string][]internalcompliance.Control{}
			for _, implReq := range controlImpl.ImplementedRequirements {
				group, ok := findControlGroup(implReq.ControlID, catalog)
				if !ok {
					group = unknownGroup
				}
				controls, ok := controlsPerGroup[group.ID]
				if !ok {
					controls = []internalcompliance.Control{}
					controlsPerGroup[group.ID] = controls
				}
				controlRefs := []string{}
				for _, prop := range listRules(implReq.Props) {
					ruleSet, ok := findRuleSetByRuleId(ruleSetMap, prop.Value)
					if ok {
						controlRefs = append(controlRefs, ruleSet.policyId)
					}
				}
				controlsPerGroup[group.ID] = append(controls, internalcompliance.Control{
					Name:        implReq.ControlID,
					ControlRefs: controlRefs,
				})
			}
			categories := []internalcompliance.Category{}
			for groupId, controls := range controlsPerGroup {
				category := internalcompliance.Category{
					Name:     groupId,
					Controls: controls,
				}
				categories = append(categories, category)
			}
			standard := internalcompliance.Standard{
				Name:       profile.Metadata.Title,
				Categories: categories,
			}
			intCompliances = append(intCompliances, internalcompliance.Compliance{
				Standard: standard,
			})
		}
	}
	return intCompliances[0]
}

func findControlGroup(controlId string, catalog typesoscal.Catalog) (typesoscal.Group, bool) {
	findCategoriesFromControls := func(controls []typesoscal.Control) bool {
		for _, control := range controls {
			_controlId := control.ID
			if controlId == _controlId {
				return true
			}
			for _, innerControl := range control.Controls {
				_innerControlId := innerControl.ID
				if controlId == _innerControlId {
					return true
				}
			}
		}
		return false
	}

	for _, group := range catalog.Groups {
		if findCategoriesFromControls(group.Controls) {
			return group, true
		}
		for _, subGroup := range group.Groups {
			if findCategoriesFromControls(subGroup.Controls) {
				return group, true
			}
		}
	}
	return typesoscal.Group{}, false
}
