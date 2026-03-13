# OPM Naming Taxonomy

## Summary

This document defines a clean naming taxonomy for the OPM controller ecosystem.

The purpose is to reduce confusion between several naming layers that serve different roles:

- Kubernetes API groups
- resource labels and annotations
- inventory and status field names
- CUE module and registry naming

Without an explicit taxonomy, it becomes easy to overload one prefix or label family for multiple unrelated purposes.

## Why this document exists

Several different names and prefixes already exist in OPM discussions and code:

- `releases.opmodel.dev`
- `module-release.opmodel.dev/...`
- `component.opmodel.dev/name`
- `app.kubernetes.io/*`
- `opmodel.dev/...` module paths in CUE and registry configuration

These all look similar, but they do different jobs.

This document defines what each namespace is supposed to represent.

## Naming layers

### 1. Kubernetes API group names

Kubernetes API group names identify CRD families.

These are for:

- `apiVersion`
- controller APIs
- Kubernetes discovery and RBAC

Recommended release API group:

- `releases.opmodel.dev`

Examples:

- `releases.opmodel.dev/v1alpha1`
- `ModuleRelease`
- `BundleRelease`

Meaning:

- `opmodel.dev` = API domain controlled by OPM
- `releases` = this API group is about release reconciliation/runtime behavior

This namespace is specifically for Kubernetes CRDs, not for workload labels or CUE module paths.

### 2. Resource label and annotation prefixes

Resource labels and annotations identify metadata carried on rendered workload resources and controller-owned objects.

These should be split into clear families.

#### 2a. Kubernetes conventional labels

These are ecosystem-standard keys used for app metadata.

Examples:

- `app.kubernetes.io/name`
- `app.kubernetes.io/instance`
- `app.kubernetes.io/managed-by`

Recommended semantics:

- `name` = component or workload name in the rendered application domain
- `instance` = release or component instance identity depending on the final OPM convention
- `managed-by` = runtime actor, e.g. `opm-cli` or `opm-controller`

#### 2b. Release-scoped metadata

These keys identify the OPM release that produced a resource.

Existing family:

- `module-release.opmodel.dev/...`

Recommended examples:

- `module-release.opmodel.dev/name`
- `module-release.opmodel.dev/namespace`
- future annotation: `module-release.opmodel.dev/uid`

Recommended meaning:

- these keys describe release identity in the rendered-resource domain
- they are not Kubernetes API group names

#### 2c. Component-scoped metadata

These keys identify the OPM component that produced a resource.

Existing family:

- `component.opmodel.dev/...`

Recommended example:

- `component.opmodel.dev/name`

Recommended meaning:

- this identifies the module component responsible for rendering the resource
- it is distinct from inventory ownership and distinct from runtime actor identity

#### 2d. Internal OPM infrastructure metadata

These keys identify OPM-internal supporting objects or infrastructure categories.

Existing family:

- `opmodel.dev/...`

Example:

- `opmodel.dev/component=inventory`

Recommended meaning:

- use this family for OPM's own supporting objects or broad infrastructure categorization
- do not use this family as a catch-all for everything else

### 3. Inventory and status field names

Inventory and status naming should use plain schema fields rather than trying to mirror label prefixes directly.

Examples:

- `status.inventory.entries[]`
- `status.source.artifactDigest`
- `status.lastAppliedRenderDigest`
- `status.history[]`

Recommended meaning:

- field names should describe controller state cleanly
- they should not inherit label prefix conventions unnecessarily

For example, use:

- `sourceDigest`
- `configDigest`
- `renderDigest`

instead of encoding label-like prefixes into status field names.

### 4. CUE module and registry naming

This naming layer is separate from Kubernetes API groups and labels.

It identifies:

- CUE modules
- CUE package paths
- OCI registry resolution prefixes

Examples from current discussions:

- `opmodel.dev/modules/jellyfin`
- `opmodel.dev/...` CUE module namespace

Recommended meaning:

- this is the language and distribution identity of OPM modules
- it should not be conflated with Kubernetes CRD API groups

