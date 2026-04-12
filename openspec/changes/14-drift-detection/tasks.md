## 1. Drift detection function

- [ ] 1.1 Define `DriftResult` and `DriftedResource` structs in `internal/apply/drift.go`
- [ ] 1.2 Implement `DetectDrift(ctx, resourceManager, resources) (DriftResult, error)` using SSA dry-run
- [ ] 1.3 Compare dry-run result against desired objects to identify drifted resources

## 2. Condition constants

- [ ] 2.1 Add `DriftedCondition` type constant to `internal/status/`
- [ ] 2.2 Add `DriftDetected` reason constant to `internal/status/`
- [ ] 2.3 Implement `MarkDrifted(obj, count int)` and `ClearDrifted(obj)` helpers

## 3. Phase 4 integration

- [ ] 3.1 Add drift detection call in Phase 4 after no-op digest check
- [ ] 3.2 Set `Drifted=True` when drift found, clear after successful apply in Phase 5
- [ ] 3.3 On dry-run failure: log warning, increment `failureCounters.drift`, continue reconcile
- [ ] 3.4 Ensure drift detection runs even on no-op reconciles

## 4. Tests

- [ ] 4.1 Write unit test: `DetectDrift` returns empty result when no drift
- [ ] 4.2 Write unit test: `DetectDrift` identifies drifted resources
- [ ] 4.3 Write envtest test: drift detected sets `Drifted=True` condition
- [ ] 4.4 Write envtest test: successful apply clears `Drifted` condition
- [ ] 4.5 Write envtest test: drift on no-op reconcile preserves `Ready=True`

## 5. Validation

- [ ] 5.1 Run `make fmt vet lint test` and verify all checks pass
