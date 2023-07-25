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

package ocm

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/composer"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/ocmk8sclients"
	"github.com/IBM/compliance-to-policy/pkg/types/internalcompliance"
	typesplacement "github.com/IBM/compliance-to-policy/pkg/types/placements"
	typespolicy "github.com/IBM/compliance-to-policy/pkg/types/policy"
	"github.com/go-logr/logr"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
)

// ControlReferenceReconciler reconciles a ControlReference object
type ControlReferenceReconciler struct {
	client.Client
	Scheme                    *runtime.Scheme
	TempDir                   string
	OcmK8ResourceInterfaceSet ocmk8sclients.OcmK8ResourceInterfaceSetType
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

//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferences,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferences/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=controlreferences/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps.open-cluster-management.io,resources=placementrules,verbs=*
//+kubebuilder:rbac:groups=policy.open-cluster-management.io,resources=placementbindings;policies;policysets,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ControlReference object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ControlReferenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cr compliancetopolicycontrollerv1alpha1.ControlReference
	err := r.Get(ctx, req.NamespacedName, &cr)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	cloneDir, path, err := utils.GitClone(cr.Spec.PolicyResources.Url, r.TempDir)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to load policy resources %v", cr))
	}

	composer := composer.NewComposer(cloneDir+"/"+path, r.TempDir)

	logger.Info("")
	logger.Info("--- Start creating files by policy-generator ---")
	intCompliance := utils.ConvertComplianceToIntCompliance(cr.Spec.Compliance)
	composedResult, err := composer.Compose(cr.Spec.Target.Namespace, intCompliance, nil)
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to compose %v", intCompliance))
	}

	_ = composedResult

	resourcesByPolicy := composedResult.ToResourcesByPolicy()

	logger.Info(fmt.Sprintf("--- Deploying generated policies to namespace '%s' ---", cr.Spec.Target.Namespace))
	for _, resources := range resourcesByPolicy {
		for _, resource := range resources {
			kind := resource.GetKind()
			yamlData, err := resource.AsYAML()
			if err != nil {
				return utils.HandleError(logger, err, fmt.Sprintf("Failed to convert resource %s to yaml", kind))
			}
			switch kind {
			case "Policy":
				var typedObj typespolicy.Policy
				if err := utilyaml.Unmarshal(yamlData, &typedObj); err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to unmarshal %s", string(yamlData)))
				}
				client := ocmk8sclients.NewPolicyClient(r.OcmK8ResourceInterfaceSet.Policy)
				_, err := client.Create(cr.Spec.Target.Namespace, typedObj)
				if err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to create %s.%s", kind, typedObj.Name))
				}
			case "PlacementBinding":
				var typedObj typesplacement.PlacementBinding
				if err := utilyaml.Unmarshal(yamlData, &typedObj); err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to unmarshal %s", string(yamlData)))
				}
				client := ocmk8sclients.NewPlacementBindingClient(r.OcmK8ResourceInterfaceSet.PlacementBinding)
				_, err := client.Create(cr.Spec.Target.Namespace, typedObj)
				if err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to create %s.%s", kind, typedObj.Name))
				}
			case "PlacementRule":
				var typedObj typesplacement.PlacementRule
				if err := utilyaml.Unmarshal(yamlData, &typedObj); err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to unmarshal %s", string(yamlData)))
				}
				client := ocmk8sclients.NewPlacementRuleClient(r.OcmK8ResourceInterfaceSet.PlacementRule)
				_, err := client.Create(cr.Spec.Target.Namespace, typedObj)
				if err != nil {
					return utils.HandleError(logger, err, fmt.Sprintf("Failed to create %s.%s", kind, typedObj.Name))
				}
			}
		}
	}

	if err := summarize(logger, intCompliance, cr); err != nil {
		logger.Error(nil, "fail to summarize stats")
	}

	return ctrl.Result{}, nil
}

func summarize(logger logr.Logger, intCompliance internalcompliance.Compliance, cr compliancetopolicycontrollerv1alpha1.ControlReference) error {
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
func (r *ControlReferenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compliancetopolicycontrollerv1alpha1.ControlReference{}).
		Complete(r)
}
