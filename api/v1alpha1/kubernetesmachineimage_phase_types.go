// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

type KubernetesMachineImagePhase string

const (
	// KubernetesMachineImagePhaseBuilding is the phase when the image is being built.
	KubernetesMachineImagePhaseBuilding = KubernetesMachineImagePhase("Building")

	// KubernetesMachineImagePhaseCreated is the phase when the image has been built.
	KubernetesMachineImagePhaseCreated = KubernetesMachineImagePhase("Created")

	// KubernetesMachineImagePhaseFailed is the phase when the image build has failed.
	KubernetesMachineImagePhaseFailed = KubernetesMachineImagePhase("Failed")
)
