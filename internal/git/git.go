// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package git

import (
	"context"
	"fmt"
	"net/url"

	fluxgit "github.com/fluxcd/pkg/git"
	"github.com/fluxcd/pkg/git/gogit"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubernetesupgraderv1 "github.com/dkoshkin/kubernetes-upgrader/api/v1alpha1"
)

// GetAuthOpts fetches the secret containing the auth options (if specified),
// constructs a git.AuthOptions object using those options along with the provided repository's URL and returns it.
func GetAuthOpts(
	ctx context.Context,
	reader client.Reader,
	spec kubernetesupgraderv1.GitCheckoutSpec,
) (*fluxgit.AuthOptions, error) {
	data, err := getSecretData(ctx, reader, spec.SecretRef.Name, spec.SecretRef.Namespace)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get auth secret '%s/%s': %w",
			spec.SecretRef.Namespace,
			spec.SecretRef.Name,
			err,
		)
	}

	u, err := url.Parse(spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL '%s': %w", spec.URL, err)
	}

	opts, err := fluxgit.NewAuthOptions(*u, data)
	if err != nil {
		return nil, fmt.Errorf("failed to configure authentication options: %w", err)
	}

	return opts, nil
}

func GetGitClientOpts(
	gitTransport fluxgit.TransportType,
	diffPushBranch bool,
) []gogit.ClientOption {
	clientOpts := []gogit.ClientOption{gogit.WithDiskStorage()}
	if gitTransport == fluxgit.HTTP {
		clientOpts = append(clientOpts, gogit.WithInsecureCredentialsOverHTTP())
	}

	// If the push branch is different from the checkout ref, we need to
	// have all the references downloaded at clone time, to ensure that
	// SwitchBranch will have access to the target branch state. fluxcd/flux2#3384
	//
	// To always overwrite the push branch, the feature gate
	// GitAllBranchReferences can be set to false, which will cause
	// the SwitchBranch operation to ignore the remote branch state.
	//
	// Disable to be able to push to a different branch than the one fetching from.
	if diffPushBranch {
		clientOpts = append(clientOpts, gogit.WithSingleBranch(false))
	}
	return clientOpts
}

func getSecretData(
	ctx context.Context,
	reader client.Reader,
	name, namespace string,
) (map[string][]byte, error) {
	key := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	var secret corev1.Secret
	if err := reader.Get(ctx, key, &secret); err != nil {
		//nolint:wrapcheck // No additional context to add.
		return nil, err
	}
	return secret.Data, nil
}
