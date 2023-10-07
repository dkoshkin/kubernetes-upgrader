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
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"time"

	fluxgit "github.com/fluxcd/pkg/git"
	"github.com/fluxcd/pkg/git/gogit"
	"github.com/fluxcd/pkg/git/repository"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"sigs.k8s.io/yaml"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/git"
)

const (
	getAuthCredentialsRequeueDelay = 1 * time.Minute
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
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch

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

	gitClient *gogit.Client

	cloneOpts    repository.CloneConfig
	switchBranch bool
	pushBranch   string

	updateStrategy kubernetesupgraderv1.UpdateStrategy
	commitSpec     kubernetesupgraderv1.CommitSpec
}

func (i *inGitUpgrader) UpgradeCluster(
	ctx context.Context,
	logger logr.Logger,
	cluster *clusterv1.Cluster,
) error {
	// Clone the git repository.
	logger.Info("Cloning git repository",
		"url", i.Spec.Git.Checkout.URL,
		"branch", i.cloneOpts.Branch,
	)
	_, err := i.gitClient.Clone(ctx, i.Spec.Git.Checkout.URL, i.cloneOpts)
	if err != nil {
		return fmt.Errorf("unable to clone git repository: %w", err)
	}

	// Switch the branch if needed.
	if i.switchBranch {
		logger.Info("Switching to branch", "branch", i.pushBranch)
		err = i.gitClient.SwitchBranch(ctx, i.pushBranch)
		if err != nil {
			return fmt.Errorf("unable to switch branch: %w", err)
		}
	}

	clusterBytes, err := yaml.Marshal(sanitizeClusterObject(cluster))
	if err != nil {
		return fmt.Errorf("unable to marshal Cluster to yaml: %w", err)
	}
	files := map[string]io.Reader{
		i.updateStrategy.Path: bytes.NewReader(clusterBytes),
	}
	message, err := templateMessage(i.commitSpec.MessageTemplate, cluster)
	if err != nil {
		return fmt.Errorf("unable to template message: %w", err)
	}
	logger.Info("Committing updated file", "file", i.updateStrategy.Path)
	_, err = i.gitClient.Commit(
		fluxgit.Commit{
			Author: fluxgit.Signature{
				Name:  i.commitSpec.Author.Name,
				Email: i.commitSpec.Author.Email,
				When:  time.Now(),
			},
			Message: message,
		},
		repository.WithFiles(files),
	)
	if err != nil {
		if strings.Contains(err.Error(), "no staged files") {
			logger.Info("No changes to commit")
			return nil
		}
		return fmt.Errorf("unable to commit changes: %w", err)
	}

	err = i.gitClient.Push(ctx, repository.PushConfig{})
	if err != nil {
		return fmt.Errorf("unable to push changes: %w", err)
	}

	return nil
}

// TODO(dkoshkin): Check if there is some utility function.
// sanitizeClusterObject removes Read-only metadata fields that should not be written to git.
func sanitizeClusterObject(cluster *clusterv1.Cluster) *clusterv1.Cluster {
	clusterToCommit := clusterv1.Cluster{
		TypeMeta: cluster.TypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:              cluster.Name,
			Namespace:         cluster.Namespace,
			Annotations:       cluster.Annotations,
			Labels:            cluster.Labels,
			CreationTimestamp: cluster.CreationTimestamp,
			Finalizers:        cluster.Finalizers,
		},
		Spec: cluster.Spec,
		Status: clusterv1.ClusterStatus{
			ControlPlaneReady:   cluster.Status.ControlPlaneReady,
			InfrastructureReady: cluster.Status.InfrastructureReady,
		},
	}

	return &clusterToCommit
}

