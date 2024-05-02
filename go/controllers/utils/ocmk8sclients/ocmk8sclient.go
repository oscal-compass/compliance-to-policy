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
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"
)

var logger logr.Logger = ctrl.Log.WithName("ocmk8sclients")

type OcmK8ResourceInterfaceSetType struct {
	Policy           dynamic.NamespaceableResourceInterface
	PlacementRule    dynamic.NamespaceableResourceInterface
	PlacementBinding dynamic.NamespaceableResourceInterface
}

func NewOcmK8sClientSet(discoveryClient *discovery.DiscoveryClient, dyClient dynamic.Interface) (OcmK8ResourceInterfaceSetType, error) {
	var ocmK8sClientSet OcmK8ResourceInterfaceSetType
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		logger.Error(err, "failed to get APIGroupResource")
		return ocmK8sClientSet, err
	}
	ocmResourceGKs := []schema.GroupKind{{
		Group: "policy.open-cluster-management.io",
		Kind:  "Policy",
	}, {
		Group: "policy.open-cluster-management.io",
		Kind:  "PlacementBinding",
	}, {
		Group: "apps.open-cluster-management.io",
		Kind:  "PlacementRule",
	}}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	for _, gk := range ocmResourceGKs {
		mapping, err := restMapper.RESTMapping(gk, "v1")
		if err != nil {
			logger.Error(err, fmt.Sprintf("failed to get restMapping %s", gk.String()))
			return ocmK8sClientSet, err
		}
		if mapping == nil {
			logger.Error(err, fmt.Sprintf("restMapping is null %s", gk.String()))
			return ocmK8sClientSet, err
		}
		k8sNamespacedClient := dyClient.Resource(mapping.Resource)
		switch gk.Kind {
		case "Policy":
			ocmK8sClientSet.Policy = k8sNamespacedClient
		case "PlacementRule":
			ocmK8sClientSet.PlacementRule = k8sNamespacedClient
		case "PlacementBinding":
			ocmK8sClientSet.PlacementBinding = k8sNamespacedClient
		}
	}
	if ocmK8sClientSet.Policy == nil {
		return ocmK8sClientSet, errors.New("Policy Client is not be initialized.")
	}
	if ocmK8sClientSet.PlacementBinding == nil {
		return ocmK8sClientSet, errors.New("PlacementBinding Client is not be initialized.")
	}
	if ocmK8sClientSet.PlacementRule == nil {
		return ocmK8sClientSet, errors.New("PlacementRule Client is not be initialized.")
	}
	return ocmK8sClientSet, nil
}
