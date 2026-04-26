---
date: 2026-04-25
topic: identity-access-foundations
origin: docs/ideation/2026-04-25-zcid-identity-access-ideation.md
---

# Identity & Access Foundations

## Problem Frame

zcid currently has enough local auth and RBAC to run the product, but not enough identity/access control to be safe as a shared CI/CD platform. The riskiest gaps are not SSO or ReBAC yet; they are simpler foundation issues that make every later identity capability harder to trust:

- A static `admin` / `admin123` bootstrap credential is shipped in migrations and documented as the default login.
- Login, logout, failed authentication, refresh-token use, role changes, and token lifecycle events are not first-class audit events.
- Programmatic callers have no scoped token primitive, so automation either relies on a human password/JWT flow or out-of-band secrets.

This requirements document deliberately narrows the larger identity/access ideation into a first foundations release: remove static bootstrap credentials, make auth/security events auditable, and add scoped API tokens for humans and projects. Teams/SSO/SCIM, policy-as-code, JIT access, and per-run workload identity remain important follow-ups, but they depend on this substrate being trustworthy.

---

## Actors

- A1. **Platform administrator**: installs zcid, creates the first admin, manages users and project automation credentials, and needs a safe path that does not rely on a known shared password.
- A2. **Security/compliance reviewer**: inspects who authenticated, who changed access, and which automation token touched CI/CD resources.
- A3. **Automation maintainer**: triggers pipeline/deployment/API workflows from CI scripts, webhooks, or external systems without binding long-lived credentials to a human login session.
- A4. **Existing local user**: continues to log in with username/password and uses the product without SSO/MFA changes in this release.

---

## Key Flows

- F1. **First admin bootstrap**
  - **Trigger:** zcid starts with an empty `users` table after migrations.
  - **Actors:** A1
  - **Steps:**
    1. zcid detects that no human-created users exist. A legacy/static bootstrap seed from older migrations must not count as completing first-admin setup.
    2. zcid creates a one-time bootstrap token with a short expiry and stores only a verifier, not the plaintext token.
    3. zcid exposes the plaintext token once through a local/server-side bootstrap channel suitable for local dev and container/Helm installs.
    4. A1 redeems the token to create the first admin user with a new password.
    5. The token is invalidated after successful redemption or expiry.
  - **Outcome:** every install receives a unique first-admin path; no known shared admin password exists in migrations or docs.
  - **Covered by:** R1, R2, R3, R4, R14

- F2. **Authentication event capture**
  - **Trigger:** a user logs in, fails login, refreshes a token, logs out, changes password/status/role, redeems bootstrap, or creates/revokes a programmatic token.
  - **Actors:** A1, A2, A4
  - **Steps:**
    1. The auth or admin code emits a typed audit event with actor, action, result, IP, and structured detail.
    2. Both successes and failures are recorded, including failure reason categories that are safe to store.
    3. The audit list API/UI can show these events alongside existing mutation audit events.
  - **Outcome:** A2 can answer “who authenticated or changed access?” without reconstructing from application logs.
  - **Covered by:** R5, R6, R7, R8, R9

- F3. **Scoped API token issue and use**
  - **Trigger:** A1/A4 creates a personal access token, or A1 creates a project access token for automation.
  - **Actors:** A1, A2, A3, A4
  - **Steps:**
    1. Creator names the token, selects type, scopes, and expiry.
    2. zcid stores a one-way token hash and shows the raw token exactly once with a distinguishable prefix.
    3. An API request using the token is authenticated by token middleware, mapped to a principal, checked for scope, and logged with `last_used_at`.
    4. Expired, revoked, disabled-owner, or wrong-scope tokens are rejected and audited.
  - **Outcome:** automation can call zcid without using a human refresh token or static shared credential, and A2 can trace token use.
  - **Covered by:** R10, R11, R12, R13, R14

---

## Requirements

