// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

// Policer is an interface representing a policy implementation type.
type Policer interface {
	Latest([]VersionedObject) (string, error)
}
