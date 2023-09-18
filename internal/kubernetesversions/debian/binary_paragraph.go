// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package debian

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// BinaryParagraph is a simple struct that holds Debian's package control files.
// See https://www.debian.org/doc/debian-policy/#s-binarycontrolfiles.
// TODO(dkoshkin): consider using a type from pault.ag/go/debian/control.
type BinaryParagraph struct {
	Package      string
	Version      string
	Architecture string
}

// GetVersion returns the version with format "vX.X.X" dropping all other version information.
func (p *BinaryParagraph) GetVersion() string {
	parsed, err := semver.NewVersion(p.Version)
	if err != nil {
		return p.Version
	}

	return fmt.Sprintf("v%d.%d.%d", parsed.Major(), parsed.Minor(), parsed.Patch())
}
