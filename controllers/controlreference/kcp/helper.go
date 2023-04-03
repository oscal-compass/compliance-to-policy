package kcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	c2pv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	edgev1alpha1 "github.com/IBM/compliance-to-policy/controllers/edge.kcp.io/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils/kcpclient"
	"github.com/IBM/compliance-to-policy/pkg"
	kcpv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/apis/v1alpha1"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var NAMESPACE_LABEL_KEY string = "c2p"
var NAMESPACE_LABEL_VALUE string = "c2p"

var _ c2pv1alpha1.Compliance

func applyC2PlabelToNamespaces(resource *unstructured.Unstructured) {
	if resource.GetKind() == "Namespace" {
		labels := resource.GetLabels()
		labels[NAMESPACE_LABEL_KEY] = NAMESPACE_LABEL_VALUE
		resource.SetLabels(labels)
	}
}

func generateEdgePlacement(ctx context.Context, name string, kcpWmwClient *kcpclient.KcpClient, resourcesByPolicy map[string][]unstructured.Unstructured, checkPoliciesByPolicy map[string][]c2pv1alpha1.CheckPolicy, matchLabel map[string]string) (edgev1alpha1.EdgePlacement, error) {
	logger := log.FromContext(ctx)

	namespaceLableSelector := map[string]string{NAMESPACE_LABEL_KEY: NAMESPACE_LABEL_VALUE}

	var edgePlacement edgev1alpha1.EdgePlacement
	edgePlacement.TypeMeta = metav1.TypeMeta{
		Kind:       "EdgePlacement",
		APIVersion: "edge.kcp.io/v1alpha1",
	}
	edgePlacement.ObjectMeta = metav1.ObjectMeta{
		Name: name,
	}
	edgePlacement.Spec.LocationSelectors = []metav1.LabelSelector{
		{
			MatchLabels: matchLabel,
		},
	}
	edgePlacement.Spec.NamespaceSelector = metav1.LabelSelector{
		MatchLabels: namespaceLableSelector,
	}

	// nonNamespace
	nnorsList, err := makeNonNamespacedObjectReferenceSets(ctx, kcpWmwClient, resourcesByPolicy)
	if err != nil {
		logger.Error(err, "Failed to make nonNamespacedObjectReferenceSets")
		return edgePlacement, err
	}
	edgePlacement.Spec.NonNamespacedObjects = nnorsList

	// upsync
	upsyncSets, err := makeUpsyncSets(ctx, kcpWmwClient, resourcesByPolicy, checkPoliciesByPolicy)
	if err != nil {
		logger.Error(err, "Failed to make upsyncSets")
		return edgePlacement, err
	}
	edgePlacement.Spec.Upsync = upsyncSets
	return edgePlacement, nil
}

