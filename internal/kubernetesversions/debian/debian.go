// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package debian

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"pault.ag/go/debian/control"

	"github.com/dkoshkin/kubernetes-upgrader/internal/kubernetesversions"
	"github.com/dkoshkin/kubernetes-upgrader/internal/policy"
)

type Source struct {
	uri  string
	opts Options

	packagesFileFunc func(context.Context, string) (io.Reader, error)
}

type Options struct {
	Architecture string
}

func NewHTTPSource(uri string, opts Options) kubernetesversions.Source {
	return Source{
		uri:              uri,
		opts:             opts,
		packagesFileFunc: remotePackagesFile,
	}
}

func (s Source) List(ctx context.Context) (policy.VersionedList, error) {
	packages, err := s.packagesFileFunc(ctx, s.uri)
	if err != nil {
		return nil, fmt.Errorf("error downloading Debian Packages file: %w", err)
	}

	binaries := []BinaryParagraph{}
	err = control.Unmarshal(&binaries, packages)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling Debian Packages file: %w", err)
	}

	// assume that if the Kubeadm binaries exist, all others will exist to for that version
	kubeadmBinaries := getKubeadmBinaries(binaries, s.opts.Architecture)

	return toVersionedList(kubeadmBinaries), nil
}

// remotePackagesFile downloads a file from a URL and returns an io.Reader with its contents.
func remotePackagesFile(ctx context.Context, url string) (io.Reader, error) {
	const httpTimeout = 10 * time.Second
	ctxWithTimeout, cancel := context.WithTimeout(ctx, httpTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxWithTimeout, http.MethodGet, url, http.NoBody)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return nil, err
	}
	defer resp.Body.Close()

	// read the response body into a byte slice to return it as an io.Reader
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		//nolint:wrapcheck // No additional context to add.
		return nil, err
	}

	fmt.Printf("resp: %+v\n", resp)
	fmt.Printf("resBody: %+v\n", string(resBody))

	return bytes.NewReader(resBody), nil
}

// getKubeadmBinaries returns a list of just kubeadm binaries.
func getKubeadmBinaries(binaries []BinaryParagraph, architecture string) []BinaryParagraph {
	const kubeadmPackageName = "kubeadm"
	var kubeadmBinaries []BinaryParagraph
	for i := range binaries {
		if binaries[i].Package == kubeadmPackageName &&
			binaries[i].Architecture == architecture {
			kubeadmBinaries = append(kubeadmBinaries, binaries[i])
		}
	}

	return kubeadmBinaries
}

func toVersionedList(binaries []BinaryParagraph) policy.VersionedList {
	var versioned policy.VersionedList
	for i := range binaries {
		versioned = append(versioned, &binaries[i])
	}

	return versioned
}
