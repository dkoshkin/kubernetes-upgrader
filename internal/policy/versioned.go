// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

type Versioned interface {
	GetVersion() string
}

type VersionedList []Versioned
