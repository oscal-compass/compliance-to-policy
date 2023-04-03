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

package kcp

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	edge "github.com/IBM/compliance-to-policy/controllers/edge.kcp.io/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/kcpclient"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/IBM/compliance-to-policy/pkg/composer"
	"github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	"github.com/go-logr/logr"
)

var logger logr.Logger = ctrl.Log.WithName("control-reference-controller-kcp")

// ControlReferenceReconciler reconciles a ControlReference object
type ControlReferenceKcpReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	TempDir string
	Cfg     *rest.Config
}

type RequiredControlId struct {
	Profile             string `yaml:"profile,omitempty"`
	ComponentDefinition string `yaml:"component-definition,omitempty"`
	Missing             string `yaml:"missing,omitempty"`
}

type Summary struct {
	Name                 string              `yaml:"Name,omitempty"`
	ComplianceDefinition string              `yaml:"Compliance Definition,omitempty"`
	Catalog              string              `yaml:"Catalog,omitempty"`
	Profile              string              `yaml:"Profile,omitempty"`
	ComponentDefinition  string              `yaml:"Component Definition,omitempty"`
	RequiredControlId    RequiredControlId   `yaml:"Required Controls,omitempty"`
	Timestamp            string              `yaml:"Timestamp Generated,omitempty"`
	GeneratedPolicies    map[string][]string `yaml:"Generated Policies,omitempty"`
}

