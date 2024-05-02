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

package policygenerator

import (
	"fmt"
	"testing"

	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/types"
)

func TestKustomize(t *testing.T) {
	testDataDir := pkg.PathFromPkgDirectory("./policygenerator/testdata")
	pflagOpts := krusty.MakeDefaultOptions()
	pflagOpts.PluginConfig = types.EnabledPluginConfig(types.BploUseStaticallyLinked)
	tests := []struct {
		path string
		opts *krusty.Options
	}{
		{testDataDir + "/input-kustomize", krusty.MakeDefaultOptions()},
		{testDataDir + "/", pflagOpts},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("kustomize %s", test.path), func(t *testing.T) {
			manifests, err := kustomizeWithOpts(test.path, test.opts)
			if err != nil {
				t.Error(err)
			}
			yml, err := manifests.AsYaml()
			if err != nil {
				t.Error(err)
			}
			t.Log(string(yml))
		})
	}
}
