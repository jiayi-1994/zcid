---
date: 2026-04-26
topic: zcid-next-wave
focus: "下一波功能完善：可观测性、流水线智能化、多租户/组织层、通知生态、开发者门户"
mode: repo-grounded
---

# Ideation: zcid Next-Wave Feature Expansion

## Grounding Context

### What the prior two ideation docs covered
- **2026-04-24 (feature-gaps)**: Manual run triggers, build matrix, test-result storage, remote cache, step marketplace, human approval gates, PR commit-status writeback, supply-chain (SLSA/cosign/SBOM), OTel CI/CD SemConv, Tekton Results, progressive delivery (Argo Rollouts), AI failure diagnosis.
- **2026-04-25 (identity-access)**: SSO/OIDC/SAML/LDAP, MFA (TOTP/WebAuthn), PAT scoping, session management, password policy/lockout, RBAC role drift fix (4 roles in README vs 3 in code), audit log completeness (login/logout/role-change events missing).

### Current codebase state (Apr 2026)
- 22 migrations (latest: access_tokens `000022`, bootstrap_tokens `000020`, step_executions `000019`)
- `internal/stepexec` — brand new step-level execution tracking module
- `internal/logarchive` — MinIO-backed log storage
- `web/src/pages/access-tokens` — PAT management UI just landed
- `internal/svcdef` — service definition module (purpose unclear from structure alone)
- `internal/crdclean` — CRD cleanup utility (likely GC for orphaned Tekton CRDs)
- No migration `000021` — gap suggests a reverted/dropped schema change
- Swagger stubs are empty — no API documentation surface

### Themes NOT yet addressed by prior ideation
1. Observability & Insights (pipeline analytics, step metrics, failure patterns)
2. Pipeline Intelligence (parameterized templates, AI-assisted authoring)
3. Multi-tenancy / Org layer (teams/orgs above projects)
4. Notification ecosystem (native Slack/email/PagerDuty, retry, routing)
5. Developer Portal (service catalog, self-service environments, token scoping)

---

## Theme 1: Observability & Insights

### Problem
`stepexec` tracks step-level execution but there is no analytics surface. Users cannot answer: "Which step fails most often?", "How long does my pipeline trend over time?", "What is my success rate this week?" The Prometheus metrics middleware exists but is not exposed in the UI.

### Ideas (ranked by impact × feasibility)

#### 1.1 Pipeline Analytics Dashboard ⭐⭐⭐⭐⭐
**What**: A per-project analytics page showing: success/failure rate over time (7d/30d), median and p95 run duration trend, top-failing steps, most-triggered pipelines.
**Why now**: `stepexec` table already captures step start/end/status. Aggregation queries are straightforward. No new infrastructure needed.
**Implementation sketch**:
- Backend: `GET /projects/:id/analytics?range=7d` — SQL aggregations over `pipeline_runs` + `step_executions`
- Frontend: New `analytics/` page under `projects/`, recharts or arco `Statistic` + line charts
- Metrics: success_rate, p50/p95 duration, failure_step_name frequency
**Effort**: M (3–5 days). **Risk**: Low.

#### 1.2 Step-Level Flame Graph / Timeline ⭐⭐⭐⭐
**What**: On the PipelineRunDetailPage, show a horizontal Gantt/flame chart of step durations within a run — which steps ran in parallel, which were the bottleneck.
**Why now**: `step_executions` has `started_at` / `finished_at`. The visual editor already uses `@xyflow/react` — a timeline view is a natural companion.
**Implementation sketch**:
- Reuse `stepexec` data already fetched for the run detail page
- Render a CSS-grid or SVG timeline below the DAG view
- Color-code by status (green/red/yellow)
**Effort**: S–M (2–3 days). **Risk**: Low.

#### 1.3 Flaky Step Detection ⭐⭐⭐
**What**: Flag steps that have a high retry/failure rate but eventually succeed — classic flakiness signal. Surface as a warning badge on the pipeline list.
**Why now**: With `stepexec` history accumulating, a simple SQL query (`failure_count / total_count > 0.3 AND success_eventually`) can detect flakiness without ML.
**Implementation sketch**:
- Scheduled job (or on-demand query) per pipeline: compute flakiness score per step
- Store in a `step_flakiness` materialized view or a lightweight cache key in Redis
- Frontend: amber badge on PipelineListPage + tooltip "Step 'unit-test' is flaky (38% failure rate)"
**Effort**: M (3–4 days). **Risk**: Low.

