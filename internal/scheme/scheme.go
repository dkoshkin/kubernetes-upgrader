// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package scheme

import (
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

//nolint:gochecknoglobals // Want this to be public and shared in other packages.
var Scheme = runtime.NewScheme()

//nolint:gochecknoinits // Upstream standard.
func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(Scheme))

	utilruntime.Must(clusterv1.AddToScheme(Scheme))

	utilruntime.Must(kubernetesupgraderv1.AddToScheme(Scheme))
	//+kubebuilder:scaffold:scheme
}
