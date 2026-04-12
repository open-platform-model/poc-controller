## 1. Counter helper functions

- [ ] 1.1 Implement `EnsureCounters(status *ModuleReleaseStatus) *FailureCounters` — initializes if nil
- [ ] 1.2 Implement `IncrementCounter(counters *FailureCounters, field string)` in `internal/status/`
- [ ] 1.3 Implement `ResetCounter(counters *FailureCounters, field string)` in `internal/status/`

## 2. Phase 7 integration

- [ ] 2.1 Collect phase outcomes during phases 1-6 (which phases succeeded/failed)
- [ ] 2.2 In Phase 7, apply counter updates: increment failed phase counters, reset succeeded phase counters
- [ ] 2.3 Reset `reconcile` counter on overall success, increment on any failure

## 3. Tests

- [ ] 3.1 Write unit test: `EnsureCounters` initializes nil counters
- [ ] 3.2 Write unit test: `IncrementCounter` increments correct field
- [ ] 3.3 Write unit test: `ResetCounter` zeros correct field without affecting others
- [ ] 3.4 Write envtest test: failed reconcile increments reconcile counter
- [ ] 3.5 Write envtest test: successful reconcile resets counters to 0

## 4. Validation

- [ ] 4.1 Run `make fmt vet lint test` and verify all checks pass
