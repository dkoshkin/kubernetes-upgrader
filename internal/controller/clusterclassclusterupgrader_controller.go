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

	"github.com/Masterminds/semver/v3"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

const (
	clusterUpgraderRequeueDelay = 1 * time.Minute
)

// ClusterClassClusterUpgraderReconciler reconciles a ClusterClassClusterUpgrader object.
type ClusterClassClusterUpgraderReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=clusterclassclusterupgraders,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=clusterclassclusterupgraders/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=clusterclassclusterupgraders/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans,verbs=get;list;watch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterClassClusterUpgrader object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
//
//nolint:dupl // Prefer readability to DRY.
func (r *ClusterClassClusterUpgraderReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("clusterclassclusterupgrader", req.Name, "namespace", req.Namespace)

	clusterUpgrader := &kubernetesupgraderv1.ClusterClassClusterUpgrader{}
	if err := r.Get(ctx, req.NamespacedName, clusterUpgrader); err != nil {
		logger.Error(
			err,
			"unable to fetch ClusterClassClusterUpgrader",
			"namespace", req.Namespace, "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(clusterUpgrader, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the MachineImage object and status after each reconciliation.
	defer func() {
		if err := patchClusterClassClusterUpgrader(ctx, patchHelper, clusterUpgrader); err != nil {
			logger.Error(err, "failed to patch ClusterClassClusterUpgrader")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !clusterUpgrader.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, clusterUpgrader)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, clusterUpgrader)
}

//nolint:funlen // TODO(dkoshkin): Refactor.
func (r *ClusterClassClusterUpgraderReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	clusterUpgrader *kubernetesupgraderv1.ClusterClassClusterUpgrader,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	// Get the Cluster referenced by this Upgrader.
	cluster, err := clusterUpgrader.GetCluster(ctx, r.Client)
	if err != nil {
		r.Recorder.Eventf(
			clusterUpgrader,
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

	plan, err := clusterUpgrader.Spec.GetPlan(ctx, r.Client)
	if err != nil {
		r.Recorder.Eventf(
			clusterUpgrader,
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
			clusterUpgrader,
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
			clusterUpgrader,
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
			clusterUpgrader,
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
		clusterUpgrader.Status.LatestSetVersion = versionString(latestVersion)

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

	// Update the Cluster topology variable value if it is set.
	if clusterUpgrader.Spec.TopologyVariable != nil &&
		*clusterUpgrader.Spec.TopologyVariable != "" {
		logger.Info("Updating Cluster variable", "variable", *clusterUpgrader.Spec.TopologyVariable)
		err = updateTopologyVariableWithID(
			cluster.Spec.Topology.Variables,
			latestMachineImageDetails.ID,
			clusterUpgrader.Spec.TopologyVariable,
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

	// Update the status with the latest version set on the Cluster.
	clusterUpgrader.Status.LatestSetVersion = versionString(latestVersion)

	return ctrl.Result{}, nil
}

func (r *ClusterClassClusterUpgraderReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgraderv1.ClusterClassClusterUpgrader,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

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

func patchClusterClassClusterUpgrader(
	ctx context.Context,
	patchHelper *patch.Helper,
	clusterUpgrader *kubernetesupgraderv1.ClusterClassClusterUpgrader,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		clusterUpgrader,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterClassClusterUpgraderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // No additional context to add.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.ClusterClassClusterUpgrader{}).
		Watches(
			&kubernetesupgraderv1.Plan{},
			handler.EnqueueRequestsFromMapFunc(r.planMapper),
		).
		Complete(r)
}

//nolint:dupl // Prefer readability to DRY.
func (r *ClusterClassClusterUpgraderReconciler) planMapper(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	plan, ok := o.(*kubernetesupgraderv1.Plan)
	logger := log.FromContext(ctx).
		WithValues("plan", plan.Name, "namespace", plan.Namespace)

	if !ok {
		//nolint:goerr113 // This is a user facing error.
		logger.Error(fmt.Errorf("expected a Plan but got a %T", plan), "failed to reconcile object")
		return nil
	}

	upgraders := &kubernetesupgraderv1.ClusterClassClusterUpgraderList{}
	listOps := &client.ListOptions{
		Namespace: plan.GetNamespace(),
	}
	err := r.List(ctx, upgraders, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0)
	for i := range upgraders.Items {
		item := upgraders.Items[i]
		if item.Spec.PlanRef.Name == plan.GetName() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			})
		}
	}
	return requests
}

func versionString(version *semver.Version) string {
	if version == nil {
		return ""
	}
	return fmt.Sprintf("v%s", version.String())
}
