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

// MachineImageSyncerSpec defines the desired state of MachineImageSyncer.
type MachineImageSyncerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// VersionRange gives a semver range of the Kubernetes version.
	// The cluster will be upgraded to the highest version within the range.
	// +required
	// +kubebuilder:validation:MinLength=1
	VersionRange string `json:"versionRange"`

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

// MachineImageSyncerStatus defines the observed state of MachineImageSyncer.
type MachineImageSyncerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LatestVersion is the highest version within the range that was found.
	// +optional
	LatestVersion string `json:"latestVersion,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-api
//+kubebuilder:printcolumn:name="Latest Version",type="string",JSONPath=`.status.latestVersion`

// MachineImageSyncer is the Schema for the machineimagesyncers API.
type MachineImageSyncer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MachineImageSyncerSpec   `json:"spec,omitempty"`
	Status MachineImageSyncerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MachineImageSyncerList contains a list of MachineImageSyncer.
type MachineImageSyncerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MachineImageSyncer `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&MachineImageSyncer{}, &MachineImageSyncerList{})
}
