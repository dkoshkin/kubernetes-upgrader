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
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

const (
	machineImageSyncerGetTemplateRequeue        = 1 * time.Second
	machineImageSyncerGetMachineImagesRequeue   = 1 * time.Minute
	machineImageSyncerCreateMachineImageRequeue = 1 * time.Minute
)

const (
	MachineImageSyncerNameLabel              = "kubernetes-upgrader.dimitrikoshkin.com/machine-image-syncer-name"
	MachineImageSyncerKubernetesVersionLabel = "kubernetes-upgrader.dimitrikoshkin.com/kubernetes-version"

	TemplateClonedFromNameAnnotation = "kubernetes-upgrader.dimitrikoshkin.com/cloned-from"
)

// MachineImageSyncerReconciler reconciles a MachineImageSyncer object.
type MachineImageSyncerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimagesyncers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimagesyncers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimagesyncers/finalizers,verbs=update
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimagetemplates,verbs=get;list;watch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=machineimages,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MachineImageSyncer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
//
//nolint:dupl // Prefer readability over DRY.
func (r *MachineImageSyncerReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("machineImageSyncer", req.Name, "namespace", req.Namespace)

	machineImageSyncer := &kubernetesupgraderv1.MachineImageSyncer{}
	if err := r.Get(ctx, req.NamespacedName, machineImageSyncer); err != nil {
		logger.Error(
			err,
			"unable to fetch MachineImageSyncer",
			"namespace", req.Namespace, "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(machineImageSyncer, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the MachineImageSyncer object and status after each reconciliation.
	defer func() {
		if err := patchMachineImageSyncer(ctx, patchHelper, machineImageSyncer); err != nil {
			logger.Error(err, "failed to patch MachineImageSyncer")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !machineImageSyncer.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, machineImageSyncer)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, machineImageSyncer)
}

//nolint:funlen // TODO(dkoshkin): Refactor.
func (r *MachineImageSyncerReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	// FIXME(dkoshkin): Implement fetching packages from a source
	kubernetesVersion := "v1.27.5"

	labels := map[string]string{
		MachineImageSyncerNameLabel:              machineImageSyncer.Name,
		MachineImageSyncerKubernetesVersionLabel: kubernetesVersion,
	}
	machineImages := &kubernetesupgraderv1.MachineImageList{}
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: labels})
	if err != nil {
		logger.Error(err, "unable to convert labels to selector")
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"ConvertLabelsToSelectorFailed",
			"Unable to convert labels to selector",
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	opts := []client.ListOption{
		client.InNamespace(machineImageSyncer.Namespace),
		client.MatchingLabelsSelector{Selector: selector},
	}
	err = r.List(ctx, machineImages, opts...)
	if err != nil {
		logger.Error(err, "unable to list MachineImages", "version", kubernetesVersion)
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"ListMachineImageFailed",
			"Unable list MachineImage for Kubernetes version %s",
			kubernetesVersion,
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{RequeueAfter: machineImageSyncerGetMachineImagesRequeue}, err
	}

	// Create a new MachineImage if none already exist for the given version
	if len(machineImages.Items) == 0 {
		var machineImageTemplate *kubernetesupgraderv1.MachineImageTemplate
		machineImageTemplate, err = machineImageSyncer.Spec.GetMachineImageTemplate(ctx, r.Client)
		if err != nil {
			logger.Error(err, "unable to get MachineImageTemplate")
			r.Recorder.Eventf(
				machineImageSyncer,
				corev1.EventTypeWarning,
				"GetMachineImageTemplateFailed",
				"Unable to get MachineImageTemplate for MachineImageSyncer",
				err,
			)
			//nolint:wrapcheck // No additional context to add.
			return ctrl.Result{RequeueAfter: machineImageSyncerGetTemplateRequeue}, err
		}

		objectMeta := metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", machineImageSyncer.Name, kubernetesVersion),
			Namespace:    machineImageSyncer.Namespace,
			Labels:       labels,
			// TODO(dkoshkin): Do we want to delete the MachineImage objects when MachineImageSyncer is deleted?
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: machineImageSyncer.APIVersion,
					Kind:       machineImageSyncer.Kind,
					Name:       machineImageSyncer.Name,
					UID:        machineImageSyncer.UID,
				},
			},
		}
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = map[string]string{}
		}
		// Add an annotation to the MachineImage to indicate which MachineImageTemplate it was cloned from
		objectMeta.Annotations[TemplateClonedFromNameAnnotation] = machineImageTemplate.Name

		// TODO(dkoshkin): Check for an upstream utility
		for k, v := range machineImageTemplate.Spec.Template.ObjectMeta.Labels {
			objectMeta.Labels[k] = v
		}
		for k, v := range machineImageTemplate.Spec.Template.ObjectMeta.Annotations {
			objectMeta.Annotations[k] = v
		}

		machineImage := &kubernetesupgraderv1.MachineImage{
			ObjectMeta: objectMeta,
			Spec: kubernetesupgraderv1.MachineImageSpec{
				Version:     kubernetesVersion,
				JobTemplate: machineImageTemplate.Spec.Template.Spec.JobTemplate,
			},
		}
		return r.createMachineImageAndWait(ctx, logger, machineImageSyncer, machineImage)
	}

	return ctrl.Result{}, nil
}

