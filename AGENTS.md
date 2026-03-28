# AGENTS.md - poc-controller repository guide

## Purpose

- This repository is a Kubebuilder-based Kubernetes controller written in Go.
- It defines and reconciles `ModuleRelease` and `BundleRelease` CRDs in `api/v1alpha1`.
- Agents should preserve controller-runtime patterns, Kubebuilder markers, and generated-file boundaries.

## Repository Rules

- Repo-specific agent guidance lives in this `AGENTS.md` and `CONSTITUTION.md`.

## Entrypoint

- Agents entering this repository should read these documents first, in order.
- `AGENTS.md`: repository-specific commands, workflows, style, and verification expectations.
- `CONSTITUTION.md`: root-level engineering principles and change-shaping rules.
- `openspec/config.yaml`: normative OpenSpec constitutional source for this repository.
- `Makefile`: authoritative build, generate, lint, and test entrypoints.
- `docs/STYLE.md`: documentation prose style rules for this repo.

## Repository Layout

- `adr/` - Architecture Decision Records
- `cmd/main.go`: manager entrypoint and controller registration.
- `api/v1alpha1/*.go`: API schemas, validation markers, and shared CRD types.
- `api/v1alpha1/zz_generated.deepcopy.go`: generated DeepCopy code; do not edit by hand.
- `internal/controller/*.go`: reconcilers and envtest-backed controller tests.
- `internal/{apply,inventory,reconcile,render,source,status}`: domain packages used by controllers.
- `config/crd/bases/*.yaml`: generated CRDs; regenerate instead of hand-editing.
- `config/rbac/role.yaml`: generated RBAC; regenerate instead of hand-editing.
- `config/samples/*.yaml`: sample manifests; update when behavior or schema changes.
- `test/e2e/*.go`: Kind-backed end-to-end tests guarded by the `e2e` build tag.
- `Makefile`: source of truth for generation, build, lint, test, and deploy workflows.
- `Taskfile.yml`: convenience wrappers that delegate directly to `make` targets.

## Architecture Decision Records

ADRs capture significant technical decisions with their context and consequences.

- Location: `adr/`
- Template: `adr/TEMPLATE.md`
- Naming: `NNN-kebab-case-title.md` (three-digit, zero-padded)

### Creating a new ADR

1. Copy `adr/TEMPLATE.md` to `adr/NNN-title.md` using the next available number.
2. Set status to `Proposed`.
3. Fill in Context, Decision, and Consequences.
4. Update status to `Accepted` once the decision is agreed on.

### Updating an ADR

- Never delete an ADR — update its status instead.
- To retire a decision: set status to `Deprecated`.
- To replace a decision: set status to `Superseded by ADR-NNN` and create the new ADR.
- One decision per ADR.

## Generated Files And Scaffold Boundaries

- Never hand-edit `api/v1alpha1/zz_generated.deepcopy.go`.
- Never hand-edit `config/crd/bases/*.yaml` or `config/rbac/role.yaml`.
- Never hand-edit `PROJECT`.
- Preserve `// +kubebuilder:scaffold:*` comments and license headers in scaffolded files.
- If you change API markers, schema fields, or `*_types.go`, run `make manifests generate`.

## Build And Dev Commands

### Core Commands

- `make help`: list available targets.
- `make manifests`: regenerate CRDs, RBAC, and webhook manifests with `controller-gen`.
- `make generate`: regenerate DeepCopy methods.
- `make fmt`: run `go fmt ./...`.
- `make vet`: run `go vet ./...`.
- `make lint-config`: verify the golangci-lint configuration.
- `make lint`: run golangci-lint.
- `make lint-fix`: run golangci-lint with auto-fixes.
- `make build`: run generation, formatting, vet, then build `bin/manager`.
- `make run`: run the controller locally against the current kubeconfig.
- `make test`: run non-e2e tests with envtest and write `cover.out`.
- `make setup-test-e2e`: create the Kind cluster if it does not exist.
- `make test-e2e`: run e2e tests against Kind, then clean the cluster up.
- `make build-installer`: render `dist/install.yaml` from `config/default`.
- `make docker-build IMG=<image>` and `make docker-push IMG=<image>`: build or publish images.
- `make deploy IMG=<image>` / `make undeploy`: install or remove the controller from a cluster.

### Single Test Commands

- There is no dedicated `make test-one` target; use `go test` directly.
- Package-level controller tests: `go test ./internal/controller`.
- Single Go test entrypoint: `go test ./internal/controller -run TestControllers`.
- Before focused envtest runs, ensure binaries exist with `make setup-envtest`.
- Reuse envtest binaries explicitly when running controller specs:
  `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.35.0 --bin-dir ./bin -p path)" go test ./internal/controller -run TestControllers`.
- Focus one Ginkgo suite or spec:
  `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.35.0 --bin-dir ./bin -p path)" go test ./internal/controller -run TestControllers -ginkgo.focus="BundleRelease Controller"`.
- Focus one spec text more narrowly:
  `KUBEBUILDER_ASSETS="$(./bin/setup-envtest use 1.35.0 --bin-dir ./bin -p path)" go test ./internal/controller -run TestControllers -ginkgo.focus="should successfully reconcile the resource"`.
