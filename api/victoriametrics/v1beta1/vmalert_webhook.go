package v1beta1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var vmalertlog = logf.Log.WithName("vmalert-resource")

func (r *VMAlert) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,admissionReviewVersions=v1,sideEffects=none,path=/validate-operator-victoriametrics-com-v1beta1-vmalert,mutating=false,failurePolicy=fail,groups=operator.acceldata.io,resources=vmalerts,versions=v1beta1,name=vvmalert.kb.io

var _ webhook.Validator = &VMAlert{}

func (r *VMAlert) sanityCheck() error {
	if r.Spec.Datasource.URL == "" {
		return fmt.Errorf("spec.datasource.url cannot be empty")
	}

	if r.Spec.Notifier != nil {
		if r.Spec.Notifier.URL == "" && r.Spec.Notifier.Selector == nil {
			return fmt.Errorf("spec.notifier.url and spec.notifier.selector cannot be empty at the same time, provide at least one setting")
		}
	}
	for idx, nt := range r.Spec.Notifiers {
		if nt.URL == "" && nt.Selector == nil {
			return fmt.Errorf("notifier.url is empty and selector is not set, provide at least once for spec.notifiers at idx: %d", idx)
		}
	}

	return nil
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *VMAlert) ValidateCreate() error {
	if r.Spec.ParsingError != "" {
		return fmt.Errorf(r.Spec.ParsingError)
	}
	if mustSkipValidation(r) {
		return nil
	}
	if err := r.sanityCheck(); err != nil {
		return err
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *VMAlert) ValidateUpdate(old runtime.Object) error {
	if r.Spec.ParsingError != "" {
		return fmt.Errorf(r.Spec.ParsingError)
	}
	if mustSkipValidation(r) {
		return nil
	}
	if err := r.sanityCheck(); err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *VMAlert) ValidateDelete() error {
	// no-op
	return nil
}
