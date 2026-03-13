# poc-controller Agent Guide

## Purpose

- This repository is a Kubebuilder-based Kubernetes controller written in Go.
- It defines and reconciles `ModuleRelease` and `BundleRelease` CRDs in `api/v1alpha1`.
- Agents should preserve Kubebuilder conventions, controller-runtime patterns, and generated-file boundaries.

## Repository Layout

- `cmd/main.go`: controller manager entrypoint and controller registration.
- `api/v1alpha1/*.go`: API schemas and shared CRD types.
- `api/v1alpha1/zz_generated.deepcopy.go`: generated DeepCopy code; do not edit.
- `internal/controller/*.go`: reconcilers and envtest-based controller tests.
- `internal/{apply,inventory,reconcile,render,source,status}`: internal domain packages.
- `config/crd/bases/*.yaml`: generated CRDs; do not edit by hand.
- `config/rbac/role.yaml`: generated RBAC; do not edit by hand.
- `config/samples/*.yaml`: sample CRs; safe to update when behavior changes.
- `test/e2e/*.go`: Kind-backed end-to-end tests guarded by the `e2e` build tag.
- `Makefile`: canonical build, generation, lint, test, and deploy entrypoints.
- `Taskfile.yml`: thin wrappers around `make` targets.

## Rules Files

- Checked for Cursor rules in `.cursor/rules/` and `.cursorrules`: none found.
- Checked for Copilot instructions in `.github/copilot-instructions.md`: none found.
- The only repo-specific agent instructions live in this `AGENTS.md`.

## Generated Files And Scaffold Boundaries

- Never hand-edit `api/v1alpha1/zz_generated.deepcopy.go`.
- Never hand-edit `config/crd/bases/*.yaml`.
- Never hand-edit `config/rbac/role.yaml`.
- Never hand-edit `PROJECT`.
- Preserve `// +kubebuilder:scaffold:*` comments.
- If you change API markers or `*_types.go`, regenerate outputs instead of editing generated artifacts.

## Build, Generate, Lint, And Test Commands

- `make help`: list supported targets.
- `make manifests`: regenerate CRDs, RBAC, and webhook manifests with `controller-gen`.
- `make generate`: regenerate DeepCopy methods.
- `make fmt`: run `go fmt ./...`.
- `make vet`: run `go vet ./...`.
- `make lint-config`: validate the golangci-lint configuration.
- `make lint`: run golangci-lint.
- `make lint-fix`: run golangci-lint with auto-fixes.
- `make build`: run generation + formatting + vet, then build `bin/manager`.
- `make run`: run the controller locally against the current kubeconfig.
- `make test`: run non-e2e tests with envtest and write `cover.out`.
- `make test-e2e`: create/use a Kind cluster, run e2e tests, then tear the cluster down.
- `make docker-build IMG=<image>`: build the controller image.
- `make deploy IMG=<image>`: deploy the controller to the current cluster.

## Single Test Commands

- Controller/unit tests use Go test plus Ginkgo v2; there is no dedicated `make test-one` target.
- Preferred package-level run: `go test ./internal/controller`.
- Run one Go test function: `go test ./internal/controller -run TestControllers`.
- Run one Ginkgo spec in controller tests:
  `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.35.0 --bin-dir ./bin -p path)" go test ./internal/controller -run TestControllers -ginkgo.focus="BundleRelease Controller"`
- Narrow further to one spec text:
  `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.35.0 --bin-dir ./bin -p path)" go test ./internal/controller -run TestControllers -ginkgo.focus="should successfully reconcile the resource"`
- If `./bin/setup-envtest` is missing or outdated, run `make setup-envtest` first.
- E2E package only: `go test -tags=e2e ./test/e2e -v -ginkgo.v`.
- Single e2e spec:
  `KIND_CLUSTER=poc-controller-test-e2e go test -tags=e2e ./test/e2e -run TestE2E -ginkgo.focus="should run successfully" -v -ginkgo.v`
- `make test` excludes `/test/e2e`; keep that behavior unless you explicitly want Kind-based coverage.

## Typical Agent Workflow

- For changes under `api/v1alpha1`, run `make manifests generate` after editing.
- For changes under `internal/` or `cmd/`, run `make fmt vet test` at minimum.
- For broader changes, prefer `make lint-fix` before final verification.
- If deployment manifests or RBAC behavior changed, consider `make build-installer` after regeneration.
- Use `task <name>` only as a convenience alias; `make` is the source of truth.

## Go Version And Tooling

- The module targets Go `1.25.3`.
- CI uses `actions/setup-go` with `go.mod` as the version source.
- Linting is driven by `golangci-lint` v2 with `gofmt` and `goimports` formatters enabled.
- A custom `logcheck` plugin is loaded from `.custom-gcl.yml` to enforce Kubernetes logging conventions.

