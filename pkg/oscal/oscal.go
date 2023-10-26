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
	"strings"

	"github.com/IBM/compliance-to-policy/pkg/decomposer"
	"github.com/IBM/compliance-to-policy/pkg/tables/resources"
	"github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	"github.com/IBM/compliance-to-policy/pkg/types/oscal"
	cd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
	"github.com/google/uuid"
)

var standardFromPolicyCollectionToOscal map[string]string = map[string]string{
	"NIST SP 800-53": "NIST_SP-800-53_rev5_catalog",
}

const (
	OscaleNamespace = "http://ibm.github.io/compliance-trestle/schemas/oscal/cd"
)

func controlIdFromPolicyCollectionToOscal(controlId string) (string, bool) {
	// CM-6 Configuration Settings => CM-6
	// AU-5 => AU-5
	cid := strings.Split(controlId, " ")
	return strings.ToLower(cid[0]), true
}

func makeProp(name string, value string, remarks string) cd.Prop {
	return cd.Prop{
		Name:    name,
		Ns:      OscaleNamespace,
		Value:   value,
		Remarks: remarks,
	}
}

func listRules(props []cd.Prop) []cd.Prop {
	newProps := []cd.Prop{}
	for _, prop := range props {
		if prop.Name == "Rule_Id" {
			newProps = append(newProps, prop)
		}
	}
	return newProps
}

func GenerateUUID() string {
	uuid, err := uuid.NewUUID()
	if err != nil {
		return "01234567-yyyy-zzzz-1111-0123456789ab"
	}
	return uuid.String()
}

func makeComponentDefinitionFromMasterData(resourceTable *resources.Table) *cd.ComponentDefinition {
	complianceLists := decomposer.GroupByComplianceInHierarchy(resourceTable)
	controlImplementations := []cd.ControlImplementation{}
	for _, complianceList := range complianceLists {
		standard, ok := standardFromPolicyCollectionToOscal[complianceList.Standard.Name]
		if !ok {
			continue
		}
		implementedRequirements := []cd.ImplementedRequirement{}
		for _, category := range complianceList.Standard.Categories {
			for _, control := range category.Controls {
				controlId, ok := controlIdFromPolicyCollectionToOscal(control.Name)
				if !ok {
					continue
				}
				props := []cd.Prop{}
				for _, controlRef := range control.ControlRefs {
					ruleProp := makeProp("Rule_Id", controlRef, "")
					props = append(props, ruleProp)
				}
				implementedRequirements = append(implementedRequirements, cd.ImplementedRequirement{
					UUID:      GenerateUUID(),
					ControlID: controlId,
					Props:     props,
				})
			}
		}
		controlImplementations = append(controlImplementations, cd.ControlImplementation{
			UUID:                    GenerateUUID(),
			Source:                  standard,
			ImplementedRequirements: implementedRequirements,
		})
	}
	component := cd.Component{
		UUID:                   "c8106bc8-5174-4e86-91a4-52f2fe0ed027",
		Type:                   "service",
		Title:                  "My Kubernetes service",
		Description:            "My Kubernetes service ...",
		ControlImplementations: controlImplementations,
	}
	return &cd.ComponentDefinition{
		UUID: "c14d8812-7098-4a9b-8f89-cba41b6ff0d8",
		Metadata: cd.Metadata{
			Title: "Component definition",
		},
		Components: []cd.Component{component},
	}
}

type TrestleCsvRow struct {
	TrestleComponentProps
	ControlIdList        []string
	RuleId               string
	RuleDescription      string
	ParameterId          string
	ParameterDescription string
	ProfileSource        string
	ProfileDescription   string
	Namespace            string
}

type TrestleComponentProps struct {
	ComponentTitle       string
	ComponentDescription string
	ComponentType        string
}

func (t *TrestleCsvRow) Header() []string {
	return []string{
		"$$Component_Title",
		"$$Component_Type",
		"$$Control_Id_List",
		"$$Rule_Id",
		"$$Rule_Description",
		"$Parameter_Id",
		"$Parameter_Description",
		"$$Profile_Source",
		"$$Profile_Description",
		"$$Namespace",
		"$$Component_Description",
	}
}

func (t *TrestleCsvRow) ToStringList() []string {
	controlIdList := []string{}
	for _, control := range t.ControlIdList {
		controlId, _ := controlIdFromPolicyCollectionToOscal(control)
		controlIdList = append(controlIdList, controlId)
	}
	return []string{
		t.ComponentTitle,
		t.ComponentType,
		strings.Join(controlIdList, " "),
		t.RuleId,
		t.RuleDescription,
		t.ParameterId,
		t.ParameterDescription,
		t.ProfileSource,
		t.ProfileDescription,
		t.Namespace,
		t.ComponentDescription,
	}
}

