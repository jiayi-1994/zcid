---
date: 2026-04-24
topic: zcid-feature-gaps
focus: "基于当前功能，还有什么缺陷，有什么好补充完善的功能"
mode: repo-grounded
---

# Ideation: zcid Feature Gaps & Improvements

## Grounding Context

### Codebase Context
- Go/Gin backend + React 19/TypeScript frontend; Tekton (CI) + ArgoCD (CD) with mock fallback
- 17 `internal/` modules, 12 `pkg/` modules, 36 migrations, ~34 React TSX components
- Prometheus metrics already wired (`pkg/middleware/metrics.go`); slog w/ secret masking; AES-256-GCM variable encryption
- HMAC-SHA256 webhook signature verification exists (`pkg/gitprovider/webhook.go:36-54`); rate-limit middleware IS wired to webhook handlers (`cmd/server/main.go:441-444`)
- Gaps: no distributed tracing (no OTel), no SSO/OIDC/LDAP/SAML, no multi-tenancy (RBAC is role-only), no webhook notification retries, no deployment strategy abstraction beyond plain ArgoCD sync, no SBOM/image signing/cosign, no test-result ingestion, no remote build cache, no K8s Events surfaced on stuck TaskRuns
- `docs/solutions/` does not exist — no prior institutional learnings to honor

### External Context (CI/CD 2025-2026)
- **Table-stakes gaps vs competitors**: manual run w/ param overrides, build matrix, test-result storage + trending, remote/layer cache, step marketplace (Artifact Hub), parallel test splitting, human approval gate, PR commit-status writeback
- **Tekton-specific pains**: TaskRun stuck Pending with no UI signal; no cache primitives; no manual trigger button; verbose params with no schema validation
- **Emerging**: AI pipeline failure diagnosis, PR preview environments via ArgoCD ApplicationSet PR Generator, Argo Rollouts progressive delivery, feature-flag gates, per-run cost attribution
- **Supply-chain must-haves**: Tekton Chains (SLSA), Syft SBOM, cosign sign+verify, Kyverno/OPA policy gates
- **Observability**: OpenTelemetry cicd.* SemConv (v1.27.0 stable), Tekton Results for PipelineRun history, JUnit flaky detection, duration trends, cost attribution

Sources (selected): JetBrains State of CI/CD 2025, Tekton Chains docs, OpenTelemetry CI/CD SIG, ArgoCD ApplicationSet PR Generator, Artifact Hub, SLSA/Sigstore docs.

## Ranked Ideas

### 1. Per-Step Execution Twin Table
**Description:** Add `step_executions` table capturing per-Tekton-step: resolved inputs (params, env, secret-refs), output hashes, content-addressed cache key, artifact digests, duration, resource usage. Populate from existing Tekton watcher (`internal/ws/k8s_watcher.go` + `pkg/tekton/`). Data currently flows past and is discarded.
**Warrant:** `direct:` Tekton watcher path already observes step state transitions; MinIO already present for artifact side; 36-migration cadence makes table-adds cheap.
**Rationale:** One table unlocks remote build cache (input-hash hits), replay from step N, SLSA provenance (in-toto = inputs + outputs + signer), flaky-step detection (same inputs → different outputs), per-team/pipeline cost attribution, determinism check, waterfall view. Each dependent feature becomes 30-40% of its standalone cost once the twin exists.
**Downsides:** Storage growth (30-90 day retention needed). Must store secret refs not values (leak risk). Adds per-step-transition write load.
**Confidence:** 85%
**Complexity:** Medium
**Status:** Unexplored

