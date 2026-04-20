# Impersonation and Privilege Escalation

## Summary

This document analyzes the privilege-escalation surface created by `spec.serviceAccountName` on `ModuleRelease` and `BundleRelease`. It is an analysis document, not a decision record.

The controller currently uses ServiceAccount impersonation to apply rendered resources (see [ssa-ownership-and-drift-policy.md](ssa-ownership-and-drift-policy.md) and the archived change `2026-04-13-15-serviceaccount-impersonation`). Impersonation keeps the controller's own RBAC narrow, but it shifts the privilege boundary onto *which ServiceAccount a release is allowed to reference*. If that boundary is not enforced, a tenant who can create a `ModuleRelease` can cause the controller to apply arbitrary resources under a privileged identity, including `cluster-admin`.

This is the same class of problem Kubernetes already has for `Pod.spec.serviceAccountName`. It is worth solving deliberately rather than by accident.

> **Warning:** The options in this document are exploratory. None are implemented beyond the current impersonation mechanism. Do not treat the Tension Map as a design decision.

## Related Non-Goal

[`scope-and-non-goals.md`](scope-and-non-goals.md) §9 explicitly defers "advanced impersonation design" and "security boundary guarantees across tenants" to post-PoC. This document exists so that when those deferred items are picked up, the option space is already mapped.

## Threat Model

### The escalation gadget

```
Tenant creates ModuleRelease:
  spec.serviceAccountName: cluster-admin-sa   ← the step that matters
  rendered manifests: <anything the tenant wants>

Controller:
  1. Checks that cluster-admin-sa exists in the release namespace.
  2. Builds an impersonated client as system:serviceaccount:<ns>:cluster-admin-sa.
  3. Server-side-applies the rendered manifests under that identity.

Result:
  Tenant has caused arbitrary workloads and RBAC to be applied as cluster-admin.
```

The rendered manifests do not have to contain RBAC. A `DaemonSet` that runs a privileged container, or a `ClusterRoleBinding` granting the tenant's user `cluster-admin`, are equally terminal. **The escalation vector is the SA reference, not the manifest contents.**

This mirrors the classic Kubernetes problem of "who may run a Pod with `serviceAccountName: X`". Kubernetes has historically resolved that through admission: a principal may reference an SA in a Pod only if they can use that SA. OPM inherited the problem the moment `spec.serviceAccountName` was added.

### What does *not* mitigate the attack

- **Narrow controller RBAC.** The controller itself holds only `impersonate` on `serviceaccounts`. That is the point of impersonation. It does not limit which SA is targeted.
- **Module admission on rendered RBAC kinds.** Rejecting modules that render `ClusterRoleBinding` blocks one payload. It does not block privileged DaemonSets, Secret reads, or exec-into-kube-system Pods.
- **Pre-flight `SubjectAccessReview` against the target SA.** SAR will return `allowed: true` for every verb, because the target SA legitimately has those verbs. That is the whole problem.

### What the controller already enforces

From the current `serviceaccount-impersonation` spec:

| Requirement | Current behavior |
|-------------|------------------|
| Missing SA | Reconcile stalls; `Ready=False`. |
| Controller lacks `impersonate` RBAC | Reconcile stalls; `Ready=False`. |
| Cross-namespace SA reference | Rejected; SA must be in same namespace as the release. |
| SA empty (unset) | Controller uses its own identity. Apply fails because the controller's RBAC does not include workload verbs. |
| SA missing during deletion cleanup | Release stalls with `Ready=False, reason=DeletionSAMissing`. Finalizer retained. Operator must restore the SA, set `spec.prune=false`, or set annotation `opm.dev/force-delete-orphan=true` to drop the finalizer and orphan the inventory. Controller NEVER falls back to its own client on the deletion path. |

Same-namespace and "no fallback identity" are useful but insufficient. A cluster-admin SA can exist in any namespace, including a tenant namespace if an operator placed one there by mistake or by installer default.

## Current RBAC Surface

From `config/rbac/role.yaml`:

