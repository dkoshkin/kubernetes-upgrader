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

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

const (
	planRequeueDelay = 1 * time.Minute
)

// PlanReconciler reconciles a Plan object.
type PlanReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans/finalizers,verbs=update
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
//
//nolint:dupl // Prefer readability over DRY.
func (r *PlanReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("plan", req.Name, "namespace", req.Namespace)

	plan := &kubernetesupgraderv1.Plan{}
	if err := r.Get(ctx, req.NamespacedName, plan); err != nil {
		logger.Error(
			err,
			"unable to fetch Plan",
			"namespace", req.Namespace, "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(plan, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the PreprovisionedMachine object and status after each reconciliation.
	defer func() {
		if err := patchPlan(ctx, patchHelper, plan); err != nil {
			logger.Error(err, "failed to patch Plan")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !plan.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, plan)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, plan)
}

//nolint:funlen // This function is mostly calling other functions.
func (r *PlanReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	plan *kubernetesupgraderv1.Plan,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	// Get the Cluster referenced by this Plan.
	cluster := &clusterv1.Cluster{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: plan.Namespace,
		Name:      plan.Spec.ClusterName,
	}, cluster); err != nil {
		logger.Error(err,
			"unable to fetch Cluster",
			"namespace", plan.Namespace, "cluster", plan.Spec.ClusterName)
		r.Recorder.Eventf(
			plan,
			corev1.EventTypeWarning,
			"GetClusterFailed",
			"Unable to get Cluster %q: %s",
			plan.Spec.ClusterName,
			plan.Name,
			err,
		)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{RequeueAfter: planRequeueDelay}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Get the latest version from MachineImages.
	latestVersion, err := r.latestMachineImageVersion(ctx, logger, plan)
	if err != nil {
		return ctrl.Result{}, err
	}

	if latestVersion == nil {
		logger.Info("Did not find a suitable MachineImage, requeueing")
		r.Recorder.Eventf(
			plan,
			corev1.EventTypeWarning,
			"LatestVersionNotFound",
			"Did not find a suitable MachineImage, requeueing",
		)
		return ctrl.Result{RequeueAfter: planRequeueDelay}, nil
	}

	// Update the status with the latest version.
	plan.Status.LatestVersion = latestVersion.GetVersion()

	// If the latest version is the same as the current version, there is nothing to do.
	if cluster.Spec.Topology.Version == latestVersion.GetVersion() {
		logger.Info("Cluster is already at the latest version", "version", latestVersion)
		r.Recorder.Eventf(
			plan,
			corev1.EventTypeNormal,
			"ClusterUsingLatestVersion",
			"Cluster is already using the latest version %q",
			latestVersion.GetVersion(),
		)
		return ctrl.Result{}, nil
	}

	// If the latest version is not the same as the current version, update the Cluster.
	logger.Info("Updating Cluster version", "version", latestVersion)
	r.Recorder.Eventf(
		plan,
		corev1.EventTypeNormal,
		"SettingClusterVersion",
		"Setting cluster to the latest version %q",
		latestVersion.GetVersion(),
	)
	cluster.Spec.Topology.Version = latestVersion.GetVersion()

	// Update the Cluster topology variable value if it is set.
	if plan.Spec.TopologyVariable != nil && *plan.Spec.TopologyVariable != "" {
		logger.Info("Updating Cluster variable", "variable", *plan.Spec.TopologyVariable)
		err = updateTopologyVariable(
			cluster.Spec.Topology.Variables,
			latestVersion,
			plan.Spec.TopologyVariable,
		)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf(
				"error updating Cluster topology variable value: %w",
				err,
			)
		}
	}

	// Update the Cluster after updating the version and topology variable value.
	err = r.Update(ctx, cluster)
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

	// Update the Plan status with the latest MachineImage.
	plan.Status.MachineImageRef = latestVersion.GetObjectReference()

	return ctrl.Result{}, nil
}

func (r *PlanReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgraderv1.Plan,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func (r *PlanReconciler) latestMachineImageVersion(
	ctx context.Context,
	logger logr.Logger,
	plan *kubernetesupgraderv1.Plan,
) (policy.VersionedObject, error) {
	// List all MachineImages in the same namespace as the Plan.
	// TODO(dkoshkin): Use a label selector to filter MachineImages.
	machineImages := &kubernetesupgraderv1.MachineImageList{}
	err := r.List(ctx, machineImages, client.InNamespace(plan.Namespace))
	if err != nil {
		return nil, fmt.Errorf("unable to list MachineImages: %w", err)
	}

	// Filter all MachineImages that have an ID set.
	possibleMachineImages := machineImagesWithIDs(machineImages.Items)
	if len(possibleMachineImages) == 0 {
		logger.Info("No MachineImages with IDs found")
		return nil, nil
	}

	policer, err := policy.NewSemVer(plan.Spec.VersionRange)
	if err != nil {
		return nil, fmt.Errorf("invalid versionRange policy: %w", err)
	}

	// Get the latest version from the MachineImages that match VersionRange.
	latestVersion, err := policer.Latest(possibleMachineImages)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version from MachineImages: %w", err)
	}

	return latestVersion, nil
}

// createdMachineImages returns a list of MachineImages that have spec.ID set.
func machineImagesWithIDs(
	allMachines []kubernetesupgraderv1.MachineImage,
) []policy.VersionedObject {
	var machineImages []policy.VersionedObject
	for i := range allMachines {
		machineImage := allMachines[i]
		if machineImage.Spec.ID != "" {
			machineImages = append(machineImages, &machineImage)
		}
	}
	return machineImages
}

// updateTopologyVariable will update the value of the variableToUpdate with the MachineImage ID.
func updateTopologyVariable(
	variables []clusterv1.ClusterVariable,
	latestVersion policy.VersionedObject,
	variableToUpdate *string,
) error {
	for i, variable := range variables {
		if variable.Name == pointer.StringDeref(variableToUpdate, "") {
			value, err := toJSON(latestVersion)
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

func patchPlan(
	ctx context.Context,
	patchHelper *patch.Helper,
	plan *kubernetesupgraderv1.Plan,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		plan,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.Plan{}).
		Complete(r)
}