//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferencekcps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferencekcps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferencekcps/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControlReference object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ControlReferenceKcpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cr compliancetopolicycontrollerv1alpha1.ControlReferenceKcp
	err := r.Get(ctx, req.NamespacedName, &cr)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	cloneDir, path, err := utils.GitClone(cr.Spec.PolicyResources.Url, r.TempDir)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to load policy resources %v", cr))
	}

	composer := composer.NewComposer(cloneDir+"/"+path, r.TempDir)

	logger.Info("--- Create EdgePlacement ---")
	if err := r.createEdgePlacement(ctx, cr); err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create EdgePlacement %v", cr))
	}

	logger.Info("--- Get SinglePlacementSlice ---")
	sts, err := utils.GetSts(ctx, *r.Cfg, cr)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get SinglePlacementSlice %v", cr))
		return ctrl.Result{}, err
	}

	_ = sts

	logger.Info("--- Calculate workspace list to which the policies are delivered from clusterGroups ---")
	workspaceObjs, err := utils.GetWorkspaces(ctx, *r.Cfg, "root:espw")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to get workspaces %v", cr))
	}
	logger.Info(fmt.Sprintf("  Workspaces: %s", workspaceObjs))

	logger.Info("")
	logger.Info("--- Start creating files by policy-generator ---")
	intCompliance := utils.ConvertComplianceToIntCompliance(cr.Spec.Compliance)
	composedResult, err := composer.Compose("default", intCompliance, nil)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to compose %v", intCompliance))
	}

	resourcesByPolicy, err := composedResult.ToPrimitiveResourcesByPolicy()
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to extract config policies %v", intCompliance))
	}

	logger.Info("--- Deploying generated policies to workspaces ---")
	workspaces := []string{}
	for _, w := range workspaceObjs {
		workspaces = append(workspaces, w.Name)
	}
	workspaces = append(workspaces, "root:wmw")
	for _, workspace := range workspaces {
		logger.Info(fmt.Sprintf("\nworkspace: %s", workspace))
		kcpClient, err := kcpclient.NewKcpClient(*r.Cfg, workspace)
		if err != nil {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to create KcpClient for workspace '%s'", workspace))
		}
		var cumurativeError error
		for policyId, resources := range resourcesByPolicy {
			logger.V(4).Info(fmt.Sprintf("  poicy: %s", policyId))
			for _, resource := range resources {
				gvk := resource.GroupVersionKind()
				fullName := fmt.Sprintf("%s.%s/%s/%s", resource.GetKind(), resource.GetAPIVersion(), resource.GetNamespace(), resource.GetName())
				logger.V(4).Info(fmt.Sprintf("    deploy resource: %s", fullName))
				dyClient, namespaced, err := kcpClient.GetDyClientWithScopeInfo(gvk.Group, gvk.Kind, gvk.Version)
				if err != nil {
					logger.V(1).Info(fmt.Sprintf("Failed to create KcpClient for '%s' in workspace '%s', will be retried in reconciliation loop", fullName, workspace))
					cumurativeError = err
					continue
				}
				var err2 error
				if namespaced {
					ns := resource.GetNamespace()
					if ns == "" {
						ns = "default"
					}
					_, err2 = dyClient.Namespace(ns).Create(ctx, &resource, v1.CreateOptions{})
				} else {
					_, err2 = dyClient.Create(ctx, &resource, v1.CreateOptions{})
				}
				if err2 != nil {
					if errors.IsAlreadyExists(err2) {
						logger.Info(fmt.Sprintf("'%s' already exists", fullName))
						continue
					}
					logger.V(1).Info(fmt.Sprintf("Failed to deploy '%s', will be retried in reconciliation loop", fullName))
					cumurativeError = err2
					continue
				}
			}
		}
		if cumurativeError != nil {
			return ctrl.Result{}, cumurativeError
		}
	}

	if err := summarize(logger, intCompliance, cr); err != nil {
		logger.Error(nil, "fail to summarize stats")
	}

	// Kyverno Policy Reports per policy
	policyReportRefs := []compliancetopolicycontrollerv1alpha1.PolicyReportRef{}
	clusterPolicyReportRefs := []compliancetopolicycontrollerv1alpha1.PolicyReportRef{}
	for policyId, resources := range resourcesByPolicy {
		for _, resource := range resources {
			kind := resource.GetKind()
			apiVersion := resource.GetAPIVersion()
			if kind == "Policy" && apiVersion == "kyverno.io/v1" {
				ns := resource.GetNamespace()
				if ns == "" {
					ns = "default"
				}
				policyReportRefs = append(policyReportRefs, compliancetopolicycontrollerv1alpha1.PolicyReportRef{
					PolicyId:  policyId,
					Name:      "pol-" + resource.GetName(),
					Namespace: ns,
				})
			} else if kind == "ClusterPolicy" && apiVersion == "kyverno.io/v1" {
				clusterPolicyReportRefs = append(clusterPolicyReportRefs, compliancetopolicycontrollerv1alpha1.PolicyReportRef{
					PolicyId: policyId,
					Name:     "cpol-" + resource.GetName(),
				})
			}
		}
	}

	resultCollectorCR := compliancetopolicycontrollerv1alpha1.ResultCollector{
		ObjectMeta: v1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: compliancetopolicycontrollerv1alpha1.ResultCollectorSpec{
			ComplianceDeployment: cr.Name,
			ControlReference:     cr.Name,
			PolicyReports:        policyReportRefs,
			ClusterPolicyReports: clusterPolicyReportRefs,
			Interval:             "10s",
		},
	}
	namespacedName := types.NamespacedName{Namespace: resultCollectorCR.Namespace, Name: resultCollectorCR.Name}
	var existingResultCollectorCR compliancetopolicycontrollerv1alpha1.ResultCollector
	if err := r.Get(ctx, namespacedName, &existingResultCollectorCR); err != nil {
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Not found so create resultCollectorCR %v", namespacedName))
			if err := r.Create(ctx, &resultCollectorCR, &client.CreateOptions{}); err != nil {
				return utils.HandleError(logger, err, fmt.Sprintf("Failed to create resultCollectorCR %v", namespacedName))
			}
		} else {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to get resultCollectorCR %v", namespacedName))
		}
	} else {
		logger.Info(fmt.Sprintf("Found so update resultCollectorCR %v", namespacedName))
		resultCollectorCR.SetResourceVersion(existingResultCollectorCR.GetResourceVersion())
		resultCollectorCR.SetUID("")
		if err := r.Update(ctx, &resultCollectorCR, &client.UpdateOptions{}); err != nil {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to update resultCollectorCR %v", namespacedName))
		}
	}

	return ctrl.Result{}, nil
}

