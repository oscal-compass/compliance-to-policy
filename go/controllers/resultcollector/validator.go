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

package resultcollector

import (
	"context"
	"fmt"
	"strings"

	c2pv1alpha1 "github.com/oscal-compass/compliance-to-policy/go/api/v1alpha1"
	"github.com/oscal-compass/compliance-to-policy/go/controllers/utils/kcpclient"
	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func validate(ctx context.Context, policyValidationRequests []c2pv1alpha1.PolicyValidationRequest, kcpClient kcpclient.KcpClient) ([]PolicyValidationResult, error) {
	policyValidationResults := []PolicyValidationResult{}
	for _, validationRequest := range policyValidationRequests {
		policyValidationResult := PolicyValidationResult{policyId: validationRequest.PolicyId}
		checkPolicyResults := []CheckPolicyResult{}
		policyId := validationRequest.PolicyId
		checkPolicies := validationRequest.CheckPolicies
		for _, checkPolicy := range checkPolicies {
			checkPolicyResult := CheckPolicyResult{policyId: policyId, checkPolicy: checkPolicy, testResults: []CheckPolicyTestResult{}}
			for _, objectTemplate := range checkPolicy.Spec.ObjectTemplates {
				var unstObjDef unstructured.Unstructured
				err := pkg.LoadByteToK8sTypedObject(objectTemplate.ObjectDefinition.Raw, &unstObjDef)
				if err != nil {
					logger.Error(err, "Failed to load ObjectDefinition.Raw.")
					return nil, err
				}
				kind := unstObjDef.GetKind()
				apiVersion := unstObjDef.GetAPIVersion()
				group := strings.Split(apiVersion, "/")[0]
				version := strings.Split(apiVersion, "/")[1]
				name := unstObjDef.GetName()
				namespace := unstObjDef.GetNamespace()

				dyClient, namespaced, err := kcpClient.GetDyClientWithScopeInfo(group, kind, version)
				if err != nil {
					logger.Error(err, "Failed to get Client.")
					return nil, err
				}
				if namespaced {
					if namespace == "*" {
						dyNsClient, err := kcpClient.GetDyClient("", "Namespace", "v1")
						if err != nil {
							logger.Error(err, "Failed to get namespace dyClient.")
							return nil, err
						}
						nsUnstList, err := dyNsClient.List(ctx, v1.ListOptions{})
						if err != nil {
							logger.Error(err, "Failed to list namespace.")
							return nil, err
						}
						for _, ns := range nsUnstList.Items {
							_unstObjDef := unstObjDef.DeepCopy()
							_unstObjDef.SetNamespace(ns.GetName()) // Replace namespace for showing exact namespace instead of "*" in report
							testResults := validateObject(ctx, name, *_unstObjDef, dyClient.Namespace(ns.GetName()))
							checkPolicyResult.testResults = append(checkPolicyResult.testResults, testResults...)
						}
					} else {
						testResults := validateObject(ctx, name, unstObjDef, dyClient.Namespace(namespace))
						checkPolicyResult.testResults = append(checkPolicyResult.testResults, testResults...)
					}
				} else {
					testResults := validateObject(ctx, name, unstObjDef, dyClient)
					checkPolicyResult.testResults = append(checkPolicyResult.testResults, testResults...)
				}
			}
			checkPolicyResults = append(checkPolicyResults, checkPolicyResult)
		}
		policyValidationResult.checkPolicyResults = checkPolicyResults
		policyValidationResults = append(policyValidationResults, policyValidationResult)
	}
	return policyValidationResults, nil
}

func validateObject(ctx context.Context, name string, objectDefinition unstructured.Unstructured, dyClient dynamic.ResourceInterface) []CheckPolicyTestResult {
	testResults := []CheckPolicyTestResult{}
	if name == "*" {
		unstList, err := dyClient.List(ctx, v1.ListOptions{})
		if err != nil {
			logger.Error(err, "Failed to list objects.")
			objectDefinition.SetName(name) // Replace name for showing exact name instead of "*" in report
			message := fmt.Sprintf("CheckPolicy errored for %s/%s. %v", objectDefinition.GetNamespace(), objectDefinition.GetName(), err)
			testResult := &CheckPolicyTestResult{pass: false, error: err, objectDefinition: objectDefinition, message: message}
			testResults = append(testResults, *testResult)
			return testResults
		}
		for _, unstObj := range unstList.Items {
			testResult := compareObject(objectDefinition, unstObj)
			if testResult != nil {
				testResults = append(testResults, *testResult)
			}
		}
	} else {
		unstObj, err := dyClient.Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				message := fmt.Sprintf("CheckPolicy is skipped for %s/%s due to nothing report", objectDefinition.GetNamespace(), objectDefinition.GetName())
				logger.Info(message)
				testResult := &CheckPolicyTestResult{pass: true, error: nil, objectDefinition: objectDefinition, message: message}
				testResults = append(testResults, *testResult)
				return testResults
			} else {
				logger.Error(err, "Failed to list objects.")
				message := fmt.Sprintf("CheckPolicy errored for %s/%s. %v", objectDefinition.GetNamespace(), objectDefinition.GetName(), err)
				testResult := &CheckPolicyTestResult{pass: false, error: err, objectDefinition: objectDefinition, message: message}
				testResults = append(testResults, *testResult)
				return testResults
			}
		}
		testResult := compareObject(objectDefinition, *unstObj)
		if testResult != nil {
			testResults = append(testResults, *testResult)
		}
	}
	return testResults
}

