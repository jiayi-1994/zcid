---
date: 2026-04-24
topic: execution-twin
---

# Per-Step Execution Twin

## Problem Frame

Today zcid runs Tekton pipelines through `internal/ws/k8s_watcher.go`, which reads each `PipelineRun`'s `status.childReferences` and emits an in-memory `StepStatus` (StepID, Name, Status, StartedAt, FinishedAt) over WebSocket to the live UI. No per-step data is persisted. Once a run ends, the only durable artifacts are the aggregate `pipeline_runs` row (status, started_at, finished_at, params, artifacts JSONB) and archived logs in MinIO.

Consequences:
- No data backing remote build caching (no input-hash lookup).
- No data backing SLSA/in-toto provenance (no per-step input/output fingerprint).
- No data backing flaky-step detection, replay from step N, cost attribution, waterfall duration analysis, or determinism checks.
- Every downstream feature that needs per-step history must build its own collection layer.

The Twin fills this by persisting a forward-compatible per-step "fact record" sourced from the existing Tekton watcher path, plus a minimum query surface (API + waterfall UI) to prove the collection works and demonstrate first-order value.

---

## Actors

- A1. **Pipeline viewer** (human, end user of zcid): opens a PipelineRun detail page, wants to see per-step timing to answer "which step made this run slow?"
- A2. **Platform developer** (human, zcid contributor): builds future downstream features (cache / flaky / SLSA / cost) on top of the Twin, queries the captured records from Go services or SQL.
- A3. **Tekton watcher** (internal subsystem, `internal/ws/k8s_watcher.go`): observes `PipelineRun` + `TaskRun` state transitions from the K8s API; is the source of truth for step execution signals.

---

## Key Flows

- F1. **Step record capture**
  - **Trigger:** Tekton TaskRun status transition observed by the watcher (a step enters Running, or a step finishes).
  - **Actors:** A3
  - **Steps:**
    1. Watcher receives a PipelineRun/TaskRun update event via the existing dynamic-client watch loop.
    2. For each step inside each TaskRun, extract identity (pipeline_run_id, task_run_name, step_name, step_index), image reference, resolved command+args, resolved params, env keys + non-secret values, secret refs (by scoped name only, never values), started_at, finished_at, exit code, output artifact digests when present, Tekton-reported resource limits, and any Tekton-reported result variables.
    3. Compute an input_hash over a canonicalized subset of the captured fields (algorithm details deferred to planning).
    4. Upsert a row into the new `step_executions` table, keyed by (pipeline_run_id, task_run_name, step_name).
  - **Outcome:** one row per Tekton step per run, created on first-seen and updated on finish; visible via API within the watcher's normal lag.
  - **Covered by:** R1, R2, R3, R4, R7

- F2. **Waterfall view of a PipelineRun**
  - **Trigger:** A1 opens a PipelineRun detail page and switches to the "Steps" / "Waterfall" tab.
  - **Actors:** A1
  - **Steps:**
    1. Frontend requests `GET /api/v1/pipeline-runs/:id/step-executions` (exact path deferred to planning).
    2. Backend returns ordered step records with step_name, started_at, finished_at, duration_ms, status, and enough identity for click-through.
    3. UI renders horizontal bars per step, proportionally sized to duration, grouped by TaskRun (zcid stage), ordered by pipeline DAG order.
    4. In-flight steps (no finished_at yet) render with the bar extending to "now" and a pulsing style.
  - **Outcome:** A1 can answer "which step took longest" at a glance for any run ≤ 90 days old.
  - **Covered by:** R6, R8, R9

---

## Requirements

**Capture**
- R1. Persist one record per Tekton step execution per PipelineRun into a new `step_executions` table. Granularity is the Tekton step (container), not TaskRun and not pipeline. A 10-stage × 3-steps pipeline produces 30 rows per run.
- R2. Capture, at minimum: pipeline_run_id, task_run_name, step_name, step_index within its TaskRun, image reference (including digest when Tekton resolved one), command, args, resolved params, env keys, non-secret env values, secret refs (scoped name only), workspace mounts, started_at, finished_at, duration_ms, status, exit code, output artifact digests when reported, Tekton-reported resource limits, Tekton result variables, input_hash with a versioned hash_version field, trace_id (nullable placeholder for future OpenTelemetry integration).
- R3. Source all capture from the existing `internal/ws/k8s_watcher.go` watch loop. Do not introduce a separate K8s watch for step-level data.
- R4. Coverage is **real Tekton runs only**. Mock-mode runs do not produce step_executions rows in v1. This is deliberate; see Scope Boundaries.
- R5. Capture is idempotent: replaying the watcher over an already-observed PipelineRun must not create duplicate rows. Upsert on `(pipeline_run_id, task_run_name, step_name)`.

