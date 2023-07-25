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
	"github.com/IBM/compliance-to-policy/pkg/types/policygenerator"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/api/resmap"
	kustomizetypes "sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type PolicyGeneratorManifestClaim struct {
	Namespace        string
	Standards        []string
	Categories       []string
	Controls         []string
	Policies         []policygenerator.PolicyConfig
	ClusterSelectors map[string]string
}

func Kustomize(path string) (resmap.ResMap, error) {
	pflagOpts := krusty.MakeDefaultOptions()
	pflagOpts.PluginConfig = kustomizetypes.EnabledPluginConfig(kustomizetypes.BploUseStaticallyLinked)
	return kustomizeWithOpts(path, pflagOpts)
}

func kustomizeWithOpts(path string, opts *krusty.Options) (resmap.ResMap, error) {
	fs := filesys.MakeFsOnDisk()
	k := krusty.MakeKustomizer(opts)
	return k.Run(fs, path)
}

func GeneratePolicyGeneratorManifest(policyGeneratorManifestClaim PolicyGeneratorManifestClaim) policygenerator.PolicyGenerator {
	standards := policyGeneratorManifestClaim.Standards
	categories := policyGeneratorManifestClaim.Categories
	controls := policyGeneratorManifestClaim.Controls
	policyDefault := policygenerator.PolicyDefaults{
		Namespace: policyGeneratorManifestClaim.Namespace,
		PolicyOptions: policygenerator.PolicyOptions{
			Standards:                standards,
			Controls:                 controls,
			Categories:               categories,
			InformKyvernoPolicies:    false,
			InformGatekeeperPolicies: false,
			GeneratePolicyPlacement:  true,
			Placement: policygenerator.PlacementConfig{
				ClusterSelectors: policyGeneratorManifestClaim.ClusterSelectors,
			},
		},
		ConfigurationPolicyOptions: policygenerator.ConfigurationPolicyOptions{
			RemediationAction: "inform",
			Severity:          "low",
			ComplianceType:    "musthave",
			NamespaceSelector: policygenerator.NamespaceSelector{
				Exclude: []string{"kube-*"},
			},
		},
	}
	return BuildPolicyGeneratorManifest("policy", policyDefault, policyGeneratorManifestClaim.Policies)
}

func BuildPolicyGeneratorManifest(policyGeneratorName string, policyDefault policygenerator.PolicyDefaults, policies []policygenerator.PolicyConfig) policygenerator.PolicyGenerator {
	return policygenerator.PolicyGenerator{
		APIVersion: "policy.open-cluster-management.io/v1",
		Kind:       "PolicyGenerator",
		Metadata: struct {
			Name string "json:\"name,omitempty\" yaml:\"name,omitempty\""
		}{policyGeneratorName},
		PolicyDefaults: policyDefault,
		Policies:       policies,
	}
}
