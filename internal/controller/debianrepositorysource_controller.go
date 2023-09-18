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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/kubernetesversions/debian"
)

const (
	debianRepositorySourceRequeueDelay = 1 * time.Minute
)

// DebianRepositorySourceReconciler reconciles a DebianRepositorySource object.
type DebianRepositorySourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=debianrepositorysources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=debianrepositorysources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=debianrepositorysources/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DebianRepositorySource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
//
//nolint:dupl // Prefer readability over DRY.
func (r *DebianRepositorySourceReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("debianrepositorysource", req.Name, "namespace", req.Namespace)

	source := &kubernetesupgraderv1.DebianRepositorySource{}
	if err := r.Get(ctx, req.NamespacedName, source); err != nil {
		logger.Error(
			err,
			"unable to fetch DebianRepositorySource",
			"namespace", req.Namespace, "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(source, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the DebianRepositorySource object and status after each reconciliation.
	defer func() {
		if err := patchDebianRepositorySource(ctx, patchHelper, source); err != nil {
			logger.Error(err, "failed to patch DebianRepositorySource")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	// Handle deleted clusters
	if !source.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, source)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, source)
}

func (r *DebianRepositorySourceReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	source *kubernetesupgraderv1.DebianRepositorySource,
) (ctrl.Result, error) {
	logger.Info("Reconciling normal")

	opts := debian.Options{Architecture: source.Spec.Architecture}
	versionsSource := debian.NewHTTPSource(source.Spec.URL, opts)
	versionedList, err := versionsSource.List(ctx)
	if err != nil {
		r.Recorder.Eventf(
			source,
			corev1.EventTypeWarning,
			"ListVersionsFailed",
			"Unable to list versions for source: %s",
			err,
		)

		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{RequeueAfter: debianRepositorySourceRequeueDelay}, err
	}

	fmt.Printf("versionsSource: %+v\n", versionedList)

	versions := make([]string, len(versionedList))
	for i := range versionedList {
		versions[i] = versionedList[i].GetVersion()
	}

	// Update the status with the latest versions.
	source.Status.Versions = versions

	// Requeue periodically to check for new versions.
	return ctrl.Result{RequeueAfter: debianRepositorySourceRequeueDelay}, nil
}

func (r *DebianRepositorySourceReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgraderv1.DebianRepositorySource,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func patchDebianRepositorySource(
	ctx context.Context,
	patchHelper *patch.Helper,
	source *kubernetesupgraderv1.DebianRepositorySource,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		source,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DebianRepositorySourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//nolint:wrapcheck // This is generated code.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.DebianRepositorySource{}).
		Complete(r)
}
