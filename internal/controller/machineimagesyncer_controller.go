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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/kubernetes"
	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

const (
	machineImageSyncerGetVersionsFromSourceRequeue = 1 * time.Minute
	machineImageSyncerCreateMachineImageRequeue    = 1 * time.Minute
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
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch

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
//nolint:dupl // Prefer readability to DRY.
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

	if machineImageSyncer.Spec.Paused {
		logger.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

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

	latestVersion, err := latestVersionFromSource(ctx, r.Client, machineImageSyncer)
	if err != nil {
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"GetVersionsFromSourceFailed",
			"Unable to get latest version from source",
			err,
		)
		return ctrl.Result{RequeueAfter: machineImageSyncerGetVersionsFromSourceRequeue}, err
	}

	latestVersionString := latestVersion.GetVersion()
	if latestVersionString == "" {
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"NoLatestVersionFromSource",
			"No version found from source",
		)
		return ctrl.Result{RequeueAfter: machineImageSyncerGetVersionsFromSourceRequeue}, nil
	}

	// These labels will be used to find MachineImages created by this MachineImageSyncer.
	labels := map[string]string{
		MachineImageSyncerNameLabel:              machineImageSyncer.Name,
		MachineImageSyncerKubernetesVersionLabel: latestVersionString,
	}
	machineImages, err := listMachineImagesWithLabels(
		ctx,
		r.Client,
		machineImageSyncer.Namespace,
		labels,
	)
	if err != nil {
		logger.Error(err, "unable to list MachineImages", "version", latestVersionString)
		r.Recorder.Eventf(
			machineImageSyncer,
			corev1.EventTypeWarning,
			"ListMachineImageFailed",
			"Unable list MachineImage for Kubernetes version %s",
			latestVersionString,
			err,
		)

		return ctrl.Result{}, err
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
			return ctrl.Result{}, err
		}

		// Build a new MachineImage for the given version.
		machineImage := machineImageForKubernetesVersion(
			machineImageSyncer,
			machineImageTemplate,
			labels,
			latestVersionString,
			r.Scheme,
		)
		return r.reconcileCreateMachineImageAndWait(ctx, logger, machineImageSyncer, machineImage)
	}

	// TODO(dkoshkin) how to watch for changes to SourceRef?
	return ctrl.Result{
		RequeueAfter: pointer.DurationDeref(
			machineImageSyncer.Spec.Interval,
			kubernetesupgraderv1.MachineImageSyncerInterval),
	}, nil
}

func latestVersionFromSource(
	ctx context.Context,
	k8sClient client.Client,
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
) (policy.Versioned, error) {
	versions, err := machineImageSyncer.Spec.GetVersionsFromSource(ctx, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("error getting versions from source: %w", err)
	}

	policer, err := policy.NewSemVer(machineImageSyncer.Spec.VersionRange)
	if err != nil {
		return nil, fmt.Errorf("invalid versionRange policy: %w", err)
	}

	latestVersion, err := policer.Latest(policy.VersionedStrings(versions...))
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version from source: %w", err)
	}

	return latestVersion, nil
}

func listMachineImagesWithLabels(
	ctx context.Context,
	k8sClient client.Client,
	namespace string,
	labels map[string]string,
) (*kubernetesupgraderv1.MachineImageList, error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: labels})
	if err != nil {
		return nil, fmt.Errorf("error converting labels to selector: %w", err)
	}

	machineImages := &kubernetesupgraderv1.MachineImageList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: selector},
	}
	err = k8sClient.List(ctx, machineImages, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to list MachineImages with labels %v: %w", labels, err)
	}

	return machineImages, nil
}

func machineImageForKubernetesVersion(
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
	machineImageTemplate *kubernetesupgraderv1.MachineImageTemplate,
	labels map[string]string,
	kubernetesVersion string,
	scheme *runtime.Scheme,
) *kubernetesupgraderv1.MachineImage {
	objectMeta := metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", machineImageSyncer.Name, kubernetesVersion),
		Namespace:    machineImageSyncer.Namespace,
		Labels:       labels,
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

	// TODO(dkoshkin): Do we want to delete the MachineImage objects when MachineImageSyncer is deleted?
	// This function is generating a new MachineImage object, shouldn never error.
	_ = controllerutil.SetControllerReference(machineImageSyncer, machineImage, scheme)

	return machineImage
}

func (r *MachineImageSyncerReconciler) reconcileCreateMachineImageAndWait(
	ctx context.Context,
	logger logr.Logger,
	machineImageSyncer *kubernetesupgraderv1.MachineImageSyncer,
	machineImage *kubernetesupgraderv1.MachineImage,
) (ctrl.Result, error) {
	kubernetesVersion := machineImage.Spec.Version
	err := kubernetes.CreateAndWait(ctx, r.Client, machineImage)
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
func (r *MachineImageSyncerReconciler) SetupWithManager(
	ctx context.Context,
	mgr ctrl.Manager,
) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.MachineImageSyncer{}).
		WithEventFilter(ResourceNotPaused(ctrl.LoggerFrom(ctx))).
		Owns(&kubernetesupgraderv1.MachineImage{}).
		Watches(
			&kubernetesupgraderv1.MachineImageTemplate{},
			handler.EnqueueRequestsFromMapFunc(r.machineImageTemplateMapper),
		).
		Complete(r)
}

//nolint:dupl // Prefer readability to DRY.
func (r *MachineImageSyncerReconciler) machineImageTemplateMapper(
	ctx context.Context,
	o client.Object,
) []reconcile.Request {
	template, ok := o.(*kubernetesupgraderv1.MachineImageTemplate)
	logger := log.FromContext(ctx).
		WithValues("plan", template.Name, "namespace", template.Namespace)

	if !ok {
		//nolint:goerr113 // This is a user facing error.
		logger.Error(
			fmt.Errorf("expected a MachineImageTemplate but got a %T", template),
			"failed to reconcile object",
		)
		return nil
	}

	syncers := &kubernetesupgraderv1.MachineImageSyncerList{}
	listOps := &client.ListOptions{
		Namespace: template.GetNamespace(),
	}
	err := r.List(ctx, syncers, listOps)
	if err != nil {
		return []reconcile.Request{}
	}

	requests := make([]reconcile.Request, 0)
	for i := range syncers.Items {
		item := syncers.Items[i]
		if item.Spec.MachineImageTemplateRef.Name == template.GetName() {
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