func makeNonNamespacedObjectReferenceSets(ctx context.Context, kcpWmwClient *kcpclient.KcpClient, resourcesByPolicy map[string][]unstructured.Unstructured) ([]edgev1alpha1.NonNamespacedObjectReferenceSet, error) {
	type GVRN struct {
		gvr  schema.GroupVersionResource
		name string
	}
	gvrnList := []GVRN{}
	mappings := map[schema.GroupVersionKind]*meta.RESTMapping{}
	for _, resources := range resourcesByPolicy {
		for _, resource := range resources {
			gvk := resource.GroupVersionKind()
			if gvk.Kind == "Namespace" {
				continue
			}
			_, ok := mappings[gvk]
			if !ok {
				mapping, err := kcpWmwClient.GetMapping(gvk.Group, gvk.Version, gvk.Kind)
				if err != nil {
					logger.Error(err, "Failed to get gvk-gvr mapping")
					return nil, err
				}
				mappings[gvk] = mapping
			}
			if mappings[gvk].Scope == meta.RESTScopeRoot {
				gvr := mappings[gvk].Resource
				gvrnList = append(gvrnList, GVRN{
					gvr:  gvr,
					name: resource.GetName(),
				})
			}
		}
	}
	nnorsList := []edgev1alpha1.NonNamespacedObjectReferenceSet{}
	gvrnGroupedByGroup := map[string][]GVRN{}
	for _, gvrn := range gvrnList {
		_, ok := gvrnGroupedByGroup[gvrn.gvr.Group]
		if !ok {
			gvrnGroupedByGroup[gvrn.gvr.Group] = []GVRN{gvrn}
		} else {
			gvrnGroupedByGroup[gvrn.gvr.Group] = append(gvrnGroupedByGroup[gvrn.gvr.Group], gvrn)
		}
	}
	for group, gvrnList := range gvrnGroupedByGroup {
		resources := sets.String{}
		resourceNames := sets.String{}
		for _, gvrn := range gvrnList {
			resources.Insert(gvrn.gvr.Resource)
			resourceNames.Insert(gvrn.name)
		}
		nnors := edgev1alpha1.NonNamespacedObjectReferenceSet{
			APIGroup:      group,
			Resources:     resources.List(),
			ResourceNames: resourceNames.List(),
		}
		nnorsList = append(nnorsList, nnors)
	}

	nnorsList = append(nnorsList, edgev1alpha1.NonNamespacedObjectReferenceSet{
		APIGroup:      "apis.kcp.io",
		Resources:     []string{"apibindings"},
		ResourceNames: []string{"bind-kube"},
	})
	return nnorsList, nil
}

func makeUpsyncSets(ctx context.Context, kcpWmwClient *kcpclient.KcpClient, resourcesByPolicy map[string][]unstructured.Unstructured, checkPoliciesByPolicy map[string][]c2pv1alpha1.CheckPolicy) ([]edgev1alpha1.UpsyncSet, error) {
	upsyncSets := []edgev1alpha1.UpsyncSet{}
	type NN struct {
		names      sets.String
		namespaces sets.String
	}
	nnPerGVK := map[schema.GroupVersionKind]NN{}
	for policy, resources := range resourcesByPolicy {
		checkPolicies, ok := checkPoliciesByPolicy[policy]
		if ok {
			for _, checkPolicy := range checkPolicies {
				for _, objectTemplate := range checkPolicy.Spec.ObjectTemplates {
					var unstObj unstructured.Unstructured
					err := pkg.LoadByteToK8sTypedObject(objectTemplate.ObjectDefinition.Raw, &unstObj)
					if err != nil {
						logger.Error(err, "Failed to load ObjectDefinition.Raw.")
						return nil, err
					}
					kind := unstObj.GetKind()
					apiVersion := unstObj.GetAPIVersion()
					group := strings.Split(apiVersion, "/")[0]
					version := strings.Split(apiVersion, "/")[1]
					name := unstObj.GetName()
					namespace := unstObj.GetNamespace()
					found := false
					for _, resource := range resources {
						if resource.GetKind() == kind && resource.GetAPIVersion() == apiVersion && resource.GetNamespace() == namespace && resource.GetName() == name {
							logger.Info("It's part of downsynced resources. No need to be added in upsync.")
							found = true
							break
						}
					}
					if !found {
						gvk := schema.GroupVersionKind{Group: group, Version: version, Kind: kind}
						_, ok := nnPerGVK[gvk]
						if !ok {
							nnPerGVK[gvk] = NN{
								names:      sets.NewString(name),
								namespaces: sets.NewString(namespace),
							}
						} else {
							nnPerGVK[gvk].names.Insert(name)
							nnPerGVK[gvk].namespaces.Insert(namespace)
						}
					}
				}
			}
		}
	}
	type RNN struct {
		resources  sets.String
		names      sets.String
		namespaces sets.String
	}
	rnnPerGroup := map[string]RNN{}
	for gvk, nn := range nnPerGVK {
		mapping, err := kcpWmwClient.GetMapping(gvk.Group, gvk.Version, gvk.Kind)
		if err != nil {
			logger.Error(err, "Failed to get gvk-gvr mapping")
			return nil, err
		}
		resource := mapping.Resource
		_, ok := rnnPerGroup[gvk.Group]
		if !ok {
			rnnPerGroup[gvk.Group] = RNN{
				resources:  sets.NewString(resource.Resource),
				names:      nn.names,
				namespaces: nn.namespaces,
			}
		} else {
			rnnPerGroup[gvk.Group].resources.Insert(resource.Resource)
			rnnPerGroup[gvk.Group].names.Insert(nn.names.List()...)
			rnnPerGroup[gvk.Group].namespaces.Insert(nn.namespaces.List()...)
		}
	}
	for group, rnn := range rnnPerGroup {
		upsyncSet := edgev1alpha1.UpsyncSet{
			APIGroup:   group,
			Resources:  rnn.resources.List(),
			Namespaces: rnn.namespaces.List(),
			Names:      rnn.names.List(),
		}
		upsyncSets = append(upsyncSets, upsyncSet)
	}
	return upsyncSets, nil
}

