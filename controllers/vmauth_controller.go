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
	"fmt"
	"sync"

	"github.com/VictoriaMetrics/operator/controllers/factory"
	"github.com/VictoriaMetrics/operator/controllers/factory/finalize"
	"github.com/VictoriaMetrics/operator/internal/config"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1beta1 "github.com/VictoriaMetrics/operator/api/v1beta1"
)

var vmAuthSyncMU = sync.Mutex{}

// VMAuthReconciler reconciles a VMAuth object
type VMAuthReconciler struct {
	client.Client
	BaseConf     *config.BaseOperatorConf
	Log          logr.Logger
	OriginScheme *runtime.Scheme
}

// Scheme implements interface.
func (r *VMAuthReconciler) Scheme() *runtime.Scheme {
	return r.OriginScheme
}

// Reconcile implements interface
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmauths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=operator.acceldata.io,resources=vmauths/status,verbs=get;update;patch
func (r *VMAuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	l := r.Log.WithValues("vmauth", req.NamespacedName)

	var instance operatorv1beta1.VMAuth

	if err := r.Get(ctx, req.NamespacedName, &instance); err != nil {
		return handleGetError(req, "vmauth", err)
	}

	RegisterObjectStat(&instance, "vmauth")

	if !instance.DeletionTimestamp.IsZero() {
		if err := finalize.OnVMAuthDelete(ctx, r, &instance); err != nil {
			return result, fmt.Errorf("cannot remove finalizer from vmauth: %w", err)
		}
		return result, nil
	}
	if instance.Spec.ParsingError != "" {
		return handleParsingError(instance.Spec.ParsingError, &instance)
	}

	if err := finalize.AddFinalizer(ctx, r.Client, &instance); err != nil {
		return result, err
	}

	if err := factory.CreateOrUpdateVMAuth(ctx, &instance, r, r.BaseConf); err != nil {
		return result, fmt.Errorf("cannot create or update vmauth deploy: %w", err)
	}

	svc, err := factory.CreateOrUpdateVMAuthService(ctx, &instance, r)
	if err != nil {
		return result, fmt.Errorf("cannot create or update vmauth service :%w", err)
	}
	if err := factory.CreateOrUpdateVMAuthIngress(ctx, r, &instance); err != nil {
		return result, fmt.Errorf("cannot create or update ingress for vmauth: %w", err)
	}

	if !r.BaseConf.DisableSelfServiceScrapeCreation {
		if err := factory.CreateVMServiceScrapeFromService(ctx, r, svc, instance.Spec.ServiceScrapeSpec, instance.MetricPath()); err != nil {
			l.Error(err, "cannot create serviceScrape for vmauth")
		}
	}
	if r.BaseConf.ForceResyncInterval > 0 {
		result.RequeueAfter = r.BaseConf.ForceResyncInterval
	}

	return
}

// SetupWithManager inits object.
func (r *VMAuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1beta1.VMAuth{}).
		Owns(&v1.Secret{}, builder.OnlyMetadata).
		Owns(&appsv1.Deployment{}, builder.OnlyMetadata).
		Owns(&v1.Service{}, builder.OnlyMetadata).
		Owns(&v1.ServiceAccount{}, builder.OnlyMetadata).
		Owns(&operatorv1beta1.VMServiceScrape{}, builder.OnlyMetadata).
		WithOptions(getDefaultOptions()).
		Complete(r)
}
