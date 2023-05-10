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

package kcpclient

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	apix "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var logger logr.Logger = ctrl.Log.WithName("kcpclient")

type KcpClient struct {
	K8sClient       client.Client
	dyClient        dynamic.Interface
	discoveryClient *discovery.DiscoveryClient
	apixClientSet   *apix.Clientset
}

func NewKcpClient(cfg rest.Config, workspace string) (KcpClient, error) {

	var kcpClient KcpClient

	cfg.Burst = 1024
	hosts := strings.Split(cfg.Host, "/")
	hosts[len(hosts)-1] = workspace
	cfg.Host = strings.Join(hosts, "/")

	k8sClient, err := client.New(&cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		return kcpClient, err
	}
	apixClientSet, err := apix.NewForConfig(&cfg)
	if err != nil {
		return kcpClient, err
	}

	dyClient, err := dynamic.NewForConfig(&cfg)
	if err != nil {
		return kcpClient, err
	}

	kcpClient = KcpClient{
		K8sClient:       k8sClient,
		dyClient:        dyClient,
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(&cfg),
		apixClientSet:   apixClientSet,
	}
	return kcpClient, nil
}

func (c *KcpClient) GetDyClientWithScopeInfo(group string, kind string, version string) (dynamic.NamespaceableResourceInterface, bool, error) {
	var client dynamic.NamespaceableResourceInterface
	mapping, err := c.GetMapping(group, version, kind)
	if err != nil {
		return client, false, err
	}
	client = c.dyClient.Resource(mapping.Resource)
	return client, mapping.Scope == meta.RESTScopeNamespace, nil
}

func (c *KcpClient) GetDyClient(group string, kind string, version string) (dynamic.NamespaceableResourceInterface, error) {
	client, _, err := c.GetDyClientWithScopeInfo(group, kind, version)
	return client, err
}

func (c *KcpClient) GetMapping(group string, version string, kind string) (*meta.RESTMapping, error) {
	gk := schema.GroupKind{
		Group: group,
		Kind:  kind,
	}
	groupResources, err := restmapper.GetAPIGroupResources(c.discoveryClient)
	if err != nil {
		logger.Error(err, "failed to get APIGroupResource")
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
	mapping, err := restMapper.RESTMapping(gk, version)
	if err != nil {
		logger.Error(err, fmt.Sprintf("failed to get restMapping %s", gk.String()))
		return nil, err
	}
	if mapping == nil {
		logger.Error(err, fmt.Sprintf("failed to get restMapping %s", gk.String()))
		return nil, err
	}
	return mapping, nil
}
