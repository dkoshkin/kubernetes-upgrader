// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package kubernetesversions

import (
	"context"

	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

type Source interface {
	// List returns a list of Kubernetes versions available in the source.
	List(ctx context.Context) (policy.VersionedList, error)
}