func makeTrestleCsvFromMasterData(componentProps TrestleComponentProps, namespace string, resourceTable *resources.Table) map[string][]TrestleCsvRow {
	rowsMap := map[string][]TrestleCsvRow{}
	groupedByStandard := resourceTable.GroupBy("standard")
	for standard, resourcesByStandard := range groupedByStandard {
		rows := []TrestleCsvRow{}
		groupedByPolicy := resourcesByStandard.GroupBy("policy")
		for policy, resourcesByPolicy := range groupedByPolicy {
			groupedByControl := resourcesByPolicy.GroupBy("control")
			controls := []string{}
			for control := range groupedByControl {
				controls = append(controls, control)
			}
			row := TrestleCsvRow{
				TrestleComponentProps: componentProps,
				RuleId:                policy,
				RuleDescription:       "Description of " + policy,
				ProfileSource:         standard,
				ProfileDescription:    "Description of " + standard,
				ControlIdList:         controls,
				Namespace:             namespace,
			}
			rows = append(rows, row)
		}
		rowsMap[standard] = rows
	}
	return rowsMap
}

func IntersectProfileWithCD(compDef cd.ComponentDefinition, profile oscal.Profile) cd.ComponentDefinition {
	components := []cd.Component{}
	for _, component := range compDef.Components {
		controlImplementations := []cd.ControlImplementation{}
		for _, controlImpl := range component.ControlImplementations {
			implReqs := []cd.ImplementedRequirement{}
			for _, implReq := range controlImpl.ImplementedRequirements {
				if findControlId(profile, implReq.ControlID) {
					implReqs = append(implReqs, implReq)
				}
			}
			controlImplementations = append(controlImplementations, cd.ControlImplementation{
				UUID:                    controlImpl.UUID,
				Source:                  profile.Metadata.Title,
				Description:             controlImpl.Description,
				Props:                   controlImpl.Props,
				SetParameters:           controlImpl.SetParameters,
				ImplementedRequirements: implReqs,
			})
		}
		components = append(components, cd.Component{
			UUID:                   component.UUID,
			Type:                   component.Type,
			Title:                  component.Title,
			Description:            component.Description,
			Props:                  component.Props,
			ControlImplementations: controlImplementations,
		})
	}
	return cd.ComponentDefinition{
		UUID:       "cdfd629a-bd62-11ed-afa1-0242ac120002",
		Metadata:   compDef.Metadata,
		Components: components,
	}
}

func findControlId(profile oscal.Profile, controlId string) bool {
	for _, profileImport := range profile.Imports {
		for _, includeControl := range profileImport.IncludeControls {
			for _, id := range includeControl.WithIds {
				if id == controlId {
					return true
				}
			}
		}
	}
	return false
}

type OscalStandard struct {
	Id         string
	Categories []OscalCategory
}

type OscalCategory struct {
	Id         string
	ControlIds []string
}

func makeInternalOscalFormat(catalog oscal.Catalog, profile oscal.Profile) []OscalStandard {

	unknownGroup := oscal.Group{ID: "unknown", Title: "unknown", Class: "unknown"}

	oscalStandards := []OscalStandard{}
	for _, profileImport := range profile.Imports {
		controlIdsPerGroup := map[string][]string{}
		for _, includeControls := range profileImport.IncludeControls {
			for _, controlId := range includeControls.WithIds {
				group, ok := findControlGroup(controlId, catalog)
				if !ok {
					group = unknownGroup
				}
				controlIds, ok := controlIdsPerGroup[group.ID]
				if !ok {
					controlIds = []string{}
					controlIdsPerGroup[group.ID] = controlIds
				}
				controlIdsPerGroup[group.ID] = append(controlIds, controlId)
			}
		}
		oscalCategories := []OscalCategory{}
		for group, controlIds := range controlIdsPerGroup {
			oscalCategories = append(oscalCategories, OscalCategory{
				Id:         group,
				ControlIds: controlIds,
			})
		}
		oscalStandards = append(oscalStandards, OscalStandard{
			Id:         profileImport.Href,
			Categories: oscalCategories,
		})
	}
	return oscalStandards
}

func MakeInternalCompliance(catalog oscal.Catalog, profile oscal.Profile, cd cd.ComponentDefinition) internalcompliance.Compliance {

	unknownGroup := oscal.Group{ID: "unknown", Title: "unknown", Class: "unknown"}

	intCompliances := []internalcompliance.Compliance{}
	for _, component := range cd.Components {
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
					controlRefs = append(controlRefs, prop.Value)
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

func findControlGroup(controlId string, catalog oscal.Catalog) (oscal.Group, bool) {
	findCategoriesFromControls := func(controls []oscal.Control) bool {
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
	return oscal.Group{}, false
}