| API group | Resources | Verbs | Purpose |
|-----------|-----------|-------|---------|
| `""` | `events` | `create`, `patch` | Emit controller events. |
| `""` | `serviceaccounts` | `get`, `impersonate` | Build impersonated clients. |
| `events.k8s.io` | `events` | `create`, `patch`, `update` | Modern events API. |
| `releases.opmodel.dev` | `bundlereleases`, `modulereleases`, `releases` | full | Own CRDs. |
| `releases.opmodel.dev` | `*/finalizers` | `update` | Finalizer management. |
| `releases.opmodel.dev` | `*/status` | `get`, `patch`, `update` | Status subresource. |
| `source.toolkit.fluxcd.io` | `buckets`, `gitrepositories`, `ocirepositories` | `get`, `list`, `watch` | Resolve Flux sources. |

The controller has no wildcard verbs and no direct write access to workload kinds. Workload mutation only happens through impersonation. This is a strong baseline; the escalation problem is layered on top of it, not caused by it.

## Option Space

Eight distinct approaches, grouped at the end into three philosophical stances.

### Option A — Sudoers model

The ServiceAccount opts in to being used by releases. An annotation or label on the SA (for example `opm.dev/allowed-release-selector`) declares which releases may impersonate it. The controller refuses to impersonate unless the release matches. Default: deny.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: deploy-sa
  namespace: team-a
  annotations:
    opm.dev/allowed-releases: "team-a/*"
```

- **Analogy:** `/etc/sudoers`. The target identity declares who may assume it.
- **Who decides:** The SA owner (often the platform admin who created it).
- **Blocks the gadget?** Yes, for any SA that is not annotated. A cluster-admin SA with no annotation cannot be used at all.
- **Failure mode:** Admins annotate an SA too permissively.

### Option B — SubjectAccessReview on the creator

Before impersonating SA X, the controller runs a `SubjectAccessReview` asking "does the principal that created or last updated this ModuleRelease have `impersonate` on `serviceaccounts/X`?". If the creator could not `kubectl --as=system:serviceaccount:<ns>:<sa>` themselves, the controller will not do it for them.

- **Analogy:** AWS `sts:AssumeRole` trust policy. You can only delegate privileges you already hold.
- **Who decides:** Cluster RBAC, via `impersonate` verbs granted to users/groups.
- **Blocks the gadget?** Yes, unless the tenant already has `impersonate` on the target SA, in which case the escalation exists independently of OPM.
- **Failure mode:** CI bots that create releases on behalf of tenants need `impersonate` grants, which is a large blast radius.

### Option C — Capability negotiation

The CUE module declares the Kubernetes verbs and resources it needs. The controller (or a sibling component) synthesizes a per-release ServiceAccount bound to a Role containing only those verbs. Cluster-admin is never reachable because the module never asked for it. Admission compares the rendered resource set against the declared capabilities and rejects mismatches.

```yaml
# hypothetical module declaration
capabilities:
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["create", "update", "patch"]
  - apiGroups: [""]
    resources: ["services", "configmaps"]
    verbs: ["create", "update", "patch"]
