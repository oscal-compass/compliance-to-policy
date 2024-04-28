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

package ocmk8sclients

import (
	"context"

	"github.com/oscal-compass/compliance-to-policy/go/pkg"
	typespolicy "github.com/oscal-compass/compliance-to-policy/go/pkg/types/policy"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type policyClient struct {
	client dynamic.NamespaceableResourceInterface
}

func NewPolicyClient(client dynamic.NamespaceableResourceInterface) policyClient {
	return policyClient{
		client: client,
	}
}

func (c *policyClient) List(namespace string) ([]*typespolicy.Policy, error) {
	unstList, err := c.client.Namespace(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	typedList := []*typespolicy.Policy{}
	for _, unstPolicy := range unstList.Items {
		typedObj := typespolicy.Policy{}
		if err := pkg.ToK8sTypedObject(&unstPolicy, &typedObj); err != nil {
			return nil, err
		}
		typedList = append(typedList, &typedObj)
	}
	return typedList, nil
}

func (c *policyClient) Get(namespace string, name string) (*typespolicy.Policy, error) {
	unstObj, err := c.client.Namespace(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	typedObj := typespolicy.Policy{}
	if err := pkg.ToK8sTypedObject(unstObj, &typedObj); err != nil {
		return nil, err
	}
	return &typedObj, nil
}

func (c *policyClient) Create(namespace string, typedObj typespolicy.Policy) (*typespolicy.Policy, error) {
	unstObj, err := pkg.ToK8sUnstructedObject(&typedObj)
	if err != nil {
		return nil, err
	}
	_unstObj, err := c.client.Namespace(namespace).Create(context.TODO(), &unstObj, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	_typedObj := typespolicy.Policy{}
	if err := pkg.ToK8sTypedObject(_unstObj, &_typedObj); err != nil {
		return nil, err
	}
	return &_typedObj, nil
}

func (c *policyClient) Delete(namespace string, name string) error {
	return c.client.Namespace(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
}
