---
date: 2026-04-25
topic: zcid-identity-access
focus: 帮我继续完善这个zcid的整体项目功能 — 补充能力 (narrowed to identity & access: RBAC/SSO/MFA)
mode: repo-grounded
---

# Ideation: zcid Identity & Access Capability Gaps

## Grounding Context (Codebase)

### Auth surface today
- `internal/auth/service.go`: JWT HS256, 30-min access + 7-day refresh stored in Redis. Login/refresh/logout endpoints. bcrypt password hash only. No password policy, no reset, no lockout, no failed-attempt tracking.
- `internal/auth/handler.go`: `/login`, `/refresh`, `/logout`, admin `/admin/users` CRUD.
- `internal/auth/repo.go`: refresh token keyed per-user (single slot — second login silently evicts the first).

### RBAC
- `internal/rbac/enforcer.go`: Casbin gorm adapter; 4-tuple `(sub, proj, obj, act)`; regex action match + keyMatch resource. Role inheritance `admin > project_admin > member` (3 in code).
- README claims 4 roles (admin/owner/maintainer/member) — **doc/code drift**.
- Watcher: Redis pub/sub on `rbac:policy:update`, triggered when role changes.
- Middleware: `RequireCasbinRBAC` reads `X-Project-ID`; `RequireAdminRBAC` enforces `role=="admin"`.

### User management
- Admin-only CRUD via `/admin/users`. No self-signup, no password reset, no account lockout.
- `web/src/pages/admin-users/AdminUsersPage.tsx` — list, create, edit, status toggle. No MFA UI, no PAT UI, no session UI, no SSO UI.

### Audit
- `internal/audit/middleware.go:14-17` — only logs mutations (POST/PUT/PATCH/DELETE). Skips login/logout/failed-auth/role-change.
- Fields: UserID, Action, ResourceType/ID, Result, IP. No `detail` usage, 5XX-skipped.

### Migrations relevant to identity
- `000001_init_schema` — `users` table with `password_hash`, `role`, `status`.
- `000003_create_casbin_rule` — Casbin policy table.
- `000004_seed_admin_user` — **hardcoded `admin/admin123` bcrypt at line 8**.
- `000005_seed_admin_rbac` — admin role policies.

### External baseline (peers in 2026)
- SSO + SCIM 2.0 are baseline at Premium tier (GitLab/GitHub/Buildkite); SCIM auto-deprovision = SOC2 CC6.3 evidence.
- WebAuthn/passkeys are 2026 baseline; SMS OTP deprecated.
- Token taxonomy norm: PAT (user-bound), Project Access Token (project bot, no seat), Group Access Token, Deploy Token (registry-only), CI/CD Job Token (short-lived auto-issued). GitHub fine-grained PATs require ≤1y expiry.
- Workload identity: GitHub Actions OIDC tokens with claims (`sub/repo/ref/environment`) → cloud STS short-lived creds. SPIFFE/SPIRE for in-cluster SVIDs. Sigstore/Fulcio for keyless signing.
- ReBAC (Zanzibar): SpiceDB (Apache-2.0, ZedTokens, caveats for ABAC), OpenFGA (CNCF, conditions, contextual tuples), Permify.
- JIT/break-glass: Teleport Access Requests, Indent, Sym, ConductorOne, AWS IAM TEAM, GitLab protected env approvals.
- Audit/SIEM: GitHub Audit Log Streaming → Splunk/Datadog/S3/EventHubs in OCSF format. Tamper-evident hash-chained logs (Vault, AWS CloudTrail Lake).
- Step-up auth (sudo-mode 3h re-auth window): GitHub for sensitive ops.
- Pipeline secrets: HashiCorp Vault dynamic secrets, External Secrets Operator (CNCF), Tekton Chains for SLSA L2/L3.

### Pain points seeded into ideation
1. Role-model doc drift (README 4 roles ≠ code 3).
2. Hardcoded `admin/admin123` in migration is a critical bootstrap risk.
3. Auth-event audit blackout = SOC2 gap + zero forensic capability.
4. Pipeline runs use a single global K8s/Tekton/ArgoCD secret — no scoped/short-lived identity per run.
5. No team/group concept — N×M user-project rows that drift on every joiner/leaver.
6. Casbin's 4-tuple too coarse for env-level "deploy to prod" gating.
7. RBAC pubsub watcher is good infra, currently underused.
8. Multi-project ownership graph already in DB — natural fit for ReBAC migration later.

## Ranked Ideas