// createMachineImageAndWait creates a new MachineImage.
// It waits for the cache to be updated with the newly created MachineImage.
//
//nolint:funlen // TODO(dkoshkin): Refactor to a shorter functions.
func (r *MachineImageSyncerReconciler) createMachineImageAndWait(
	ctx context.Context,
	logger logr.Logger,
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
	machineImage *kubernetesupgraderv1.MachineImage,
) (ctrl.Result, error) {
	kubernetesVersion := machineImage.Spec.Version
	err := r.Create(ctx, machineImage)
	if err != nil {
		logger.Error(err, "unable to create MachineImage", "version", kubernetesVersion)
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"CreateMachineImageFailed",
			"Unable create MachineImage for Kubernetes version %s",
			kubernetesVersion,
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{RequeueAfter: machineImageSyncerCreateMachineImageRequeue}, err
	}

	logger.Info("Waiting for MachineImage to be created", "version", kubernetesVersion)
	// Keep trying to get the MachineImage.
	// This will force the cache to update and prevent any future reconciliation of the MachineImageSyncer
	// to reconcile with an outdated list of MachineImage,
	// which could lead to unwanted creation of a duplicate MachineImage.
	const (
		interval = 100 * time.Millisecond
		timeout  = 10 * time.Second
	)
	var pollErrors []error
	if err = wait.PollUntilContextTimeout(
		ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
			if err = r.Client.Get(
				ctx,
				client.ObjectKeyFromObject(machineImage),
				&kubernetesupgraderv1.MachineImage{},
			); err != nil {
				// Do not return error here. Continue to poll even if we hit an error
				// so that we avoid exiting because of transient errors like network flakes.
				// Capture all the errors and return the aggregate error if the poll fails eventually.
				pollErrors = append(pollErrors, err)
				return false, nil
			}
			return true, nil
		}); err != nil {
		return ctrl.Result{},
			errors.Wrapf(
				kerrors.NewAggregate(pollErrors),
				"failed to get the MachineImage %s after creation", machineImage.Name)
	}

	r.Recorder.Eventf(
		machineImageSyncer,
		corev1.EventTypeNormal,
		"CreatedMachineImage",
		"Created MachineImage %q for Kubernetes version %s",
		machineImage.Name,
		kubernetesVersion,
	)
	machineImageSyncer.Status.LatestVersion = kubernetesVersion

	return ctrl.Result{}, nil
}

func (r *MachineImageSyncerReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgraderv1.MachineImageSyncer,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func patchMachineImageSyncer(
	ctx context.Context,
	patchHelper *patch.Helper,
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		machineImageSyncer,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MachineImageSyncerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.MachineImageSyncer{}).
		Complete(r)
}
