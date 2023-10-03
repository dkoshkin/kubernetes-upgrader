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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

// InGitUpgradeAutomationReconciler reconciles a InGitUpgradeAutomation object.
type InGitUpgradeAutomationReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//nolint:lll // This is generated code.
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=ingitupgradeautomations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=ingitupgradeautomations/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=ingitupgradeautomations/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans,verbs=get;list;watch
//+kubebuilder:rbac:groups=kubernetesupgraded.dimitrikoshkin.com,resources=plans/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the InGitUpgradeAutomation object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
//
//nolint:dupl // Prefer readability to DRY.
func (r *InGitUpgradeAutomationReconciler) Reconcile(
	ctx context.Context,
	req ctrl.Request,
) (_ ctrl.Result, rerr error) {
	logger := log.FromContext(ctx).
		WithValues("igitupgradeautomation", req.Name, "namespace", req.Namespace)

	upgrader := &kubernetesupgraderv1.InGitUpgradeAutomation{}
	if err := r.Get(ctx, req.NamespacedName, upgrader); err != nil {
		logger.Error(
			err,
			"unable to fetch InGitUpgradeAutomation",
			"namespace", req.Namespace, "name", req.Name)
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}

	// Initialize the patch helper
	patchHelper, err := patch.NewHelper(upgrader, r.Client)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	// Always attempt to Patch the MachineImage object and status after each reconciliation.
	defer func() {
		if err := patchInGitUpgradeAutomation(ctx, patchHelper, upgrader); err != nil {
			logger.Error(err, "failed to patch InGitUpgradeAutomation")
			if rerr == nil {
				rerr = err
			}
		}
	}()

	if upgrader.Spec.Paused {
		logger.Info("Reconciliation is paused for this object")
		return ctrl.Result{}, nil
	}

	// Handle deleted clusters
	if !upgrader.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, logger, upgrader)
	}

	// Handle non-deleted clusters
	return r.reconcileNormal(ctx, logger, upgrader)
}

type inGitUpgrader struct {
	*kubernetesupgraderv1.InGitUpgradeAutomation
}

func (i inGitUpgrader) UpgradeCluster(
	_ context.Context,
	_ *clusterv1.Cluster,
) error {
	panic("implement me")
}

func (r *InGitUpgradeAutomationReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	inGitUpgraderAutomation *kubernetesupgraderv1.InGitUpgradeAutomation,
) (ctrl.Result, error) {
	genericUpgrader := &genericUpgradeAutomationReconciler{
		Client:   r.Client,
		Scheme:   r.Scheme,
		Recorder: r.Recorder,
	}

	upgrader := &inGitUpgrader{
		InGitUpgradeAutomation: inGitUpgraderAutomation,
	}
	return genericUpgrader.reconcileNormal(ctx, logger, upgrader)
}

func (r *InGitUpgradeAutomationReconciler) reconcileDelete(
	_ context.Context,
	logger logr.Logger,
	_ *kubernetesupgraderv1.InGitUpgradeAutomation,
) (ctrl.Result, error) {
	logger.Info("Reconciling delete")

	return ctrl.Result{}, nil
}

func patchInGitUpgradeAutomation(
	ctx context.Context,
	patchHelper *patch.Helper,
	upgrader *kubernetesupgraderv1.InGitUpgradeAutomation,
) error {
	// Patch the object, ignoring conflicts on the conditions owned by this controller.
	//nolint:wrapcheck // This is generated code.
	return patchHelper.Patch(
		ctx,
		upgrader,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *InGitUpgradeAutomationReconciler) SetupWithManager(
	ctx context.Context,
	mgr ctrl.Manager,
) error {
	//nolint:wrapcheck // No additional context to add.
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubernetesupgraderv1.InGitUpgradeAutomation{}).
		WithEventFilter(ResourceNotPaused(ctrl.LoggerFrom(ctx))).
		Watches(
			&kubernetesupgraderv1.Plan{},
			handler.EnqueueRequestsFromMapFunc(r.planMapper),
		).
		Complete(r)
}

//nolint:dupl // Prefer readability to DRY.
func (r *InGitUpgradeAutomationReconciler) planMapper(
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

	upgraders := &kubernetesupgraderv1.InGitUpgradeAutomationList{}
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
