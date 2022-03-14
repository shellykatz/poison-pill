/*
Copyright 2021.

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

package v1alpha1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"time"
)

const (
	WebhookCertDir  = "/apiserver.local.config/certificates"
	WebhookCertName = "apiserver.crt"
	WebhookKeyName  = "apiserver.key"
)

//minimal time durations allowed string format
const (
	MinDurPeerApiServerTimeout = "10ms"
	MinDurApiServerTimeout = "10ms"
	MinDurPeerDialTimeout = "10ms"
	MinDurPeerRequestTimeout = "10ms"
	MinDurApiCheckInterval = "1s"
	MinDurPeerUpdateInterval = "10s"
)

const (
	ErrPeerApiServerTimeout = "PeerApiServerTimeout " + MinDurPeerApiServerTimeout
	ErrApiServerTimeout     = "ApiServerTimeout " + MinDurApiServerTimeout
	ErrPeerDialTimeout      = "PeerDialTimeout " + MinDurPeerDialTimeout
	ErrPeerRequestTimeout   = "PeerRequestTimeout " + MinDurPeerRequestTimeout
	ErrApiCheckInterval     = "ApiCheckInterval can't be less than " + MinDurApiCheckInterval
	ErrPeerUpdateInterval   = "PeerUpdateInterval can't be less than " + MinDurPeerUpdateInterval
)

type field struct {
	fieldDurValue *metav1.Duration
	fieldMinValueString string
	errorMessage string
}

// log is for logging in this package.
var poisonpillconfiglog = logf.Log.WithName("poisonpillconfig-resource")

func (r *PoisonPillConfig) SetupWebhookWithManager(mgr ctrl.Manager) error {

	// check if OLM injected certs
	certs := []string{filepath.Join(WebhookCertDir, WebhookCertName), filepath.Join(WebhookCertDir, WebhookKeyName)}
	certsInjected := true
	for _, fname := range certs {
		if _, err := os.Stat(fname); err != nil {
			certsInjected = false
			break
		}
	}
	if certsInjected {
		server := mgr.GetWebhookServer()
		server.CertDir = WebhookCertDir
		server.CertName = WebhookCertName
		server.KeyName = WebhookKeyName
	} else {
		poisonpillconfiglog.Info("OLM injected certs for webhooks not found")
	}
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-poison-pill-medik8s-io-v1alpha1-poisonpillconfig,mutating=false,failurePolicy=fail,sideEffects=None,groups=poison-pill.medik8s.io,resources=poisonpillconfigs,verbs=create;update,versions=v1alpha1,name=vpoisonpillconfig.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &PoisonPillConfig{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *PoisonPillConfig) ValidateCreate() error {
	poisonpillconfiglog.Info("validate create", "name", r.Name)

	return r.ValidateTimes()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *PoisonPillConfig) ValidateUpdate(old runtime.Object) error {
	poisonpillconfiglog.Info("validate update", "name", r.Name)

	return r.ValidateTimes()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *PoisonPillConfig) ValidateDelete() error {
	poisonpillconfiglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

// ValidateTimes validates each time field in the PoisonPillConfig CR doesn't go below the minimum time
// that was defined to it
func (r *PoisonPillConfig) ValidateTimes() error {

	s := r.Spec
	fields := []field{
		{s.PeerApiServerTimeout, MinDurPeerApiServerTimeout, ErrPeerApiServerTimeout},
		{s.ApiServerTimeout, MinDurApiServerTimeout, ErrApiServerTimeout},
		{s.PeerDialTimeout, MinDurPeerDialTimeout, ErrPeerDialTimeout},
		{s.PeerRequestTimeout, MinDurPeerRequestTimeout, ErrPeerRequestTimeout},
		{s.ApiCheckInterval, MinDurApiCheckInterval, ErrApiCheckInterval},
		{s.PeerUpdateInterval, MinDurPeerUpdateInterval, ErrPeerUpdateInterval},
	}

	for _, f := range fields {
		err := checkField(f.fieldDurValue.Milliseconds(), f.fieldMinValueString, f.errorMessage)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkField(inputTime int64, minValidDur string, errMessage string) error {
	minValidDurMS, err := toMS(minValidDur)
	if err != nil {
		return err
	}

	if inputTime < minValidDurMS {
		err := fmt.Errorf(errMessage)
		poisonpillconfiglog.Error(err, errMessage, "time given (in milliseconds) was:", inputTime)
		return err
	}

	return nil
}

func toMS(value string) (int64, error) {
	d, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}
	return d.Milliseconds(), nil
}