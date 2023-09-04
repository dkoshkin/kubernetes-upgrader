// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"

	"github.com/Masterminds/semver/v3"
	"github.com/fluxcd/pkg/version"
)

// SemVer represents a SemVer policy.
type SemVer struct {
	Range string

	constraint *semver.Constraints
}

// NewSemVer constructs a SemVer object validating the provided semver constraint.
func NewSemVer(r string) (*SemVer, error) {
	constraint, err := semver.NewConstraint(r)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return nil, err
	}

	return &SemVer{
		Range:      r,
		constraint: constraint,
	}, nil
}

// Latest returns latest version from a provided list of strings.
func (p *SemVer) Latest(versions []VersionedObject) (VersionedObject, error) {
	if len(versions) == 0 {
		//nolint:goerr113 // This error type is not expected to be checked.
		return nil, errors.New("version list argument cannot be empty")
	}

	var latestSemver *semver.Version
	var latestVersioned VersionedObject
	for i := range versions {
		o := versions[i]
		if v, err := version.ParseVersion(o.GetVersion()); err == nil {
			if p.constraint.Check(v) && (latestSemver == nil || v.GreaterThan(latestSemver)) {
				latestSemver = v
				latestVersioned = o
			}
		}
	}

	if latestSemver != nil {
		return latestVersioned, nil
	}
	//nolint:goerr113 // This error type is not expected to be checked.
	return nil, errors.New("unable to determine latest version")
}