func (r *ControlReferenceKcpReconciler) createEdgePlacement(
	ctx context.Context,
	cr compliancetopolicycontrollerv1alpha1.ControlReferenceKcp,
) error {

	kcpClient, err := kcpclient.NewKcpClient(*r.Cfg, "root:wmw")
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get kcpClient %s", cr.Name))
		return err
	}
	epClient, err := kcpClient.GetDyClient("edge.kcp.io", "EdgePlacement", "v1alpha1")
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to get epClient %s", cr.Name))
		return err
	}
	edgePlacement := edge.EdgePlacement{
		TypeMeta: v1.TypeMeta{
			Kind:       "EdgePlacement",
			APIVersion: "edge.kcp.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: cr.Name,
		},
		Spec: edge.EdgePlacementSpec{
			LocationSelectors: []v1.LabelSelector{{
				MatchLabels: *cr.Spec.Target.ClusterGroups[0].MatchLabels,
			}},
		},
	}
	if err := upsertUnstObj(ctx, epClient, edgePlacement.Name, &edgePlacement); err != nil {
		logger.Error(err, fmt.Sprintf("Failed to upsert edgePlacement %s", cr.Name))
		return err
	}

	return nil
}

func summarize(logger logr.Logger, intCompliance internalcompliance.Compliance, cr compliancetopolicycontrollerv1alpha1.ControlReferenceKcp) error {
	generatedPolicies := map[string][]string{}
	for _, category := range intCompliance.Standard.Categories {
		for _, control := range category.Controls {
			_, ok := generatedPolicies[control.Name]
			if ok {
				generatedPolicies[control.Name] = append(generatedPolicies[control.Name], control.ControlRefs...)
			} else {
				generatedPolicies[control.Name] = control.ControlRefs
			}
		}
	}
	cdSummary := cr.Spec.Summary
	summary := Summary{
		Name:                 cdSummary["name"],
		ComplianceDefinition: fmt.Sprintf("%s/%s", cdSummary["compliance-definition-namespace"], cdSummary["compliance-definition-name"]),
		Catalog:              cdSummary["catalog"],
		Profile:              cdSummary["profile"],
		ComponentDefinition:  cdSummary["component-definition"],
		RequiredControlId: RequiredControlId{
			Profile:             cdSummary["controlIdsInProfile"],
			ComponentDefinition: cdSummary["controlIdsInCD"],
			Missing:             cdSummary["excludedControlIds"],
		},
		Timestamp:         time.Now().String(),
		GeneratedPolicies: generatedPolicies,
	}
	yamlDate, err := yaml.Marshal(summary)
	if err != nil {
		return err
	}

	logger.Info("")
	logger.Info("--- Summary --- \n\n" + string(yamlDate) + "\n")
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ControlReferenceKcpReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compliancetopolicycontrollerv1alpha1.ControlReferenceKcp{}).
		Complete(r)
}

func upsertUnstObj(ctx context.Context, dyClient dynamic.NamespaceableResourceInterface, name string, typedObject interface{}) error {
	unstObj, err := pkg.ToK8sUnstructedObject(typedObject)
	if err != nil {
		logger.Error(err, fmt.Sprintf("Failed to convert %s to unstObj", name))
		return err
	}
	fetchedUnst, err := dyClient.Get(ctx, name, v1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			if _, err := dyClient.Create(ctx, &unstObj, v1.CreateOptions{}); err != nil {
				logger.Error(err, fmt.Sprintf("Failed to create %s to unstObj", name))
				return err
			}
		}
	} else {
		unstObj.SetResourceVersion(fetchedUnst.GetResourceVersion())
		if _, err := dyClient.Update(ctx, &unstObj, v1.UpdateOptions{}); err != nil {
			logger.Error(err, fmt.Sprintf("Failed to update %s", name))
			return err
		}
	}
	return nil
}
