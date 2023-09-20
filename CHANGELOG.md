# Changelog

## 0.5.0 (2023-09-20)

## What's Changed
* chore: use a MachineImageTemplate object by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/39
* chore: rename example files by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/41
* fix: MachineImageSyncer reconcile race by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/42
* chore: refactor machineimagesyncer controller to be more testable by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/43
* fix: watch for Jobs in MachineImage reconciler by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/44
* chore(deps): Bump actions/checkout from 3 to 4 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/45
* chore(deps): Bump k8s.io/apiextensions-apiserver from 0.27.2 to 0.28.1 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/46
* chore: refactor policy package to be more generic by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/47
* feat: refactor to use ClusterClassClusterUpgrader type by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/48
* fix: use builtin Owns kubebuilder watcher func by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/49
* fix: watch MachineImage in Plan controller by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/50
* feat: implement Debian repository source type by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/51
* chore(deps): Bump k8s.io/apiextensions-apiserver from 0.28.1 to 0.28.2 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/52
* chore(deps): Bump github.com/stretchr/testify from 1.8.2 to 1.8.4 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/53
* chore(deps): Bump docker/login-action from 2 to 3 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/54
* docs: add architecture details in README by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/56
* docs: wait for clusterctl providers by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/57
* fix: remove debug print lines by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/59
* docs: minor fixes in the README by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/58


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.4.0...v0.5.0

## 0.4.0 (2023-09-06)

## What's Changed
* feat: add MachineImageSyncer type with an initial test implementation by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/37


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.3.0...v0.4.0

## 0.3.0 (2023-09-04)

## What's Changed
* chore: rename MachineImage API by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/28
* chore: rename spec.ImageID to ID by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/30
* chore(deps): Bump k8s.io/client-go from 0.27.2 to 0.27.5 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/33
* chore(deps): Bump sigs.k8s.io/controller-runtime from 0.15.0 to 0.15.2 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/32
* chore(deps): Bump sigs.k8s.io/cluster-api from 1.5.0 to 1.5.1 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/31
* feat: support auto upgrading Clusters with a Plan type by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/36


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.2.1...v0.3.0

## 0.2.1 (2023-09-02)

## What's Changed
* chore: revert "don't publish archives in releases" by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/25
* chore: more README fixes by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/26


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.2.0...v0.2.1

## 0.2.0 (2023-09-02)

## What's Changed
* chore: don't publish archives in releases by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/20
* chore: include sample in releases by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/22
* chore: update README with usage instructions by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/23
* fix: set status on success by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/24


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.1.3...v0.2.0

## 0.1.3 (2023-09-02)

## What's Changed
* build: use out/ directory for components yaml by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/18


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.1.2...v0.1.3

## 0.1.2 (2023-09-02)

## What's Changed
* chore: login to ghcr before release by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/16


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.1.1...v0.1.2

## 0.1.1 (2023-09-02)

## What's Changed
* chore: add upx tool by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/14


**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/compare/v0.1.0...v0.1.1

## 0.1.0 (2023-09-02)

## What's Changed
* feat: Initial implementation building images by @dkoshkin in https://github.com/dkoshkin/kubernetes-upgrader/pull/1
* chore(deps): Bump asdf-vm/actions from 2.1.0 to 2.2.0 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/12
* chore(deps): Bump github.com/onsi/ginkgo/v2 from 2.11.0 to 2.12.0 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/11
* chore(deps): Bump guyarb/golang-test-annotations from 0.6.0 to 0.7.0 by @dependabot in https://github.com/dkoshkin/kubernetes-upgrader/pull/3

## New Contributors
* @dkoshkin made their first contribution in https://github.com/dkoshkin/kubernetes-upgrader/pull/1

**Full Changelog**: https://github.com/dkoshkin/kubernetes-upgrader/commits/v0.1.0