#### 1.4 OpenTelemetry Trace Export ⭐⭐⭐
**What**: Emit OTLP traces for pipeline runs (cicd.* SemConv v1.27.0) so users can correlate CI failures with their own Jaeger/Grafana Tempo/Honeycomb.
**Why now**: The prior ideation doc identified this gap. `stepexec` now provides the span boundaries needed.
**Implementation sketch**:
- Add `pkg/otel/` with `go.opentelemetry.io/otel` SDK
- On run start/step start/step end: emit spans with `cicd.pipeline.name`, `cicd.pipeline.run.id`, `cicd.task.name`, `cicd.task.run.status`
- Config: `otel.endpoint` in `config.yaml`, disabled by default
**Effort**: L (5–7 days). **Risk**: Medium (SDK integration, config surface).

---

## Theme 2: Pipeline Intelligence

### Problem
Pipeline templates exist (Go/Java/Node/Docker/JAR) but are static blobs. Users cannot parameterize them, share custom templates, or get authoring assistance. The visual editor is powerful but has no guardrails or suggestions.

### Ideas

#### 2.1 Parameterized Pipeline Templates ⭐⭐⭐⭐⭐
**What**: Templates declare typed input parameters (string, secret-ref, enum) that users fill in when instantiating. Example: the Go template asks for `GO_VERSION` (enum: 1.22/1.23/1.25) and `TEST_FLAGS` (string).
**Why now**: The `pipelines` table already stores JSON config. Adding a `template_params` JSON column and a param-fill UI is incremental.
**Implementation sketch**:
- Migration: add `template_params jsonb` to `pipelines` table
- Backend: `POST /projects/:id/pipelines/from-template` accepts `{template_id, params}`; validates param schema; substitutes into pipeline JSON
- Frontend: `TemplateSelectPage` gains a param-fill step (form generated from JSON Schema)
**Effort**: M (4–5 days). **Risk**: Low.

#### 2.2 Custom Template Publishing ⭐⭐⭐⭐
**What**: Any pipeline can be "published as template" — scoped to project, org (future), or global (admin-only). Templates appear in the template picker alongside built-ins.
**Why now**: Complements 2.1. Enables teams to codify their own best practices.
**Implementation sketch**:
- Migration: add `is_template bool`, `template_scope enum(project,global)`, `template_category text` to `pipelines`
- Backend: `POST /pipelines/:id/publish-template`, `GET /templates?scope=global`
- Frontend: "Publish as Template" button in PipelineEditorPage toolbar; template picker shows custom + built-in tabs
**Effort**: M (3–4 days). **Risk**: Low.

#### 2.3 AI-Assisted Step Suggestion ⭐⭐⭐
**What**: When a user adds a stage, offer AI-generated step suggestions based on the project's detected language/framework (from git repo file scan or user-declared service type).
**Why now**: `svcdef` module likely holds service type metadata. Git integration already connects to repos. LLM call can be optional/configurable.
**Implementation sketch**:
- Backend: `POST /projects/:id/pipelines/suggest-steps` — reads `svcdef` + git file tree (package.json? go.mod? pom.xml?) → calls configured LLM API → returns ranked step suggestions
- Frontend: "✨ Suggest steps" button in stage panel; renders suggestions as one-click insertions
- Config: `ai.provider` + `ai.api_key` in config.yaml; feature disabled if not configured
**Effort**: L (5–7 days). **Risk**: Medium (LLM API dependency, prompt quality).

#### 2.4 Pipeline Lint / Validation Rules ⭐⭐⭐
**What**: Before saving or running, validate the pipeline DAG for common mistakes: empty stages, missing git-clone step, duplicate step names, secret references to non-existent variables, circular dependencies.
**Why now**: The Tekton translator already does some validation implicitly. Making it explicit and surfacing errors in the editor prevents confusing runtime failures.
**Implementation sketch**:
- Backend: `POST /projects/:id/pipelines/:pid/validate` — runs lint rules, returns structured errors with step/stage references
- Frontend: "Validate" button in editor toolbar; inline error markers on offending nodes (red border + tooltip)
**Effort**: S–M (2–3 days). **Risk**: Low.

---

## Theme 3: Multi-Tenancy / Organization Layer

### Problem
RBAC is per-project. There is no concept of a team or organization above projects. Admins must manage every project individually. Cross-project visibility (e.g., "show me all failing pipelines across my team's projects") is impossible.

