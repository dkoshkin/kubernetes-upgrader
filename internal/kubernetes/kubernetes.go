// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateAndWait creates a new object and waits for the cache to be updated with the newly created object.
func CreateAndWait(
	ctx context.Context,
	k8sClient client.Client,
	object client.Object,
) error {
	err := k8sClient.Create(ctx, object)
	if err != nil {
		return fmt.Errorf("error creating MachineImage: %w", err)
	}

	// Keep trying to get the object.
	// This will force the cache to update and prevent any future reconciliation
	// to reconcile with an outdated list of object,
	// which could lead to unwanted creation of a duplicate object.
	const (
		interval = 100 * time.Millisecond
		timeout  = 10 * time.Second
	)
	var pollErrors []error
	if err = wait.PollUntilContextTimeout(
		ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
			if err = k8sClient.Get(
				ctx,
				client.ObjectKeyFromObject(object),
				object,
			); err != nil {
				// Do not return error here. Continue to poll even if we hit an error
				// so that we avoid exiting because of transient errors like network flakes.
				// Capture all the errors and return the aggregate error if the poll fails eventually.
				pollErrors = append(pollErrors, err)
				return false, nil
			}
			return true, nil
		}); err != nil {
		return errors.Wrapf(
			kerrors.NewAggregate(pollErrors),
			"failed to get the MachineImage %s after creation", object.GetName(),
		)
	}

	return nil
}
