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

// PlanSpec defines the desired state of Plan.
type PlanSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// VersionRange gives a semver range of the Kubernetes version.
	// The cluster will be upgraded to the highest version within the range.
	// +required
	// +kubebuilder:validation:MinLength=1
	VersionRange string `json:"versionRange"`

	// MachineImageSelector can be used to select MachineImages to apply to the cluster plan.
	// Defaults to the empty LabelSelector, which matches all objects.
	// +optional
	MachineImageSelector *metav1.LabelSelector `json:"machineImageSelector,omitempty"`
}

// PlanStatus defines the observed state of Plan.
type PlanStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Phase represents the current phase of image building
	// E.g. NoSuitableMachineImage, FoundMachineImage.
	// +optional
	Phase PlanPhase `json:"phase,omitempty"`

	// MachineImageDetails holds the details for a MachineImage with the highest version within the range.
	MachineImageDetails *MachineImageDetails `json:"machineImageDetails"`

	// MachineImageRef is a reference to the MachineImage that was applied to the cluster upgrade.
	MachineImageRef *corev1.ObjectReference `json:"machineImageRef"`
}

type MachineImageDetails struct {
	// Version is the version of the Kubernetes image to build
	Version string `json:"version,omitempty"`

	// ID is the unique name or another identifier of the image that was built.
	ID string `json:"id"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-upgrader
//+kubebuilder:printcolumn:name="Version Range",type="string",JSONPath=`.spec.versionRange`
//+kubebuilder:printcolumn:name="Latest Version",type="string",JSONPath=`.status.machineImageDetails.version`
//+kubebuilder:printcolumn:name="Phase",type="string",JSONPath=`.status.phase`

// Plan is the Schema for the plans API.
type Plan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlanSpec   `json:"spec,omitempty"`
	Status PlanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlanList contains a list of Plan.
type PlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Plan `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&Plan{}, &PlanList{})
}
