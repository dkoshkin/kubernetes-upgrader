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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KubernetesMachineImageSpec defines the desired state of KubernetesMachineImage.
type KubernetesMachineImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version is the version of the Kubernetes image to build
	Version string `json:"version"`

	// ImageID is the ID of the image that was built
	// +optional
	ImageID string `json:"imageID,omitempty"`

	// JobTemplate is the template for the job that builds the image
	JobTemplate JobTemplate `json:"jobTemplate"`
}

// JobTemplate defines the template for the job that builds the image.
type JobTemplate struct {
	// Spec is the spec for the job that builds the image
	Spec corev1.PodSpec `json:"spec"`
}

// KubernetesMachineImageStatus defines the observed state of KubernetesMachineImage.
type KubernetesMachineImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready indicates if the image has been built
	// +optional
	Ready bool `json:"ready"`

	// Phase represents the current phase of image building
	// E.g. Building, Created, Failed.
	// +optional
	Phase KubernetesMachineImagePhase `json:"phase,omitempty"`

	// JobRef is a reference to the job that builds the image
	// +optional
	JobRef *corev1.ObjectReference `json:"jobRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KubernetesMachineImage is the Schema for the kubernetesmachineimages API.
type KubernetesMachineImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubernetesMachineImageSpec   `json:"spec,omitempty"`
	Status KubernetesMachineImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KubernetesMachineImageList contains a list of KubernetesMachineImage.
type KubernetesMachineImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KubernetesMachineImage `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&KubernetesMachineImage{}, &KubernetesMachineImageList{})
}