func compareObject(objectDefinition unstructured.Unstructured, fetchedObject unstructured.Unstructured) *CheckPolicyTestResult {
	expected, ok, err := unstructured.NestedInt64(objectDefinition.Object, "summary", "error")
	if !ok || err != nil {
		logger.Info("No summary error such field")
	} else {
		actual, ok, err := unstructured.NestedInt64(fetchedObject.Object, "summary", "error")
		if !ok || err != nil {
			message := fmt.Sprintf("CheckPolicy failed for %s/%s since the expected field is not found.", objectDefinition.GetNamespace(), objectDefinition.GetName())
			return &CheckPolicyTestResult{pass: false, objectDefinition: objectDefinition, message: message}
		} else {
			pass := expected == actual
			var message string
			if pass {
				message = fmt.Sprintf("CheckPolicy passed for %s/%s", objectDefinition.GetNamespace(), objectDefinition.GetName())
			} else {
				message = fmt.Sprintf("CheckPolicy failed for %s/%s since the expected field is not found.", objectDefinition.GetNamespace(), objectDefinition.GetName())
			}
			return &CheckPolicyTestResult{pass: pass, objectDefinition: objectDefinition, message: message}
		}
	}
	expected, ok, err = unstructured.NestedInt64(objectDefinition.Object, "status", "readyReplicas")
	if !ok || err != nil {
		logger.Info("No status field")
	} else {
		actual, ok, err := unstructured.NestedInt64(fetchedObject.Object, "status", "readyReplicas")
		if !ok || err != nil {
			message := fmt.Sprintf("CheckPolicy failed for %s/%s since the expected field is not found.", objectDefinition.GetNamespace(), objectDefinition.GetName())
			return &CheckPolicyTestResult{pass: false, objectDefinition: objectDefinition, message: message}
		} else {
			pass := expected == actual
			var message string
			if pass {
				message = fmt.Sprintf("CheckPolicy passed for %s/%s", objectDefinition.GetNamespace(), objectDefinition.GetName())
			} else {
				message = fmt.Sprintf("CheckPolicy failed for %s/%s", objectDefinition.GetNamespace(), objectDefinition.GetName())
			}
			return &CheckPolicyTestResult{pass: pass, objectDefinition: objectDefinition, message: message}
		}
	}
	return nil
}
