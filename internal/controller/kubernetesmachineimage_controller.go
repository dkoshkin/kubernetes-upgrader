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

// KubernetesMachineImageReconciler reconciles a KubernetesMachineImage object.
type KubernetesMachineImageReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=kubernetesmachineimages,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=kubernetesmachineimages/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=kubernetesmachineimages/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *KubernetesMachineImageReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("kubernetesmachineimage", req.Name, "namespace", req.Namespace)

	kubernetesMachineImage := &kubernetesupgradedv1alpha1.KubernetesMachineImage{}
	if err := r.Get(ctx, req.NamespacedName, kubernetesMachineImage); err != nil {
		logger.Error(err, "unable to fetch KubernetesMachineImage")
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(kubernetesMachineImage, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the PreprovisionedMachine object and status after each reconciliation.
	defer func() {
		if err := patchKubernetesMachineImage(ctx, patchHelper, kubernetesMachineImage); err != nil {
			logger.Error(err, "failed to patch KubernetesMachineImage")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !kubernetesMachineImage.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, kubernetesMachineImage)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, kubernetesMachineImage)
}

func (r *KubernetesMachineImageReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	kubernetesMachineImage *kubernetesupgradedv1alpha1.KubernetesMachineImage,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	logger.Info("Checking if job already exists")
	if kubernetesMachineImage.Status.JobRef != nil {
		return r.handleJob(ctx, logger, kubernetesMachineImage, jobs.NewJobManager(r.Client))
	}

	if kubernetesMachineImage.Spec.ImageID != "" {
		logger.Info("Image already built, nothing to do")
		kubernetesMachineImage.Status.Ready = true
		kubernetesMachineImage.Status.Phase = kubernetesupgradedv1alpha1.KubernetesMachineImagePhaseCreated
		return ctrl.Result{}, nil
	}

	jobManager := jobs.NewJobManager(r.Client)

	logger.Info("Job reference not set, creating a new job")
	spec := &kubernetesMachineImage.Spec.JobTemplate.Spec
	jobRef, err := jobManager.Create(ctx, kubernetesMachineImage, spec)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	kubernetesMachineImage.Status.JobRef = jobRef
	kubernetesMachineImage.Status.Phase = kubernetesupgradedv1alpha1.KubernetesMachineImagePhaseBuilding

	return ctrl.Result{}, nil
}

func (r *KubernetesMachineImageReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgradedv1alpha1.KubernetesMachineImage,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func (r *KubernetesMachineImageReconciler) handleJob(
	ctx context.Context,
	logger logr.Logger,
	kubernetesMachineImage *kubernetesupgradedv1alpha1.KubernetesMachineImage,
	jobManager jobs.JobManager,
) (ctrl.Result, error) {
	logger.Info("Job reference set, checking status")
	status, imageID, err := jobManager.Status(ctx, kubernetesMachineImage.Status.JobRef)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Referenced Job not found, setting reference to nil to recreate")
			kubernetesMachineImage.Status.JobRef = nil
			return ctrl.Result{RequeueAfter: imageBuilderJobRequeueNow}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	switch {
	case status.Active > 0:
		logger.Info("Job is still active, requeuing")
		kubernetesMachineImage.Status.Phase = kubernetesupgradedv1alpha1.KubernetesMachineImagePhaseBuilding
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
		kubernetesMachineImage.Spec.ImageID = imageID
		return ctrl.Result{}, nil
	case status.Failed > 0:
		logger.Info("Job failed, will not requeue")
		kubernetesMachineImage.Status.Phase = kubernetesupgradedv1alpha1.KubernetesMachineImagePhaseFailed
		//nolint:goerr113 // This is a user facing error.
		return ctrl.Result{}, errors.New("job failed, delete the job to retry")
	}

	return ctrl.Result{}, nil
}

func patchKubernetesMachineImage(
	ctx context.Context,
	patchHelper *patch.Helper,
	kubernetesMachineImage *kubernetesupgradedv1alpha1.KubernetesMachineImage,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		kubernetesMachineImage,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubernetesMachineImageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgradedv1alpha1.KubernetesMachineImage{}).
		Complete(r)
}
