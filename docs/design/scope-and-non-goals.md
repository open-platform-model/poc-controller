# Experimental Scope and Non-Goals

## Summary

This document defines the intentionally minimal scope of the OPM proof-of-concept controller.

The controller is an experiment whose job is to prove one narrow architecture:

- Flux `source-controller` can resolve and store native CUE OCI module artifacts for OPM.
- OPM can fetch that stored artifact, recover a valid CUE module tree, evaluate it, and reconcile the resulting Kubernetes resources.
- `ModuleRelease` can act as the durable ledger for source, render, apply, prune, and ownership status.

Anything not required to prove that loop is intentionally out of scope.

## Why This Boundary Exists

The project already has enough uncertainty in a few critical areas:

- native CUE OCI artifact handoff through Flux
- controller-owned source recovery and unpacking
- the right initial `ModuleRelease` status and inventory contract
- the right division of responsibility between Flux and OPM

Because of that, the proof-of-concept should optimize for clarity and fast learning, not completeness.

The controller should therefore be judged by whether it proves the core control loop, not by whether it includes every feature expected from a production release manager.

## In Scope

The proof-of-concept controller is in scope for the following behavior only.

### Primary resource

- `ModuleRelease` is the only serious reconciliation target in the initial implementation.
- `ModuleRelease` references a Flux `OCIRepository`.
- `ModuleRelease` owns detailed digests, conditions, inventory, and bounded history.

### Source contract

- only Flux `OCIRepository` is supported as the source object
- only native CUE OCI module artifacts are supported
- the controller assumes the `OCIRepository` preserves the native CUE zip layer rather than relying on default tar+gzip extraction
- the controller validates and consumes `OCIRepository.status.artifact`

### Core reconcile loop

- fetch the Flux artifact
- recover and validate the CUE module tree
- evaluate the selected module using release values
- render desired Kubernetes objects
- apply them with server-side apply
- optionally prune stale previously owned objects
- record conditions, digests, inventory, and bounded history in status

### Ownership and metadata

- `status.inventory` is the authoritative ownership record
- runtime-owned metadata such as `app.kubernetes.io/managed-by` is injected by the controller path
- labels are helpful live metadata, not the prune authority

### Verification level

- enough tests and experiments to prove the source-to-render-to-apply path works
- enough validation to fail clearly on incompatible source configuration

## Explicit Non-Goals

The following items are intentionally left out of the proof-of-concept.

### 1. Bundle orchestration

The proof-of-concept does not attempt to make `BundleRelease` production-capable.

Out of scope:

- child `ModuleRelease` orchestration semantics
- cross-module dependency ordering
- aggregate readiness semantics for bundles
- bundle-level inventory as a production contract

`BundleRelease` may remain a sketch or placeholder while `ModuleRelease` is proven end to end.

### 2. Advanced rollout behavior

The proof-of-concept does not implement progressive delivery.

Out of scope:

- canary rollout
- blue/green rollout
- partitioned rollout
- staged workload promotion
- automated rollback
- rollout pause or approval checkpoints

The initial controller only needs a straightforward reconcile/apply/prune flow.

### 3. Cross-release dependency management

The proof-of-concept does not model release graphs.

Out of scope:

- `dependsOn`
- dependency-aware scheduling
- release graph validation
- readiness gating between different releases

Each `ModuleRelease` should be understandable and reconcilable on its own.

### 4. Workload health orchestration

The proof-of-concept does not try to become a full workload health engine.

Out of scope:

- custom readiness semantics per workload type
- rollout success based on Deployment, StatefulSet, or Job health
- post-apply workload convergence checks beyond the basic reconcile result
- controller-driven health remediation

For the experiment, successful source handling and successful apply/prune are enough.

### 5. Rich drift detection and preview UX

The proof-of-concept does not need a full drift product surface.

Out of scope:

- human-readable diff views
- status fields dedicated to drift reports
- preview mode surfaced through the API
- advanced reconciliation planning UX

The controller may compute what it needs internally, but it does not need to expose a rich drift interface.

### 6. Multi-source support

The proof-of-concept intentionally narrows the source story.

Out of scope:

- `GitRepository`
- `Bucket`
- `HelmRepository`
- direct OCI references in `ModuleRelease`
- mixing multiple source types behind one release API

Supporting one good source contract is more valuable than abstracting several incomplete ones.

### 7. Alternate artifact formats

The proof-of-concept does not support multiple publication contracts.

Out of scope:

- Flux-specific tarball artifact publication for OPM modules
- YAML-first publication contracts
- format negotiation across module types
- multiple artifact shapes for the same logical release API

The experiment is specifically about native CUE OCI modules.

### 8. Complex values composition

The proof-of-concept should keep release inputs simple.

Out of scope:

- layered values files
- multiple values sources merged by precedence rules
- values pulled from Secrets or ConfigMaps
- environment/profile selection systems
- values import/include mechanisms

One inline values payload is sufficient to prove the release contract.

### 9. Multi-tenant and policy features

The proof-of-concept does not attempt to solve controller tenancy or policy integration comprehensively.

Out of scope:

- namespace tenancy strategies
- policy engine integration
- admission webhook enforcement
- advanced impersonation design
- security boundary guarantees across tenants

`serviceAccountName` may still exist as a narrow execution input, but not as part of a full multi-tenant model.

### 10. Shared-resource ownership edge cases

The proof-of-concept does not solve multi-writer ownership.

Out of scope:

- shared-resource coordination across releases
- arbitrary resource adoption workflows
- label-only ownership models
- complex conflict resolution between OPM and other tools

The controller may safely assume one release owns the resources recorded in its inventory.

### 11. Production hardening and scale guarantees

The proof-of-concept is not a production-readiness exercise.

Out of scope:

- HA tuning
- bespoke queue tuning and backoff policy design
- performance claims for large-scale reconcile volume
- controller SLOs and operational benchmarks
- disaster recovery strategy for controller state

The implementation should be correct and coherent, but not optimized prematurely.

### 12. Full observability product surface

The proof-of-concept does not need an exhaustive observability contract.

Out of scope:

- production metrics taxonomy
- rich audit streams
- detailed event design beyond normal controller events
- dashboards or UI-oriented status contracts

Basic logs, standard conditions, and a useful status surface are enough.

### 13. Advanced history and rollback ledger

The proof-of-concept does not need long-term release bookkeeping.

Out of scope:

- manifest archives
- external history storage
- large retention policies
- rollback-to-history workflows
- exact parity with Helm-style release ledgers

Bounded status history is enough to support debugging and prove the core model.

### 14. CLI migration and interoperability workflows

The proof-of-concept does not need a formal migration story.

Out of scope:

- migration from CLI-managed releases to controller-managed releases
- compatibility guarantees for experimental status fields
- automatic adoption of existing CLI inventory state
- dual-writer coordination between CLI and controller

The experiment only needs to prove the controller path on its own terms.

## Practical Minimal Controller Contract

If the controller can do the following reliably, the experiment succeeds:

1. Observe a `ModuleRelease`.
2. Resolve its referenced Flux `OCIRepository`.
3. Validate that the source artifact contract is compatible with native CUE module recovery.
4. Fetch and recover the native CUE module payload.
5. Evaluate and render the module with release values.
6. Apply desired objects and optionally prune stale owned objects.
7. Persist clear status, digests, and ownership inventory.

That is the intended end state of the proof-of-concept.

## Deferred Future Work

The items listed as non-goals are not rejected permanently. They are deferred until the minimal controller loop is proven.

The most likely follow-up areas after the experiment are:

- `BundleRelease` orchestration semantics
- richer rollout and health handling
- broader source support if needed
- stronger validation and policy enforcement
- better observability and operational hardening

Those should only be revisited after the experiment proves the core architecture is sound.

## Exit Criteria For The Experiment

The proof-of-concept is complete enough to evaluate when all of the following are true:

- a native CUE OCI module can be published and reconciled through Flux `OCIRepository`
- the controller can recover and evaluate the module payload reliably
- `ModuleRelease` can apply and optionally prune rendered resources
- `ModuleRelease.status` gives a coherent account of source, render, apply, and ownership state
- the implementation is small enough that the next design decisions are informed by real controller behavior rather than speculation