```

- **Analogy:** OCI image capabilities, OPA bundle signing, Nix derivations. The contract is explicit and diffable.
- **Who decides:** The module author declares; the platform admin may still gate by capability class (for example, reject modules that declare `rbac.authorization.k8s.io/*`).
- **Blocks the gadget?** Yes. The SA derived for the release never has verbs the module did not declare.
- **Failure mode:** Implementation cost is high. The capability surface needs versioning and diffing. Legitimate edge cases (CRDs, webhooks, cluster-scoped resources) require careful policy.

### Option D — RBAC-manifest quarantine

Rendered `Role`, `RoleBinding`, `ClusterRole`, and `ClusterRoleBinding` objects are not applied in the same pass as workloads. They are held in `status` as a pending set and require an out-of-band approval resource (`ReleaseEscalation` or similar) signed by a platform admin before the controller applies them.

- **Analogy:** CI pipelines requiring manual approval for production deploys. Two-person rule.
- **Who decides:** A platform admin, per release, per RBAC payload.
- **Blocks the gadget?** Partial. It stops rendered RBAC from being applied silently, but does not stop a privileged impersonated SA from running workloads that escalate by other means.
- **Failure mode:** Admin burden scales with release count.

### Option E — Label-gated ServiceAccounts

Only SAs with a specific label (for example `opm.dev/deployer=true`) are eligible for impersonation. Admins control who can set that label via a dedicated RBAC rule. Tenants choose from the approved pool.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: deploy-sa
  namespace: team-a
  labels:
    opm.dev/deployer: "true"
```

- **Analogy:** PodSecurityStandards-style enforcement via labels.
- **Who decides:** Platform admins, by writing RBAC that restricts who may set the label.
- **Blocks the gadget?** Only if the admin never labels a cluster-admin SA as `deployer=true`. The discipline problem is not solved by the mechanism.
- **Failure mode:** Labels are easy to apply and easy to get wrong. Does not catch a labeled-but-overprivileged SA.

### Option F — Creator-identity apply

The controller drops impersonation of an arbitrary SA and instead executes apply under the identity of the principal that created the release, using `TokenRequest` or user-extras impersonation. A release only succeeds if the creator could have applied the manifests themselves.

- **Analogy:** `kubectl apply` run by the user directly. No delegation.
- **Who decides:** Kubernetes RBAC on the creating principal.
- **Blocks the gadget?** Yes. There is no SA to choose; the identity is fixed to the creator.
- **Failure mode:** Kills the CI-bot-deploys-for-tenant pattern. A bot that creates releases on behalf of many tenants needs every tenant's privileges. Breaks the usual GitOps model.

### Option G — Reject RBAC kinds in modules

Admission rejects any module whose rendered set contains `rbac.authorization.k8s.io/*` (and possibly other sensitive kinds: `CustomResourceDefinition`, `MutatingWebhookConfiguration`, `ValidatingAdmissionPolicy`). RBAC is explicitly out-of-scope for modules; it is provisioned by Helm, Terraform, or platform tooling separately.

- **Analogy:** Kubernetes Policy Libraries; namespace-scoped controllers like Argo Rollouts that refuse to manage cluster-scoped objects.
- **Who decides:** A platform policy author, once.
- **Blocks the gadget?** Blocks RBAC payloads. Does not block privileged workloads. Useful as a *defense in depth* layer, not a standalone answer.
- **Failure mode:** Modules that legitimately install their own operator (CRD + controller + RBAC) cannot do so. Requires an escape hatch for trusted modules.

### Option H — Scoped impersonation

The controller strips cluster-scoped verbs and certain namespaced verbs from any impersonated apply. It runs `SelfSubjectRulesReview` against the impersonated identity, computes the effective verb set, and refuses to apply resources whose required verbs are not in an allowlist (for example, deny `create clusterrolebindings` even when the SA holds that verb).

- **Analogy:** `setuid` dropping capabilities; sandbox profiles.
- **Who decides:** The controller, via a static policy.
- **Blocks the gadget?** Yes for the RBAC payload. Still permits anything in the allowlist, which must be curated.
- **Failure mode:** Allowlist drift. Hard to reason about which kinds are safely in the allowlist when CRDs are involved.

## Tension Map

| Axis | A. Sudoers | B. SAR on creator | C. Capabilities | D. Quarantine | E. Labels | F. Creator-apply | G. No-RBAC | H. Scoped |
|------|-----------|-------------------|-----------------|---------------|-----------|------------------|-----------|-----------|
| Blocks cluster-admin SA trick | ✔ if SA not annotated | ✔ if creator lacks verb | ✔ always | △ RBAC only | △ admin discipline | ✔ | △ RBAC only | ✔ |
| Works with CI bot creating releases | ✔ | ✘ | ✔ | ✔ | ✔ | ✘ | ✔ | ✔ |
| Module-author friction | low | low | **medium** (declare caps) | high | low | low | medium (RBAC split out) | low |
| Platform-admin burden | medium (annotate SAs) | low | low | **high** (approve each) | medium (label discipline) | low | medium (RBAC elsewhere) | medium (curate allowlist) |
| Audit legibility | good (opt-in visible) | **excellent** (SAR trail) | excellent (cap diff) | excellent | poor | excellent | good | good |
| Implementation cost | S | S | **L** | M | S | M | S | M |
| Defense in depth | layers with B/H | layers with A | self-contained | layers with G | layers with A | self-contained | layers with H | layers with G |

Legend: ✔ blocks, ✘ does not block, △ partial or conditional. S/M/L: small/medium/large implementation cost.

## Three Philosophical Stances

The eight options reduce to three stances on *where identity is controlled*.

### Stance 1 — Identity is the tenant's problem

OPM impersonates whatever the release says. The escalation boundary is enforced by Kubernetes-native mechanisms: admission policy, `SubjectAccessReview`, labels. The controller stays thin.

- Options: B, E, H.
- Fits: environments where RBAC is already the primary control plane and operators want OPM to be a thin execution layer.

### Stance 2 — Identity is the module's contract

The module declares what it needs. The platform derives a least-privilege SA per release. The module author cannot escape declared capabilities; the platform can refuse modules whose declarations are too broad.

- Options: C.
- Fits: environments where modules are treated as first-class artifacts with signed contracts.
- This is the direction Crossplane's `rbac-manager` leans, but it is more open-ended: rbac-manager aggregates capabilities without constraining their content.

### Stance 3 — Privilege moves are a separate resource

Applying workloads and granting privileges are different operations with different approvers. RBAC payloads are quarantined or rejected outright. Workloads apply through a constrained identity.

- Options: A, D, G.
- Fits: regulated environments where every privilege change should have an auditable approval event.

## Comparison to Crossplane's `rbac-manager`

Crossplane's design document (https://github.com/crossplane/crossplane/blob/main/design/design-doc-rbac-manager.md) proposes a sibling controller that:

1. Aggregates capabilities declared by installed Providers.
2. Creates per-Provider ServiceAccounts and `ClusterRoleBinding`s.
3. Grants aggregated roles to built-in `cluster-admin`/`admin`/`edit`/`view` so kubectl users inherit the capabilities.

This addresses **operator UX** (no manual RBAC per Provider) and **auditability** (capabilities are declared in-tree). It does **not** solve the escalation gadget described here. A Provider that declares `*/*` gets `*/*`. The rbac-manager itself holds `bind` on ClusterRoles, which is a privilege-equivalent operation, so a compromised rbac-manager can fabricate arbitrary bindings.

Options A, B, D, and F in this document target escalation directly. Option C targets the same UX problem Crossplane targets, but with a tighter contract: the module's declared capabilities are a ceiling the controller enforces, not a floor that can grow.

## Open Questions

1. **Who creates ModuleReleases?** Tenants directly, or a pipeline/CI system on tenants' behalf? Option F is incompatible with the CI-bot pattern; Options A and B tolerate it only if the bot is treated as the tenant.
2. **What is the ServiceAccount's owner in the mental model?** The tenant (namespace SA), the module author (ships an SA manifest), or the platform (provisioned centrally)? Each answer shifts which option is natural.
3. **Is the PoC bound by `scope-and-non-goals.md` §9?** If yes, the immediate answer may be a simple narrowing (for example, require `spec.serviceAccountName` and document the escalation surface explicitly) with full design deferred. If no, Option C is the long-term fit.
4. **Does a legitimate module ever need to install RBAC?** If yes (operator-in-a-module), Option G needs an escape hatch and the escape hatch becomes the new escalation surface. If no, Option G is cheap and strong.
5. **How is `cluster-admin` prevented as a target?** Every option above can be bypassed by an admin who labels, annotates, or approves the wrong SA. Mechanism without policy does not prevent human error. At minimum, a deny-list for known-privileged roles (`cluster-admin`, `system:*`) is worth adding regardless of which option is chosen.

## Interim Recommendation

None. This document intentionally does not pick a winner.

If an interim narrowing is wanted while the full design is deferred, the cheapest useful step is:

1. Require `spec.serviceAccountName` (remove the controller-identity fallback entirely).
2. Add a static deny-list that refuses impersonation of SAs bound to `cluster-admin` or any `system:*` role, discovered via `SubjectAccessReview` against a canary verb.
3. Document the escalation surface in operator docs so admins understand that creating a privileged SA in a tenant namespace is equivalent to granting that privilege to anyone who can create a release in that namespace.

This is not a solution. It is a smaller foot-gun while the real answer is designed.
