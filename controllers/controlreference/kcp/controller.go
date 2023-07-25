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
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c2pv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/composer"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/kcpclient"
	"github.com/IBM/compliance-to-policy/pkg"
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

	var cr c2pv1alpha1.ControlReferenceKcp
	err := r.Get(ctx, req.NamespacedName, &cr)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	targetWorkspace := cr.Spec.ComplianceDeployment.Target.Workspace
	logger.Info(fmt.Sprintf("--- Create Workload Management Workspace '%s' ---", targetWorkspace))
	kcpClient, err := kcpclient.NewKcpClient(*r.Cfg, "root")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create KcpClient for workspace '%s'", "root"))
	}
	_ = kcpClient
	err = createWorkspace(ctx, kcpClient, strings.Split(targetWorkspace, ":")[1])
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create workload management workspace '%s'", targetWorkspace))
	}

	logger.Info("Initialize Workload Management Workspace")
	kcpWmwClient, err := kcpclient.NewKcpClient(*r.Cfg, targetWorkspace)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create KcpClient for workspace '%s'", targetWorkspace))
	}

	err = createApiBinding(ctx, kcpWmwClient, "bind-kube", "root:compute", "kubernetes")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create bind-kube APIBinding '%s'", targetWorkspace))
	}
	err = createApiBinding(ctx, kcpWmwClient, "bind-espw", "root:espw", "edge.kcp.io")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create bind-espw APIBinding '%s'", targetWorkspace))
	}

	cloneDir, path, err := utils.GitClone(cr.Spec.ComplianceDeployment.PolicyResources.Url, r.TempDir)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to load policy resources %v", cr))
	}
	composer := composer.NewComposer(cloneDir+"/"+path, r.TempDir)

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
	unstCheckPoliciesByPolicy, err := composedResult.ToCheckPoliciesByPolicy()
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to extract check policies %v", intCompliance))
	}
	checkPoliciesByPolicy := map[string][]c2pv1alpha1.CheckPolicy{}
	for policy, unstCheckPolicies := range unstCheckPoliciesByPolicy {
		checkPolicies := []c2pv1alpha1.CheckPolicy{}
		for _, unstCheckPolicy := range unstCheckPolicies {
			var checkPolicy c2pv1alpha1.CheckPolicy
			err := pkg.ToK8sTypedObject(&unstCheckPolicy, &checkPolicy)
			if err != nil {
				return utils.HandleError(logger, err, fmt.Sprintf("Failed to convert unstObj to checkPolicy %v", intCompliance))
			}
			checkPolicies = append(checkPolicies, checkPolicy)
		}
		checkPoliciesByPolicy[policy] = checkPolicies
	}

	logger.Info("--- Deploying generated policies to workspaces ---")
	var cumurativeError error
	for policyId, resources := range resourcesByPolicy {
		logger.V(4).Info(fmt.Sprintf("  poicy: %s", policyId))
		for _, resource := range resources {
			applyC2PlabelToNamespaces(&resource)
			gvk := resource.GroupVersionKind()
			fullName := fmt.Sprintf("%s.%s/%s/%s", resource.GetKind(), resource.GetAPIVersion(), resource.GetNamespace(), resource.GetName())
			logger.V(4).Info(fmt.Sprintf("    deploy resource: %s", fullName))
			dyClient, namespaced, err := kcpWmwClient.GetDyClientWithScopeInfo(gvk.Group, gvk.Kind, gvk.Version)
			if err != nil {
				logger.V(1).Info(fmt.Sprintf("Failed to create KcpClient for '%s' in workspace '%s', will be retried in reconciliation loop", fullName, targetWorkspace))
				cumurativeError = err
				continue
			}
			var err2 error
			if namespaced {
				_, err2 = dyClient.Namespace(resource.GetNamespace()).Create(ctx, &resource, v1.CreateOptions{})
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

	if err := summarize(logger, intCompliance, cr); err != nil {
		logger.Error(nil, "fail to summarize stats")
	}

	logger.Info("--- Create EdgePlacement ---")
	edgePlacement, err := generateEdgePlacement(ctx, cr.Name, &kcpWmwClient, resourcesByPolicy, checkPoliciesByPolicy, *cr.Spec.ComplianceDeployment.ClusterGroups[0].MatchLabels)
	if err != nil {
		return utils.HandleError(logger, err, "Failed to create EdgePlacement")
	}
	epClient, err := kcpWmwClient.GetDyClient("edge.kcp.io", "EdgePlacement", "v1alpha1")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to get epClient %s", cr.Name))
	}
	if err := upsertUnstObj(ctx, epClient, edgePlacement.Name, &edgePlacement); err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to upsert edgePlacement %s", cr.Name))
	}

	policyValidationRequests := []c2pv1alpha1.PolicyValidationRequest{}
	for policyId, checkPolicies := range checkPoliciesByPolicy {
		validationRequest := c2pv1alpha1.PolicyValidationRequest{PolicyId: policyId, CheckPolicies: checkPolicies}
		policyValidationRequests = append(policyValidationRequests, validationRequest)
	}
	resultCollectorCR := c2pv1alpha1.ResultCollector{
		ObjectMeta: v1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: c2pv1alpha1.ResultCollectorSpec{
			ComplianceDeployment:     cr.Spec.ComplianceDeployment,
			PolicyValidationRequests: policyValidationRequests,
			Compliance:               cr.Spec.Compliance,
			Interval:                 "10s",
		},
	}

	if err := utils.CreateOrUpdate(ctx, r.Client, &resultCollectorCR, &c2pv1alpha1.ResultCollector{}); err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to create or update ResultCollector CR %s", resultCollectorCR.Name))
	}

	return ctrl.Result{}, nil
}

func summarize(logger logr.Logger, intCompliance internalcompliance.Compliance, cr c2pv1alpha1.ControlReferenceKcp) error {
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
		For(&c2pv1alpha1.ControlReferenceKcp{}).
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
