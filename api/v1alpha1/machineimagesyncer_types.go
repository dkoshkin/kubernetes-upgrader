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
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/cluster-api/controllers/external"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MachineImageSyncerSpec defines the desired state of MachineImageSyncer.
type MachineImageSyncerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// VersionRange gives a semver range of the Kubernetes version.
	// New MachineImages will be created for all versions within the range.
	// +required
	// +kubebuilder:validation:MinLength=1
	VersionRange string `json:"versionRange"`

	// SourceRef is a reference to a type that satisfies the Source contract.
	// Expects status.versions to be a list of version strings.
	// +required
	SourceRef corev1.ObjectReference `json:"sourceRef"`

	// MachineImageTemplateRef is a reference to a MachineImageTemplate object.
	// +required
	MachineImageTemplateRef corev1.ObjectReference `json:"machineImageTemplateRef"`

	// Interval is the time between checks for new versions from the source.
	// Defaults to 1h.
	// +optional
	Interval *time.Duration `json:"interval,omitempty"`
}

func (s *MachineImageSyncerSpec) GetVersionsFromSource(
	ctx context.Context,
	reader client.Reader,
) ([]string, error) {
	sourceRef := s.SourceRef
	source, err := external.Get(
		ctx,
		reader,
		&sourceRef,
		sourceRef.Namespace,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get object from ref %v %q: %w",
			sourceRef.GroupVersionKind(),
			sourceRef.Name,
			err,
		)
	}

	versionsField := []string{"status", "versions"}
	versions, _, err := unstructured.NestedStringSlice(source.Object, versionsField...)
	if err != nil {
		return nil, fmt.Errorf("error getting %s: %w", strings.Join(versionsField, "."), err)
	}

	return versions, nil
}

func (s *MachineImageSyncerSpec) GetMachineImageTemplate(
	ctx context.Context,
	reader client.Reader,
) (*MachineImageTemplate, error) {
	machineImageTemplate := &MachineImageTemplate{}
	key := client.ObjectKey{
		Name:      s.MachineImageTemplateRef.Name,
		Namespace: s.MachineImageTemplateRef.Namespace,
	}
	err := reader.Get(ctx, key, machineImageTemplate)
	if err != nil {
		return nil, fmt.Errorf("error getting MachineImageTemplate for MachineImageSyncer: %w", err)
	}

	return machineImageTemplate, nil
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
//+kubebuilder:resource:categories=kubernetes-upgrader
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
