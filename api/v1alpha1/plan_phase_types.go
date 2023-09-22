// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

type PlanPhase string

const (
	// PlanPhaseNoSuitableMachineImage is the phase when there is no suitable machine image.
	PlanPhaseNoSuitableMachineImage = PlanPhase("NoSuitableMachineImage")

	// PlanPhaseFoundMachineImage is the phase when a suitable machine image has been found.
	PlanPhaseFoundMachineImage = PlanPhase("FoundMachineImage")
)
