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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InGitUpgradeAutomationSpec defines the desired state of InGitUpgradeAutomation.
type InGitUpgradeAutomationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	GenericUpgradeAutomationSpec `json:",inline"`

	// GitSpec contains all the git-specific definitions.
	// +required
	// +kubebuilder:validation:Required
	Git GitSpec `json:"git"`

	// Update gives the specification for how to update the files in the repository.
	// +required
	// +kubebuilder:validation:Required
	Update UpdateStrategy `json:"update"`
}

// UpdateStrategyName is the type for names that go in
// .update.strategy. NB the value in the const immediately below.
// +kubebuilder:validation:Enum=Setters
type UpdateStrategyName string

const (
	// UpdateStrategySetters is the name of the update strategy that
	// uses kyaml setters. NB the value in the enum annotation for the
	// type, above.
	UpdateStrategySetters UpdateStrategyName = "Setters"
)

// UpdateStrategy is a union of the various strategies for updating the Git repository.
// Parameters for each strategy (if any) can be inlined here.
type UpdateStrategy struct {
	// Strategy names the strategy to be used.
	// +required
	// +kubebuilder:default=Setters
	Strategy UpdateStrategyName `json:"strategy"`

	// Path to the directory containing the manifests to be updated.
	// Defaults to 'None', which translates to the root path of the GitRepositoryRef.
	// +optional
	Path string `json:"path,omitempty"`
}

// InGitUpgradeAutomationStatus defines the observed state of InGitUpgradeAutomation.
type InGitUpgradeAutomationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	GenericUpgradeAutomationStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-upgrader
//+kubebuilder:printcolumn:name="Cluster Name",type="string",JSONPath=`.spec.clusterName`
//+kubebuilder:printcolumn:name="Latest Set Version",type="string",JSONPath=`.status.latestSetVersion`

// InGitUpgradeAutomation is the Schema for the ingitupgradeautomations API.
type InGitUpgradeAutomation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InGitUpgradeAutomationSpec   `json:"spec,omitempty"`
	Status InGitUpgradeAutomationStatus `json:"status,omitempty"`
}

func (r *InGitUpgradeAutomation) GetCluster(
	ctx context.Context,
	reader client.Reader,
) (*clusterv1.Cluster, error) {
	return r.Spec.GetCluster(ctx, reader, r.Namespace)
}

func (r *InGitUpgradeAutomation) GetPlan(
	ctx context.Context,
	reader client.Reader,
) (*Plan, error) {
	return r.Spec.GetPlan(ctx, reader, r.Namespace)
}

func (r *InGitUpgradeAutomation) GetTopologyVariable() *string {
	return r.Spec.GetTopologyVariable()
}

func (r *InGitUpgradeAutomation) SetStatus(latestVersion string) {
	r.Status.SetStatus(latestVersion)
}

//+kubebuilder:object:root=true

// InGitUpgradeAutomationList contains a list of InGitUpgradeAutomation.
type InGitUpgradeAutomationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InGitUpgradeAutomation `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&InGitUpgradeAutomation{}, &InGitUpgradeAutomationList{})
}
