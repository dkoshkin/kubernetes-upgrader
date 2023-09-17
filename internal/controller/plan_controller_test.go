// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
	"github.com/dkoshkin/kubernetes-upgrader/internal/scheme"
)

//nolint:gochecknoglobals // Shared across the tests.
var (
	testPlan = &kubernetesupgraderv1.Plan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plan",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: kubernetesupgraderv1.PlanSpec{
			VersionRange:         "v1.27.x",
			MachineImageSelector: &metav1.LabelSelector{},
		},
	}

	testPlanWithSelector = &kubernetesupgraderv1.Plan{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-plan-with-selector",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: kubernetesupgraderv1.PlanSpec{
			VersionRange: "v1.27.x",
			MachineImageSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"imageBuilderVersion": "v1",
				},
			},
		},
	}
)

//nolint:funlen // Long tests are ok.
func Test_latestMachineImageVersion(t *testing.T) {
	tests := []struct {
		name        string
		plan        *kubernetesupgraderv1.Plan
		allMachines *kubernetesupgraderv1.MachineImageList
		want        *kubernetesupgraderv1.MachineImage
		wantError   error
	}{
		{
			name: "no machines with matching version",
			plan: testPlan,
			allMachines: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-1",
							Namespace: metav1.NamespaceDefault,
						},
						Spec: kubernetesupgraderv1.MachineImageSpec{
							Version: "v1.26.99",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-2",
							Namespace: metav1.NamespaceDefault,
						},
						Spec: kubernetesupgraderv1.MachineImageSpec{
							Version: "v1.28.99",
						},
					},
				},
			},
		},
		{
			name: "all machines with matching version",
			plan: testPlan,
			allMachines: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-1",
							Namespace: metav1.NamespaceDefault,
						},
						Spec: kubernetesupgraderv1.MachineImageSpec{
							ID:      "id1",
							Version: "v1.27.1",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-2",
							Namespace: metav1.NamespaceDefault,
						},
						Spec: kubernetesupgraderv1.MachineImageSpec{
							ID:      "id2",
							Version: "v1.27.3",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-3",
							Namespace: metav1.NamespaceDefault,
						},
						Spec: kubernetesupgraderv1.MachineImageSpec{
							ID:      "id3",
							Version: "v1.27.2",
						},
					},
				},
			},
			want: &kubernetesupgraderv1.MachineImage{
				ObjectMeta: metav1.ObjectMeta{
					Name:            "machine-image-2",
					Namespace:       metav1.NamespaceDefault,
					ResourceVersion: "999",
				},
				Spec: kubernetesupgraderv1.MachineImageSpec{
					ID:      "id2",
					Version: "v1.27.3",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithLists(tt.allMachines).
				WithObjects(tt.plan).
				Build()

			got, err := latestMachineImageVersion(context.Background(), c, logr.Discard(), tt.plan)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

//nolint:funlen // Long tests are ok.
func Test_machineImagesForPlan(t *testing.T) {
	tests := []struct {
		name        string
		plan        *kubernetesupgraderv1.Plan
		allMachines *kubernetesupgraderv1.MachineImageList
		want        *kubernetesupgraderv1.MachineImageList
		wantError   error
	}{
		{
			name:        "no machines",
			plan:        testPlan,
			allMachines: &kubernetesupgraderv1.MachineImageList{},
			want: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{},
			},
		},
		{
			name: "all machines in a different namespace",
			plan: testPlan,
			allMachines: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-1",
							Namespace: "other-namespace",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-2",
							Namespace: "other-namespace",
						},
					},
				},
			},
			want: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{},
			},
		},
		{
			name: "some machines in the namespace",
			plan: testPlan,
			allMachines: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-1",
							Namespace: "other-namespace",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-2",
							Namespace: metav1.NamespaceDefault,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-3",
							Namespace: "other-namespace",
						},
					},
				},
			},
			want: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "machine-image-2",
							Namespace:       metav1.NamespaceDefault,
							ResourceVersion: "999",
						},
					},
				},
			},
		},
		{
			name: "some machines in the namespace with a selector",
			plan: testPlanWithSelector,
			allMachines: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-1",
							Namespace: "other-namespace",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-2",
							Namespace: metav1.NamespaceDefault,
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-3",
							Namespace: "other-namespace",
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "machine-image-4",
							Namespace: metav1.NamespaceDefault,
							Labels: map[string]string{
								"imageBuilderVersion": "v1",
							},
						},
					},
				},
			},
			want: &kubernetesupgraderv1.MachineImageList{
				TypeMeta: metav1.TypeMeta{
					Kind:       "MachineImageList",
					APIVersion: "kubernetesupgraded.dimitrikoshkin.com/v1alpha1",
				},
				Items: []kubernetesupgraderv1.MachineImage{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:            "machine-image-4",
							Namespace:       metav1.NamespaceDefault,
							ResourceVersion: "999",
							Labels: map[string]string{
								"imageBuilderVersion": "v1",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithLists(tt.allMachines).
				WithObjects(tt.plan).
				Build()

			got, err := machineImagesForPlan(context.Background(), c, tt.plan)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_machineImagesWithIDs(t *testing.T) {
	tests := []struct {
		name        string
		allMachines []kubernetesupgraderv1.MachineImage
		want        []kubernetesupgraderv1.MachineImage
	}{
		{
			name: "empty slice",
		},
		{
			name: "all have IDs",
			allMachines: []kubernetesupgraderv1.MachineImage{
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id1"}},
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id2"}},
			},
			want: []kubernetesupgraderv1.MachineImage{
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id1"}},
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id2"}},
			},
		},
		{
			name: "some have IDs",
			allMachines: []kubernetesupgraderv1.MachineImage{
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id1"}},
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: ""}},
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id3"}},
			},
			want: []kubernetesupgraderv1.MachineImage{
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id1"}},
				{Spec: kubernetesupgraderv1.MachineImageSpec{ID: "id3"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := machineImagesWithIDs(tt.allMachines); !reflect.DeepEqual(got, tt.want) {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