func templateMessage(
	messageTemplate string,
	cluster *clusterv1.Cluster,
) (string, error) {
	t, err := template.New("commit message").Parse(messageTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to create commit message template from spec: %w", err)
	}

	type templateData struct {
		ClusterName      string
		ClusterNamespace string
		Version          string
	}
	templateValues := &templateData{
		ClusterName:      cluster.Name,
		ClusterNamespace: cluster.Namespace,
		Version:          cluster.Spec.Topology.Version,
	}

	b := &strings.Builder{}
	err = t.Execute(b, *templateValues)
	if err != nil {
		return "", fmt.Errorf("failed to run template from spec: %w", err)
	}
	return b.String(), nil
}

//nolint:funlen // TODO(dkoshkin): Refactor.
func (r *InGitUpgradeAutomationReconciler) reconcileNormal(
	ctx context.Context,
	logger logr.Logger,
	inGitUpgraderAutomation *kubernetesupgraderv1.InGitUpgradeAutomation,
) (ctrl.Result, error) {
	tmp, err := os.MkdirTemp(
		"",
		fmt.Sprintf("%s-%s", inGitUpgraderAutomation.Namespace, inGitUpgraderAutomation.Name),
	)
	if err != nil {
		r.Recorder.Eventf(
			inGitUpgraderAutomation,
			corev1.EventTypeWarning,
			"MkdirTempError",
			"Unable to create temp directory: %s",
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			logger.Error(err, "failed to remove working directory", "path", tmp)
		}
	}()

	// pushBranch contains the branch name the commit needs to be pushed to.
	// It takes the value of the push branch if one is specified,
	// or if the push config is nil, then it takes the value of the checkout branch.
	var pushBranch string
	var switchBranch bool
	if inGitUpgraderAutomation.Spec.Git.Push != nil &&
		inGitUpgraderAutomation.Spec.Git.Push.Branch != "" {
		pushBranch = inGitUpgraderAutomation.Spec.Git.Push.Branch
		// We only need to switch branches when a branch has been specified in the push spec,
		// and it is different from the one in the checkout ref.
		if inGitUpgraderAutomation.Spec.Git.Push.Branch !=
			inGitUpgraderAutomation.Spec.Git.Checkout.Reference.Branch {
			switchBranch = true
		}
	} else {
		pushBranch = inGitUpgraderAutomation.Spec.Git.Checkout.Reference.Branch
	}
	authOpts, err := git.GetAuthOpts(ctx, r.Client, inGitUpgraderAutomation.Spec.Git.Checkout)
	if err != nil {
		r.Recorder.Eventf(
			inGitUpgraderAutomation,
			corev1.EventTypeWarning,
			"GetAuthOptsError",
			"Error getting auth options: %s",
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{RequeueAfter: getAuthCredentialsRequeueDelay}, err
	}
	clientOpts := git.GetGitClientOpts(authOpts.Transport, switchBranch)
	gitClient, err := gogit.NewClient(tmp, authOpts, clientOpts...)
	if err != nil {
		r.Recorder.Eventf(
			inGitUpgraderAutomation,
			corev1.EventTypeWarning,
			"CreateGitClientError",
			"Error creating git client: %s",
			err,
		)
		//nolint:wrapcheck // No additional context to add.
		return ctrl.Result{}, err
	}
	defer gitClient.Close()

	cloneOpts := repository.CloneConfig{}
	cloneOpts.Branch = inGitUpgraderAutomation.Spec.Git.Checkout.Reference.Branch

	upgrader := &inGitUpgrader{
		InGitUpgradeAutomation: inGitUpgraderAutomation,
		gitClient:              gitClient,
		cloneOpts:              cloneOpts,
		switchBranch:           switchBranch,
		pushBranch:             pushBranch,
		updateStrategy:         inGitUpgraderAutomation.Spec.UpdateStrategy,
		commitSpec:             inGitUpgraderAutomation.Spec.Git.Commit,
	}

	genericUpgrader := &genericUpgradeAutomationReconciler{
		Client:   r.Client,
		Scheme:   r.Scheme,
		Recorder: r.Recorder,
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