### 1. Bootstrap-Token Replaces Static `admin/admin123` Seed
**Description.** Drop `migrations/000004_seed_admin_user.up.sql`. On first boot when `users` is empty: generate a one-shot token, print to stderr and write `/var/run/zcid/bootstrap-token` mode 0600, force password + MFA enrollment on redemption, self-destruct on use or after 15 minutes. Helm chart accepts `--set bootstrap.adminPasswordSecret=...` for GitOps installs.
**Warrant:** `direct:` `migrations/000004_seed_admin_user.up.sql:8` ships bcrypt of literal `admin123` to every install. Anyone with read access to the migration history at any past point knows the bootstrap secret. Every prod install that did not manually rotate is currently compromised by inspection.
**Rationale.** Removes the highest-severity finding by deletion not patching. Closes SOC2 CC6.1 by default. Net-negative LOC. Pattern is well-established (kubeadm `--token-ttl`, Grafana `GF_SECURITY_ADMIN_PASSWORD`, Vault unseal, GitLab `gitlab-rake "gitlab:password:reset"`).
**Downsides.** First-run UX adds a copy-paste step; helm/docker-compose/migration-tool ownership of the ceremony needs a one-time decision.
**Confidence.** 95%
**Complexity.** Low
**Status.** Unexplored

### 2. Auth-Event-First Audit Log with Hash-Chained Tamper Evidence
**Description.** Extend `internal/audit` to record login success/failure, logout, refresh-token rotation, password changes, role-grant/revoke, MFA-enroll, and policy-publish events. Add `prev_hash` and `row_hash` columns (Vault-style) so any deletion or backdating in `audit_logs` is detectable. Revoke `UPDATE` from the app DB role on the audit table. Nightly job publishes a Merkle root to S3 Object Lock or Sigstore Rekor as an external witness. Add an OCSF-shaped streaming sink to Splunk / Datadog / Azure EventHubs.
**Warrant:** `direct:` `internal/audit/middleware.go:14-17` returns early for non-mutation methods; `internal/auth/service.go:62-95` (Login) never calls the audit service. Today the audit table cannot answer "who logged into the deploy-prod admin at 03:00."
**Rationale.** Tamper-evident is qualitatively different from tamper-resistant — an auditor verifies inclusion proofs without DB access. One mechanism closes the auth-event blackout, satisfies SOC2 CC6.1/CC7.2, and unlocks SIEM streaming. Foundational for survivors #5 and #7, both of which want a durable event substrate.
**Downsides.** Hash-chain repair semantics on out-of-order writes; retention join with Postgres VACUUM; one extra column on every audit row; one external-witness dependency.
**Confidence.** 90%
**Complexity.** Medium
**Status.** Unexplored

### 3. Per-Pipeline-Run OIDC Workload Identity
**Description.** Mint a 5–15 min JWT per `PipelineRun` with claims `{run_id, project_id, env, repo, ref, triggered_by, exp}` signed by a zcid-internal issuer. Tekton/ArgoCD verify via OIDC discovery and exchange the token for a short-lived K8s token via `TokenRequest` API and `ProjectedServiceAccountToken` trust binding. Cluster-side audit then attributes every kubectl/Tekton call to a specific zcid run.
**Warrant:** `direct:` `pkg/k8s/client.go:14-41` holds one global `Clients`, used by every caller in `internal/deployment` and `internal/pipelinerun` — zero per-run attribution at the cluster API audit layer. `external:` GitHub Actions OIDC, SPIFFE/SPIRE, Tekton Chains + Fulcio.
**Rationale.** Compromised-pipeline-secret blast radius drops from "everything zcid has ever deployed" to "≤15 min of one run on one project/env." Run-id becomes the natural subject for the survivor #2 audit log — deploy events are cryptographically linked to a specific run, not a static admin. Unlocks per-environment scoping (run cannot escape its declared namespace) without zcid having to enforce that itself.
**Downsides.** OIDC discovery endpoint to host; JWKS rotation; mock-mode path needs a no-op verifier; cross-cluster trust setup is per-tenant.
**Confidence.** 80%
**Complexity.** High
**Status.** Unexplored

### 4. Teams + IdP-Group-Driven Role Membership (SCIM-Ready)
**Description.** Add `teams`, `team_members(user_id, team_id)`, and `team_project_roles(team_id, project_id, role)` tables. Sync `team_members` from the OIDC `groups` claim on every login (JIT) and accept SCIM 2.0 push for lifecycle. Admin UI shifts to managing group→role mappings, not user→role. Resolves the README↔code role-count drift because roles become data, not enum-in-code.
**Warrant:** `direct:` `internal/rbac/enforcer.go:34` matcher already uses `g()` (transitive grouping policy) — no Casbin model change needed. Pain-point #5 (no group/team) is the single biggest onboarding-drag problem at >20 users.
**Rationale.** Collapses N×M user-project rows that drift on every joiner/leaver. Off-boarding becomes "fire from Okta" — automatic, evidenced, SOC2 CC6.3 by construction. Precondition for any later passkey-only or full ReBAC migration. Subsumes 80% of the access-request volume that survivor #7 would otherwise have to handle.
**Downsides.** Schema migration with backfill; team-vs-project cardinality decision (does a team own a project, or just have roles in it?); SCIM bearer-token issuance flow; UI rewrite for the admin-users page.
**Confidence.** 85%
**Complexity.** Medium
**Status.** Unexplored

