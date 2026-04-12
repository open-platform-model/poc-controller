## 1. Metric definitions

- [ ] 1.1 Create `internal/metrics/metrics.go` with Prometheus metric variable definitions
- [ ] 1.2 Define `ReconcileTotal` counter vec with labels `name`, `namespace`, `outcome`
- [ ] 1.3 Define `ReconcileDuration` histogram vec with labels `name`, `namespace`
- [ ] 1.4 Define `ApplyResourcesTotal` counter vec with labels `name`, `namespace`, `action`
- [ ] 1.5 Define `PruneResourcesTotal` counter vec with labels `name`, `namespace`
- [ ] 1.6 Define `InventorySize` gauge vec with labels `name`, `namespace`

## 2. Metric registration

- [ ] 2.1 Register all metrics with `controller-runtime/pkg/metrics.Registry` in `init()`
- [ ] 2.2 Import `internal/metrics` in `cmd/main.go` to trigger registration

## 3. Reconcile integration

- [ ] 3.1 Record reconcile outcome and duration in Phase 7 status commit
- [ ] 3.2 Record apply resource counts after Phase 5
- [ ] 3.3 Record prune resource count after Phase 6
- [ ] 3.4 Update inventory size gauge after Phase 7

## 4. Tests

- [ ] 4.1 Write unit test: verify metrics registered with correct names and labels
- [ ] 4.2 Write unit test: verify metric recording functions update counters/gauges correctly

## 5. Validation

- [ ] 5.1 Run `make fmt vet lint test` and verify all checks pass
