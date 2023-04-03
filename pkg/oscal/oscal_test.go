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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/tables/resources"
	"github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	"github.com/IBM/compliance-to-policy/pkg/types/oscal"
	cd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := os.MkdirAll(pkg.PathFromPkgDirectory("./oscal/_test"), os.ModePerm); err != nil {
		panic(err)
	}
}
func TestOscalCD(t *testing.T) {
	f, err := os.Open(pkg.PathFromPkgDirectory("./oscal/testdata/resources.csv"))
	if err != nil {
		panic(err)
	}
	resourceTable := resources.FromCsv(f)
	compDef := makeComponentDefinitionFromMasterData(resourceTable)
	t.Log(compDef)

	compDefRoot := cd.ComponentDefinitionRoot{ComponentDefinition: *compDef}
	odir := pkg.PathFromPkgDirectory("./oscal/_test")
	err = pkg.WriteObjToJsonFile(odir+"/component-definition.json", compDefRoot)
	assert.NoError(t, err, "Should be no error")
}

func TestOscalCDWithProfile(t *testing.T) {
	var compDefRoot cd.ComponentDefinitionRoot
	err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory("./oscal/testdata/component-definition.json"), &compDefRoot)
	assert.NoError(t, err, "Should be no error")

	profileResources := map[string]string{
		"moderate": "./oscal/testdata/NIST_SP-800-53_rev5_MODERATE-baseline_profile.json",
		"high":     "./oscal/testdata/NIST_SP-800-53_rev5_HIGH-baseline_profile.json",
		"low":      "./oscal/testdata/NIST_SP-800-53_rev5_LOW-baseline_profile.json",
	}
	for level, file := range profileResources {
		var profileRoot oscal.ProfileRoot
		err = pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory(file), &profileRoot)
		assert.NoError(t, err, "Should be no error")

		newCompDefRoot := cd.ComponentDefinitionRoot{
			ComponentDefinition: IntersectProfileWithCD(compDefRoot.ComponentDefinition, profileRoot.Profile),
		}
		odir := pkg.PathFromPkgDirectory("./oscal/_test")
		err = pkg.WriteObjToJsonFile(fmt.Sprintf("%s/component-definition.%s.json", odir, level), newCompDefRoot)
		assert.NoError(t, err, "Should be no error")
	}
}

func TestCD2InternalCompliance(t *testing.T) {
	cdjson, err := os.ReadFile(pkg.PathFromPkgDirectory("./oscal/testdata/component-definition.json"))
	if err != nil {
		panic(err)
	}

	cdobj := cd.ComponentDefinitionRoot{}
	if err := json.Unmarshal(cdjson, &cdobj); err != nil {
		panic(err)
	}

	intCompliance := makeInternalComplianceFromComponentDefinition(cdobj.ComponentDefinition)
	if err := pkg.MakeDirAndWriteObjToYamlFile(
		pkg.PathFromPkgDirectory("./oscal/testdata/_test"), "internal-compliance-from-cd.yaml", intCompliance,
	); err != nil {
		panic(err)
	}
}

func TestTrestleCsv(t *testing.T) {
	f, err := os.Open(pkg.PathFromPkgDirectory("./oscal/testdata/resources.csv"))
	if err != nil {
		panic(err)
	}
	resourceTable := resources.FromCsv(f)
	componentProps := TrestleComponentProps{
		ComponentTitle:       "OCM",
		ComponentDescription: "Description ...",
		ComponentType:        "Service",
	}
	namespace := "https://github.com/IBM/compliance-to-policy"
	trestlCsvRowsMap := makeTrestleCsvFromMasterData(componentProps, namespace, resourceTable)
	t.Log(trestlCsvRowsMap)

	for standard, trestlCsvRows := range trestlCsvRowsMap {
		odir := pkg.PathFromPkgDirectory("./oscal/_test")
		_standard := strings.Replace(standard, "/", "_", -1)
		_standard = strings.Replace(_standard, " ", "_", -1)
		csvFile, err := os.Create(fmt.Sprintf("%s/trestle.rule-mapping.%s.csv", odir, _standard))
		assert.NoError(t, err, "Should be no error")

		wr := csv.NewWriter(csvFile)
		err = wr.Write(trestlCsvRows[0].Header())
		assert.NoError(t, err, "Should be no error")
		for _, row := range trestlCsvRows {
			err := wr.Write(row.ToStringList())
			assert.NoError(t, err, "Should be no error")
		}
		wr.Flush()
	}
}