### 5. Policy-as-Code (YAML in Git) + Decision Logger
**Description.** Replace migration-based Casbin seeds with `policies/*.yaml` reviewed in PRs. A reconciler at boot (and on PR merge) diffs declared vs DB rows and applies the delta. Add a Decision Logger interceptor in front of every authz check that records `(principal, action, resource, decision, reason, policy_version, latency)` for every check — allowed and denied. Denies feed friendly 403 UX, anomaly hooks, and shadow-mode comparison for any future ReBAC migration. CI runs a policy-test harness (`assert can("alice","deploy:prod") == false`).
**Warrant:** `direct:` today RBAC is mutated only via API/SQL with no review trail beyond migration history; missing `RequireCasbinRBAC` on a route is a silent privilege escalation. `external:` ArgoCD AppProject yaml, OPA/Rego, AWS Cedar.
**Rationale.** Permission changes become reviewed PRs with diff/blame/revert. CI test harness catches regressions. Decision Logger turns the authz layer from a black box into a tappable surface that debugging, UX, and security all consume. Makes a future ReBAC swap (SpiceDB/OpenFGA) shadow-comparable instead of a flag-day rewrite. Resolves the role-drift bug because the YAML is the doc.
**Downsides.** Reconciler edge cases (rows mutated outside YAML); decision-log volume on hot paths needs sampling or async write; commit-as-policy-change introduces a new merge-discipline expectation.
**Confidence.** 80%
**Complexity.** Medium
**Status.** Unexplored

### 6. Token Taxonomy — PATs + Project Access Tokens with Mandatory Expiry
**Description.** Add `personal_access_tokens(user_id, name, scopes, expires_at, last_used_at)` and `project_access_tokens(project_id, name, scopes, expires_at, last_used_at)`. Explicit scopes (`pipelines:trigger`, `variables:read`, `deployments:create`, `runs:read`). Mandatory ≤1-year expiry per token. Project tokens consume no user seat and survive creator offboarding. UI exposes rotation, last-used-at, scope inspection. Tokens use a distinct `Authorization: Bearer zcid_pat_...` prefix so middleware can route to the token validator instead of the human JWT path.
**Warrant:** `direct:` `internal/auth/service.go:62-95` only issues human-tied access/refresh JWT — every external caller (CI script, webhook poller, integration) currently must paste a personal JWT issued from a human password login. `external:` GitLab 5-token taxonomy, GitHub fine-grained PATs.
**Rationale.** Off-boarding-an-engineer SLA stops requiring a sweep across every CI system. Scope strings become the natural targets for fine-grained Casbin policies (and later ReBAC relations). Mandatory expiry is now an explicit SOC2 audit checklist item. Pairs with survivor #4 — project tokens issued *to teams* are the natural multi-engineer credential.
**Downsides.** Token-rotation UX; secret-display-once flow on issuance; rate-limit / abuse story for project-scoped tokens; documentation refresh.
**Confidence.** 90%
**Complexity.** Medium
**Status.** Unexplored

### 7. Just-In-Time Access Requests with TTL-Bound Grants
**Description.** Add `POST /access-requests` flow: a `member` requests temporary `project_admin` (or scoped capability) on project X for N hours with a written justification; one or more approvers confirm via Slack / email / PagerDuty; on approval, a Casbin row is inserted with `expires_at`; a sweeper job revokes at TTL. All four steps emit auth events into survivor #2's log so the elevation is a self-contained audit packet. UI shows pending requests on the approver's dashboard.
**Warrant:** `direct:` `internal/rbac/enforcer.go:69-100` Redis-pubsub watcher already provides the dynamic-reload primitive — no new infra needed for grants to take effect across replicas. `external:` Teleport Access Requests, Indent, Sym, ConductorOne, AWS IAM Identity Center TEAM.
**Rationale.** Converts standing privilege into ephemeral privilege — the alternative is "everyone is admin all the time because filing a ticket is friction." Auto-deprovisioning evidence for SOC2 CC6.3 falls out for free. Reuses survivor #2's audit log as the request/approve/grant/revoke ledger. Pairs with survivor #4 — teams can be configured with default approver chains, eliminating the per-request approver lookup.
**Downsides.** Approver-chain abstraction needs a one-time design decision (Slack first? PagerDuty integration? in-app only?); TTL-mid-deploy semantics (hard kill or grace period?); request UI surface; Slack/email integration adds outbound dependencies.
**Confidence.** 80%
**Complexity.** Medium
**Status.** Unexplored

