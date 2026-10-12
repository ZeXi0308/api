# Changelog

All notable changes to api will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.2.0] - 2026-10-12

### Added
- Add predicted NUMA memory groups to BindRequests and memory-group annotation constants [#20](https://github.com/kai-scheduler/api/issues/20) [itsomri](https://github.com/itsomri)

## [v0.1.4] - 2026-10-04

### Added
- Sync API types with KAI-Scheduler main: NvFractions GPU sharing, semi-preemptible, fractional GPU groups, admission service name [#18](https://github.com/kai-scheduler/api/issues/18) [SiorMeir](https://github.com/SiorMeir)

## [v0.1.3] - 2026-09-14

### Added
- Add GangScheduleDeployment option to kai/v1 pod-grouper Args

## [v0.1.2] - 2026-09-14

### Added
- Add PriorityClassName, GoMemLimit(Ratio) and ExtendedResourceClaimAllocation types
- Add FIPSOnly, InPlacePodResize, and PodGroup minMember=0 support

## [v0.1.1] - 2026-07-26

### Added
- `kai.scheduler/v1` API group (`kai/v1`): `Config` and `SchedulingShard` types plus their component sub-packages, with generated CRD manifests. `usagedb` config structs (`UsageDBConfig`, `UsageParams`, `WindowType`), `podgrouper/constants`, and `utilities/math` (`SaturatingAdd`).

### Changed
- Re-synced `utilities/resources` (DRA GPU-count overflow fix), `scheduling/v2alpha2` PodGroup types, and `podgrouper/constants` with KAI-Scheduler main.

## [v0.1.0]

Initial release of the standalone `github.com/kai-scheduler/api` module, seeded from KAI-Scheduler `main`.

### Added

- **API types** (client-backed CRD contracts):
  - Queue (`scheduling.run.ai/v2`) — hierarchical queue resource
  - PodGroup (`scheduling.run.ai/v2alpha2`) — gang scheduling with subgroups, plus validating webhook
  - BindRequest, NumaPlacementRequest (`scheduling.run.ai/v1alpha2`) — pod binding coordination
  - Topology (`kai.scheduler/v1alpha1`) — topology-aware scheduling input

- **Generated clients** — single clientset, informer factory, and listers spanning both the
  `scheduling.run.ai` and `kai.scheduler/v1alpha1` groups.

- **CRD manifests** (`config/crd/`):
  - `scheduling.run.ai_queues.yaml`
  - `scheduling.run.ai_podgroups.yaml`
  - `scheduling.run.ai_bindrequests.yaml`
  - `kai.scheduler_topologies.yaml`

- **Utilities**:
  - `utilities/resources/` — GPU fraction/memory extraction, DRA support, resource list operations
  - `utilities/podgroup/` — `CalculatePreemptibility()`

- **Constants** — GPU annotation keys, labels, and priority constants.

- **Code generation** — `make generate` (deepcopy), `make manifests` (CRDs), `make clients`
  (clientset/informers/listers) reproduce all generated artifacts.

### Notes

- Seeded fresh from `main` (no git history preserved); Apache 2.0 / NVIDIA attribution retained via per-file headers.
- Scheduler-internal code stays in kai-scheduler: `kai.scheduler/v1` config types (Config, SchedulingShard),
  scheduler-framework glue (`k8s_utils`), and feature gates.
- Dependencies: Kubernetes v0.35.4, controller-runtime v0.23.3, Go 1.26.3.

[v0.1.0]: https://github.com/kai-scheduler/api/releases/tag/v0.1.0
