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
	"time"

	compliancetopolicycontrollerv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/IBM/compliance-to-policy/controllers/utils/kcpclient"
	wgpolicyk8sv1alpha2 "github.com/IBM/compliance-to-policy/controllers/wgpolicyk8s.io/v1alpha2"
	"github.com/IBM/compliance-to-policy/pkg"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger logr.Logger = ctrl.Log.WithName("result-collector-controller")

type ResultCollectorReconciler struct {
	client.Client
	Cfg         *rest.Config
	ctxBatch    context.Context
	cancelBatch context.CancelFunc
}

//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=wgpolicyk8s.io,resources=policyreports;clusterpolicyreports,verbs=get;list;watch;create;update;patch;delete

func (r *ResultCollectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	if r.cancelBatch != nil {
		r.cancelBatch()
	}
	r.ctxBatch = context.Background()
	r.ctxBatch, r.cancelBatch = context.WithCancel(ctx)

	var cr compliancetopolicycontrollerv1alpha1.ResultCollector
	err := r.Get(ctx, req.NamespacedName, &cr)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	compDeployNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      cr.Spec.ComplianceDeployment,
	}
	var compDeploy compliancetopolicycontrollerv1alpha1.ComplianceDeployment
	err = r.Get(ctx, compDeployNamespacedName, &compDeploy)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	controlReferenceNamespacedName := types.NamespacedName{
		Namespace: req.Namespace,
		Name:      cr.Spec.ControlReference,
	}
	var controlReference compliancetopolicycontrollerv1alpha1.ControlReferenceKcp
	err = r.Get(ctx, controlReferenceNamespacedName, &controlReference)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	workspaceObjs, err := utils.GetWorkspaces(ctx, *r.Cfg, "root:espw")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to get workspaces %v", cr))
	}
	logger.V(3).Info("--- Start collecting generated reports from workspaces ---")

	run := func() {
		interval, err := cr.Spec.GetInterval()
		if err != nil {
			logger.Info(fmt.Sprintf("Failed to set interval. Use default (10s) : %v", err))
			interval = 10 * time.Second
		}
		t := time.NewTicker(interval)
		for {
			select {
			case <-r.ctxBatch.Done():
				t.Stop()
				return
			case <-t.C:
				logger.Info("Collect...")
				if err := r.collect(ctx, cr, controlReference, workspaceObjs); err != nil {
					logger.Error(err, "Failed to collect results. Retry later...")
				}
			}
		}
	}
	go run()
	return ctrl.Result{}, nil
}

