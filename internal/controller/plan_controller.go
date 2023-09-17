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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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
	// Always attempt to Patch the Plan object and status after each reconciliation.
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

func (r *PlanReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	plan *kubernetesupgraderv1.Plan,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	// Get the latest version from MachineImages.
	latestVersion, err := latestMachineImageVersion(ctx, r.Client, logger, plan)
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

	// Update the status with the latest found version.
	plan.Status.MachineImageDetails = &kubernetesupgraderv1.MachineImageDetails{
		Version: latestVersion.GetVersion(),
		ID:      latestVersion.GetID(),
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

func latestMachineImageVersion(
	ctx context.Context,
	k8sClient client.Client,
	logger logr.Logger,
	plan *kubernetesupgraderv1.Plan,
) (*kubernetesupgraderv1.MachineImage, error) {
	logger.Info("Listing MachineImages for Plan")
	machineImages, err := machineImagesForPlan(ctx, k8sClient, plan)
	if err != nil {
		return nil, fmt.Errorf("error listing MachineImages for Plan: %w", err)
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
	latestVersion, err := policer.Latest(
		kubernetesupgraderv1.MachineImagesToVersioned(possibleMachineImages),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version from MachineImages: %w", err)
	}

	return latestVersion.(*kubernetesupgraderv1.MachineImage), nil
}

func machineImagesForPlan(
	ctx context.Context,
	k8sClient client.Client,
	plan *kubernetesupgraderv1.Plan,
) (*kubernetesupgraderv1.MachineImageList, error) {
	// List all MachineImages in the same namespace as the Plan.
	machineImages := &kubernetesupgraderv1.MachineImageList{}

	// Use the optional MachineImageSelector to filter the list of MachineImages.
	selector, err := metav1.LabelSelectorAsSelector(plan.Spec.MachineImageSelector)
	if err != nil {
		return nil, fmt.Errorf("unable to convert MachineImageSelector to selector: %w", err)
	}
	opts := []client.ListOption{
		client.InNamespace(plan.Namespace),
		client.MatchingLabelsSelector{Selector: selector},
	}
	err = k8sClient.List(ctx, machineImages, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to list MachineImages: %w", err)
	}

	return machineImages, nil
}

// createdMachineImages returns a list of MachineImages that have spec.ID set.
func machineImagesWithIDs(
	allMachines []kubernetesupgraderv1.MachineImage,
) []kubernetesupgraderv1.MachineImage {
	var machineImages []kubernetesupgraderv1.MachineImage
	for i := range allMachines {
		machineImage := allMachines[i]
		if machineImage.Spec.ID != "" {
			machineImages = append(machineImages, machineImage)
		}
	}
	return machineImages
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
		Watches(
			&kubernetesupgraderv1.MachineImage{},
			handler.EnqueueRequestsFromMapFunc(r.machineImageMapper),
		).
		Complete(r)
}

func (r *PlanReconciler) machineImageMapper(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	machineImage, ok := o.(*kubernetesupgraderv1.MachineImage)
	logger := log.FromContext(ctx).
		WithValues("machineImage", machineImage.Name, "namespace", machineImage.Namespace)

	if !ok {
		//nolint:goerr113 // This is a user facing error.
		logger.Error(
			fmt.Errorf("expected a MachineImage but got a %T", machineImage),
			"failed to reconcile object",
		)
		return nil
	}

	plans := &kubernetesupgraderv1.PlanList{}
	listOps := &client.ListOptions{
		Namespace: machineImage.GetNamespace(),
	}
	err := r.List(ctx, plans, listOps)
	if err != nil {
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(plans.Items))
	for i := range plans.Items {
		requests[i] = reconcile.Request{
			NamespacedName: client.ObjectKey{
				Namespace: plans.Items[i].Namespace,
				Name:      plans.Items[i].Name,
			},
		}
	}
	return requests
}
