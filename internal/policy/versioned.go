// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package policy

import corev1 "k8s.io/api/core/v1"

type VersionedObject interface {
	GetID() string
	GetVersion() string
	GetObjectReference() *corev1.ObjectReference
}
