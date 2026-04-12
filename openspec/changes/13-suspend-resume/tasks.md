## 1. Suspend check in Phase 0

- [ ] 1.1 Add suspend check in Phase 0 after finalizer/deletion handling, before Phase 1
- [ ] 1.2 When `spec.suspend=true`: set `Ready=False` reason `Suspended`, remove `Reconciling` and `Stalled` conditions
- [ ] 1.3 Patch status and return `ctrl.Result{}` (no requeue) when suspended
- [ ] 1.4 Log "Reconciliation is suspended" at info level with name/namespace keys

## 2. Resume behavior

- [ ] 2.1 Verify that spec generation change on unsuspend triggers reconcile via default predicate (no custom logic needed)
- [ ] 2.2 Log "Reconciliation resumed" at info level when previously-suspended resource enters normal reconcile

## 3. Tests

- [ ] 3.1 Write envtest test: suspend=true skips reconciliation and sets correct conditions
- [ ] 3.2 Write envtest test: suspend=true preserves existing status (inventory, digests, history)
- [ ] 3.3 Write envtest test: unsuspend triggers full reconcile
- [ ] 3.4 Write envtest test: deletion proceeds despite suspend=true (covered by finalizer-and-deletion but verify here)

## 4. Validation

- [ ] 4.1 Run `make fmt vet lint test` and verify all checks pass
