// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

type MachineImagePhase string

const (
	// MachineImagePhaseBuilding is the phase when the image is being built.
	MachineImagePhaseBuilding = MachineImagePhase("Building")

	// MachineImagePhaseCreated is the phase when the image has been built.
	MachineImagePhaseCreated = MachineImagePhase("Created")

	// MachineImagePhaseFailed is the phase when the image build has failed.
	MachineImagePhaseFailed = MachineImagePhase("Failed")
)