**Retention**
- R6. Retain full-fidelity records for 90 days by default. Retention window must be configurable via the same config surface as other zcid limits.
- R7. After the retention window, records are rolled up into a per-month summary (aggregate only: step_name, count, p50/p95 duration, success/failure counts, total cost if later added) and the full-fidelity rows deleted. Rollup is a v2 deliverable; in v1 records older than retention are hard-deleted without rollup.

**Query surface**
- R8. Expose a read API that returns all step_executions for a given PipelineRun, ordered by (task_run order, step_index), with fields sufficient to render the waterfall and to support later downstream features. Exact path and shape deferred to planning.
- R9. Ship a waterfall view in the PipelineRun detail page: per-step horizontal bars proportional to duration, grouped by stage (TaskRun), in DAG order, with in-flight steps extending to "now".

**Privacy / security**
- R10. Never persist secret values. Persist only the secret reference (the scoped variable name) and never a hash or fingerprint of the secret value.
- R11. input_hash canonicalization must exclude resolved secret values. If secret content participates in determinism, the reference identity + version participates, not the value.
- R12. Logs are already handled by `internal/logarchive/` and MinIO. Do not copy step stdout/stderr into `step_executions`; store only a reference to the existing log location when needed for cross-linking.

---

## Acceptance Examples

- AE1. **Covers R1, R2, R5.** Given a Tekton pipeline with two stages (TaskRuns), each containing a `git-clone` + `shell` step pair. When the run completes, then `step_executions` contains exactly 4 rows, each with a non-null image reference, started_at, finished_at, exit_code, and input_hash. Re-running the watcher against the same PipelineRun (e.g., after a restart) produces the same 4 rows — no duplicates.
- AE2. **Covers R9.** Given a completed PipelineRun with steps of duration 10s, 180s, 5s, 30s. When A1 opens the Steps tab, then a horizontal bar chart renders with bar widths in a 10:180:5:30 ratio, ordered by stage then step index.
- AE3. **Covers R9.** Given a currently running PipelineRun where step 2 is in-flight for 42s with no finished_at. When A1 opens the Steps tab, then step 2's bar extends from its started_at to "now" and visibly pulses; on the next watcher tick after step 2 completes, the bar freezes at the observed finished_at.
- AE4. **Covers R10, R11.** Given a step that references secret variable `github_token` (scoped at project level). When its record is persisted, then the `secret_refs` field contains the string `github_token` with its scope, the `env` field does not contain the token value, and the `input_hash` computed over the record is identical to a later run that uses the same secret reference even if the secret value was rotated in between (the hash depends on the reference, not the value).
- AE5. **Covers R6, R7 (v1 half).** Given records older than 90 days exist. When the retention worker runs, then those records are deleted. No monthly rollup rows are created in v1 — query results over historical windows beyond 90 days return empty (rollup is v2).

---

## Success Criteria

**Human outcome (A1):** For any real-Tekton PipelineRun ≤ 90 days old, a user opening its detail page can identify the longest-duration step within ~3 seconds without reading any log line.

**Downstream handoff (A2):** A platform developer building the next downstream (cache, flaky, SLSA, cost, replay) can query `step_executions` directly and find the fields they need already present or clearly slotted in the schema — no new capture path, no new K8s watch, no backfill of historical data. `input_hash` canonicalization is documented well enough that two developers independently computing it for the same step produce the same hash.

**Operational:** step_executions row-write failures are logged but never block the watcher from advancing its next PipelineRun event — capture is best-effort per row.

---

## Scope Boundaries

- **Mock-mode runs.** Deliberately excluded from v1 capture. Mock is for dev convenience; Twin is about real execution history. Revisit if mock users need downstream features for local development.
- **Downstream feature wiring.** v1 ships the schema, capture path, query API, and waterfall only. No cache lookup logic. No SLSA attestation emission. No flaky detection job. No cost calculation. No replay-from-step-N. Each downstream is a separate brainstorm.
- **Resource usage (CPU/memory-seconds).** Not captured in v1 — requires k8s metrics-server integration which is a separate plumbing exercise. Schema reserves the shape (the Tekton-reported `resources` block) but no metric collector runs v1.
- **Monthly rollup.** Hard-delete on expiry in v1; rollup table and job are v2.
- **Cross-run correlation.** Schema supports per-run queries. Cross-run aggregates (trending pass-rate, per-step cost trend) not delivered v1 beyond what the raw step_executions table allows via ad-hoc SQL.
- **Plugin / external consumer API.** v1 exposes an internal read API sufficient for the waterfall UI. Stable external contract and typed SDK are not v1 deliverables.
- **Backfill.** Historical PipelineRuns that completed before Twin shipped are not backfilled. Rows start accumulating from deploy forward.

