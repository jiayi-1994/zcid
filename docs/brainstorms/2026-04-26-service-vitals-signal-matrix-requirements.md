---
date: 2026-04-26
topic: service-vitals-signal-matrix
source-ideation: docs/ideation/2026-04-26-zcid-service-vitals-signal-matrix-ideation.md
status: ready-for-planning
---

# Requirements: Service Vitals & Signal Matrix

## Problem Frame

zcid already has the raw building blocks of a CI/CD platform: services, environments, deployments, pipelines, step executions, logs, audit, access tokens, integrations, and notifications. The user experience is still fragmented: a service is only a shallow repo row, environment health is inferred from names in the frontend, integration health is partly TODO/mock-backed, and deployments still require manual image/run correlation.

The next product step is to make zcid answer one operational question well: **“What is the health and delivery posture of this service right now, and what evidence supports that answer?”**

The chosen direction combines two ideas:

- **Service Vitals:** a richer service catalog page that shows ownership, pipelines, deployments, environments, recent health, delivery trends, and current warnings.
- **Signal Matrix:** a generic but bounded way to ingest and roll up health/status signals from zcid subsystems into service/environment/integration health.

## Actors

- A1. **Service developer:** wants to know whether their service is healthy, what failed recently, and what to fix next.
- A2. **Platform engineer:** wants to see systemic health and whether failures come from application code, CI infrastructure, or external integrations.
- A3. **Project admin:** wants trustworthy service ownership, environment, and deployment status without manually stitching together multiple pages.

## Requirements

- R1. zcid must show a service-level vitals page for each service, not just a service list row.
- R2. The vitals page must aggregate existing facts from pipelines, deployments, environments, and step executions before introducing any external observability dependency.
- R3. Environment health shown in the UI must come from persisted or computed signals, not name-based frontend heuristics.
- R4. The system must store health/status signals with source, target entity, severity/status, reason/message, freshness, and timestamps.
- R5. Services must be linkable to relevant pipelines and deployments enough for a vitals page to summarize delivery health.
- R6. The first version must avoid becoming a general observability platform; it should only track delivery-platform signals needed by zcid.
- R7. The vitals page must degrade gracefully when data is missing, stale, or insufficient.
- R8. The feature must preserve existing service, environment, deployment, and pipeline APIs unless explicitly extended.

## Key Flows

- F1. **View service vitals:** user opens a service and sees owner/repo metadata, linked pipelines, recent run success rate, recent deployments, environment health, warnings, and freshness indicators.
- F2. **Inspect a warning:** user clicks a warning such as “prod environment signal stale” or “pipeline test step degraded” and sees the source signal and linked run/deployment/integration.
- F3. **Refresh environment/integration health:** zcid records a new signal from an internal check or existing deployment status and updates the visible rollup.
- F4. **Handle missing data:** new service with no runs/deployments shows empty-state explanations instead of fake healthy/degraded values.

## Acceptance Examples

- AE1. Given a service with linked pipeline runs and deployments, when the user opens its vitals page, then zcid shows recent success rate, latest deployment per environment, and active warning count.
- AE2. Given an environment with no recent health signal, when it appears in Service Vitals, then its health is shown as stale/unknown with the last signal timestamp, not healthy by default.
- AE3. Given a failed pipeline step recorded in `step_executions`, when the service vitals page summarizes warnings, then the warning links back to the run detail and identifies the affected step.
- AE4. Given a service with no linked pipeline/deployment history, when the page loads, then the user sees a clear “no delivery data yet” state and actions/links to configure pipelines or deployments.
- AE5. Given an integration probe fails, when a signal is recorded for that integration or related environment, then the affected service/environment rollup reflects degraded/unknown status with the failure reason.

## Scope Boundaries

- Include a bounded signal model and rollup for delivery health.
- Include service vitals UI inside existing project service surfaces.
- Include a path to link services to pipelines/deployments using existing repo URL, project, or explicit IDs.
- Do not build full org/team multi-tenancy in this slice.
- Do not build a full Backstage replacement.
- Do not build generic metrics storage, tracing, or arbitrary Prometheus querying.
- Do not implement feature flags, canary orchestration, or air-traffic-control deployment queues in v1.
- Do not require external observability products for the first useful release.

## Product Decisions

- **Primary object:** service, not pipeline. Pipelines and deployments become evidence attached to a service.
- **Health semantics:** explicit statuses are `healthy`, `warning`, `degraded`, `unknown`, and `stale`. `unknown` means no signal; `stale` means signal exists but is too old to trust.
- **Signal scope:** v1 only accepts zcid-owned signals from existing subsystems: deployment sync/health, step execution summaries, integration/admin health checks, and optional registry check result.
- **UI posture:** show evidence and freshness, not just a colored badge. Users should be able to tell why a status exists.
- **Data absence:** absence of data is never treated as healthy.

## Success Criteria

- A developer can open one page and understand whether a service is healthy enough to deploy or investigate.
- Environment health no longer uses frontend name heuristics.
- At least three signal sources feed the initial rollup: deployment status, step execution/run status, and integration/admin health.
- Empty and stale states are explicit and understandable.
- Existing service list, environment list, deployment list, and run detail pages continue to work.

## Assumptions

- The current `services.repo_url` is sufficient for first-pass automatic association with pipelines that use matching Git connections or repo URLs; explicit pipeline/service linking can be added when inference is not reliable.
- The first version can compute some rollups on read or via simple SQL without needing a separate stream processor.
- Real health checks can start as internal probes and existing statuses before adopting Tekton Results, OTel, or external observability inputs.

## Deferred for Later

- Org-level service catalog and cross-project scorecards.
- Preview environment leases and auto-reaping.
- Context-aware deployment picker.
- Progressive delivery state machine / air traffic control.
- Backstage import/export compatibility.
- External signal ingestion APIs for third-party tools.
