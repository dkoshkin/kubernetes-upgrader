// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/dkoshkin/kubernetes-upgrader/internal/constants"
)

func ResourceNotPaused(logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfNotPaused(
				logger.WithValues("predicate", "ResourceNotPaused", "eventType", "update"),
				e.ObjectNew,
			)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfNotPaused(
				logger.WithValues("predicate", "ResourceNotPaused", "eventType", "create"),
				e.Object,
			)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfNotPaused(
				logger.WithValues("predicate", "ResourceNotPaused", "eventType", "delete"),
				e.Object,
			)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfNotPaused(
				logger.WithValues("predicate", "ResourceNotPaused", "eventType", "generic"),
				e.Object,
			)
		},
	}
}

func processIfNotPaused(logger logr.Logger, obj client.Object) bool {
	kind := strings.ToLower(obj.GetObjectKind().GroupVersionKind().Kind)
	log := logger.WithValues("namespace", obj.GetNamespace(), kind, obj.GetName())
	objMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		log.Error(err, "error converting object to unstructured")
		return true
	}
	paused, found, err := unstructured.NestedBool(objMap, "spec", "paused")
	if err != nil {
		log.Error(err, "error checking value spec.paused from object")
		return true
	}
	if !found {
		log.V(constants.LogLevelDebug).
			Info("Resource does not have spec.paused, will attempt to map resource")
		return true
	}
	if paused {
		log.V(constants.LogLevelDebug).Info("Resource is paused, will not attempt to map resource")
		return false
	}
	return true
}