### 2. Internal Event Bus atop Audit Service
**Description:** Promote `internal/audit/` from write-only log to typed CloudEvents pub/sub. Every state change (run.started, step.succeeded, deploy.promoted, variable.rotated) emitted as typed event. Postgres audit table = durable log; in-process fan-out drives subscribers first; outbox worker later for at-least-once external delivery. WebSocket hub and webhook-notify become subscribers.
**Warrant:** `direct:` audit service already has `service.go`, `repo.go`, `middleware.go` — is a causal stream in everything but name. AES-256-GCM available for signing sensitive payloads.
**Rationale:** Fixes notification-retry gap (subscriber with backoff policy, not ad-hoc code). Every future integration (Jira, PagerDuty, SIEM, status page) becomes `nats sub zcid.>` or an outbox row, not a PR. AI hooks (failure diagnosis, anomaly scoring) attach passively.
**Downsides:** Requires schema discipline (breaking event-type changes hurt). In-process fan-out OK v0; external bus (NATS/Kafka) adds ops burden. Event-ordering guarantees need explicit design.
**Confidence:** 80%
**Complexity:** Medium
**Status:** Unexplored

### 3. Unified Policy Engine (OPA/Rego or Cedar)
**Description:** Replace 4-tier RBAC (admin/owner/maintainer/member) with single PDP evaluating `(subject, action, resource, context)`. Built-in roles become Rego bundles. Every authz call = `pdp.Allow(ctx, req)`. Same engine powers: deploy-approval rules, variable/secret access scoping, supply-chain gates (image must be cosign-signed), webhook egress rules. Policies live in Postgres (auditable, diffable).
**Warrant:** `external:` Kubernetes, GitHub, AWS IAM (→Cedar) all converged here. `reasoned:` 4-tier RBAC fails day one of real customer ("QA can promote to staging after-hours if no migration touched") — that is a policy, not a role.
**Rationale:** Unblocks compliance procurement (SOC2/ISO27001 want auditable authz logic). Also collapses `if user.Role == "admin"` scattered across services into one inspectable decision point.
**Downsides:** Rego has learning curve. Migration from Casbin rules → Rego bundles is a project. PDP on hot path needs caching.
**Confidence:** 70%
**Complexity:** High
**Status:** Unexplored

### 4. Intent-First Pipeline Compilation
**Description:** Users declare intent in `zcid.yaml` at repo root (~15 lines: `build: go`, `test: true`, `deploy: staging`). Compiler reads repo (go.mod / package.json / Dockerfile / Chart.yaml), auto-detects toolchain, auto-wires secrets from `.env.example` keys, emits Tekton PipelineRun. Pipeline authoring UI becomes diff preview / override surface, not primary. Cache keys auto-derived from lockfile hash.
**Warrant:** `external:` Dagger, Earthly, Nixpacks, Railway, Vercel all independently converged on "infer pipeline from repo." `direct:` `pkg/tekton/serializer.go` + `pkg/tekton/buildchain.go` already do intent→Tekton translation ad-hoc; the inversion is making it the primary path instead of a helper.
**Rationale:** Collapses the onboarding dead zone (blank YAML canvas). Eliminates 3 pain points at once: manual param wiring, template picker friction, pipeline/deploy config split. Intent is backend-agnostic — pairs with any future Executor-driver refactor.
**Downsides:** Existing pipelines must keep working (migration path). Power users want raw YAML escape hatch. Intent schema design is high-stakes — early wrong choices rot hard.
**Confidence:** 75%
**Complexity:** High
**Status:** Unexplored

### 5. Failure-as-Investigation Artifact
**Description:** On PipelineRun failure, auto-generate structured Investigation: (a) failing step container state snapshot, (b) diff vs last green run (env, deps, image digest, git range), (c) LLM root-cause hypothesis from tail logs + diff, (d) "re-run with fix" button that regenerates step spec. Run card never shows bare red X — shows "Investigation: dep pin drift (87% confidence)".
**Warrant:** `external:` Honeycomb BubbleUp, Sentry AI triage, Replay.io, incident.io auto-RCA. `reasoned:` red X discards the most info-dense moment in CI; pairing with last-green baseline is nearly free and collapses "broke → why" cycle.
**Rationale:** CI failure triage is the biggest dev-hours sink in a CI platform. Structured investigation with one-click fix is a demo-friendly differentiator competitors with flat text logs can't match without rebuilding their log pipeline. Pairs with Idea #1 (Twin) which provides the "last green" baseline data.
**Downsides:** LLM dependency (API cost + latency + hallucination risk). Ideally depends on #1. Scope creep risk — start with "show diff vs last green", add LLM as v2.
**Confidence:** 72%
**Complexity:** Medium (without LLM) / High (full)
**Status:** Unexplored

