---
date: 2026-04-26
topic: zcid-service-vitals-signal-matrix
focus: "下一波功能完善：服务生命体征、真实环境/集成健康、深度服务目录"
mode: repo-grounded
---

# Ideation: zcid Service Vitals & Signal Matrix

## Grounding Context

### Prior ideation already covered

- `docs/ideation/2026-04-24-zcid-feature-gaps-ideation.md` already explored execution twin, event bus, policy engine, intent-first pipeline compilation, failure investigation, supply-chain invariants, and CI trust fundamentals.
- `docs/ideation/2026-04-25-zcid-identity-access-ideation.md` already explored bootstrap token, tamper-evident auth audit, per-run workload identity, teams/SCIM, policy-as-code, PAT/project tokens, and JIT access.
- `docs/ideation/2026-04-26-zcid-next-wave-ideation.md` already captured sprint-sized next-wave items such as pipeline analytics, step timeline, scoped tokens, Slack notifications, and parameterized templates.

This pass intentionally looks beyond those into a product substrate that makes zcid feel like an internal developer platform, not only a pipeline runner.

### Codebase context

- `internal/svcdef/model.go` defines `services` with only `name`, `description`, `repo_url`, and status. The frontend `web/src/pages/projects/services/ServiceListPage.tsx` mirrors this shallow model.
- `web/src/pages/projects/environments/EnvironmentListPage.tsx` derives environment health from the environment name (`staging` → syncing, otherwise healthy), so the current health UI is not signal-backed.
- `internal/admin/handler.go` returns integration status for `k8s` with `Detail: "TODO: integrate real K8s/Tekton health check"`.
- `web/src/pages/projects/deployments/DeploymentListPage.tsx` requires manual image input and an optional Pipeline Run ID, which makes deployment correlation a human copy/paste task.
- `internal/stepexec/model.go` now records rich per-step facts: image refs/digests, command args, public env, secret refs, resolved params, resources, Tekton results, output digests, trace ID, timings, and exit code.
- Access tokens already exist (`migrations/000022_create_access_tokens.up.sql`, `web/src/pages/access-tokens`), so automation and external integrations can be authenticated with scoped tokens.

### External context

- Backstage-style internal developer platforms use a service catalog, ownership metadata, golden paths, and scorecards to reduce cognitive load and make standards visible.
- GitLab-style environments and release evidence connect deployments, approvals, artifacts, and environment state into an auditable product surface.
- Buildkite/CircleCI-style analytics show that teams value trends, bottlenecks, flakiness, and cross-run comparisons rather than isolated run detail pages.
- Tekton Results, Tekton Chains, and Argo Rollouts show the native ecosystem direction: persistent delivery history, provenance, and progressive delivery decisions based on signals.

## Ranked Ideas

### 1. Service Vitals & Patient History

**Description:** Turn `services` from passive repo rows into a service health chart. Each service gets ownership metadata, linked pipelines, deployment targets, latest deployed versions, deployment frequency, failure rate, flaky/bottleneck steps, active integration warnings, and recent interventions such as rollbacks or manual deploys.

**Warrant:** `direct:` `internal/svcdef/model.go` stores only `name`, `description`, and `repo_url`; `ServiceListPage.tsx` displays the same shallow fields. `external:` Backstage software catalogs and scorecards use service ownership and maturity data as the core IDP surface.

**Rationale:** zcid already has enough delivery data to make services the primary product object. A service vitals page gives developers and platform owners one place to answer “is this service healthy, who owns it, what runs/deploys it, and what needs attention?” without jumping across pipeline, deployment, and environment pages.

**Downsides:** Requires careful aggregation to avoid slow pages. Ownership/team data is thin until org/team work lands, so v1 should support simple owner strings or user IDs without waiting for full SCIM.

**Confidence:** 88%

**Complexity:** Medium

**Status:** Unexplored

### 2. Universal Signal Matrix

**Description:** Add a generic signal ingestion and scoring layer that records health/status signals from deployments, ArgoCD, Tekton, registry checks, notification delivery, synthetic probes, and future observability integrations. Signals attach to first-class zcid entities (`service`, `environment`, `pipeline`, `deployment`, `integration`) and roll up into confidence-scored health instead of hand-authored or frontend-derived status.

**Warrant:** `direct:` `EnvironmentListPage.tsx` derives health from names, and `internal/admin/handler.go` still has a TODO for real K8s/Tekton health. `external:` modern CI/CD and IDP tools converge on health surfaces fed by multiple signals rather than single binary checks.

**Rationale:** This is the substrate that makes Service Vitals real. Once every tool can emit “green/yellow/red + reason + freshness,” zcid can show trustworthy health, degrade stale signals, power deployment gates, and surface integration failures before a pipeline hits them.

**Downsides:** Generic signal models can become vague if over-abstracted. The v1 must keep a narrow set of signal kinds and avoid trying to become a full observability platform.

**Confidence:** 84%

**Complexity:** Medium-High

**Status:** Unexplored

### 3. Context-Aware Deployment Picker

**Description:** Replace manual image and Pipeline Run ID input with a selectable release timeline that correlates successful pipeline runs, image digests, git commits, and prior deployments. Users choose “deploy this known-good release” rather than typing image strings.

**Warrant:** `direct:` `DeploymentListPage.tsx` currently requires `image` and optional `pipelineRunId` fields in a modal. `direct:` `step_executions` and `pipeline_runs.artifacts` already carry enough artifact/run data to begin correlation.

