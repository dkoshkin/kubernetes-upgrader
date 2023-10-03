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

// InClusterUpgradeAutomationSpec defines the desired state of InClusterUpgradeAutomation.
type InClusterUpgradeAutomationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	GenericUpgradeAutomationSpec `json:",inline"`
}

// InClusterUpgradeAutomationStatus defines the observed state of InClusterUpgradeAutomation.
type InClusterUpgradeAutomationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	GenericUpgradeAutomationStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=cluster-upgrader
//+kubebuilder:printcolumn:name="Cluster Name",type="string",JSONPath=`.spec.clusterName`
//+kubebuilder:printcolumn:name="Latest Set Version",type="string",JSONPath=`.status.latestSetVersion`

// InClusterUpgradeAutomation is the Schema for the inclusterupgradeautomations API.
type InClusterUpgradeAutomation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InClusterUpgradeAutomationSpec   `json:"spec,omitempty"`
	Status InClusterUpgradeAutomationStatus `json:"status,omitempty"`
}

func (r *InClusterUpgradeAutomation) GetCluster(
	ctx context.Context,
	reader client.Reader,
) (*clusterv1.Cluster, error) {
	return r.Spec.GetCluster(ctx, reader, r.Namespace)
}

func (r *InClusterUpgradeAutomation) GetPlan(
	ctx context.Context,
	reader client.Reader,
) (*Plan, error) {
	return r.Spec.GetPlan(ctx, reader, r.Namespace)
}

func (r *InClusterUpgradeAutomation) GetTopologyVariable() *string {
	return r.Spec.GetTopologyVariable()
}

func (r *InClusterUpgradeAutomation) SetStatus(latestVersion string) {
	r.Status.SetStatus(latestVersion)
}

//+kubebuilder:object:root=true

// InClusterUpgradeAutomationList contains a list of InClusterUpgradeAutomation.
type InClusterUpgradeAutomationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InClusterUpgradeAutomation `json:"items"`
}

//nolint:gochecknoinits // This is required for the Kubebuilder tooling.
func init() {
	SchemeBuilder.Register(&InClusterUpgradeAutomation{}, &InClusterUpgradeAutomationList{})
}