- E2E package only: `go test -tags=e2e ./test/e2e -v -ginkgo.v`.
- Single e2e spec:
  `KIND_CLUSTER=poc-controller-test-e2e go test -tags=e2e ./test/e2e -run TestE2E -ginkgo.focus="should run successfully" -v -ginkgo.v`.
- `make test` intentionally excludes `/test/e2e`; do not fold Kind-based tests into the default unit path.

## Working Style for Agents

- For edits under `api/v1alpha1`, run `make manifests generate` after changing schema or markers.
- For Go changes under `cmd/` or `internal/`, run `make fmt vet test` at minimum.
- For non-trivial changes, prefer `make lint` or `make lint-fix` before finishing.
- If manifests or RBAC behavior changed, consider `make build-installer` so generated deploy output stays aligned.
- Use `task <name>` only as a shorthand; `make` remains the authoritative interface.

## Go Version And Tooling

- The module targets Go `1.25.3`.
- Primary frameworks are controller-runtime, Kubebuilder-generated APIs, Flux source types, Ginkgo v2, and Gomega.
- Linting is driven by `golangci-lint` v2 with `gofmt` and `goimports` enabled as formatters.
- A custom `logcheck` plugin is built from `.custom-gcl.yml` and enforces Kubernetes logging conventions.
- Envtest and related binaries are installed into `./bin` through the Makefile targets.

## Formatting And Imports

- Let `gofmt` and `goimports` own layout, spacing, and import grouping.
- Use standard Go formatting with tabs; do not vertically align fields or assignments manually.
- Keep imports grouped as: standard library, third-party dependencies, then local module imports.
- Preserve blank lines between import groups.
- Use aliases only when they improve clarity or match existing conventions, such as `ctrl`, `logf`, `metav1`, or versioned API aliases.
- Avoid unused helpers and speculative abstractions; keep files idiomatic and compact.

## Naming And API Design

- Exported identifiers use `PascalCase`; unexported helpers use `camelCase`.
- Keep receiver names short and conventional; reconcilers use `r`.
- Package names should be lowercase and concise.
- Follow Kubernetes naming patterns like `Spec`, `Status`, `Conditions`, and `ObservedGeneration`.
- JSON field names should be explicit lowerCamelCase.
- Prefer concrete structs over `map[string]any`, especially in APIs and reconciliation flow.
- Reuse Kubernetes and Flux reference types where the repo already does so.
- Use `[]metav1.Condition` for status conditions rather than introducing custom condition structs.
- For API timestamps and durations, prefer Kubernetes types such as `metav1.Time` and `metav1.Duration`.
- Maintain `omitempty` and `omitzero` tags consistently with the existing API style.

## Controller And Reconcile Style

- Keep reconciliation idempotent and safe to retry.
- Prefer controller-runtime helpers and patterns over ad hoc Kubernetes client logic.
- Fetch fresh objects before mutating state that may have changed concurrently.
- Declare watches with builder methods like `.For(...)` and `.Owns(...)`.
- Keep `Reconcile` readable; move complex domain logic into `internal/*` packages.
- Return explicit `ctrl.Result{}` values in non-trivial branches.
- Keep RBAC markers accurate whenever a reconciler starts reading or writing new resources.

## Error Handling And Logging

- Wrap returned errors with context using `%w`, for example `fmt.Errorf("failed to render bundle: %w", err)`.
- Keep error messages lowercase unless they start with a proper noun or identifier.
- Do not silently swallow errors; either return them or log intentional best-effort failures clearly.
- Use sentinel errors only when callers need to branch on them.
- Use structured controller-runtime logging with balanced key/value pairs.
- Follow Kubernetes log style: capitalized message, no trailing period, meaningful action wording.
- Include identifying keys such as object name and namespace when logging reconciliation events.

## Testing Style

- Tests use Ginkgo v2 and Gomega.
- Envtest-backed controller tests live in `internal/controller` and load CRDs from `config/crd/bases`.
- E2E tests require Kind and may install CertManager unless `CERT_MANAGER_INSTALL_SKIP=true` is set.
- Prefer descriptive `Describe`, `Context`, and `It` text that reads well with `-ginkgo.focus`.
- Use `Eventually` for asynchronous Kubernetes behavior instead of sleeps.
- In tests, prefer `Expect(err).NotTo(HaveOccurred())` or `Expect(...).To(Succeed())`.
- Add package-local helpers when assertions repeat, but keep setup readable.

## Lint Expectations

- Enabled linters include `errcheck`, `ginkgolinter`, `gocyclo`, `govet`, `misspell`, `modernize`, `revive`, `staticcheck`, `unused`, and others.
- `gofmt` and `goimports` are enforced via golangci-lint formatters.
- `logcheck` validates structured logging calls and balanced key/value parameters.
- `lll` and `dupl` are relaxed in parts of `api/*` and `internal/*`, but do not rely on those exclusions unnecessarily.
- Write Ginkgo code idiomatically; `ginkgolinter` will flag non-idiomatic patterns.

## Verification Checklist For Agents

- Ran `make manifests generate` after API or marker changes.
- Ran `make fmt vet test` after meaningful Go changes.
- Ran `make lint` or `make lint-fix` for non-trivial edits.
- Avoided manual edits to generated files and scaffold markers.
- Mentioned if e2e verification was skipped because Kind or cluster setup was unavailable.
