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

package reporter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/IBM/compliance-to-policy/pkg"
	typeplacementdecision "github.com/IBM/compliance-to-policy/pkg/types/placementdecision"
	typepolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	"github.com/stretchr/testify/assert"
	typepolr "sigs.k8s.io/wg-policy-prototypes/policy-report/pkg/api/wgpolicyk8s.io/v1beta1"
)

func TestConvertToPolicyReport(t *testing.T) {
	policies := []typepolicy.Policy{}
	traverseFunc := genTraverseFunc(
		func(policy typepolicy.Policy) { policies = append(policies, policy) },
		func(policySet typepolicy.PolicySet) {},
		func(placementDecision typeplacementdecision.PlacementDecision) {},
	)
	err := filepath.Walk(pkg.PathFromPkgDirectory("./testdata/ocm/policy-results"), traverseFunc)
	assert.NoError(t, err, "Should not happen")

	tempDirPath := pkg.PathFromPkgDirectory("./testdata/_test")
	err = os.MkdirAll(tempDirPath, os.ModePerm)
	assert.NoError(t, err, "Should not happen")
	tempDir := pkg.NewTempDirectory(tempDirPath)

	for _, policy := range policies {
		policyReport := ConvertToPolicyReport(policy)
		assert.NotNil(t, policyReport, "Should not happen")
		fname := fmt.Sprintf("policy-report.%s.%s.yaml", policy.Namespace, policy.Name)
		err = pkg.WriteObjToYamlFile(tempDir.GetTempDir()+"/"+fname, policyReport)
		assert.NoError(t, err, "Should not happen")

		var expected typepolr.PolicyReport
		err = pkg.LoadYamlFileToK8sTypedObject(pkg.PathFromPkgDirectory("./testdata/reports")+"/"+fname, &expected)
		assert.NoError(t, err, "Should not happen")

		// Workaround: I'm not sure why empty map field is treated as nil for loaded object. Explicitly set as empty map.
		for idx := range expected.Results {
			if expected.Results[idx].Properties == nil {
				expected.Results[idx].Properties = map[string]string{}
			}
		}
		// Timestamp is currently set by Now(). Since the timestamp should be always different from expected one, reset creationTimestamp of expected one to actual one.
		expected.CreationTimestamp = policyReport.CreationTimestamp
		assert.Equal(t, expected, policyReport)
	}
}
