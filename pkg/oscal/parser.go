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

package oscal

import (
	. "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
)

type RuleObject struct {
	RuleId               string
	RuleDescription      string
	PolicyId             string
	ParameterId          string
	ParameterDescription string
}

type ControlObject struct {
	ControlId string
	RuleIds   []string
}

type ControlImpleObject struct {
	SetParameters  []SetParameter
	ControlObjects []ControlObject
}

type ComponentObject struct {
	ComponentTitle      string
	ComponentType       string
	RuleObjects         []RuleObject
	ControlImpleObjects []ControlImpleObject
}

func FindRulesByRuleId(ruleId string, rules []RuleObject) (RuleObject, bool) {
	for _, rule := range rules {
		if rule.RuleId == ruleId {
			return rule, true
		}
	}
	return RuleObject{}, false
}

func GetComponentWideRules(component Component) []RuleObject {
	ruleMap := map[string]*RuleObject{}
	for _, prop := range component.Props {
		ruleId := prop.Remarks
		rule, ok := ruleMap[ruleId]
		if !ok {
			rule = &RuleObject{}
			ruleMap[ruleId] = rule
		}
		switch prop.Name {
		case "Rule_Id":
			rule.RuleId = prop.Value
		case "Rule_Description":
			rule.RuleDescription = prop.Value
		case "Policy_Id":
			rule.PolicyId = prop.Value
		case "Parameter_Id":
			rule.ParameterId = prop.Value
		case "Parameter_Description":
			rule.ParameterDescription = prop.Value
		}
	}
	rules := []RuleObject{}
	for _, rule := range ruleMap {
		rules = append(rules, *rule)
	}
	return rules
}

func ParseComponentDefinition(cd ComponentDefinitionRoot) []ComponentObject {
	componentObjects := []ComponentObject{}
	for _, component := range cd.Components {
		ruleObjects := GetComponentWideRules(component)
		controlImpleObjects := []ControlImpleObject{}
		for _, controlImpl := range component.ControlImplementations {
			controlObjects := []ControlObject{}
			for _, implReq := range controlImpl.ImplementedRequirements {
				ruleIds := []string{}
				for _, prop := range listRules(implReq.Props) {
					ruleIds = append(ruleIds, prop.Value)
				}
				controlObjects = append(controlObjects, ControlObject{
					ControlId: implReq.ControlID,
					RuleIds:   ruleIds,
				})
			}
			controlImpleObjects = append(controlImpleObjects, ControlImpleObject{
				SetParameters:  controlImpl.SetParameters,
				ControlObjects: controlObjects,
			})
		}
		componentObjects = append(componentObjects, ComponentObject{
			ComponentTitle:      component.Title,
			ComponentType:       component.Type,
			RuleObjects:         ruleObjects,
			ControlImpleObjects: controlImpleObjects,
		})
	}
	return componentObjects
}