## Cross-Cutting Structure

- **Event spine.** Survivors #2, #5, #7 share an event log — #2 first makes #5 (decision events) and #7 (request/approve/grant/revoke events) nearly free.
- **Implicit Polymorphic Principal.** #3 (run-actor), #4 (team-actor), #6 (token-actor) collectively force the actor sum-type without a separate refactor ticket. The `actors`/`principals` table can emerge from doing the work, not from a planning meeting.
- **Bootstrap unblocks the rest.** #1 is a precondition for #4 (clean IdP integration without a hardcoded local-admin fallback) and for #2 (the bootstrap-redeem event is the first row of the new audit log).
- **Sequencing suggestion.** #1 → #2 → #6 → #4 → #5 → #7 → #3. The first three are independent low-risk wins. #4 is the foundation for SCIM/SSO. #5 is the substrate for ReBAC-readiness. #7 lives on top of #2 and #4. #3 is the largest architectural lift and benefits from #2 being in place to receive run-attributed events.
- **Strategic future direction (cut as premature).** Full ReBAC migration to SpiceDB / OpenFGA / Permify — the right move once #4 (Teams) and #5 (Decision Logger) make the migration shadow-comparable, but multi-quarter and not a survivor today.

## Rejection Summary

| # | Idea | Reason Rejected |
|---|------|-----------------|
| 1 | ReBAC migration to SpiceDB / OpenFGA / Permify | High payoff but multi-quarter; survivors #4 (Teams) and #5 (Decision Logger) are the right precursors before this lift |
| 2 | Sessions UI + sudo-mode re-auth window | Narrower benefit standalone; folded into survivor #1 (bootstrap creates the first audited session) and #5 (sudo-mode is a Decision Logger annotation) |
| 3 | Polymorphic Principal abstraction as a separate ticket | Implicit outcome of survivors #3+#4+#6+#7 doing their work; calling it out separately adds no shippable capability |
| 4 | Postgres RLS at the data layer (RBAC in DB) | Defense-in-depth but doubles per-query complexity at current zcid scale; cost > value today |
| 5 | Passkey-only / no-password column | Presupposes IdP-first migration (survivor #4) is mature; too dependent to ship standalone |
| 6 | Solo mode (zero-RBAC tier for homelab) | Packaging story not capability; outside the "补充能力" framing |
| 7 | ISPS posture ratchet (NORMAL / HEIGHTENED / LOCKDOWN) | Reactive control; survivors #2+#7 cover pre-incident better; revisit once audit pipeline lands |
| 8 | WHO surgical-checklist read-back step-up on prod deploy | Strong novelty but presupposes env-level gating from cluster D — defer until ReBAC lands |
| 9 | Doc/code role-drift fix as standalone item | Subsumed by survivor #5 — once policy is YAML, the YAML is the doc |
| 10 | Kill JWT refresh, opaque-cookie sessions only | Simplification not capability; rides along survivor #1's auth refactor if pursued |
| 11 | Logout = `DELETE /sessions/<id>` (no special endpoint) | Minor cleanup, fails the meeting test |
| 12 | `zcid bootstrap` CLI as separate binary | Subsumed into survivor #1's first-boot flow |
| 13 | Sigstore Rekor specifically as the witness substrate | Pattern adopted in survivor #2's hash-chain + external witness — naming the specific witness is an implementation choice, not a separate idea |
| 14 | Event-sourced IAM (audit log as source of truth, DB as projection) | Architecturally elegant but rebuild risk; survivor #2 captures the practical 80% (events durable + tamper-evident) without the rebuild blast radius |
| 15 | "Zero-human" / every actor is a workload as a separate model | Subsumed by survivors #3+#6 forcing the abstraction implicitly |
| 16 | Typed Permission DSL compiling to Casbin + audit + TS | Survivor #5's YAML-to-Casbin reconciler covers the high-value half (review trail + tests); a full DSL with TS codegen is leverage at a later stage |
| 17 | Nuclear PAL / sealed-authenticator framing | Same idea as survivor #1 with a different analogy — kept the simpler bootstrap-token framing |
| 18 | Vertebrate immune MHC framing | Same idea as the rejected ReBAC migration with a different analogy |
| 19 | State Bar CLE periodic-attest auto-decay | Subset of survivor #7 (TTL grants); periodic re-attestation is the long-form variant |
| 20 | Identity Bootstrap CLI for IdP wiring + key rotation | Subsumed into survivor #1 (bootstrap) and survivor #4 (IdP wiring); admin-CLI exists in `cmd/migrate` already |
