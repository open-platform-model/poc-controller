## ADDED Requirements

### Requirement: Impersonated client for apply and prune
When `spec.serviceAccountName` is set, the controller MUST use an impersonated client for apply (Phase 5) and prune (Phase 6) operations.

#### Scenario: Apply with impersonation
- **GIVEN** a ModuleRelease with `spec.serviceAccountName=deploy-sa` in namespace `team-a`
- **WHEN** the controller runs Phase 5 (Apply)
- **THEN** SSA apply operations use the identity `system:serviceaccount:team-a:deploy-sa`
- **AND** the apply succeeds only if `deploy-sa` has sufficient RBAC permissions

#### Scenario: Prune with impersonation
- **GIVEN** a ModuleRelease with `spec.serviceAccountName=deploy-sa` and stale resources to prune
- **WHEN** the controller runs Phase 6 (Prune)
- **THEN** delete operations use the impersonated identity

### Requirement: Default behavior without serviceAccountName
When `spec.serviceAccountName` is empty, the controller MUST use its own client.

#### Scenario: No impersonation when SA not specified
- **GIVEN** a ModuleRelease with `spec.serviceAccountName` empty or unset
- **WHEN** the controller reconciles
- **THEN** all operations use the controller's own service account

### Requirement: Missing ServiceAccount stalls reconcile
If the specified ServiceAccount does not exist, the reconcile MUST fail with a stalled condition.

#### Scenario: ServiceAccount not found
- **GIVEN** a ModuleRelease with `spec.serviceAccountName=nonexistent-sa`
- **WHEN** the controller attempts to build an impersonated client
- **THEN** the reconcile is classified as `FailedStalled`
- **AND** `Ready=False` with reason indicating SA not found

### Requirement: Impersonation RBAC failure stalls reconcile
If the controller lacks impersonation permissions, the reconcile MUST fail with a stalled condition.

#### Scenario: Impersonation unauthorized
- **GIVEN** a ModuleRelease with `spec.serviceAccountName=deploy-sa` and the controller lacking `impersonate` RBAC
- **WHEN** the controller attempts an apply operation
- **THEN** the reconcile is classified as `FailedStalled`
- **AND** the error message indicates impersonation was denied

### Requirement: Same-namespace only
The ServiceAccount MUST be in the same namespace as the ModuleRelease.

#### Scenario: SA resolved in same namespace
- **GIVEN** a ModuleRelease in namespace `team-a` with `spec.serviceAccountName=deploy-sa`
- **WHEN** the controller builds the impersonated client
- **THEN** it impersonates `system:serviceaccount:team-a:deploy-sa` (same namespace)

### Requirement: Impersonation includes standard SA group set
The impersonated client MUST be configured with both `UserName` and `Groups` on the `rest.ImpersonationConfig`. The Kubernetes apiserver does not derive group membership from the impersonated `UserName`; it reads `Impersonate-Group` headers independently. Without explicit groups, the impersonated identity belongs to no groups, and any RBAC binding whose subject targets a group (`system:serviceaccounts`, `system:serviceaccounts:<namespace>`, or `system:authenticated`) silently fails — even though the same SA succeeds when authenticating with its own token.

The `Groups` slice MUST contain the standard set that the apiserver's `serviceaccount.TokenAuthenticator` would inject for an SA in the given namespace:

- `system:serviceaccounts`
- `system:serviceaccounts:<namespace>` (where `<namespace>` is the SA's namespace)
- `system:authenticated`

This matches the behavior of Flux's `runtime/client/impersonation` and gives impersonated identity parity with token-based identity for the same SA.

#### Scenario: Impersonation config carries standard groups
- **GIVEN** a ModuleRelease in namespace `team-a` with `spec.serviceAccountName=deploy-sa`
- **WHEN** the controller builds the impersonated client
- **THEN** the underlying `rest.ImpersonationConfig.Groups` is exactly `["system:serviceaccounts", "system:serviceaccounts:team-a", "system:authenticated"]`
- **AND** `rest.ImpersonationConfig.UserName` is `system:serviceaccount:team-a:deploy-sa`

#### Scenario: Group-subject RoleBinding authorizes apply
- **GIVEN** a ModuleRelease in namespace `team-a` with `spec.serviceAccountName=deploy-sa` and a RoleBinding in `team-a` whose subjects are `[{Kind: "Group", Name: "system:serviceaccounts:team-a"}]` granting permissions on the resources to be applied
- **WHEN** the controller runs Phase 5 (Apply)
- **THEN** the apply succeeds (the impersonated identity is recognized as a member of `system:serviceaccounts:team-a`)
- **AND** `Ready=True` is set on the ModuleRelease

#### Scenario: Authenticated-group binding authorizes read access
- **GIVEN** a ClusterRoleBinding granting `view` on a CRD to the group `system:authenticated`
- **WHEN** the controller's impersonated client lists instances of that CRD
- **THEN** the request is authorized (the impersonated identity is a member of `system:authenticated`)
