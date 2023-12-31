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

	corev1 "k8s.io/api/core/v1"
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

	// Paused can be used to prevent controllers from processing the object.
	// +optional
	Paused bool `json:"paused,omitempty"`

	// ClusterName is the name of the cluster to upgrade.
	// +required
	// +kubebuilder:validation:MinLength=1
	ClusterName string `json:"clusterName"`

	// TopologyVariable is the name of the topology variable to set with the MachineImage's ID.
	// +optional
	TopologyVariable *string `json:"topologyVariable,omitempty"`

	// PlanRef is a reference to a Plan object.
	// +required
	PlanRef corev1.ObjectReference `json:"planRef"`
}

func (s *InClusterUpgradeAutomationSpec) GetPlan(
	ctx context.Context,
	reader client.Reader,
) (*Plan, error) {
	plan := &Plan{}
	key := client.ObjectKey{
		Name:      s.PlanRef.Name,
		Namespace: s.PlanRef.Namespace,
	}
	err := reader.Get(ctx, key, plan)
	if err != nil {
		return nil, fmt.Errorf("error getting Plan for InClusterUpgradeAutomation: %w", err)
	}

	return plan, nil
}

// InClusterUpgradeAutomationStatus defines the observed state of InClusterUpgradeAutomation.
type InClusterUpgradeAutomationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LatestFoundVersion is the highest version within the range that was set on the Cluster.
	// +optional
	LatestSetVersion string `json:"latestSetVersion,omitempty"`
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
	cluster := &clusterv1.Cluster{}
	key := client.ObjectKey{
		Name:      r.Spec.ClusterName,
		Namespace: r.GetNamespace(),
	}
	err := reader.Get(ctx, key, cluster)
	if err != nil {
		return nil, fmt.Errorf("error getting Cluster for InClusterUpgradeAutomation: %w", err)
	}

	return cluster, nil
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
