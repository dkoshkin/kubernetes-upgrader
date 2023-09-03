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
	"errors"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubernetesupgradedv1alpha1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/jobs"
)

const (
	imageBuilderJobRequeueDelay = 1 * time.Minute
	imageBuilderJobRequeueNow   = 1 * time.Second
)

// MachineImageReconciler reconciles a MachineImage object.
type MachineImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *MachineImageReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("machineimage", req.Name, "namespace", req.Namespace)

	machineImage := &kubernetesupgradedv1alpha1.MachineImage{}
	if err := r.Get(ctx, req.NamespacedName, machineImage); err != nil {
		logger.Error(err, "unable to fetch MachineImage")
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(machineImage, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the PreprovisionedMachine object and status after each reconciliation.
	defer func() {
		if err := patchMachineImage(ctx, patchHelper, machineImage); err != nil {
			logger.Error(err, "failed to patch MachineImage")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !machineImage.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, machineImage)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, machineImage)
}

func (r *MachineImageReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	machineImage *kubernetesupgradedv1alpha1.MachineImage,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	logger.Info("Checking if job already exists")
	if machineImage.Status.JobRef != nil {
		return r.handleJob(ctx, logger, machineImage, jobs.NewJobManager(r.Client))
	}

	if machineImage.Spec.ImageID != "" {
		logger.Info("Image already built, nothing to do")
		machineImage.Status.Ready = true
		machineImage.Status.Phase = kubernetesupgradedv1alpha1.MachineImagePhaseCreated
		return ctrl.Result{}, nil
	}

	jobManager := jobs.NewJobManager(r.Client)

	logger.Info("Job reference not set, creating a new job")
	spec := &machineImage.Spec.JobTemplate.Spec
	jobRef, err := jobManager.Create(ctx, machineImage, spec)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	machineImage.Status.JobRef = jobRef
	machineImage.Status.Phase = kubernetesupgradedv1alpha1.MachineImagePhaseBuilding

	return ctrl.Result{}, nil
}

func (r *MachineImageReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgradedv1alpha1.MachineImage,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func (r *MachineImageReconciler) handleJob(
	ctx context.Context,
	logger logr.Logger,
	machineImage *kubernetesupgradedv1alpha1.MachineImage,
	jobManager jobs.JobManager,
) (ctrl.Result, error) {
	logger.Info("Job reference set, checking status")
	status, imageID, err := jobManager.Status(ctx, machineImage.Status.JobRef)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Referenced Job not found, setting reference to nil to recreate")
			machineImage.Status.JobRef = nil
			return ctrl.Result{RequeueAfter: imageBuilderJobRequeueNow}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	switch {
	case status.Active > 0:
		logger.Info("Job is still active, requeuing")
		machineImage.Status.Phase = kubernetesupgradedv1alpha1.MachineImagePhaseBuilding
		return ctrl.Result{RequeueAfter: imageBuilderJobRequeueDelay}, nil
	case status.Succeeded > 0:
		if imageID == "" {
			logger.Info("Job succeeded but Image ID is empty, will not requeue")
			//nolint:goerr113 // This is a user facing error.
			return ctrl.Result{}, errors.New(
				"job completed but image id is empty, delete the job to retry",
			)
		}
		logger.Info(fmt.Sprintf("Job succeeded, updating with image id: %s", imageID))
		machineImage.Spec.ImageID = imageID
		machineImage.Status.Ready = true
		machineImage.Status.Phase = kubernetesupgradedv1alpha1.MachineImagePhaseCreated
		return ctrl.Result{}, nil
	case status.Failed > 0:
		logger.Info("Job failed, will not requeue")
		machineImage.Status.Phase = kubernetesupgradedv1alpha1.MachineImagePhaseFailed
		//nolint:goerr113 // This is a user facing error.
		return ctrl.Result{}, errors.New("job failed, delete the job to retry")
	}

	return ctrl.Result{}, nil
}

func patchMachineImage(
	ctx context.Context,
	patchHelper *patch.Helper,
	machineImage *kubernetesupgradedv1alpha1.MachineImage,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		machineImage,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgradedv1alpha1.MachineImage{}).
		Complete(r)
}
