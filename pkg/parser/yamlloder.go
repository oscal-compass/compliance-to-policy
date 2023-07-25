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
	"errors"
	"io"

	"github.com/IBM/compliance-to-policy/pkg/types/placements"
	"github.com/IBM/compliance-to-policy/pkg/types/policy"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	k8syaml "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

func loadYaml(r io.Reader) ([]interface{}, error) {
	dec := yaml.NewDecoder(r)

	var objects []interface{}
	for {
		var objInput interface{}
		err := dec.Decode(&objInput)
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return objects, err
		} else if objInput == nil {
			continue
		}
		objects = append(objects, &objInput)
	}
	return objects, nil
}

func loadAndUnmarshal(r io.Reader) ([]*policy.Policy, []*unstructured.Unstructured, []*unstructured.Unstructured, error) {
	var policies []*policy.Policy
	var pbs []*unstructured.Unstructured
	var prs []*unstructured.Unstructured
	k8sdec := k8syaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	objects, err := loadYaml(r)
	if err != nil {
		return policies, pbs, prs, err
	}
	for _, object := range objects {
		data, err := yaml.Marshal(object)
		if err != nil {
			logger.Sugar().Errorf("%v", err)
		}
		unst := &unstructured.Unstructured{}
		_, gvk, err := k8sdec.Decode(data, nil, unst)
		if err != nil {
			logger.Sugar().Errorf("%v", err)
		}
		switch gvk.Kind {
		case "Policy":
			var policy policy.Policy
			if err := utilyaml.Unmarshal(data, &policy); err != nil {
				logger.Sugar().Errorf("%v", err)
			}
			policies = append(policies, &policy)
		case "PlacementBinding":
			var pb placements.PlacementBinding
			if err := utilyaml.Unmarshal(data, &pb); err != nil {
				logger.Sugar().Errorf("%v", err)
			}
			pbs = append(pbs, unst)
		case "PlacementRule":
			var pr placements.PlacementRule
			if err := utilyaml.Unmarshal(data, &pr); err != nil {
				logger.Sugar().Errorf("%v", err)
			}
			prs = append(prs, unst)
		}
	}
	return policies, pbs, prs, nil
}
