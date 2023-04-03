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

package compliancedeployment

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils"
)

// ComplianceDeploymentReconciler reconciles a ComplianceDeployment object
type ComplianceDeploymentReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	TempDir string
}

//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=compliancedeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=compliancedeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=compliancedeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ComplianceDeployment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *ComplianceDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var compDeploy compliancetopolicycontrollerv1alpha1.ComplianceDeployment
	err := r.Get(ctx, req.NamespacedName, &compDeploy)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	logger.Info("")
	logger.Info(fmt.Sprintf("--- Starting processing compliance-deployment CR '%s' ---", compDeploy.Name))

	var cr compliancetopolicycontrollerv1alpha1.ControlReference
	cdComposit, err := utils.MakeControlReference(r.TempDir, compDeploy)
	if err != nil {
		return utils.HandleError(logger, err, "Failed to create CR manifest")
	}
	cr = cdComposit.ControlReference

	nsNamespaceName := types.NamespacedName{
		Name: compDeploy.Spec.Target.Namespace,
	}

	if compDeploy.Spec.Target.Namespace != "" && compDeploy.Spec.Target.Workspace == "" {
		var ns corev1.Namespace
		err = r.Get(ctx, nsNamespaceName, &ns, &client.GetOptions{})
		if errors.IsNotFound(err) {
			if err := r.Create(ctx, &corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: compDeploy.Spec.Target.Namespace,
				},
			}, &client.CreateOptions{}); err != nil {
				return utils.HandleError(logger, err, fmt.Sprintf("Failed to create target namespace %s", compDeploy.Spec.Target.Namespace))
			}
		} else if err != nil {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to get target namespace %s", compDeploy.Spec.Target.Namespace))
		}
		if err := r.Create(ctx, &cr, &client.CreateOptions{}); err != nil {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to create CR %v", compDeploy))
		}
	} else if compDeploy.Spec.Target.Namespace == "" && compDeploy.Spec.Target.Workspace != "" {
		if err := r.Create(ctx, &cdComposit.ControlReferenceKcp, &client.CreateOptions{}); err != nil {
			return utils.HandleError(logger, err, fmt.Sprintf("Failed to create KCP CR %v", compDeploy))
		}
	} else {
		return utils.HandleError(logger, errors.NewBadRequest("Should select either Namespace or Workspace"), "Should select either Namespace or Workspace")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ComplianceDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compliancetopolicycontrollerv1alpha1.ComplianceDeployment{}).
		Complete(r)
}
