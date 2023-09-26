// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/jobs"
)

//nolint:wrapcheck // No need to wrap in a test file.
var _ = Describe("MachineImage controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	var (
		podSpec = corev1.PodSpec{
			// For simplicity, we only fill out the required fields.
			Containers: []corev1.Container{
				{
					Name:  "test-container",
					Image: "test-image",
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		}

		machineImageLookupKey = types.NamespacedName{}
		createdMachineImage   = &kubernetesupgraderv1.MachineImage{}
	)

	const (
		machineImageName = "test-machine-image"
		buildImageID     = "test-id"
	)

	BeforeEach(func() {
		ctx := context.Background()
		By("Creating a new Namespace")
		ns := &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: machineImageName + "-",
			},
		}
		Expect(k8sClient.Create(ctx, ns)).Should(Succeed())
		machineImageNamespace := ns.Name

		By("Creating a new MachineImage")
		machineImage := &kubernetesupgraderv1.MachineImage{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				Kind:       "MachineImage",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      machineImageName,
				Namespace: machineImageNamespace,
			},
			Spec: kubernetesupgraderv1.MachineImageSpec{
				Version: "v1.27.99",
				JobTemplate: kubernetesupgraderv1.JobTemplate{
					Spec: podSpec,
				},
			},
		}

		Expect(k8sClient.Create(ctx, machineImage)).Should(Succeed())

		machineImageLookupKey = types.NamespacedName{
			Name:      machineImageName,
			Namespace: machineImageNamespace,
		}
		createdMachineImage = &kubernetesupgraderv1.MachineImage{}

		// We'll need to retry getting this newly created MachineImage, given that creation may not immediately happen.
		Eventually(func() error {
			return k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
		}, timeout, interval).Should(BeNil())
	})

	Context("When creating a MachineImage object", func() {
		It(
			"Should create a Job and the id value from the Job's annotation on the MachineImage spec.id",
			func() {
				By("Checking the MachineImage has status.jobRef set")
				var jobRef *corev1.ObjectReference
				Eventually(func() (*corev1.ObjectReference, error) {
					err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
					if err != nil {
						return nil, err
					}
					jobRef = createdMachineImage.Status.JobRef
					return jobRef, nil
				}, timeout, interval).Should(Not(BeNil()), "MachineImage status.jobRef should be set")

				By("Checking the MachineImage has status.phase set to Building")
				Eventually(func() (kubernetesupgraderv1.MachineImagePhase, error) {
					err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
					if err != nil {
						return "", err
					}
					return createdMachineImage.Status.Phase, nil
				}, timeout, interval).Should(
					Equal(kubernetesupgraderv1.MachineImagePhaseBuilding),
					"MachineImage status.phase should be set to Building",
				)

				By("Updating the Job")
				job := &batchv1.Job{}
				Expect(
					k8sClient.Get(ctx, objectReferenceToNamespacedName(jobRef), job),
				).ShouldNot(HaveOccurred())
				job.Status.Succeeded = int32(1)
				job.Annotations[jobs.IDAnnotation] = buildImageID
				Expect(
					k8sClient.Status().Update(ctx, job),
				).ShouldNot(HaveOccurred(), "error setting job status")
				Expect(
					k8sClient.Update(ctx, job),
				).ShouldNot(HaveOccurred(), "error setting job annotation")

				By("Checking the MachineImage has status.phase set to Created")
				Eventually(func() (kubernetesupgraderv1.MachineImagePhase, error) {
					err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
					if err != nil {
						return "", err
					}
					return createdMachineImage.Status.Phase, nil
				}, timeout, interval).Should(
					Equal(kubernetesupgraderv1.MachineImagePhaseCreated),
					"MachineImage status.phase should be set to Created",
				)

				By("Checking the MachineImage has spec.id set")
				Eventually(func() (string, error) {
					err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
					if err != nil {
						return "", err
					}
					return createdMachineImage.Spec.ID, nil
				}, timeout, interval).Should(Equal(buildImageID), "MachineImage spec.id should be set")
			},
		)

		It("Should create only 1 new Job if the Job gets deleted before succeeding", func() {
			By("Checking the MachineImage has status.jobRef set")
			var jobRef *corev1.ObjectReference
			Eventually(func() (*corev1.ObjectReference, error) {
				err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
				if err != nil {
					return nil, err
				}
				jobRef = createdMachineImage.Status.JobRef
				return jobRef, nil
			}, timeout, interval).Should(Not(BeNil()), "MachineImage status.jobRef should be set")

			By("Checking the MachineImage has status.phase set to Building")
			Eventually(func() (kubernetesupgraderv1.MachineImagePhase, error) {
				err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
				if err != nil {
					return "", err
				}
				return createdMachineImage.Status.Phase, nil
			}, timeout, interval).Should(
				Equal(kubernetesupgraderv1.MachineImagePhaseBuilding),
				"MachineImage status.phase should be set to Building",
			)

			By("Deleting the Job")
			job := &batchv1.Job{}
			Expect(
				k8sClient.Get(ctx, objectReferenceToNamespacedName(jobRef), job),
			).ShouldNot(HaveOccurred())
			Expect(k8sClient.Delete(ctx, job)).ShouldNot(HaveOccurred())

			// Wait for the Job to be actually deleted
			Eventually(func() error {
				err := k8sClient.Get(ctx, objectReferenceToNamespacedName(jobRef), job)
				if err == nil {
					// TODO(dkoshkin): is removing the Finalizers necessary, why didn't the Job get deleted?
					job.ObjectMeta.Finalizers = nil
					return k8sClient.Update(ctx, job)
				}
				return err
			}, timeout, interval).Should(MatchError(ContainSubstring("not found")))

			By("Checking the controller only created a single Job")
			jobList := &batchv1.JobList{}
			Consistently(func() ([]batchv1.Job, error) {
				err := k8sClient.List(ctx, jobList, &client.ListOptions{
					Namespace: machineImageLookupKey.Namespace,
				})
				return jobList.Items, err
			}, duration, interval).Should(HaveLen(1), "Expect only 1 Job to be created for the MachineImage")

			By("Checking the MachineImage has status.jobRef set")
			Eventually(func() (*corev1.ObjectReference, error) {
				err := k8sClient.Get(ctx, machineImageLookupKey, createdMachineImage)
				if err != nil {
					return nil, err
				}
				jobRef = createdMachineImage.Status.JobRef
				return jobRef, nil
			}, timeout, interval).Should(Not(BeNil()), "MachineImage status.jobRef should be set")

			By("Checking the MachineImage status.jobRef equals the Job's name")
			Expect(
				jobList.Items[0].Name,
			).Should(Equal(jobRef.Name), "MachineImage status.jobRef should equal the Job's name")
		})
	})
})

func objectReferenceToNamespacedName(ref *corev1.ObjectReference) types.NamespacedName {
	return types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}
}