## Formatting And Imports

- Let `gofmt` and `goimports` own formatting and import grouping.
- Use tabs and standard Go formatting; do not align fields manually.
- Keep imports grouped as: standard library, external dependencies, then local module imports.
- Use import aliases only when they clarify intent or avoid collisions.
- Common aliases already in use include `ctrl`, `logf`, `metav1`, `fluxmeta`, and versioned API aliases.
- Preserve blank lines between logical import groups.

## Naming Conventions

- Exported Go types use `PascalCase`; unexported helpers use `camelCase`.
- Receiver names are short and conventional; reconcilers use `r`.
- Kubernetes API structs use canonical names like `Spec`, `Status`, `Conditions`, `ObservedGeneration`.
- Acronyms generally follow Go conventions already present in the repo, such as `URL`, `RBAC`, `CRD`, `OCI`.
- JSON field names are lower camel case and explicit on all serialized fields.
- Keep package names short, lowercase, and singular when practical.

## Types And API Design

- Prefer concrete structs over `map[string]any` in API and controller code.
- For CRD status conditions, use `[]metav1.Condition` rather than custom condition structs.
- For timestamps and durations, use Kubernetes types like `metav1.Time` and `metav1.Duration`.
- Reuse Flux and Kubernetes reference types where the repo already does so.
- Keep API comments compatible with Kubebuilder markers and CRD generation.
- Add or update validation markers when introducing schema constraints.
- Maintain `omitempty` and `omitzero` tags consistently with existing types.

## Controller And Reconcile Style

- Keep reconciliation idempotent.
- Prefer controller-runtime patterns over ad hoc client logic.
- Fetch from the API server before mutating objects that may have changed.
- Use `.For(...)`, `.Owns(...)`, and related builder methods to declare watches.
- Keep `Reconcile` readable; push domain logic into `internal/*` packages as complexity grows.
- Return explicit `ctrl.Result{}` values rather than relying on zero-value ambiguity in complex branches.
- Keep RBAC markers on reconcilers accurate whenever API usage changes.

## Error Handling

- Return errors with context using `%w` when wrapping: `fmt.Errorf("failed to ...: %w", err)`.
- Use package-level sentinel errors only when callers need to branch on them.
- Do not swallow errors silently; either return them or log with clear context.
- Keep error messages lowercase unless they start with a proper noun or identifier.
- In tests, prefer `Expect(err).NotTo(HaveOccurred())` or `Expect(...).To(Succeed())`.
- In cleanup paths, it is acceptable to ignore best-effort errors intentionally, but do so explicitly.

## Logging Conventions

- Use structured logging via controller-runtime loggers.
- Follow Kubernetes log style: capitalize the message, no trailing period, use active/past-tense wording.
- Include balanced key/value pairs such as object name and namespace.
- Prefer messages like `"Failed to create controller"` over generic text.
- Keep high-volume logs out of hot paths unless they materially aid debugging.

## Testing Conventions

- Unit and controller tests use Ginkgo v2 and Gomega.
- Envtest-backed tests live in `internal/controller` and bootstrap CRDs from `config/crd/bases`.
- E2E tests require Kind and may install CertManager unless `CERT_MANAGER_INSTALL_SKIP=true` is set.
- Keep test names descriptive; match the existing `Describe` / `Context` / `It` style.
- Use `Eventually` for asynchronous Kubernetes behavior instead of sleeps.
- Prefer package-local helpers when assertions are repeated.

## Lint Expectations

- Current linters include `errcheck`, `ginkgolinter`, `gocyclo`, `govet`, `misspell`, `nakedret`, `revive`, `staticcheck`, `unused`, and others.
- `lll` and `dupl` are relaxed in some paths, but do not rely on exemptions unnecessarily.
- `logcheck` will flag malformed structured logging.
- `ginkgolinter` means Ginkgo tests should follow idiomatic Ginkgo patterns.

## Kubernetes And Kubebuilder Specific Guidance

- Use `kubebuilder` scaffolding commands for new APIs or webhooks instead of creating scaffold files manually.
- Do not remove license headers from scaffolded Go files.
- Keep API comments and markers near the fields and types they describe.
- When changing CRD schemas, verify sample manifests still make sense.
- The project is single-group today; APIs live under `api/v1alpha1` with group `releases.opmodel.dev`.

## Verification Checklist For Agents

- Ran `make manifests generate` after API or marker changes.
- Ran `make fmt vet test` after Go code changes.
- Ran `make lint` or `make lint-fix` for non-trivial edits.
- Avoided edits to generated files and scaffold markers.
- Mentioned if e2e verification was skipped because Kind or cluster setup was not available.