**Bootstrap**
- R1. Remove the static seeded `admin/admin123` credential from the default bootstrap path. New installs must not create a usable admin account with a known shared password, even if historical migrations remain present in the repository.
- R2. When no real first admin has been configured, zcid must support a one-time bootstrap-token redemption flow that creates the first admin account. The implementation must explicitly handle the current migration ordering where `000004_seed_admin_user.up.sql` would otherwise make `users` non-empty before server startup.
- R3. Bootstrap tokens must expire quickly by default, be single-use, and be stored server-side as a verifier/hash rather than plaintext.
- R4. The local development and deployment documentation must replace “login with admin/admin123” with the new first-admin bootstrap flow, including a non-interactive/admin-secret option for GitOps-style deployment if supported by implementation.

**Audit events**
- R5. Record typed auth/security audit events for bootstrap redemption, login success, login failure, refresh-token use/failure, logout, user create/update/disable, password change, role assignment, PAT/project-token create/use/revoke/expiry rejection.
- R6. Audit logging must record both successful and failed auth/security events. Failure detail must be categorical and safe to retain; never store plaintext passwords, refresh tokens, PAT values, or bootstrap tokens.
- R7. Audit detail must be structured enough to distinguish event type, principal type, token type, token id/name when safe, target user/project, result, and reason category.
- R8. Existing mutation audit behavior must continue. This release extends the audit surface rather than replacing it with a full event bus, SIEM streaming, or hash-chain design.
- R9. The audit log UI/API must let administrators identify auth/security events without reading raw database rows.

**Programmatic tokens**
- R10. Add personal access tokens (user-bound) and project access tokens (project-bound automation credentials) with explicit scopes and mandatory expiry.
- R11. Tokens must be displayed only once at creation, stored only as one-way hashes, revocable, and distinguishable by prefix from human JWT access tokens.
- R12. Token authentication must enforce token status, expiry, owner/project status where applicable, and requested scope before allowing access.
- R13. Token metadata must include name, type, scopes, expiry, created actor, revoked status, and last-used timestamp so administrators can rotate and investigate usage.
- R14. Programmatic token lifecycle and use must integrate with the auth/security audit event model from R5-R7.

---

## Acceptance Examples

- AE1. **Covers R1-R4.** Given a fresh database and completed migrations, when zcid starts, then there is no usable `admin/admin123` account and the historical static seed cannot suppress bootstrap-token generation. When A1 redeems the one-time bootstrap token before expiry with a username/password, then exactly one active admin user is created, the token cannot be reused, and a bootstrap audit event is recorded.
- AE2. **Covers R3.** Given a bootstrap token was generated but not redeemed before its expiry, when A1 attempts redemption after expiry, then the request is rejected, no admin user is created, and a failed bootstrap audit event records reason `expired` without storing the token value.
- AE3. **Covers R5-R7.** Given a known user tries to log in with a wrong password, when login fails, then the audit log records a failed login event with user/username context where safe, client IP, result `failure`, reason category `invalid_credentials`, and no password or token material.
- AE4. **Covers R5-R9.** Given an admin changes another user's role from `member` to `admin`, when the operation succeeds, then the audit log records the actor user, target user, old/new role values, result `success`, and the audit UI/API can show it as an auth/security event.
- AE5. **Covers R10-R14.** Given A4 creates a personal access token with scope `pipelines:trigger` and 30-day expiry, when A3 uses it to trigger a pipeline, then the request is accepted, `last_used_at` updates, and an audit event links the request to the token id/name and user principal without exposing the token value.
- AE6. **Covers R10-R14.** Given a project access token has scope `runs:read` only, when it attempts a deployment-creating endpoint, then the request is rejected for insufficient scope and a failed token-use audit event is recorded.

---

## Success Criteria

- **Security:** a fresh zcid install no longer has a universal default admin credential, and no new token-like secret is stored plaintext server-side.
- **Forensics:** an administrator can answer “who logged in, who failed login, who changed access, and which token called the API?” from zcid audit data.
- **Automation usability:** CI scripts and integrations can use scoped, expiring zcid tokens without relying on a human refresh token or shared admin account.
- **Compatibility:** existing username/password login and existing project/admin routes continue to work for human users.

---

## Scope Boundaries

