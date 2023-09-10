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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MachineImageTemplateSpec defines the desired state of MachineImageTemplate.
type MachineImageTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Template is the MachineImage template to use when building the image.
	// +required
	Template MachineImageSyncerResource `json:"template"`
}

// MachineImageSyncerResource defines the Template structure.
type MachineImageSyncerResource struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	ObjectMeta clusterv1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the MachineImage spec to use when building the image.
	// +required
	Spec MachineImageSpec `json:"spec"`
}

// MachineImageTemplateStatus defines the observed state of MachineImageTemplate.
type MachineImageTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-api

// MachineImageTemplate is the Schema for the machineimagetemplates API.
type MachineImageTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineImageTemplateSpec   `json:"spec,omitempty"`
	Status MachineImageTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineImageTemplateList contains a list of MachineImageTemplate.
type MachineImageTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineImageTemplate `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&MachineImageTemplate{}, &MachineImageTemplateList{})
}
