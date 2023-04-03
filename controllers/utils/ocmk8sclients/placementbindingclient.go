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

	"github.com/IBM/compliance-to-policy/pkg"
	typesplacement "github.com/IBM/compliance-to-policy/pkg/types/placements"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type placementBindingClient struct {
	client dynamic.NamespaceableResourceInterface
}

func NewPlacementBindingClient(client dynamic.NamespaceableResourceInterface) placementBindingClient {
	return placementBindingClient{
		client: client,
	}
}

func (c *placementBindingClient) List(namespace string) ([]*typesplacement.PlacementBinding, error) {
	unstList, err := c.client.Namespace(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}
	typedList := []*typesplacement.PlacementBinding{}
	for _, unstPolicy := range unstList.Items {
		typedObj := typesplacement.PlacementBinding{}
		if err := pkg.ToK8sTypedObject(&unstPolicy, &typedObj); err != nil {
			return nil, err
		}
		typedList = append(typedList, &typedObj)
	}
	return typedList, nil
}

func (c *placementBindingClient) Get(namespace string, name string) (*typesplacement.PlacementBinding, error) {
	unstObj, err := c.client.Namespace(namespace).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	typedObj := typesplacement.PlacementBinding{}
	if err := pkg.ToK8sTypedObject(unstObj, &typedObj); err != nil {
		return nil, err
	}
	return &typedObj, nil
}

func (c *placementBindingClient) Create(namespace string, typedObj typesplacement.PlacementBinding) (*typesplacement.PlacementBinding, error) {
	unstObj, err := pkg.ToK8sUnstructedObject(&typedObj)
	if err != nil {
		return nil, err
	}
	_unstObj, err := c.client.Namespace(namespace).Create(context.TODO(), &unstObj, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	_typedObj := typesplacement.PlacementBinding{}
	if err := pkg.ToK8sTypedObject(_unstObj, &_typedObj); err != nil {
		return nil, err
	}
	return &_typedObj, nil
}

func (c *placementBindingClient) Delete(namespace string, name string) error {
	return c.client.Namespace(namespace).Delete(context.TODO(), name, v1.DeleteOptions{})
}
