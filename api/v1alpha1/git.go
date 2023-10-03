// Copyright 2023 Dimitri Koshkin. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
)

type GitSpec struct {
	// Checkout gives the parameters for cloning the git repository, ready to make changes.
	// +required
	Checkout GitCheckoutSpec `json:"checkout"`

	// Commit specifies how to commit to the git repository.
	// +required
	Commit CommitSpec `json:"commit"`

	// Push specifies how and where to push commits made by the automation.
	// If missing, commits are pushed (back) to `.spec.checkout.branch` or its default.
	// +optional
	Push *PushSpec `json:"push,omitempty"`
}

type GitCheckoutSpec struct {
	// URL specifies the Git repository URL, it can be an HTTP/S or SSH address.
	// +kubebuilder:validation:Pattern="^(http|https|ssh)://.*$"
	// +required
	URL string `json:"url"`

	// SecretRef specifies the Secret containing authentication credentials.
	// For HTTPS repositories the Secret must contain 'username' and 'password' fields for basic auth
	// or 'bearerToken' field for token auth.
	// For SSH repositories the Secret must contain 'identity' and 'known_hosts' fields.
	// +required
	SecretRef corev1.LocalObjectReference `json:"secretRef"`

	// Branch to check out, defaults to 'main' if no other field is defined.
	// +required
	Reference GitRepositoryRef `json:"ref"`
}

// GitRepositoryRef specifies the Git reference to resolve and checkout.
type GitRepositoryRef struct {
	// Branch to check out, defaults to 'main' if no other field is defined.
	// +optional
	// +kubebuilder:default=main
	Branch string `json:"branch,omitempty"`
}

// CommitSpec specifies how to commit changes to the git repository.
type CommitSpec struct {
	// Author gives the email and optionally the name to use as the author of commits.
	// +required
	Author CommitUser `json:"author"`
	// SigningKey provides the option to sign commits with a GPG key.
	// +optional
	SigningKey *SigningKey `json:"signingKey,omitempty"`
	// MessageTemplate provides a template for the commit message,
	// into which will be interpolated the details of the change made.
	// +optional
	MessageTemplate string `json:"messageTemplate,omitempty"`
}

type CommitUser struct {
	// Name gives the name to provide when making a commit.
	// +optional
	Name string `json:"name,omitempty"`
	// Email gives the email to provide when making a commit.
	// +required
	Email string `json:"email"`
}

// SigningKey references a Kubernetes secret that contains a GPG keypair.
type SigningKey struct {
	// SecretRef holds the name to a secret that contains a 'git.asc' key
	// corresponding to the ASCII Armored file containing the GPG signing keypair as the value.
	// +required
	SecretRef corev1.LocalObjectReference `json:"secretRef,omitempty"`
}

// PushSpec specifies how and where to push commits.
type PushSpec struct {
	// Branch specifies that commits should be pushed to the branch named.
	// The branch is created using `.spec.checkout.branch` as the starting point, if it doesn't already exist.
	// +optional
	Branch string `json:"branch,omitempty"`

	// Refspec specifies the Git Refspec to use for a push operation.
	// If both Branch and Refspec are provided, then the commit is pushed to the branch
	// and also using the specified refspec.
	// For more details about Git Refspecs, see: https://git-scm.com/book/en/v2/Git-Internals-The-Refspec
	// +optional
	// Refspec string `json:"refspec,omitempty"`

	// Options specifies the push options that are sent to the Git server when performing a push operation.
	// For details, see: https://git-scm.com/docs/git-push#Documentation/git-push.txt---push-optionltoptiongt
	// +optional
	// Options map[string]string `json:"options,omitempty"`
}