func makeInternalComplianceFromComponentDefinition(cd cd.ComponentDefinition) internalcompliance.Compliance {

	controls := []internalcompliance.Control{}
	for _, controlImpl := range cd.Components[0].ControlImplementations {
		for _, implReq := range controlImpl.ImplementedRequirements {
			controlRefs := []string{}
			ruleProps := listRules(implReq.Props)
			for _, prop := range ruleProps {
				controlRefs = append(controlRefs, prop.Value)
			}
			controls = append(controls, internalcompliance.Control{
				Name:        implReq.ControlID,
				ControlRefs: controlRefs,
			})
		}
	}
	standard := internalcompliance.Standard{
		Categories: []internalcompliance.Category{{
			Name:     "category",
			Controls: controls,
		}},
	}
	intCompliance := internalcompliance.Compliance{
		Standard: standard,
	}
	return intCompliance
}

func TestOscalToIntCompliance(t *testing.T) {
	catalogObj := oscal.CatalogRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory("./oscal/testdata/NIST_SP-800-53_rev5_catalog.json"), &catalogObj); err != nil {
		panic(err)
	}
	profileObj := oscal.ProfileRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory("./oscal/testdata/NIST_SP-800-53_rev5_HIGH-baseline_profile.json"), &profileObj); err != nil {
		panic(err)
	}
	cdObj := cd.ComponentDefinitionRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory("./oscal/testdata/component-definition.json"), &cdObj); err != nil {
		panic(err)
	}

	intCompliance := MakeInternalCompliance(catalogObj.Catalog, profileObj.Profile, cdObj.ComponentDefinition)
	if err := pkg.MakeDirAndWriteObjToYamlFile(
		pkg.PathFromPkgDirectory("./oscal/testdata/_test"), "internal-compliance.yaml", intCompliance,
	); err != nil {
		panic(err)
	}
}

func TestOscalToIntOscalFormat(t *testing.T) {
	oscalToIntOscalFormat("./oscal/testdata/NIST_SP-800-53_rev5_catalog.json", "./oscal/testdata/NIST_SP-800-53_rev5_HIGH-baseline_profile.json", "interna-oscal-standards-high.yaml")
	oscalToIntOscalFormat("./oscal/testdata/NIST_SP-800-53_rev5_catalog.json", "./oscal/testdata/NIST_SP-800-53_rev5_LOW-baseline_profile.json", "interna-oscal-standards-low.yaml")
	oscalToIntOscalFormat("./oscal/testdata/NIST_SP-800-53_rev5_catalog.json", "./oscal/testdata/NIST_SP-800-53_rev5_MODERATE-baseline_profile.json", "interna-oscal-standards-moderate.yaml")
}

func oscalToIntOscalFormat(catalog string, profile string, resultFilename string) {
	catalogObj := oscal.CatalogRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory(catalog), &catalogObj); err != nil {
		panic(err)
	}
	profileObj := oscal.ProfileRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory(profile), &profileObj); err != nil {
		panic(err)
	}
	cdObj := cd.ComponentDefinitionRoot{}
	if err := pkg.LoadJsonFileToObject(pkg.PathFromPkgDirectory("./oscal/testdata/component-definition.json"), &cdObj); err != nil {
		panic(err)
	}

	intOscalStandards := makeInternalOscalFormat(catalogObj.Catalog, profileObj.Profile)
	if err := pkg.MakeDirAndWriteObjToYamlFile(
		pkg.PathFromPkgDirectory("./oscal/testdata/_test"), resultFilename, intOscalStandards,
	); err != nil {
		panic(err)
	}
}
