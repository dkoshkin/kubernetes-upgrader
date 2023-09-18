// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package debian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetVersion(t *testing.T) {
	tests := []struct {
		name   string
		binary BinaryParagraph
		want   string
	}{
		{
			name: "1.27.0-2.1",
			binary: BinaryParagraph{
				Package:      "kubeadm",
				Version:      "1.27.0-2.1",
				Architecture: amdArch,
			},
			want: "v1.27.0",
		},
		{
			name: "1.27.1-1.1",
			binary: BinaryParagraph{
				Package:      "kubeadm",
				Version:      "1.27.1-1.1",
				Architecture: amdArch,
			},
			want: "v1.27.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.binary.GetVersion()
			assert.Equal(t, tt.want, got)
		})
	}
}
