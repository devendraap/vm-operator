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

	operatorv1beta1 "github.com/VictoriaMetrics/operator/api/v1beta1"
	"github.com/VictoriaMetrics/operator/controllers/factory"
	"github.com/VictoriaMetrics/operator/internal/config"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// VMProbeReconciler reconciles a VMProbe object
type VMProbeReconciler struct {
	client.Client
	Log          logr.Logger
	OriginScheme *runtime.Scheme
	BaseConf     *config.BaseOperatorConf
}

// Scheme implements interface.
func (r *VMProbeReconciler) Scheme() *runtime.Scheme {
	return r.OriginScheme
}

// Reconcile - syncs VMProbe
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmprobes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmprobes/status,verbs=get;update;patch
func (r *VMProbeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {

	reqLogger := r.Log.WithValues("vmprobe", req.NamespacedName)

	// Fetch the VMPodScrape instance
	instance := &operatorv1beta1.VMProbe{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		return handleGetError(req, "vmprobescrape", err)
	}

	RegisterObjectStat(instance, "vmprobescrape")
	if vmAgentReconcileLimit.MustThrottleReconcile() {
		// fast path, rate limited
		return
	}

	vmAgentSync.Lock()
	defer vmAgentSync.Unlock()

	vmAgentInstances := &operatorv1beta1.VMAgentList{}
	err = r.List(ctx, vmAgentInstances, config.MustGetNamespaceListOptions())
	if err != nil {
		return result, err
	}

	for _, vmagent := range vmAgentInstances.Items {
		if !vmagent.DeletionTimestamp.IsZero() || vmagent.Spec.ParsingError != "" {
			continue
		}
		currentVMagent := &vmagent
		reqLogger = reqLogger.WithValues("vmagent", vmagent.Name)
		match, err := isSelectorsMatches(instance, currentVMagent, currentVMagent.Spec.ProbeSelector)
		if err != nil {
			reqLogger.Error(err, "cannot match vmagent and vmProbe")
			continue
		}
		// fast path
		if !match {
			continue
		}
		if err := factory.CreateOrUpdateVMAgent(ctx, currentVMagent, r, r.BaseConf); err != nil {
			reqLogger.Error(err, "cannot create or update vmagent")
			continue
		}
	}
	return
}

// SetupWithManager - setups VMProbe manager
func (r *VMProbeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1beta1.VMProbe{}).
		WithOptions(getDefaultOptions()).
		Complete(r)
}
