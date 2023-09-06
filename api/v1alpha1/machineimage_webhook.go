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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

const (
	kubernetesVersionEnv = "KUBERNETES_VERSION"
)

// log is for logging in this package.
//
//nolint:gochecknoglobals // This came from the Kubebuilder project.
var machineimagelog = logf.Log.WithName("machineimage-resource")

func (r *MachineImage) SetupWebhookWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This came from the Kubebuilder project.
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
//nolint:lll // This is generated code.
//+kubebuilder:webhook:path=/mutate-kubernetesupgraded-dimitrikoshkin-com-v1alpha1-machineimage,mutating=true,failurePolicy=fail,sideEffects=None,groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages,verbs=create;update,versions=v1alpha1,name=mmachineimage.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &MachineImage{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *MachineImage) Default() {
	machineimagelog.Info("default", "name", r.Name)

	// inject the Kubernetes version into the job template spec
	injectKubernetesVersionEnv(&r.Spec.JobTemplate.Spec, r.Spec.Version)
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//nolint:lll // This is generated code.
//+kubebuilder:webhook:path=/validate-kubernetesupgraded-dimitrikoshkin-com-v1alpha1-machineimage,mutating=false,failurePolicy=fail,sideEffects=None,groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages,verbs=create;update,versions=v1alpha1,name=vmachineimage.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &MachineImage{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *MachineImage) ValidateCreate() (admission.Warnings, error) {
	machineimagelog.Info("validate create", "name", r.Name)

	if r.Spec.Version == "" {
		return nil,
			//nolint:goerr113 // This is a user facing error.
			fmt.Errorf("spec.version: Required value")
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *MachineImage) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	machineimagelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *MachineImage) ValidateDelete() (admission.Warnings, error) {
	machineimagelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}

func injectKubernetesVersionEnv(spec *corev1.PodSpec, kubernetesVersion string) {
	envVar := corev1.EnvVar{
		Name:  kubernetesVersionEnv,
		Value: kubernetesVersion,
	}
	for i := range spec.InitContainers {
		spec.InitContainers[i].Env = append(spec.InitContainers[i].Env, envVar)
	}
	for i := range spec.Containers {
		spec.Containers[i].Env = append(spec.Containers[i].Env, envVar)
	}
}
