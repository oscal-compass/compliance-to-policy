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
	"sync"
	"time"

	c2pv1alpha1 "github.com/IBM/compliance-to-policy/api/v1alpha1"
	"github.com/IBM/compliance-to-policy/controllers/utils"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var logger logr.Logger = ctrl.Log.WithName("result-collector-controller")

type ResultCollectorReconciler struct {
	sync.Mutex
	client.Client
	Cfg     *rest.Config
	workers map[string]*Worker
}

func NewResultCollectorReconciler(client client.Client, cfg *rest.Config) *ResultCollectorReconciler {
	return &ResultCollectorReconciler{
		Client:  client,
		Cfg:     cfg,
		workers: map[string]*Worker{},
	}
}

type PolicyValidationResult struct {
	policyId           string
	checkPolicyResults []CheckPolicyResult
}

type CheckPolicyResult struct {
	testResults []CheckPolicyTestResult
	checkPolicy c2pv1alpha1.CheckPolicy
	policyId    string
}

type CheckPolicyTestResult struct {
	objectDefinition unstructured.Unstructured
	pass             bool
	message          string
	error            error
}

//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=compliance-to-policy.io,resources=resultcollectors/finalizers,verbs=update
//+kubebuilder:rbac:groups=wgpolicyk8s.io,resources=policyreports;clusterpolicyreports,verbs=get;list;watch;create;update;patch;delete

func (r *ResultCollectorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var cr c2pv1alpha1.ResultCollector
	err := r.Get(ctx, req.NamespacedName, &cr)
	if errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	}

	workspaceObjs, err := utils.GetWorkspaces(ctx, *r.Cfg, "root:espw")
	if err != nil {
		return utils.HandleError(logger, err, fmt.Sprintf("Failed to get workspaces %v", cr))
	}
	logger.V(3).Info("--- Start collecting generated reports from workspaces ---")

	r.run(ctx, cr, workspaceObjs)
	return ctrl.Result{}, nil
}

func (r *ResultCollectorReconciler) run(ctx context.Context, cr c2pv1alpha1.ResultCollector, workspaceObjs []utils.Workspace) {
	r.Lock()
	defer r.Unlock()
	worker, ok := r.workers[cr.Name]
	if !ok {
		r.workers[cr.Name] = newWorker()
	} else {
		worker.delete()
		r.workers[cr.Name] = newWorker()
	}
	go r.workers[cr.Name].start(r, cr, workspaceObjs)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ResultCollectorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&c2pv1alpha1.ResultCollector{}).
		Complete(r)
}

type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newWorker() *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *Worker) delete() {
	w.cancel()
}

func (w *Worker) start(r *ResultCollectorReconciler, cr c2pv1alpha1.ResultCollector, workspaceObjs []utils.Workspace) {
	interval, err := cr.Spec.GetInterval()
	if err != nil {
		logger.Info(fmt.Sprintf("Failed to set interval. Use default (10s) : %v", err))
		interval = 10 * time.Second
	}
	t := time.NewTicker(interval)
	for {
		select {
		case <-w.ctx.Done():
			t.Stop()
			return
		case <-t.C:
			logger.Info("Collect...")
			if err := r.collect(w.ctx, cr, workspaceObjs, cr.Spec.PolicyValidationRequests); err != nil {
				logger.Error(err, "Failed to collect results. Retry later...")
			}
		}
	}
}
