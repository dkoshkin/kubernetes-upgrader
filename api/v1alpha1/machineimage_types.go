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

	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MachineImageSpec defines the desired state of MachineImage.
type MachineImageSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version is the version of the Kubernetes image to build
	Version string `json:"version,omitempty"`

	// ID is the unique name or another identifier of the image that was built.
	// +optional
	ID string `json:"id,omitempty"`

	// JobTemplate is the template for the job that builds the image
	// +required
	// +kubebuilder:validation:Required
	JobTemplate JobTemplate `json:"jobTemplate"`
}

// JobTemplate defines the template for the job that builds the image.
type JobTemplate struct {
	// Spec is the spec for the job that builds the image
	Spec corev1.PodSpec `json:"spec"`
}

// MachineImageStatus defines the observed state of MachineImage.
type MachineImageStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Ready indicates if the image has been built
	// +optional
	Ready bool `json:"ready"`

	// Phase represents the current phase of image building
	// E.g. Building, Created, Failed.
	// +optional
	Phase MachineImagePhase `json:"phase,omitempty"`

	// JobRef is a reference to the job that built or is building the image.
	// +optional
	JobRef *corev1.ObjectReference `json:"jobRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-upgrader
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=`.spec.id`
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.ready`
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`

// MachineImage is the Schema for the machineimages API.
type MachineImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineImageSpec   `json:"spec,omitempty"`
	Status MachineImageStatus `json:"status,omitempty"`
}

// GetID returns the version of the MachineImage.
//

func (r *MachineImage) GetID() string {
	return r.Spec.ID
}

// GetVersion returns the version of the MachineImage.
//

func (r *MachineImage) GetVersion() string {
	return r.Spec.Version
}

// GetObjectReference returns a Kubernetes ObjectReference for this object.
//

func (r *MachineImage) GetObjectReference() *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:            r.Kind,
		APIVersion:      r.APIVersion,
		Namespace:       r.GetNamespace(),
		Name:            r.GetName(),
		UID:             r.GetUID(),
		ResourceVersion: r.GetResourceVersion(),
	}
}

//+kubebuilder:object:root=true

// MachineImageList contains a list of MachineImage.
type MachineImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineImage `json:"items"`
}

func MachineImagesToVersioned(versions []MachineImage) policy.VersionedList {
	result := make([]policy.Versioned, len(versions))
	for i := range versions {
		result[i] = &versions[i]
	}
	return result
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&MachineImage{}, &MachineImageList{})
}
