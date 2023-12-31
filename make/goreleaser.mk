# Copyright 2023 Dimitri Koshkin. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

export GOMODULENAME ?= $(shell go list -m)

GORELEASER_PARALLELISM ?= $(shell nproc --ignore=1)
GORELEASER_DEBUG ?= false

ifndef GORELEASER_CURRENT_TAG
export GORELEASER_CURRENT_TAG=$(GIT_TAG)
endif

.PHONY: build-snapshot
build-snapshot: ## Builds a snapshot with goreleaser
build-snapshot: manifests generate fmt vet
build-snapshot: install-tool.goreleaser install-tool.golang ; $(info $(M) building snapshot $*)
	goreleaser --debug=$(GORELEASER_DEBUG) \
		build \
		--snapshot \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		$(if $(BUILD_ALL),,--single-target) \
		--skip-post-hooks

.PHONY: release
release: ## Builds a release with goreleaser
release: manifests generate fmt vet
release: install-tool.goreleaser install-tool.golang ; $(info $(M) building release $*)
	goreleaser --debug=$(GORELEASER_DEBUG) \
		release \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		--timeout=60m \
		$(GORELEASER_FLAGS)

.PHONY: release-snapshot
release-snapshot: ## Builds a snapshot release with goreleaser
release-snapshot: manifests generate fmt vet
release-snapshot: install-tool.goreleaser install-tool.golang ; $(info $(M) building snapshot release $*)
	goreleaser --debug=$(GORELEASER_DEBUG) \
		release \
		--snapshot \
		--skip-publish \
		--clean \
		--parallelism=$(GORELEASER_PARALLELISM) \
		--timeout=60m