func createKcpResource(ctx context.Context, client dynamic.NamespaceableResourceInterface, typedObject interface{}) error {
	wsUnst, err := pkg.ToK8sUnstructedObject(typedObject)
	if err != nil {
		return err
	}
	_, err = client.Get(ctx, wsUnst.GetName(), metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			_, err := client.Create(ctx, &wsUnst, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func createWorkspace(ctx context.Context, kcpClient kcpclient.KcpClient, workspaceName string) error {
	logger := log.FromContext(ctx)
	wsClient, err := kcpClient.GetDyClient(tenancyv1alpha1.SchemeGroupVersion.Group, "Workspace", tenancyv1alpha1.SchemeGroupVersion.Version)
	if err != nil {
		return err
	}
	err = createKcpResource(ctx, wsClient, makeWorkspace(workspaceName))
	if err != nil {
		return err
	}
	interval := time.Second
	return wait.PollImmediateInfiniteWithContext(ctx, time.Second, func(ctx context.Context) (bool, error) {
		fetch, err := wsClient.Get(ctx, workspaceName, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to get workspace %s. Retry in %v", workspaceName, interval))
			return false, nil
		}
		phase, ok, err := unstructured.NestedString(fetch.Object, "status", "phase")
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to get status of workspace %s. Retry in %v", workspaceName, interval))
			return false, nil
		}
		if !ok {
			logger.Error(err, fmt.Sprintf("No phase field in status of workspace %s. Retry in %v", workspaceName, interval))
			return false, nil
		}
		if phase != "Ready" {
			logger.Error(err, fmt.Sprintf("Phase is not 'Ready' in workspace %s. Retry in %v", workspaceName, interval))
			return false, nil
		}
		return true, nil
	})
}

func makeWorkspace(name string) *tenancyv1alpha1.Workspace {
	return &tenancyv1alpha1.Workspace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: tenancyv1alpha1.SchemeGroupVersion.String(),
			Kind:       "Workspace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: tenancyv1alpha1.WorkspaceSpec{
			Type: tenancyv1alpha1.WorkspaceTypeReference{
				Name: tenancyv1alpha1.WorkspaceTypeName("universal"),
			},
		},
	}
}

func createApiBinding(ctx context.Context, kcpClient kcpclient.KcpClient, name string, exportWorkspace string, exportName string) error {
	client, err := kcpClient.GetDyClient(kcpv1alpha1.SchemeGroupVersion.Group, "APIBinding", kcpv1alpha1.SchemeGroupVersion.Version)
	if err != nil {
		return err
	}
	return createKcpResource(ctx, client, makeApiBinding(name, exportWorkspace, exportName))
}

func makeApiBinding(name string, exportWorkspace string, exportName string) *kcpv1alpha1.APIBinding {
	return &kcpv1alpha1.APIBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "APIBinding",
			APIVersion: kcpv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: kcpv1alpha1.APIBindingSpec{
			Reference: kcpv1alpha1.BindingReference{
				Export: &kcpv1alpha1.ExportBindingReference{
					Path: exportWorkspace,
					Name: exportName,
				},
			},
		},
	}
}