- **No SSO/OIDC/MFA/passkeys in this release.** These remain important, but first-admin bootstrap, audit, and token substrate should land before adding identity-provider complexity.
- **No SCIM/team/group model in this release.** Project access tokens are project-bound only; team-owned credentials are deferred until teams exist.
- **No full tamper-evident hash chain or external witness in v1.** Audit rows become more complete and structured now; hash chaining/SIEM streaming is follow-up hardening.
- **No policy-as-code or ReBAC migration.** Token scopes should be compatible with future policy work, but this release does not replace Casbin.
- **No per-pipeline-run OIDC workload identity.** Project tokens improve external automation; short-lived run identity remains a later, larger architecture effort.
- **No account lockout/password reset/password policy bundle unless needed as a small supporting detail.** The primary value is bootstrap, audit, and programmatic tokens.

---

## Key Decisions

- **Ship foundations as one release.** Bootstrap, auth-event audit, and API tokens reinforce each other: bootstrap creates the first sensitive auth event; tokens need audit to be governable; audit needs token/bootstrap event coverage to be useful.
- **Defer SSO/SCIM/JIT/OIDC.** They are strategically strong, but each adds external integration and product-surface complexity. The current repo lacks the safer primitives those features should build on.
- **Use explicit token taxonomy but keep v1 narrow.** Personal and project access tokens cover the two highest-frequency automation cases without prematurely adding group/deploy/job tokens.
- **Extend audit before making it tamper-evident.** Completeness is the first failure: today auth events are absent. Hash chains and external witnesses are valuable after the event vocabulary and write paths exist.
- **Keep local password login working.** This release hardens local auth; it does not replace it with IdP-first login.

---

## Dependencies / Assumptions

- Current local auth is in `internal/auth/service.go`, `internal/auth/handler.go`, and `internal/auth/repo.go`; JWT access tokens last 30 minutes and refresh tokens last 7 days.
- Current refresh-token storage is Redis key `auth:refresh:<userID>`, one token per user.
- Current RBAC roles are `admin`, `project_admin`, and `member` in `internal/auth/model.go` and `internal/rbac/enforcer.go`; README currently claims a different role set.
- Current audit rows are defined by `migrations/000017_create_audit_logs.up.sql` and `internal/audit/model.go`; middleware only records successful POST/PUT/PATCH/DELETE requests.
- Current project membership middleware reads `ContextKeyUserID`; project access tokens will need an explicit project-principal path or they will fail membership checks by design.
- The frontend admin user surface is `web/src/pages/admin-users/AdminUsersPage.tsx`; there is no token or auth-event management surface today.
- External grounding: GitHub Enterprise supports audit log streaming and PAT lifetime policy defaults around 366 days; GitLab uses scoped personal/group/project tokens and short-lived CI job tokens; Kubernetes TokenRequest supports short-lived, audience-bound service-account tokens. These validate token expiry, audit streaming readiness, and later workload identity direction.

---

## Outstanding Questions

### Resolve Before Planning

*(empty — scope has been narrowed enough to plan.)*

### Deferred to Planning

- [Affects R1-R4][Technical] Exact compatibility strategy for `000004_seed_admin_user.up.sql`: edit migration for greenfield installs, add a neutralizing migration, mark `admin-bootstrap-001` unusable until redemption, or another explicit path.
- [Affects R2-R4][Technical] Exact bootstrap-token delivery channel for Docker, local dev, and Helm: stderr log, mounted file, admin-secret value, or a combination.
- [Affects R5-R9][Technical] Whether auth/security audit should be emitted through `internal/audit.Service` directly, a new typed wrapper, or a lightweight event helper.
- [Affects R10-R14][Technical] Exact scope list and endpoint-to-scope mapping for v1 tokens.
- [Affects R12][Technical] Middleware composition order for human JWT auth, project scope checks, and token auth.
- [Affects R12][Technical] How project-token principals satisfy project scope without pretending to be a human `user_id`.
- [Affects R9][UX] Whether token and auth-event UI ships as new admin pages or smaller additions to existing user/audit pages.

---

## Next Steps

-> `/ce-plan` for structured implementation planning.
