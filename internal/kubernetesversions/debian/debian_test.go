// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package debian

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

const amdArch = "amd64"

//nolint:funlen // Long tests are ok.
func TestList(t *testing.T) {
	tests := []struct {
		name                 string
		packagesFileContents []byte
		want                 policy.VersionedList
	}{
		{
			name: "empty file",
		},
		{
			name: "with no kubeadm binaries",
			packagesFileContents: []byte(`
Package: cri-tools
Version: 1.25.0-1.1
Architecture: amd64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 49175
Filename: amd64/cri-tools_1.25.0-1.1_amd64.deb
Size: 17953728
MD5sum: 7e2c2ced8842dd52af887ab13c6c5020
SHA1: 7b7ec28a0053527d4ea7bde631e0f0151289cfe7
SHA256: 7bd71a8a93d38dbcf6901cde1ee9f8415cf3dfddabf793dd37728198e3b26c23
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for interacting with a container runtime
 Command-line utility for interacting with a container runtime.
`),
		},
		{
			name: "with different architecture",
			packagesFileContents: []byte(`
Package: cri-tools
Version: 1.25.0-1.1
Architecture: amd64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 49175
Filename: amd64/cri-tools_1.25.0-1.1_amd64.deb
Size: 17953728
MD5sum: 7e2c2ced8842dd52af887ab13c6c5020
SHA1: 7b7ec28a0053527d4ea7bde631e0f0151289cfe7
SHA256: 7bd71a8a93d38dbcf6901cde1ee9f8415cf3dfddabf793dd37728198e3b26c23
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for interacting with a container runtime
 Command-line utility for interacting with a container runtime.

Package: kubeadm
Version: 1.27.0-2.1
Architecture: arm64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 45052
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: arm64/kubeadm_1.27.0-2.1_arm64.deb
Size: 8483808
MD5sum: c33ce42b8ec919c185bae7b719d7f40b
SHA1: 5ea14acee043e9167267f4539ae51fb15f62475a
SHA256: d27cd4562154ddbaad4c4967b1295a9d3e4586ccb2895ed0c8c591c95dea1487
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.
`),
		},
		{
			name: "with multiple kubeadm versions",
			packagesFileContents: []byte(`
Package: cri-tools
Version: 1.25.0-1.1
Architecture: amd64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 49175
Filename: amd64/cri-tools_1.25.0-1.1_amd64.deb
Size: 17953728
MD5sum: 7e2c2ced8842dd52af887ab13c6c5020
SHA1: 7b7ec28a0053527d4ea7bde631e0f0151289cfe7
SHA256: 7bd71a8a93d38dbcf6901cde1ee9f8415cf3dfddabf793dd37728198e3b26c23
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for interacting with a container runtime
 Command-line utility for interacting with a container runtime.

Package: kubeadm
Version: 1.27.0-2.1
Architecture: amd64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 47088
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: amd64/kubeadm_1.27.0-2.1_amd64.deb
Size: 9934640
MD5sum: 144349a308aaa35d4d8d320e61d3820f
SHA1: 9dd14572e2c2374b1ab9d57d99f7ae10028e32a2
SHA256: 942b9d3c17efa014abfa21c89cd0e3f129a9910f0af7ca850d979635c10cdf9b
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.0-2.1
Architecture: ppc64el
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 46076
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: ppc64el/kubeadm_1.27.0-2.1_ppc64el.deb
Size: 8098700
MD5sum: 17a3a8476ae3238befe5c858956eaebb
SHA1: 65299277fea5603d220c3de73bf3a82eddbbe1b7
SHA256: 0f6761d4349e26badf7a05de4f4f3d4c615fd10ba9c37d25b756e929716b917c
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.0-2.1
Architecture: arm64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 45052
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: arm64/kubeadm_1.27.0-2.1_arm64.deb
Size: 8483808
MD5sum: c33ce42b8ec919c185bae7b719d7f40b
SHA1: 5ea14acee043e9167267f4539ae51fb15f62475a
SHA256: d27cd4562154ddbaad4c4967b1295a9d3e4586ccb2895ed0c8c591c95dea1487
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.0-2.1
Architecture: s390x
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 49788
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: s390x/kubeadm_1.27.0-2.1_s390x.deb
Size: 8974020
MD5sum: cecec9cbf45668694b2aa0baab79c09a
SHA1: ef4044d0b2ad521ebf0f7fb923d55a6730a0f0a4
SHA256: 118fa04e51ce7e4576e5bf47c7f720628abd9ea5348bc2448171ac25070ba979
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.1-1.1
Architecture: amd64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 47088
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: amd64/kubeadm_1.27.1-1.1_amd64.deb
Size: 9934392
MD5sum: b495daec55689d79b15adb9919793b15
SHA1: 2eac5255edfeacb9cfc9e7da36d45065c731a5ef
SHA256: 8268229a12dba5b6ece21608fe44ba6c59b2ed9f29edcc7c4708ca3bec8eb859
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.1-1.1
Architecture: ppc64el
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 46076
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: ppc64el/kubeadm_1.27.1-1.1_ppc64el.deb
Size: 8098788
MD5sum: af633536723b05a2b068468122d56f84
SHA1: 19b7c9b69efd059981b2a62727fcb577d1e3b670
SHA256: 2327ccdd7298da157b639f342c6df16205b3acfe7eff35baa87a8fb1b4d7da5d
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.1-1.1
Architecture: arm64
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 45052
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: arm64/kubeadm_1.27.1-1.1_arm64.deb
Size: 8486436
MD5sum: 3b5474655ab11ae5e4b6741bcaaac031
SHA1: 8ed9b57642d27a317b560e86600d7eec79f906b1
SHA256: 050dcc55b4bd3ef5d4a5898e5a9db9df6eb65debdcb79a4f21b9e56743aff896
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.

Package: kubeadm
Version: 1.27.1-1.1
Architecture: s390x
Maintainer: Kubernetes Authors <dev@kubernetes.io>
Installed-Size: 49788
Depends: kubelet (>= 1.19.0),kubectl (>= 1.19.0),kubernetes-cni (>= 1.1.1),cri-tools (>= 1.25.0)
Filename: s390x/kubeadm_1.27.1-1.1_s390x.deb
Size: 8968464
MD5sum: f27be9e755335c4a36b9e6361ee68ad9
SHA1: 9cc0686739bde15a20621ea56fbb1a5544883db7
SHA256: a2e8b267b72ae668ec72ccd5d42194bccb238be96e7c9db1f731761b02dc0d0d
Section: admin
Priority: optional
Homepage: https://kubernetes.io
Description: Command-line utility for administering a Kubernetes cluster
 Command-line utility for administering a Kubernetes cluster.
`),
			want: policy.VersionedList{
				&BinaryParagraph{
					Package:      "kubeadm",
					Version:      "1.27.0-2.1",
					Architecture: amdArch,
				},
				&BinaryParagraph{
					Package:      "kubeadm",
					Version:      "1.27.1-1.1",
					Architecture: amdArch,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := newTestSource(bytes.NewReader(tt.packagesFileContents))
			got, err := source.List(context.TODO())
			assert.NoError(t, err, "error listing versions")
			assert.Equal(t, tt.want, got, "versions do not match")
		})
	}
}

func newTestSource(packagesFileContents io.Reader) Source {
	return Source{
		opts: Options{
			Architecture: amdArch,
		},
		packagesFileFunc: func(_ context.Context, _ string) (io.Reader, error) {
			return packagesFileContents, nil
		},
	}
}
