## ADDED Requirements

### Requirement: Finalizer registration
The controller MUST add the finalizer `releases.opmodel.dev/cleanup` to a ModuleRelease during Phase 0 if it is not already present. The reconcile call that adds the finalizer MUST request an immediate requeue (`ctrl.Result{Requeue: true}`) rather than relying on the watch event produced by the finalizer patch — that watch event is filtered by the controller's `predicate.GenerationChangedPredicate`, because finalizer patches modify `metadata.finalizers` but do not bump `metadata.generation`. Without an explicit requeue, the next reconcile would be deferred to the periodic resync (default 10h), leaving freshly-created ModuleReleases without status conditions for an unacceptable duration.

#### Scenario: First reconcile adds finalizer
- **GIVEN** a ModuleRelease without the `releases.opmodel.dev/cleanup` finalizer
- **WHEN** the controller reconciles the resource
- **THEN** the finalizer is added to `metadata.finalizers`

#### Scenario: Subsequent reconciles preserve finalizer
- **GIVEN** a ModuleRelease that already has the `releases.opmodel.dev/cleanup` finalizer
- **WHEN** the controller reconciles the resource
- **THEN** the finalizer remains unchanged

#### Scenario: Finalizer-add reconcile requests immediate requeue
- **GIVEN** a freshly-created ModuleRelease without the `releases.opmodel.dev/cleanup` finalizer
- **WHEN** the controller's first `Reconcile` call adds the finalizer
- **THEN** the reconcile returns a result with `Requeue=true` so the workqueue re-enqueues the request directly (bypassing the predicate)
- **AND** the next reconcile starts within the rate limiter's normal cadence (typically under one second), not the periodic resync window

#### Scenario: Status conditions appear after creation without manual triggers
- **GIVEN** a ModuleRelease created via the API server
- **WHEN** the controller observes the create event and reconciles via the manager
- **THEN** within a small bounded window (e.g., 10 seconds in envtest, single-digit seconds in production) `status.conditions` contains at least one of `Ready`, `Reconciling`, or `Stalled`
- **AND** no manual `kubectl edit`, periodic-resync, or other external trigger is required to produce these conditions

### Requirement: Deletion cleanup with prune enabled
When a ModuleRelease with `spec.prune=true` is deleted, the controller MUST delete all resources listed in `status.inventory.entries`, respecting safety exclusions.

#### Scenario: Delete all owned resources on CR deletion
- **GIVEN** a ModuleRelease with `spec.prune=true`, a non-zero `DeletionTimestamp`, and inventory entries for ConfigMap `foo` and Deployment `bar`
- **WHEN** the controller reconciles the resource
- **THEN** ConfigMap `foo` and Deployment `bar` are deleted from the cluster
- **AND** the `releases.opmodel.dev/cleanup` finalizer is removed
- **AND** the ModuleRelease deletion completes

#### Scenario: Safety exclusions during deletion
- **GIVEN** a ModuleRelease with `spec.prune=true`, a non-zero `DeletionTimestamp`, and inventory entries including a Namespace and a CRD
- **WHEN** the controller reconciles the resource
- **THEN** the Namespace and CRD are NOT deleted
- **AND** all other inventory entries are deleted
- **AND** the finalizer is removed

### Requirement: Deletion with prune disabled orphans resources
When a ModuleRelease with `spec.prune=false` is deleted, the controller MUST remove the finalizer without deleting any resources.

#### Scenario: Orphan resources when prune is false
- **GIVEN** a ModuleRelease with `spec.prune=false` and a non-zero `DeletionTimestamp`
- **WHEN** the controller reconciles the resource
- **THEN** no resources are deleted
- **AND** the `releases.opmodel.dev/cleanup` finalizer is removed
- **AND** the ModuleRelease deletion completes

### Requirement: Suspend does not block deletion
The controller MUST perform deletion cleanup even when `spec.suspend=true`.

#### Scenario: Cleanup proceeds despite suspend
- **GIVEN** a ModuleRelease with `spec.suspend=true` and a non-zero `DeletionTimestamp`
- **WHEN** the controller reconciles the resource
- **THEN** deletion cleanup proceeds normally (prune if enabled, then remove finalizer)

### Requirement: Partial cleanup failure blocks finalizer removal
If some resources fail to delete, the controller MUST NOT remove the finalizer.

#### Scenario: Failed cleanup retains finalizer
- **GIVEN** a ModuleRelease being deleted with inventory entries, where one resource fails to delete (e.g., RBAC insufficient)
- **WHEN** the controller reconciles the resource
- **THEN** successfully deletable resources are deleted
- **AND** the finalizer is NOT removed
- **AND** the controller requeues for retry