**Rationale:** This removes one of the most error-prone CI/CD handoffs. It also feeds Service Vitals because each deployment becomes a reliable edge in the service history graph.

**Downsides:** Needs artifact production consistency. For pipelines that do not publish image artifacts yet, v1 needs a fallback to manual input.

**Confidence:** 86%

**Complexity:** Medium

**Status:** Unexplored

### 4. Integration Reconciliation Ledger

**Description:** Treat external integrations like financial accounts: every outbound action and health probe records an expected state, observed state, status, error, and reconciliation timestamp. The integration dashboard shows whether Slack/webhook/Git/registry/K8s/ArgoCD are actually in balance with zcid’s expectations.

**Warrant:** `direct:` admin integration health includes a TODO for real checks, registry testing is mock-backed, and notification delivery currently lacks durable retry/confirmation. `reasoned:` silent integration failure is worse than explicit failure because users believe automation happened when it did not.

**Rationale:** Integration health becomes an auditable product surface instead of a static settings page. It is a natural signal source for the Signal Matrix and helps explain whether failures are user-code failures or platform/integration failures.

**Downsides:** True reconciliation differs per provider, so v1 should start with common probes and delivery records rather than pretending every integration supports exact state comparison.

**Confidence:** 78%

**Complexity:** Medium

**Status:** Unexplored

### 5. Pipeline Statistical Process Control

**Description:** Use `step_executions` history to detect pipeline drift: steps that are slower than their historical control limits, steps that pass but are becoming unstable, and pipelines whose p95 duration or failure rate is degrading. Surface these as warnings on service vitals and pipeline analytics.

**Warrant:** `direct:` `step_executions` records timing, status, image digest, params, resources, and exit code. `external:` Buildkite Test Analytics and CircleCI Insights show product demand for flakiness and performance analytics.

**Rationale:** This turns raw execution records into proactive maintenance signals. It supports the service chart (“this service is getting slower/flakier”) without waiting for a full AI failure diagnosis system.

**Downsides:** Needs enough history before scores are meaningful. v1 should show “insufficient data” rather than producing fake precision.

**Confidence:** 82%

**Complexity:** Medium

**Status:** Unexplored

### 6. Ephemeral Lease Broker

**Description:** Introduce TTL/lease semantics for preview environments, temporary deployments, and future self-service resources. A resource exists while a PR, branch, user request, or heartbeat justifies it; otherwise zcid warns and garbage-collects it.

**Warrant:** `external:` ArgoCD ApplicationSet PR generators and preview environments are common GitOps patterns; cost control requires lifecycle ownership. `direct:` zcid already has environment/deployment tables and CRD cleanup utilities, making leases a natural extension.

**Rationale:** Preview environments are valuable only if cleanup is automatic. A lease primitive also provides future cost and quota control without baking preview-specific cleanup into every feature.

**Downsides:** Requires careful deletion safety. The first version should be report-only or opt-in cleanup for non-prod resources.

**Confidence:** 74%

**Complexity:** Medium-High

**Status:** Unexplored

### 7. Deployment Air Traffic Control

**Description:** Treat environments as runways and deployments as flights. zcid queues deployments, checks environment “weather” from the Signal Matrix, shows holding patterns, and only clears deployment when health and concurrency rules allow it.

**Warrant:** `direct:` deployments already have status, sync status, health status, rollback, and resync flows; environment health is currently not real. `reasoned:` teams need to know “is staging/prod clear?” before pushing changes.

**Rationale:** This makes deployment safety visible and operationally intuitive. It combines environment health, deployment queues, progressive delivery, and approvals without forcing every user to understand Argo internals.

**Downsides:** Can overlap with full progressive delivery state machines. Keep v1 as visibility + queueing, not canary orchestration.

**Confidence:** 76%

**Complexity:** High

**Status:** Unexplored

## Recommended Direction

Start with **Service Vitals & Signal Matrix** as one cohesive feature slice:

1. Add real signal records and entity health rollups.
2. Extend services into richer catalog entries.
3. Build a Service Vitals page that aggregates pipeline, deployment, environment, and integration signals.
4. Use that page as the foundation for deployment picker, pipeline SPC, integration ledger, and air-traffic-control follow-ups.

This is the best next step because it is not a duplicate of the existing execution twin or analytics ideas; it creates a product-level surface that connects the infrastructure already landing in the repo.

## Rejection Summary

| # | Idea | Reason Rejected |
|---|------|-----------------|
| 1 | Sentient Service Personas | Memorable but too gimmicky; the useful core is Service Vitals without anthropomorphic messaging |
| 2 | Holographic Capability Injection | Too close to full feature-flag/product experimentation platform; outside near-term zcid identity |
| 3 | Autonomous Service Negotiation | High-upside but depends on service dependency graph, policies, and real health signals first |
| 4 | Cryptographic Auto-Approvals | Mostly repeats prior policy/supply-chain ideation; better as later use of release evidence |
| 5 | Zero-YAML Pipeline Telepathy | Already covered by intent-first pipeline compilation and parameterized templates |
| 6 | Smart Alert Routing & Batching | Valuable but belongs under notification ecosystem, not the selected service-vitals substrate |
| 7 | Blast Radius Topography | Strong visualization concept but premature before service graph and signal matrix exist |