### Ideas

#### 3.1 Organization / Team Entity ⭐⭐⭐⭐⭐
**What**: Introduce an `organizations` table. Projects belong to an org. Users are org members with org-level roles (org_admin, org_member). Org admins can manage all projects within the org.
**Why now**: This is the foundational multi-tenancy primitive. Without it, every other cross-project feature is blocked.
**Implementation sketch**:
- Migration: `organizations` (id, name, slug, created_at), `org_members` (org_id, user_id, role), add `org_id` FK to `projects`
- Backend: `internal/org/` module with CRUD + membership management
- RBAC: extend Casbin policy to `(sub, org, proj, obj, act)` — org_admin implicitly has project_admin on all org projects
- Frontend: org switcher in sidebar header (like GitHub's org dropdown); org settings page
**Effort**: XL (10–14 days). **Risk**: High (schema migration, RBAC refactor, UI restructure). **Prerequisite for 3.2–3.4.**

#### 3.2 Cross-Project Dashboard ⭐⭐⭐⭐
**What**: An org-level dashboard showing: all projects' pipeline health at a glance, recent failures across the org, active runs count, deployment status summary.
**Why now**: Directly addresses the "I manage 10 projects, I need one view" pain. Depends on 3.1.
**Implementation sketch**:
- Backend: `GET /orgs/:id/dashboard` — aggregates across all org projects
- Frontend: new top-level `OrgDashboardPage` replacing or augmenting the current single-user dashboard
**Effort**: M (3–4 days after 3.1). **Risk**: Low.

#### 3.3 Project Transfer & Forking ⭐⭐⭐
**What**: Transfer a project between orgs/owners. Fork a project (copy pipeline definitions) as a starting point for a new project.
**Why**: Common workflow when teams reorganize or when bootstrapping similar services.
**Effort**: M. **Risk**: Medium (data integrity, permission checks).

#### 3.4 Usage Quotas & Limits ⭐⭐⭐
**What**: Per-org limits on concurrent pipeline runs, artifact storage (MinIO bytes), and number of projects. Admins can set limits; usage is tracked and surfaced in org settings.
**Why**: Essential for shared/SaaS deployments. Prevents one team from starving others.
**Effort**: M–L. **Risk**: Medium.

---

## Theme 4: Notification Ecosystem

### Problem
Notification rules exist (webhook-only) but have no retry logic, no native integrations, no routing rules, and no per-event granularity beyond build/deploy events. The audit log doesn't capture notification delivery failures.

### Ideas

#### 4.1 Native Slack Integration ⭐⭐⭐⭐⭐
**What**: First-class Slack notification channel: configure a Slack Bot Token + channel, get rich message cards (pipeline name, status, duration, commit SHA, link to run) on build events.
**Why now**: Slack is the dominant team communication tool. Webhook-only forces users to build their own Slack adapter. Slack Block Kit enables rich, actionable messages.
**Implementation sketch**:
- Backend: `internal/notification/slack.go` — Slack Web API client (`chat.postMessage`), Block Kit message builder
- Config: notification rule gains `type: slack` with `bot_token` (stored as encrypted variable) + `channel`
- Frontend: notification rule form gains Slack channel type with token + channel fields + "Test" button
**Effort**: M (3–4 days). **Risk**: Low.

#### 4.2 Email Notifications ⭐⭐⭐⭐
**What**: Send email on pipeline failure/success to configured recipients. Support SMTP config in system settings.
**Why now**: Email is the lowest-common-denominator notification channel. Many teams still rely on it for on-call alerting.
**Implementation sketch**:
- Backend: `pkg/mailer/` with SMTP client; notification rule gains `type: email` with `recipients []string`
- System settings: SMTP host/port/user/password/TLS config
- Frontend: email type in notification rule form
**Effort**: S–M (2–3 days). **Risk**: Low.

#### 4.3 Notification Retry & Delivery Tracking ⭐⭐⭐⭐
**What**: Webhook/Slack/email deliveries are retried with exponential backoff (3 attempts). Delivery status (delivered/failed/pending) is stored and visible in the notification rule detail page.
**Why now**: Currently, a failed webhook silently disappears. Users have no way to know their notifications aren't working.
**Implementation sketch**:
- Migration: `notification_deliveries` table (rule_id, event_type, payload, status, attempts, last_error, delivered_at)
- Backend: async delivery worker (goroutine pool or Redis queue); retry on 5xx/timeout
- Frontend: "Delivery History" tab on notification rule detail
**Effort**: M–L (4–6 days). **Risk**: Medium (async worker reliability).

#### 4.4 Notification Routing Rules ⭐⭐⭐
**What**: Route notifications based on conditions: "only notify on failure", "only notify if duration > 10min", "only notify for branch=main". Currently rules are all-or-nothing.
**Why**: Reduces alert fatigue. Teams want failure-only alerts on main, but all-events on feature branches.
**Implementation sketch**:
- Extend notification rule JSON schema with `conditions: [{field: "status", op: "eq", value: "failed"}, {field: "branch", op: "matches", value: "main"}]`
- Backend: evaluate conditions before dispatching
- Frontend: condition builder UI in notification rule form
**Effort**: M (3–4 days). **Risk**: Low.

#### 4.5 PagerDuty / OpsGenie Integration ⭐⭐⭐
**What**: On-call alerting integration for critical pipeline failures. Trigger PagerDuty incidents directly from zcid.
**Why**: Ops teams need CI/CD failures to page on-call engineers, not just send Slack messages.
**Effort**: M (3–4 days). **Risk**: Low (well-documented APIs).

---

## Theme 5: Developer Portal

### Problem
`svcdef` (service definitions) exists as a module but its purpose is unclear from structure alone. There is no self-service environment provisioning, no service catalog, and the new access tokens feature lacks scoping granularity. The platform is CI/CD-centric but lacks the "inner loop" developer experience features.

### Ideas

#### 5.1 Service Catalog (svcdef Surface) ⭐⭐⭐⭐
**What**: A browsable catalog of service definitions — each service has: name, type (API/worker/frontend), language, owner team, linked pipelines, deployment targets, current health status. Developers can discover services and navigate to their CI/CD config.
**Why now**: `svcdef` module already exists. This is about surfacing it with a proper UI.
**Implementation sketch**:
- Backend: `GET /projects/:id/services` already exists (services page in frontend). Extend with `GET /services?org=:id` for cross-project catalog
- Frontend: new `ServiceCatalogPage` at org level — card grid with service type icons, owner, health badge, "View Pipelines" link
**Effort**: M (3–5 days). **Risk**: Low (depends on 3.1 for org scope).

#### 5.2 Scoped Access Tokens ⭐⭐⭐⭐
**What**: Access tokens (just landed in `000022`) should support fine-grained scopes: `pipelines:read`, `pipelines:trigger`, `deployments:read`, `variables:read` — not just a blanket API key.
**Why now**: The token table is brand new — the right time to add scopes before the API surface hardens.
**Implementation sketch**:
- Migration: add `scopes text[]` to `access_tokens` table
- Backend: token validation middleware checks required scope per endpoint
- Frontend: token creation form gains scope checkboxes (grouped by resource type)
**Effort**: S–M (2–3 days). **Risk**: Low.

#### 5.3 Self-Service Environment Provisioning ⭐⭐⭐
**What**: Developers can request a new environment (dev/staging/preview) from the UI. The request triggers a pipeline that provisions the environment (via ArgoCD ApplicationSet or Helm). Environments have TTLs and auto-cleanup.
**Why**: Reduces dependency on ops/platform teams for environment setup. Pairs with ArgoCD PR Generator (identified in prior ideation).
**Implementation sketch**:
- Backend: `POST /projects/:id/environments/provision` — triggers a designated "provision" pipeline with env params
- Frontend: "Request Environment" button on environments page; shows provisioning status
- TTL: `expires_at` field on environments; cleanup job deletes expired envs
**Effort**: L–XL (7–10 days). **Risk**: High (K8s dependency, ArgoCD ApplicationSet setup).

#### 5.4 API Documentation Surface ⭐⭐⭐
**What**: Auto-generate and serve OpenAPI docs from the Gin routes. The current `swagger.json`/`swagger.yaml` are empty stubs. Expose a Swagger UI at `/docs` (dev mode only).
**Why now**: The empty swagger stubs are a clear gap. External integrators (and the access token feature) need API docs.
**Implementation sketch**:
- Add `swaggo/swag` annotations to Gin handlers; `make swagger` generates `docs/swagger.json`
- Serve Swagger UI via `github.com/swaggo/gin-swagger` at `/api/docs` (behind admin auth or dev-only flag)
**Effort**: M (3–5 days, mostly annotation work). **Risk**: Low.

#### 5.5 CLI / SDK for Pipeline Triggering ⭐⭐⭐
**What**: A `zcid` CLI tool (or Go SDK) that lets developers trigger pipelines, check run status, and stream logs from their terminal. Uses the access token for auth.
**Why**: Developers want to trigger CI from scripts, local dev hooks, or other automation without using the UI.
**Implementation sketch**:
- `cmd/zcid-cli/` — cobra CLI with `pipeline run`, `pipeline status`, `pipeline logs` subcommands
- Uses `ZCID_TOKEN` env var (access token) + `ZCID_URL`
- Streams WebSocket logs to stdout
**Effort**: M–L (4–6 days). **Risk**: Low.

---

## Priority Matrix

| ID | Idea | Impact | Effort | Priority |
|----|------|--------|--------|----------|
| 1.1 | Pipeline Analytics Dashboard | ⭐⭐⭐⭐⭐ | M | **P0** |
| 1.2 | Step-Level Timeline | ⭐⭐⭐⭐ | S | **P0** |
| 2.1 | Parameterized Templates | ⭐⭐⭐⭐⭐ | M | **P0** |
| 4.1 | Native Slack Integration | ⭐⭐⭐⭐⭐ | M | **P0** |
| 5.2 | Scoped Access Tokens | ⭐⭐⭐⭐ | S | **P0** |
| 5.4 | API Documentation | ⭐⭐⭐ | M | **P1** |
| 2.2 | Custom Template Publishing | ⭐⭐⭐⭐ | M | **P1** |
| 4.2 | Email Notifications | ⭐⭐⭐⭐ | S | **P1** |
| 4.3 | Notification Retry & Tracking | ⭐⭐⭐⭐ | M | **P1** |
| 1.3 | Flaky Step Detection | ⭐⭐⭐ | M | **P1** |
| 2.4 | Pipeline Lint / Validation | ⭐⭐⭐ | S | **P1** |
| 5.1 | Service Catalog | ⭐⭐⭐⭐ | M | **P2** (needs org) |
| 3.1 | Organization / Team Entity | ⭐⭐⭐⭐⭐ | XL | **P2** (foundational) |
| 3.2 | Cross-Project Dashboard | ⭐⭐⭐⭐ | M | **P2** (needs org) |
| 4.4 | Notification Routing Rules | ⭐⭐⭐ | M | **P2** |
| 2.3 | AI Step Suggestion | ⭐⭐⭐ | L | **P2** |
| 5.5 | CLI / SDK | ⭐⭐⭐ | L | **P2** |
| 1.4 | OTel Trace Export | ⭐⭐⭐ | L | **P3** |
| 4.5 | PagerDuty Integration | ⭐⭐⭐ | M | **P3** |
| 5.3 | Self-Service Env Provisioning | ⭐⭐⭐ | XL | **P3** |
| 3.3 | Project Transfer & Forking | ⭐⭐ | M | **P3** |
| 3.4 | Usage Quotas & Limits | ⭐⭐⭐ | M | **P3** |

---

## Suggested Sprint 1 (Next 2 Weeks)

Focus on high-impact, low-risk items that build on already-landed infrastructure (`stepexec`, `access_tokens`):

1. **Step-Level Timeline** (1.2) — 2 days, pure frontend, immediate UX win
2. **Scoped Access Tokens** (5.2) — 2 days, schema + middleware, hardens new feature
3. **Pipeline Analytics Dashboard** (1.1) — 4 days, SQL aggregations + new page
4. **Native Slack Integration** (4.1) — 3 days, new notification channel
5. **Parameterized Templates** (2.1) — 4 days, unlocks template ecosystem

Total: ~15 dev-days. All items are self-contained with no org-layer dependency.

---

## Open Questions

1. **`svcdef` module**: What does it currently store? Is it the right foundation for the service catalog, or does it need a redesign?
2. **Migration gap `000021`**: Was this intentionally dropped? Should it be documented to avoid confusion?
3. **`crdclean` module**: Is this a background GC job? Should it be configurable (TTL, dry-run mode)?
4. **Access token scopes**: Should scopes be project-scoped (token only works for project X) or global? The current schema doesn't show a `project_id` FK on `access_tokens`.
5. **Org layer timing**: Is multi-tenancy a near-term requirement (paying customers) or a future concern? This determines whether 3.1 should be P1 or P3.
