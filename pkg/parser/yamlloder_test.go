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

package parser

import (
	"testing"

	"github.com/IBM/compliance-to-policy/pkg/types/placements"
	"github.com/IBM/compliance-to-policy/pkg/types/policy"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func TestEmptyFileSystem(t *testing.T) {
	f, _ := static.Open("testdata/policy.yaml")
	objects, err := loadYaml(f)
	defer f.Close()
	if err != nil {
		t.Error(err)
	}
	k8sdec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	for _, object := range objects {
		data, err := yaml.Marshal(object)
		if err != nil {
			t.Error(err)
		}
		unst := &unstructured.Unstructured{}
		_, gvk, err := k8sdec.Decode(data, nil, unst)
		if err != nil {
			t.Error(err)
		}
		switch gvk.Kind {
		case "Policy":
			var policy policy.Policy
			if err := utilyaml.Unmarshal(data, &policy); err != nil {
				t.Error(err)
			}
			t.Log(policy)
		case "PlacementBinding":
			var pb placements.PlacementBinding
			if err := utilyaml.Unmarshal(data, &pb); err != nil {
				t.Error(err)
			}
			t.Log(pb)
		case "PlacementRule":
			var pr placements.PlacementRule
			if err := utilyaml.Unmarshal(data, &pr); err != nil {
				t.Error(err)
			}
			t.Log(pr)
		}
	}
}