### 6. Supply-Chain Invariants by Default
**Description:** Tekton Chains + Syft SBOM + cosign signing + cosign verify-on-deploy wired as platform invariants, not optional steps. Every PipelineRun emits signed SLSA v1.0 attestation to Rekor or OCI. Every ArgoCD deploy gates on signature verify (via policy engine #3). Unsigned images cannot ship. SBOM published alongside artifact in MinIO.
**Warrant:** `external:` Tekton Chains is co-located with Tekton and already targets zcid's stack; SLSA/SBOM/Sigstore are the 2025-2026 standard stack; SSDF, EU CRA, FedRAMP-lite all require it for regulated-industry adoption.
**Rationale:** Unlocks regulated-industry procurement (finance, federal, healthcare) currently locked out. Differentiator among self-hosted open-source options. Compounds with policy engine (#3): "image unsigned" becomes a Rego rule, not bespoke code.
**Downsides:** Key management (cosign keys / KMS) adds ops burden. Some users want escape hatch for dev builds. Chains controller is a separate K8s deployment to operate.
**Confidence:** 75%
**Complexity:** Medium
**Status:** Unexplored

### 7. CI Trust Fundamentals Bundle
**Description:** Five grounded, small gaps shipped together as one "fundamentals" release:
- (a) **Surface K8s events inline** on Pending TaskRun — watch `corev1.Events` filtered by involvedObject UID of Pod/TaskRun, stream via existing WebSocket hub as a distinct event-stream track in the log viewer.
- (b) **PR commit-status writeback** to GitHub/GitLab on run lifecycle (queued/running/succeeded/failed) — add `PostStatus(sha, state, targetURL)` to `pkg/gitprovider`, invoke on PipelineRun state transitions.
- (c) **JUnit XML ingestion** — parse `**/junit.xml` / `**/*test-results.xml` from workspace into a new `test_results` table (run_id, suite, case, status, duration_ms, failure_message); trend page flags tests with pass→fail→pass flip rate >10%.
- (d) **ArgoCD sync-error detail** inline on deployment page — surface `Application.status.operationState.message` + `status.conditions[]`; deep-link to ArgoCD UI as fallback.
- (e) **Notification delivery retries** with exponential backoff — new `notification_deliveries` table (attempt count, next_retry_at, last_error); background worker with 1m/5m/15m backoff; failed-delivery admin panel.

**Warrant:** `direct:` each has a specific grep-confirmed gap: `internal/git/webhook_handler.go` has no outbound status push; `internal/notification/` grep for `retry|backoff` returns 0 matches; no `test_results*` migration exists; no `corev1.Events` watched anywhere in `internal/`.
**Rationale:** Each item alone is too small for a dedicated survivor slot; together they close the most visible table-stakes gaps vs competitors. Every user feels at least one on day one. Collectively they signal "this platform cares about the basics."
**Downsides:** 5-in-1 bundle dilutes focus; PM may want to cherry-pick top 2 (recommended: PR commit-status + K8s events first — both highest-frequency impact).
**Confidence:** 90%
**Complexity:** Low-Medium each
**Status:** Unexplored

## Synergy / Dependency Notes

- **#1 (Twin)** is the highest-leverage substrate; unlocks ideas inside #5, #6, and many rejected "reproducible capsules / replay / cache" ideas. Recommend first.
- **#2 (Event bus)** absorbs notification retries (#7e) cleanly once shipped. Order: Twin → Bus → Policy.
- **#3 (Policy)** is the authz substrate; #6 supply-chain gates ride on it.
- **#7 (fundamentals)** is independent of substrates — can ship in parallel for immediate visible wins.
- **#4 (Intent compilation)** is the strategic product bet; biggest scope; do after substrates land.

Suggested sequence: **#7 (parallel, quick wins) → #1 Twin → #2 Bus → #3 Policy → #6 Supply-Chain (rides #3) → #5 Investigation (rides #1) → #4 Intent-first (capstone).**

## Rejection Summary

| # | Idea | Reason Rejected |
|---|------|-----------------|
| F1.7 | Commit-SHA-level webhook idempotency | Narrow bug fix, below meeting-test floor; file as issue instead |
| F2.2 | PR-is-environment-spec (mandatory) | Subsumed by #4 intent compilation; preview-env as opt-in flag is lower risk |
| F2.3 | Workload-declared credentials | Depends on #3 policy engine; premature |
| F2.4 | Resource limits derived from history | Subsumed by #1 Execution Twin once shipped |
| F2.5 | Assertion-driven observability | Requires #1 + #2; premature; refine later |
| F2.6 | Commit-time approval | Controversial; implement as policy preset under #3 |
| F2.7 | Auto-generated Tekton Tasks | Subsumed by #4 intent compilation |
| F3.1 | zcid-as-library + CRDs | Strategic pivot, not feature improvement; subject-boundary |
| F3.2 | Change Record as primary entity | Beautiful reframe but architectural rewrite; too expensive vs value |
| F3.4 | Backend-agnostic executor driver | Big refactor; defer until Tekton pain forces it |
| F3.5 | Pluggable CD Deployer | Big refactor; value deferred |
| F4.3 | OpenAPI-first + generated clients | DX win but tooling-heavy; ship alongside any later frontend rewrite |
| F4.4 | Typed RunContext object | Subsumed as implementation detail of #1 + #2 |
| F4.5 | Plugin SDK (WASM/subprocess) | Depends on #4 (intent) + #1 (twin); premature |
| F4.7 | GitOps-for-zcid self-deploy | Elegant but niche; defer |
| F4.8 | Universal React Flow canvas | Nice DX; ship opportunistically when 2nd graph surface needed |
| F5.1 | SQL Savepoints | Subsumed by #1 Twin + Bus replay |
| F5.2 | Input-tape replay (RTS netcode) | Subsumed by #1 + #6 determinism |
| F5.3 | DAW step bounce/freeze | Subsumed by #1 Twin cache |
| F5.4 | Horizontal gene transfer | Speculative fitness metric; cut |
| F5.5 | What-if preview (spreadsheet) | Needs #1 Twin for accurate prediction; refine later |
| F5.6 | Cherry-pick fix into run | Needs Savepoints which needs #1 |
| F5.7 | Black-box flight-data schema | Subsumed by #1 + OTel cicd.* semconv |
| F5.8 | Adaptive RK45 parallelism | Too ML-ish; premature |
| F6.1 | zcid-solo single binary | Strategic pivot; big architectural split, not improvement |
| F6.3 | 10k-tenant hard multi-tenancy | Mega-refactor; file as `tenant_id` migration roadmap instead |
| F6.4 | Pre-warmed executor pool | Good idea but Tekton-specific; refine after benchmark proves cold start dominates p50 |
| F6.6 | Cross-tenant CAS cache | Subsumed by #1 cache key + multi-tenancy work |
| F6.7 | No-UI CLI-first mode | Strategic pivot; out of scope for improvement ideation |

## Notes on Grounding Corrections

During generation, sub-agents found two items in the grounding summary that were overstated:
- **Rate-limit middleware IS wired** to webhook handlers (`cmd/server/main.go:441-444`). The "DoS risk on webhook" framing in initial scan was incorrect.
- **HMAC-SHA256 webhook signature verification DOES exist** (`pkg/gitprovider/webhook.go:36-54`). The "webhook spoofing possible" framing was incorrect.

Both corrections were applied — ideas around rate-limiting webhooks and adding signature verification were dropped from the candidate list.
