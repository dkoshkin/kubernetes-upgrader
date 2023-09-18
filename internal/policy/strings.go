// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

type VersionedString struct {
	version string
}

func (v *VersionedString) GetVersion() string {
	return v.version
}

// VersionedStrings returns a list of VersionedString objects from a slice of strings.
// TODO(dkoshkin): Can we simplify this package?
func VersionedStrings(versions ...string) []Versioned {
	//nolint:prealloc // Copied from another repo.
	var versioned []Versioned
	for _, v := range versions {
		versioned = append(versioned, &VersionedString{v})
	}
	return versioned
}
