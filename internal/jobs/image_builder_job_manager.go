// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

const (
	ImageBuilderOwnedLabel = "image-builder.kubernetesupgraded.dimitrikoshkin.com/owned"

	IDAnnotation = "kubernetesupgraded.dimitrikoshkin.com/image-id"
)

type Manager interface {
	Create(
		ctx context.Context,
		owner *kubernetesupgraderv1.MachineImage,
		spec *corev1.PodSpec,
	) (*corev1.ObjectReference, error)
	Status(ctx context.Context, ref *corev1.ObjectReference) (*batchv1.JobStatus, string, error)
	Delete(ctx context.Context, ref *corev1.ObjectReference) error
}

type ImageBuilderJobManager struct {
	client runtimeclient.Client
	scheme *runtime.Scheme
}

func NewManager(client runtimeclient.Client, scheme *runtime.Scheme) Manager {
	return &ImageBuilderJobManager{client: client, scheme: scheme}
}

func (m *ImageBuilderJobManager) Create(
	ctx context.Context,
	owner *kubernetesupgraderv1.MachineImage,
	spec *corev1.PodSpec,
) (*corev1.ObjectReference, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", owner.Name),
			Namespace:    owner.Namespace,
			Labels: map[string]string{
				ImageBuilderOwnedLabel: "",
			},
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: *spec,
			},
		},
	}

	err := controllerutil.SetControllerReference(owner, job, m.scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to set controller reference on job: %w", err)
	}

	err = m.client.Create(ctx, job)
	if err != nil {
		return nil, fmt.Errorf("failed to create image builder job: %w", err)
	}

	return toObjectReference(job), nil
}

// Status will return the status of the job and the image ID if it was built successfully.
func (m *ImageBuilderJobManager) Status(
	ctx context.Context,
	ref *corev1.ObjectReference,
) (*batchv1.JobStatus, string, error) {
	job := &batchv1.Job{}
	err := m.client.Get(ctx, objectReferenceToNamespacedName(ref), job)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image builder job: %w", err)
	}

	return &job.Status, job.Annotations[IDAnnotation], nil
}

func (m *ImageBuilderJobManager) Delete(
	ctx context.Context,
	ref *corev1.ObjectReference,
) error {
	job := &batchv1.Job{}
	if err := m.client.Get(ctx, objectReferenceToNamespacedName(ref), job); err != nil {
		// not found, nothing to do
		if apierrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to check if image builder job exists: %w", err)
	}
	if err := m.client.Delete(ctx, job); err != nil {
		return fmt.Errorf("failed to delete image builder job: %w", err)
	}
	return nil
}

func objectReferenceToNamespacedName(ref *corev1.ObjectReference) types.NamespacedName {
	return types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
}

func toObjectReference(job *batchv1.Job) *corev1.ObjectReference {
	return &corev1.ObjectReference{
		Kind:            job.Kind,
		APIVersion:      job.APIVersion,
		Namespace:       job.GetNamespace(),
		Name:            job.GetName(),
		UID:             job.GetUID(),
		ResourceVersion: job.GetResourceVersion(),
	}
}