func (r *ResultCollectorReconciler) collect(
	ctx context.Context,
	cr compliancetopolicycontrollerv1alpha1.ResultCollector,
	controlReference compliancetopolicycontrollerv1alpha1.ControlReferenceKcp,
	workspaces []utils.Workspace,
) error {
	sts, err := utils.GetSts(ctx, *r.Cfg, controlReference)
	if err != nil {
		logger.Error(err, "Failed to get sts")
		return err
	}
	for idx, workspace := range workspaces {
		logger.V(3).Info(fmt.Sprintf("\nworkspace: %s", workspace))
		wsName := workspace.Name
		location := ""
		for _, destination := range sts.Destinations {
			if destination.SyncTargetName == workspace.SyncTargetName {
				location = destination.LocationName
			}
		}
		compoundedPolicyReport := wgpolicyk8sv1alpha2.PolicyReport{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprintf("raw-%s-%d", cr.Name, idx),
				Namespace: cr.Namespace,
				Annotations: map[string]string{
					"workspaceName": wsName,
					"locationName":  location,
				},
			},
			Results: []*wgpolicyk8sv1alpha2.PolicyReportResult{},
			Summary: wgpolicyk8sv1alpha2.PolicyReportSummary{},
		}
		kcpClient, err := kcpclient.NewKcpClient(*r.Cfg, wsName)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to create KcpClient for workspace '%s'", workspace))
			return err
		}
		polrDyClient, err := kcpClient.GetDyClient("wgpolicyk8s.io", "PolicyReport", "v1alpha2")
		if err != nil {
			logger.Error(err, fmt.Sprintf("Failed to create polrDyClient for workspace '%s'", workspace))
			return err
		}
		for _, policyReportRef := range cr.Spec.PolicyReports {
			unstObj, err := polrDyClient.Namespace(policyReportRef.Namespace).Get(ctx, policyReportRef.Name, v1.GetOptions{})
			if err != nil {
				logger.Info(fmt.Sprintf("Cant't fetch policyReport '%s/%s' from workspace '%s': %v", policyReportRef.Namespace, policyReportRef.Name, workspace, err))
				continue
			}
			var policyReport wgpolicyk8sv1alpha2.PolicyReport
			if err := pkg.ToK8sTypedObject(unstObj, &policyReport); err != nil {
				logger.Info(fmt.Sprintf("Cant't convert policyReport '%s/%s' to typed resource in '%s': %v", policyReportRef.Namespace, policyReportRef.Name, workspace, err))
				continue
			}
			standard, category, control := getComplianceProps(controlReference, policyReportRef)
			for _, result := range policyReport.Results {
				prop := map[string]string{
					"policyId": policyReportRef.PolicyId,
					"standard": standard,
					"category": category,
					"control":  control,
				}
				props := result.Properties
				if props == nil {
					result.Properties = prop
				} else {
					for key, value := range prop {
						result.Properties[key] = value
					}
				}
			}
			compoundedPolicyReport.Results = append(compoundedPolicyReport.Results, policyReport.Results...)
			compoundedPolicyReport.Summary.Error = compoundedPolicyReport.Summary.Error + policyReport.Summary.Error
			compoundedPolicyReport.Summary.Fail = compoundedPolicyReport.Summary.Fail + policyReport.Summary.Fail
			compoundedPolicyReport.Summary.Pass = compoundedPolicyReport.Summary.Pass + policyReport.Summary.Pass
			compoundedPolicyReport.Summary.Skip = compoundedPolicyReport.Summary.Skip + policyReport.Summary.Skip
			compoundedPolicyReport.Summary.Warn = compoundedPolicyReport.Summary.Warn + policyReport.Summary.Warn
		}
		var fetchedCompoundedPolicyReport wgpolicyk8sv1alpha2.PolicyReport
		namespacedName := types.NamespacedName{Namespace: compoundedPolicyReport.GetNamespace(), Name: compoundedPolicyReport.Name}
		if err := r.Get(ctx, namespacedName, &fetchedCompoundedPolicyReport, &client.GetOptions{}); err != nil {
			if errors.IsNotFound(err) {
				if err := r.Create(ctx, &compoundedPolicyReport, &client.CreateOptions{}); err != nil {
					logger.Error(err, fmt.Sprintf("Failed to create compoundedPolicyReport for workspace '%s'", workspace))
					return err
				}
			} else {
				logger.Error(err, fmt.Sprintf("Failed to fetch compoundedPolicyReport for workspace '%s'", workspace))
				return err
			}
		} else {
			compoundedPolicyReport.SetResourceVersion(fetchedCompoundedPolicyReport.GetResourceVersion())
			if err := r.Update(ctx, &compoundedPolicyReport, &client.UpdateOptions{}); err != nil {
				logger.Error(err, fmt.Sprintf("Failed to update compoundedPolicyReport for workspace '%s'", workspace))
				return err
			}
		}
	}
	return nil
}

func getComplianceProps(controlReference compliancetopolicycontrollerv1alpha1.ControlReferenceKcp, policyReportRef compliancetopolicycontrollerv1alpha1.PolicyReportRef) (standard string, category string, control string) {
	for _, category := range controlReference.Spec.Compliance.Standard.Categories {
		for _, control := range category.Controls {
			for _, controlRef := range control.ControlRefs {
				if policyReportRef.PolicyId == controlRef {
					return controlReference.Spec.Compliance.Standard.Name, category.Name, control.Name
				}
			}
		}
	}
	return controlReference.Spec.Compliance.Standard.Name, "", ""
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResultCollectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&compliancetopolicycontrollerv1alpha1.ResultCollector{}).
		Complete(r)
}