---

## Key Decisions

- **Separate table, not JSONB on pipeline_runs.** Step-level rows are N-per-run and are expected to support trending, cache lookup, and provenance queries; normalized table wins on both query performance and foreign-key hygiene vs a JSONB array.
- **High-inclusivity schema, no downstream wiring v1.** Field set optimized for future downstream re-use; v1 itself ships only waterfall. This accepts some short-term over-capture to avoid a schema migration every time a downstream lands.
- **Step-level granularity, not TaskRun-level.** Step is finer and matches zcid's DAG-editor step card abstraction. Row volume (~3k/day at 100 runs) is tolerable.
- **90d retention default, monthly rollup deferred to v2.** Tradeoff: provenance / SLSA users who need longer history can raise retention via config; hard-delete keeps v1 simple.
- **Capture lives in `internal/ws/k8s_watcher.go`.** Keeps the capture path single-sourced from the existing Tekton watch. Extracting to a standalone `stepexec` service is an implementation refactor, not a product decision — defer to planning.
- **Secrets: refs only, never values, never hash of values.** Hashing values would create a rotation-linkability leak. Hash over refs keeps cache-key stable under value rotation, which is the correct cache-correctness behavior.

---

## Dependencies / Assumptions

- Existing Tekton watcher (`internal/ws/k8s_watcher.go`) is the sole K8s watch path for PipelineRuns and will remain so. Verified against repo.
- `internal/logarchive/` already handles per-run log archival to MinIO — Twin references these locations rather than duplicating.
- Postgres 16 + golang-migrate workflow (18 existing migration up/down pairs) is the standard for schema additions. Verified.
- `pipeline_runs.id` is a UUID suitable as a foreign key for 1:N relationships from `step_executions.pipeline_run_id`. Verified (`migrations/000013_create_pipeline_runs.up.sql`).
- Secret storage is AES-256-GCM encrypted via `ZCID_ENCRYPTION_KEY`; only secret references (scoped names) are ever persisted outside that boundary.
- Tekton `childReferences` structure reliably lists TaskRuns under a PipelineRun, and TaskRun status exposes per-step state (image, timing, exit code, results). This is standard Tekton API; assumed stable across the v0.56–v0.62 range zcid supports.

---

## Outstanding Questions

### Resolve Before Planning

*(empty — all product decisions settled in brainstorm)*

### Deferred to Planning

- [Affects R2][Technical] Exact column list and JSONB-vs-normalized split for fields like `env`, `params`, `workspace_mounts`, `output_artifact_digests`, `resources`. The product decision is "capture all of these"; the layout is an implementation choice.
- [Affects R2][Technical] Exact canonicalization algorithm for `input_hash` (field ordering, null handling, image-digest-vs-tag policy, command-arg escape rules, versioning scheme). Planning documents this and writes `hash_version` accordingly.
- [Affects R1, R3][Technical] Write-path model: synchronous write inside the watcher handler vs async buffered worker vs transactional outbox. Watcher is already eventually-consistent; need to choose a model that preserves that without losing rows under Postgres pressure.
- [Affects R6, R7][Technical] Retention enforcement mechanism (scheduled job vs Postgres-native TTL partitioning vs pg_cron). Product only requires the delete happens within ~24h of expiry.
- [Affects R8, R9][Technical] Exact API path, request/response shape, and integration pattern with the existing React PipelineRun detail page + `@xyflow/react` DAG canvas.
- [Affects R7][Needs research] Future v2 rollup schema — sufficient fields to preserve without overcommitting to today's unknowns. Decided during v2 brainstorm, not v1 planning.
- [Affects R4][Needs research] Whether a future executor-driver abstraction (ideation #3.4) should move Twin capture up one layer to `internal/pipelinerun/` so future non-Tekton executors automatically produce records. Out of scope for v1 but worth noting.

---

## Next Steps

-> `/ce-plan` for structured implementation planning.
