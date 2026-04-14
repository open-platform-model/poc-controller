## Tasks

### API Changes

- [ ] Update `ModuleReference` struct: add `Version` field, update `Path` field
      documentation to describe CUE import path semantics
- [ ] Remove `SourceReference` type and `SourceRef` field from `ModuleReleaseSpec`
- [ ] Run `make manifests generate` to regenerate CRDs and DeepCopy
- [ ] Update sample CR (`config/samples/releases_v1alpha1_modulerelease.yaml`)

### Release Synthesis

- [ ] Run experiment 002 to validate CUE package synthesis approach
- [ ] Create `internal/synthesis/` package with `SynthesizeRelease` function
      that writes temporary CUE module (cue.mod/module.cue + release.cue)
- [ ] Add unit tests for synthesis: correct CUE output, template rendering,
      cleanup on error
- [ ] Add integration test: synthesize → load → verify `components` exists
      and is concrete

### Registry Configuration

- [ ] Add `--registry` flag to controller manager command
- [ ] Read `OPM_REGISTRY` env var as fallback
- [ ] Set `CUE_REGISTRY` before CUE evaluation in reconcile loop
- [ ] Add test for registry config precedence (flag > env > default)

### Render Pipeline Update

- [ ] Update `internal/render/module.go`: accept module path + version instead
      of directory path (or add new entry point alongside existing)
- [ ] Wire synthesis → load → ParseModuleRelease → ProcessModuleRelease
- [ ] Update render tests to use real module patterns (`#components` not
      `components` in test fixtures)

### Reconcile Loop Update

- [ ] Remove Flux OCIRepository watch from controller setup
- [ ] Remove source-fetching logic from reconcile loop
- [ ] Replace with synthesis call using CR fields
- [ ] Update status reporting: `ResolutionFailed` condition for registry errors
- [ ] Retain `internal/source/` package (do not delete)

### Validation & Cleanup

- [ ] Update e2e tests for new CR schema (no sourceRef)
- [ ] Update kustomize samples and overlays
- [ ] Run `make fmt vet lint test`
- [ ] Run `make test-e2e` (if Kind cluster available)
- [ ] Add note to constitution about Principle I supersession
