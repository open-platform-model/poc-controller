## 1. Event recorder wiring

- [ ] 1.1 Add `record.EventRecorder` field to `ModuleReleaseReconciler` struct
- [ ] 1.2 Create recorder via `mgr.GetEventRecorderFor("opm-controller")` in `cmd/main.go` or `SetupWithManager`
- [ ] 1.3 Pass recorder to reconcile orchestrator

## 2. Event emission points

- [ ] 2.1 Emit `Normal/Applied` event after successful Phase 5 with resource counts
- [ ] 2.2 Emit `Warning/ApplyFailed` event on Phase 5 failure
- [ ] 2.3 Emit `Normal/Pruned` event after successful Phase 6 with deleted count
- [ ] 2.4 Emit `Warning/PruneFailed` event on Phase 6 failure
- [ ] 2.5 Emit `Warning/SourceNotReady` event on Phase 1 soft-block
- [ ] 2.6 Emit `Warning/ArtifactFetchFailed` event on Phase 2 failure
- [ ] 2.7 Emit `Warning/RenderFailed` event on Phase 3 failure
- [ ] 2.8 Emit `Normal/Suspended` and `Normal/Resumed` events on suspend transitions
- [ ] 2.9 Emit `Normal/ReconciliationSucceeded` event on full reconcile success

## 3. Tests

- [ ] 3.1 Write envtest test: verify `Applied` event emitted after successful reconcile
- [ ] 3.2 Write envtest test: verify `Warning` event emitted on failure
- [ ] 3.3 Write envtest test: verify event messages include resource counts

## 4. Validation

- [ ] 4.1 Run `make fmt vet lint test` and verify all checks pass
