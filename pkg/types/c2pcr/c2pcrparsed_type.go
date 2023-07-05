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

package c2pcr

import (
	"github.com/IBM/compliance-to-policy/pkg/oscal"
	typesoscal "github.com/IBM/compliance-to-policy/pkg/types/oscal"
	typecd "github.com/IBM/compliance-to-policy/pkg/types/oscal/componentdefinition"
)

type C2PCRParsed struct {
	Namespace           string
	PolicyResoureDir    string
	Catalog             typesoscal.CatalogRoot
	Profile             typesoscal.ProfileRoot
	ComponentDefinition typecd.ComponentDefinitionRoot
	ComponentObjects    []oscal.ComponentObject
	ClusterSelectors    map[string]string
}
