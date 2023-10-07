// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

type clusterUpgrader interface {
	UpgradeCluster(ctx context.Context, logger logr.Logger, cluster *clusterv1.Cluster) error

	GetCluster(ctx context.Context, reader client.Reader) (*clusterv1.Cluster, error)
	GetPlan(ctx context.Context, reader client.Reader) (*kubernetesupgraderv1.Plan, error)

	GetTopologyVariable() *string
	SetStatus(latestVersion string)

	// implements runtime.Object
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() runtime.Object
}

// genericUpgradeAutomationReconciler is a generic Cluster upgrade reconciler.
type genericUpgradeAutomationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:funlen // TODO(dkoshkin): Refactor.
func (r *genericUpgradeAutomationReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	upgrader clusterUpgrader,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	// Get the Cluster referenced by this Upgrader.
	cluster, err := upgrader.GetCluster(ctx, r.Client)
	if err != nil {
		r.Recorder.Eventf(
			upgrader,
			corev1.EventTypeWarning,
			"GetClusterFailed",
			"Unable to get Cluster: %s",
			err,
		)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: clusterUpgraderRequeueDelay}, nil
		}

		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	plan, err := upgrader.GetPlan(ctx, r.Client)
	if err != nil {
		r.Recorder.Eventf(
			upgrader,
			corev1.EventTypeWarning,
			"GetPlanFailed",
			"Unable to get Plan: %s",
			err,
		)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	latestMachineImageDetails := plan.Status.MachineImageDetails

	if latestMachineImageDetails == nil || latestMachineImageDetails.Version == "" {
		r.Recorder.Eventf(
			upgrader,
			corev1.EventTypeWarning,
			"MachineImageDetailsVersionNotSet",
			"status.MachineImageDetails.Version is not set on the Plan, requeueing",
		)
		return ctrl.Result{}, nil
	}

	latestVersionString := latestMachineImageDetails.Version
	latestVersion, err := semver.NewVersion(latestVersionString)
	if err != nil {
		r.Recorder.Eventf(
			upgrader,
			corev1.EventTypeWarning,
			"ErrorParsingLatestFoundVersion",
			"status.LatestFoundVersion %q is not a valid semver version, requeueing",
			latestVersionString,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	currentVersionString := cluster.Spec.Topology.Version
	currentVersion, err := semver.NewVersion(currentVersionString)
	if err != nil {
		r.Recorder.Eventf(
			upgrader,
			corev1.EventTypeWarning,
			"ErrorParsingClusterVersion",
			"cluster.Spec.Topology.Version %q is not a valid semver version, requeueing",
			currentVersion,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// If the latest version is not greater than the current version, do nothing.
	if !latestVersion.GreaterThan(currentVersion) {
		logger.Info("Cluster is at the latest version", "version", currentVersion)
		r.Recorder.Eventf(
			plan,
			corev1.EventTypeNormal,
			"ClusterUsingLatestVersion",
			"Cluster is already using the latest version %q",
			currentVersion,
		)

		// Update the status with the latest version set on the Cluster even if it was already set.
		upgrader.SetStatus(versionString(latestVersion))

		return ctrl.Result{}, nil
	}

	// If the latest version is not the same as the current version, update the Cluster.
	logger.Info("Updating Cluster version", "version", latestVersion)
	r.Recorder.Eventf(
		plan,
		corev1.EventTypeNormal,
		"SettingClusterVersion",
		"Setting cluster to the latest version %q",
		latestVersion,
	)
	cluster.Spec.Topology.Version = versionString(latestVersion)

	topologyVariable := upgrader.GetTopologyVariable()
	// Update the Cluster topology variable value if it is set.
	if topologyVariable != nil &&
		*topologyVariable != "" {
		logger.Info("Updating Cluster variable", "variable", *topologyVariable)
		err = updateTopologyVariableWithID(
			cluster.Spec.Topology.Variables,
			latestMachineImageDetails.ID,
			topologyVariable,
		)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf(
				"error updating Cluster topology variable value: %w",
				err,
			)
		}
	}

	// Update the Cluster after updating the version and topology variable value.
	err = upgrader.UpgradeCluster(ctx, logger, cluster)
	if err != nil {
		r.Recorder.Eventf(
			plan,
			corev1.EventTypeWarning,
			"ErrorUpdatingVersion",
			"Error updating Cluster with a new version: %s",
			err,
		)
		return ctrl.Result{}, fmt.Errorf("error updating Cluster with a new version: %w", err)
	}

	// Update the status with the latest version set on the Cluster.
	upgrader.SetStatus(versionString(latestVersion))

	return ctrl.Result{}, nil
}

// updateTopologyVariableWithID will update the value of the variableToUpdate with the MachineImage ID.
func updateTopologyVariableWithID(
	variables []clusterv1.ClusterVariable,
	id string,
	variableToUpdate *string,
) error {
	for i, variable := range variables {
		if variable.Name == pointer.StringDeref(variableToUpdate, "") {
			value, err := toJSON(id)
			if err != nil {
				return fmt.Errorf("failed to marshal latest version: %w", err)
			}
			variables[i].Value = *value
			break
		}
	}

	return nil
}

func toJSON(obj interface{}) (*v1.JSON, error) {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal variable value: %w", err)
	}
	return &v1.JSON{Raw: marshaled}, nil
}

func versionString(version *semver.Version) string {
	if version == nil {
		return ""
	}
	return fmt.Sprintf("v%s", version.String())
}
