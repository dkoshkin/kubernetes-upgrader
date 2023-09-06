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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
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
		// TODO(dkoshkin): Do all these labels make sense?
		"app.kubernetes.io/name":       "machineimage",
		"app.kubernetes.io/instance":   machineImageSyncer.Name,
		"app.kubernetes.io/part-of":    "kubernetes-upgrader",
		"app.kubernetes.io/created-by": "kubernetes-upgrader",
		"kubernetesVersion":            kubernetesVersion,
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
		logger.Error(err, "unable to list MachineImages for version %s")
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"ListMachineImageFailed",
			"Unable list MachineImage for version %s",
			kubernetesVersion,
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Create a new MachineImage if none already exist for the given version
	if len(machineImages.Items) == 0 {
		objectMeta := metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", machineImageSyncer.Name, kubernetesVersion),
			Namespace:    machineImageSyncer.Namespace,
			Labels:       labels,
			// TODO(dkoshkin): Do we want to delete the MachinreImage objects when MachineImageSyncer is deleted?
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
		// TODO(dkoshkin): This probably has some upstream utility
		for k, v := range machineImageSyncer.Spec.Template.ObjectMeta.Labels {
			objectMeta.Labels[k] = v
		}
		for k, v := range machineImageSyncer.Spec.Template.ObjectMeta.Annotations {
			objectMeta.Annotations[k] = v
		}

		machineImage := &kubernetesupgraderv1.MachineImage{
			ObjectMeta: objectMeta,
			Spec: kubernetesupgraderv1.MachineImageSpec{
				Version:     kubernetesVersion,
				JobTemplate: machineImageSyncer.Spec.Template.Spec.JobTemplate,
			},
		}
		err = r.Create(ctx, machineImage)
		if err != nil {
			logger.Error(err, "unable to create MachineImage for version %s")
			r.Recorder.Eventf(
				machineImageSyncer,
				corev1.EventTypeWarning,
				"CreateMachineImageFailed",
				"Unable create MachineImage for version %s",
				kubernetesVersion,
				err,
			)
			//nolint:wrapcheck // No additional context to add.
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeNormal,
			"CreatedMachineImage",
			"Created MachineImage %q for version %s",
			machineImage.Name,
			kubernetesVersion,
		)
		machineImageSyncer.Status.LatestVersion = kubernetesVersion
	}

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
