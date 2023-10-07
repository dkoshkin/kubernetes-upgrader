// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
//
//nolint:gochecknoglobals // This came from the Kubebuilder project.
var ingitupgradeautomationlog = logf.Log.WithName("ingitupgradeautomation-resource")

func (r *InGitUpgradeAutomation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This came from the Kubebuilder project.
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//nolint:lll // This is generated code.
//+kubebuilder:webhook:path=/mutate-kubernetesupgraded-dimitrikoshkin-com-v1alpha1-ingitupgradeautomation,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubernetesupgraded.dimitrikoshkin.com,resources=ingitupgradeautomations,verbs=create;update,versions=v1alpha1,name=mingitupgradeautomation.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &InGitUpgradeAutomation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *InGitUpgradeAutomation) Default() {
	ingitupgradeautomationlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // This is generated code.
//+kubebuilder:webhook:path=/validate-kubernetesupgraded-dimitrikoshkin-com-v1alpha1-ingitupgradeautomation,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubernetesupgraded.dimitrikoshkin.com,resources=ingitupgradeautomations,verbs=create;update,versions=v1alpha1,name=vingitupgradeautomation.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &InGitUpgradeAutomation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *InGitUpgradeAutomation) ValidateCreate() (admission.Warnings, error) {
	ingitupgradeautomationlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *InGitUpgradeAutomation) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	ingitupgradeautomationlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *InGitUpgradeAutomation) ValidateDelete() (admission.Warnings, error) {
	ingitupgradeautomationlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
