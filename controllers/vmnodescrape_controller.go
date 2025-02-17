/*


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

package controllers

import (
	"context"
	"github.com/VictoriaMetrics/operator/controllers/factory"
	"github.com/VictoriaMetrics/operator/internal/config"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/VictoriaMetrics/operator/api/v1beta1"
)

// VMNodeScrapeReconciler reconciles a VMNodeScrape object
type VMNodeScrapeReconciler struct {
	client.Client
	Log          logr.Logger
	OriginScheme *runtime.Scheme
	BaseConf     *config.BaseOperatorConf
}

// Scheme implements interface.
func (r *VMNodeScrapeReconciler) Scheme() *runtime.Scheme {
	return r.OriginScheme
}

// Reconcile - reconciles VMNodeScrape objects.
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmnodescrapes,verbs=*
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmnodescrapes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmnodescrapes/finalizers,verbs=*
func (r *VMNodeScrapeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	reqLogger := r.Log.WithValues("vmnodescrape", req.NamespacedName)
	// Fetch the VMNodeScrape instance
	instance := &operatorv1beta1.VMNodeScrape{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return handleGetError(req, "vmnodescrape", err)
	}

	RegisterObjectStat(instance, "vmnodescrape")

	if vmAgentReconcileLimit.MustThrottleReconcile() {
		// fast path, rate limited
		return
	}

	vmAgentSync.Lock()
	defer vmAgentSync.Unlock()

	vmAgentInstances := &operatorv1beta1.VMAgentList{}
	if err := r.List(ctx, vmAgentInstances, config.MustGetNamespaceListOptions()); err != nil {
		return result, err
	}

	for _, vmagent := range vmAgentInstances.Items {
		if !vmagent.DeletionTimestamp.IsZero() || vmagent.Spec.ParsingError != "" {
			continue
		}
		currentVMagent := &vmagent
		match, err := isSelectorsMatches(instance, currentVMagent, currentVMagent.Spec.NodeScrapeSelector)
		if err != nil {
			reqLogger.Error(err, "cannot match vmagent and vmnodescrape")
			continue
		}
		// fast path
		if !match {
			continue
		}
		if err := factory.CreateOrUpdateVMAgent(ctx, currentVMagent, r, r.BaseConf); err != nil {
			reqLogger.Error(err, "cannot create or update vmagent for nodescrape")
			continue
		}
	}

	return
}

// SetupWithManager - setups manager for VMNodeScrape
func (r *VMNodeScrapeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1beta1.VMNodeScrape{}).
		WithOptions(getDefaultOptions()).
		Complete(r)
}
