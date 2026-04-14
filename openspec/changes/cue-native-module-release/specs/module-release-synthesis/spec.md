## Module Release Synthesis

### MODIFIED: ModuleRelease CR spec

The `spec.sourceRef` field is REMOVED.

The `spec.module` field is MODIFIED:

| Field | Old | New |
|-------|-----|-----|
| `spec.module.path` | Sub-path within Flux source artifact | CUE module import path (e.g. `opmodel.dev/modules/cert_manager@v0`) |
| `spec.module.version` | Not present | Required. Pinned CUE module version (e.g. `v0.2.1`) |
| `spec.sourceRef` | Required. Reference to Flux OCIRepository | REMOVED |

The controller MUST reject a ModuleRelease CR that:
- Has an empty `spec.module.path`.
- Has an empty `spec.module.version`.

### ADDED: Registry configuration

The controller MUST accept a `--registry` flag for CUE registry configuration.

The controller MUST read the `OPM_REGISTRY` environment variable as a fallback when
`--registry` is not provided.

The controller MUST set `CUE_REGISTRY` to the resolved registry value before any
CUE module evaluation.

If neither `--registry` nor `OPM_REGISTRY` is set, the controller MUST use CUE's
default registry resolution behavior.

### ADDED: Release synthesis

On reconcile, the controller MUST synthesize a temporary CUE module containing:
- A `cue.mod/module.cue` declaring dependencies on the target module and the catalog.
- A `release.cue` that imports `#ModuleRelease` from the catalog and the target module,
  binds `#module:` to the imported module, and sets `metadata.name` and
  `metadata.namespace` from the CR.

The controller MUST clean up the temporary directory after evaluation completes
(success or failure).

### MODIFIED: Reconcile behavior

**Previous**: Controller watches Flux OCIRepository, fetches artifact on source change,
extracts to local directory, loads raw `#Module`, passes to render pipeline.

**New**: Controller synthesizes `#ModuleRelease` CUE package from CR fields. CUE's
module system resolves the target module from the OCI registry. The controller loads
the synthesized package, fills values, and passes the resulting `#ModuleRelease` to
the existing render pipeline.

The controller MUST reconcile when the ModuleRelease CR is created or updated.
The controller MUST NOT poll the OCI registry for new module versions.

### MODIFIED: Status reporting

The `status.source` field MAY be updated to reflect:
- The CUE module path and version (from `spec.module`).
- Whether module resolution from the registry succeeded.

The `status.conditions` MUST report:
- `Ready=True` when the module is successfully resolved, rendered, and applied.
- `Ready=False` with reason `ResolutionFailed` when CUE cannot resolve the module
  from the registry.
- `Ready=False` with reason `RenderFailed` when CUE evaluation or rendering fails.
- `Stalled=True` when the failure is not transient (e.g. module path does not exist).

### Scenarios

#### Happy path

1. User creates ModuleRelease CR with `spec.module.path` and `spec.module.version`.
2. Controller synthesizes `#ModuleRelease` CUE package.
3. CUE resolves module from OCI registry.
4. Evaluation produces concrete `components`.
5. Render pipeline generates Kubernetes resources.
6. Resources are applied via SSA.
7. `status.conditions` shows `Ready=True`.

#### Module not found in registry

1. User creates ModuleRelease CR with invalid `spec.module.path`.
2. Controller synthesizes package.
3. CUE fails to resolve module from registry.
4. `status.conditions` shows `Ready=False`, reason `ResolutionFailed`.
5. `Stalled=True` — not transient, user must fix the CR.

#### Invalid values

1. User creates ModuleRelease CR with values that do not satisfy `#config`.
2. Controller synthesizes package, CUE resolves module.
3. Value validation fails in `ParseModuleRelease`.
4. `status.conditions` shows `Ready=False`, reason `RenderFailed`.
5. `Stalled=True` — user must fix values.

#### Version upgrade

1. User updates `spec.module.version` on existing ModuleRelease CR.
2. Controller detects CR change, re-synthesizes package with new version.
3. CUE resolves new version from registry.
4. New resources rendered and applied.
5. Previous resources pruned if no longer in inventory (when `prune: true`).
