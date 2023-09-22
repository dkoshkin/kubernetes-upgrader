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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DebianRepositorySourceSpec defines the desired state of DebianRepositorySource.
type DebianRepositorySourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// URL is the URL of the Debian repository Packages file.
	// +required
	// +kubebuilder:validation:MinLength=1
	URL string `json:"url"`

	// Set Architecture if the Packages file contains multiple architectures.
	// Otherwise, leave it empty.
	// +optional
	Architecture string `json:"architecture,omitempty"`
}

// DebianRepositorySourceStatus defines the observed state of DebianRepositorySource.
type DebianRepositorySourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Version is the list of Kubernetes versions available in the Debian repository.
	Versions []string `json:"versions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-api
//+kubebuilder:printcolumn:name="Versions",type="string",JSONPath=`.status.versions`

// DebianRepositorySource is the Schema for the debianrepositorysources API.
type DebianRepositorySource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DebianRepositorySourceSpec   `json:"spec,omitempty"`
	Status DebianRepositorySourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DebianRepositorySourceList contains a list of DebianRepositorySource.
type DebianRepositorySourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DebianRepositorySource `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&DebianRepositorySource{}, &DebianRepositorySourceList{})
}