So:

- `opmodel.dev/modules/foo` is a module path
- `releases.opmodel.dev/v1alpha1` is a Kubernetes API group/version

They share a parent domain, but they are not the same namespace.

## Recommended naming taxonomy

### Kubernetes CRDs

- API group: `releases.opmodel.dev`
- kinds:
  - `ModuleRelease`
  - `BundleRelease`

### Workload labels and annotations

- Kubernetes standard labels:
  - `app.kubernetes.io/name`
  - `app.kubernetes.io/instance`
  - `app.kubernetes.io/managed-by`
- release identity labels/annotations:
  - `module-release.opmodel.dev/name`
  - `module-release.opmodel.dev/namespace`
  - future annotation `module-release.opmodel.dev/uid`
- component identity labels:
  - `component.opmodel.dev/name`
- OPM internal/supporting-object labels:
  - `opmodel.dev/component`

### Inventory and status

- `status.inventory`
- `status.source`
- `status.lastAttempted*`
- `status.lastApplied*`
- `status.history`

### CUE modules and registries

- `opmodel.dev/...` module namespace
- registry mapping through `CUE_REGISTRY` / `OPM_REGISTRY`

## Naming rules

### Rule 1: API groups are for CRDs only

Use `releases.opmodel.dev` only for Kubernetes APIs, not as a workload label prefix.

### Rule 2: Workload labels use their own families

Use:

- `app.kubernetes.io/*` for standard app labels
- `module-release.opmodel.dev/*` for release identity on rendered resources
- `component.opmodel.dev/*` for component identity
- `opmodel.dev/*` for OPM-internal infrastructure classification only

### Rule 3: Runtime actor identity is not the same as release identity

For example:

- `app.kubernetes.io/managed-by=opm-controller`

means:

- the controller currently manages the resource

It does not replace:

- `module-release.opmodel.dev/name`
- `module-release.opmodel.dev/namespace`

which describe the release identity.

### Rule 4: Do not force status field names to mirror label prefixes

Status fields should stay human-readable and schema-oriented.

Good:

- `lastAppliedRenderDigest`

Not recommended:

- field names that attempt to encode label namespaces or Kubernetes metadata conventions directly

### Rule 5: CUE module paths stay separate from Kubernetes CRD naming

Even if both use `opmodel.dev`, the naming rules are different:

- module paths identify distributable CUE content
- API groups identify Kubernetes resources

## Example mapping

### Example ModuleRelease API object

```yaml
apiVersion: releases.opmodel.dev/v1alpha1
kind: ModuleRelease
metadata:
  name: jellyfin
  namespace: apps
```

### Example rendered workload labels

```yaml
metadata:
  labels:
    app.kubernetes.io/name: jellyfin
    app.kubernetes.io/instance: jellyfin
    app.kubernetes.io/managed-by: opm-controller
    module-release.opmodel.dev/name: jellyfin
    module-release.opmodel.dev/namespace: apps
    component.opmodel.dev/name: server
```

### Example supporting inventory Secret label

```yaml
metadata:
  labels:
    app.kubernetes.io/managed-by: opm-controller
    module-release.opmodel.dev/name: jellyfin
    module-release.opmodel.dev/namespace: apps
    opmodel.dev/component: inventory
```

### Example CUE module path

```text
opmodel.dev/modules/jellyfin
```

These are all related, but each belongs to a different naming layer.

## Open questions

- Should `module-release.opmodel.dev/uid` be standardized now as an annotation?
- Should `app.kubernetes.io/instance` always be release-scoped instead of component-scoped?
- Should `BundleRelease` rendered child metadata ever get its own label family, or should release labels remain module-release-centric on actual workloads?

## Current recommendation

The current recommendation is:

- keep `releases.opmodel.dev` as the CRD API group
- keep workload metadata in separate label families
- keep `app.kubernetes.io/managed-by` runtime-owned
- keep `module-release.opmodel.dev/*` and `component.opmodel.dev/*` as rendered-resource identity metadata
- keep CUE module naming separate from Kubernetes API naming
