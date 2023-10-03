// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GenericUpgradeAutomationSpec holds the common fields for all upgrade automation objects.
type GenericUpgradeAutomationSpec struct {
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
	PlanRef corev1.LocalObjectReference `json:"planRef"`
}

func (s *GenericUpgradeAutomationSpec) GetCluster(
	ctx context.Context,
	reader client.Reader,
	namespace string,
) (*clusterv1.Cluster, error) {
	cluster := &clusterv1.Cluster{}
	key := client.ObjectKey{
		Name:      s.ClusterName,
		Namespace: namespace,
	}
	err := reader.Get(ctx, key, cluster)
	if err != nil {
		return nil, fmt.Errorf("error getting Cluster for upgrade automation: %w", err)
	}

	return cluster, nil
}

func (s *GenericUpgradeAutomationSpec) GetPlan(
	ctx context.Context,
	reader client.Reader,
	namespace string,
) (*Plan, error) {
	plan := &Plan{}
	key := client.ObjectKey{
		Name:      s.PlanRef.Name,
		Namespace: namespace,
	}
	err := reader.Get(ctx, key, plan)
	if err != nil {
		return nil, fmt.Errorf("error getting Plan for upgrade automation: %w", err)
	}

	return plan, nil
}

func (s *GenericUpgradeAutomationSpec) GetTopologyVariable() *string {
	return s.TopologyVariable
}

// GenericUpgradeAutomationStatus holds the common status fields for all upgrade automation objects.
type GenericUpgradeAutomationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LatestFoundVersion is the highest version within the range that was set on the Cluster.
	// +optional
	LatestSetVersion string `json:"latestSetVersion,omitempty"`
}

func (s *GenericUpgradeAutomationStatus) SetStatus(latestVersion string) {
	s.LatestSetVersion = latestVersion
}
