## ADDED Requirements

### Requirement: Module rendering from directory and values
The `internal/render` package MUST provide a `RenderModule` function that takes a CUE module directory path and optional CRD values, and returns rendered Kubernetes resources with inventory entries.

#### Scenario: Successful rendering
- **WHEN** the directory contains a valid CUE module and values are compatible with the module's `#config` schema
- **THEN** `RenderModule` returns a `RenderResult` with non-empty `Resources` and `InventoryEntries`

#### Scenario: Invalid values
- **WHEN** the provided values do not conform to the module's `#config` schema
- **THEN** `RenderModule` returns an error wrapping `pkg/errors.ConfigError`

#### Scenario: No values provided
- **WHEN** `values` is nil (module has no required config or has defaults)
- **THEN** `RenderModule` succeeds using the module's default values

#### Scenario: Module has no components
- **WHEN** the CUE module evaluates to a spec with no components
- **THEN** `RenderModule` returns an error indicating no components found

### Requirement: Runtime labels injection
The render bridge MUST fill the catalog's `#TransformerContext.#runtimeName` field with the value `app.kubernetes.io/managed-by` should take on rendered resources. For the controller, this value is `core.LabelManagedByControllerValue` (`"opm-controller"`). The catalog declares `#runtimeName` as mandatory; CUE evaluation MUST fail if the field is unset, preventing silent fallback to a wrong or empty value. The render bridge MUST NOT use the deprecated `#runtimeLabels` mechanism (removed from the catalog as part of this change).

#### Scenario: Managed-by label present on resources
- **WHEN** rendering completes successfully through `internal/render` or `pkg/render`
- **THEN** every resource in the result carries `app.kubernetes.io/managed-by: opm-controller` in its labels

#### Scenario: Render fails fast when runtime identity not injected
- **WHEN** a code path constructs `#TransformerContext` (directly or through CUE evaluation) without filling `#runtimeName`
- **THEN** CUE evaluation returns an error mentioning the missing required field
- **AND** no resources are produced

#### Scenario: Runtime identity is end-to-end consistent with Go constants
- **GIVEN** the controller's render pipeline executed against a minimal valid `#ModuleRelease`
- **WHEN** the rendered resources are inspected
- **THEN** the value of `metadata.labels["app.kubernetes.io/managed-by"]` exactly equals `core.LabelManagedByControllerValue`
- **AND** the value of `metadata.labels["module-release.opmodel.dev/uuid"]` is non-empty (sanity check that ownership labels continue to flow through the catalog merge)

### Requirement: RenderResult includes inventory entries
The `RenderResult` MUST include `[]v1alpha1.InventoryEntry` built from the rendered resources, using the inventory bridge from change 1.

#### Scenario: Inventory entries match resources
- **WHEN** rendering produces N resources
- **THEN** the result contains N inventory entries with correct Group, Kind, Namespace, Name, and Component fields

### Requirement: CRD values to CUE conversion
The render bridge MUST convert `v1alpha1.RawValues` (JSON bytes) to a `cue.Value` for the CLI pipeline.

#### Scenario: Valid JSON values
- **WHEN** `RawValues` contains valid JSON
- **THEN** the JSON is compiled into a `cue.Value` and passed to `ParseModuleRelease`

#### Scenario: Invalid JSON values
- **WHEN** `RawValues` contains malformed JSON
- **THEN** `RenderModule` returns an error before attempting CUE evaluation
