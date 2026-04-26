---
date: 2026-04-26
plan-id: 2026-04-26-001
title: "Sprint 1 — Next-Wave Feature Implementation"
source-ideation: docs/ideation/2026-04-26-zcid-next-wave-ideation.md
status: completed
---

# Sprint 1 Implementation Plan — Next-Wave Features

## Overview

Five self-contained features, ~15 dev-days total. No org-layer dependency. All build on already-landed infrastructure (`stepexec`, `access_tokens`, existing template system).

**Execution order**: Items 1 and 2 are independent and can be parallelized. Items 3–5 are also independent of each other. All five can be worked in parallel by different developers.

---

## Item 1: Step-Level Timeline on Run Detail Page

**Ideation ref**: 1.2  
**Effort**: S (1.5–2 days)  
**Risk**: Low — pure frontend, data already available

### What
Add a horizontal Gantt/timeline chart below the existing DAG view on `PipelineRunDetailPage.tsx`. Each step is a bar spanning its `startedAt` → `finishedAt` window, color-coded by status.

### Data available
`StepExecution` model already has: `stepName`, `taskRunName`, `stepIndex`, `status`, `startedAt`, `finishedAt`, `durationMs`.

The run detail page already fetches step executions (or can via `GET /projects/:pid/pipeline-runs/:rid/steps` — verify endpoint exists).

### Backend changes
- **Verify**: `GET /projects/:projectId/pipeline-runs/:runId/steps` returns `[]StepExecution`. If not exposed, add handler in `internal/pipelinerun/handler.go` calling `stepexec.Repo.ListByPipelineRun`.
- No new DB queries needed.

### Frontend changes

**File**: `web/src/pages/projects/pipelines/PipelineRunDetailPage.tsx`

1. After the existing DAG/status section, add a `<StepTimeline>` component.
2. New file: `web/src/components/pipeline/StepTimeline.tsx`

```tsx
// StepTimeline.tsx — props
interface Props {
  steps: StepExecution[]   // sorted by startedAt
  runStartedAt: string     // ISO string — baseline for relative offsets
}
```

**Rendering logic**:
- Compute `totalDurationMs = max(finishedAt) - min(startedAt)` across all steps
- Each step bar: `left = (startedAt - runStart) / total * 100%`, `width = durationMs / total * 100%`
- Min bar width: 4px (so zero-duration steps are visible)
- Color map: `succeeded → green-500`, `failed → red-500`, `running → blue-400 animated`, `pending → zinc-300`, `cancelled/interrupted → zinc-400`
- Tooltip on hover: step name, status, duration formatted as `Xm Ys`
- Group by `taskRunName` (stage) — show stage label on left, steps as rows within stage

**Styling**: Use existing Arco Design `Tooltip` for hover. CSS grid for layout. No new dependencies.

### Acceptance criteria
- [ ] Timeline renders on run detail page for both mock and real runs
- [ ] Hovering a bar shows step name + duration tooltip
- [ ] Steps grouped by stage with stage label
- [ ] Color-coded by status
- [ ] No TypeScript errors (`npx tsc --noEmit`)

---

## Item 2: Scoped Access Tokens

**Ideation ref**: 5.2  
**Effort**: S–M (2–3 days)  
**Risk**: Low — schema already has `scopes TEXT`, just needs enforcement

### What
The `access_tokens` table already has a `scopes TEXT` column and `project_id` FK. Currently scopes are stored as a free-form string. This item:
1. Defines a canonical scope vocabulary
2. Enforces scopes in the auth middleware
3. Updates the token creation UI to show scope checkboxes

### Scope vocabulary

```
pipelines:read          GET /projects/:id/pipelines, /pipeline-runs
pipelines:trigger       POST /projects/:id/pipeline-runs (trigger)
deployments:read        GET /projects/:id/deployments
deployments:write       POST/PUT /projects/:id/deployments
variables:read          GET /projects/:id/variables (non-secret values)
notifications:read      GET /projects/:id/notification-rules
admin:read              GET /admin/* (admin-only tokens)
```

Scopes stored as comma-separated string in `scopes` column (already TEXT). Example: `"pipelines:read,pipelines:trigger"`.

### Backend changes

**New file**: `pkg/auth/scopes.go`
```go
package auth

const (
    ScopePipelinesRead    = "pipelines:read"
    ScopePipelinesTrigger = "pipelines:trigger"
    ScopeDeploymentsRead  = "deployments:read"
    ScopeDeploymentsWrite = "deployments:write"
    ScopeVariablesRead    = "variables:read"
    ScopeNotificationsRead = "notifications:read"
    ScopeAdminRead        = "admin:read"
)

// AllScopes for validation
var AllScopes = []string{...}

func HasScope(tokenScopes string, required string) bool {
    for _, s := range strings.Split(tokenScopes, ",") {
        if strings.TrimSpace(s) == required { return true }
    }
    return false
}
```

**Middleware**: `internal/auth/handler.go` or a new `pkg/middleware/token_scope.go`
- Add `RequireTokenScope(scope string)` middleware that:
  1. Checks if request uses Bearer token (access token, not JWT session)
  2. Looks up token from DB/cache, checks `HasScope(token.Scopes, scope)`
  3. Returns 403 if scope missing

**Token creation**: `internal/auth/` — when creating a token, validate that all requested scopes are in `AllScopes`. Reject unknown scopes with 400.

**Token lookup**: Add Redis cache for token hash → token record (TTL = 5 min) to avoid DB hit on every API call.

### Frontend changes

**File**: `web/src/pages/access-tokens/` (existing page)

Add scope selection to the token creation form:
- Group scopes by resource: Pipelines, Deployments, Variables, Notifications, Admin
- Checkbox per scope with description
- "Select all" per group
- Show selected scopes as tags on the token list

### Acceptance criteria
- [ ] Token creation validates scope strings
- [ ] `RequireTokenScope` middleware returns 403 for missing scope
- [ ] Token list shows scopes as readable tags
- [ ] Token creation form has scope checkboxes grouped by resource
- [ ] No TypeScript errors, no Go build errors

---

## Item 3: Pipeline Analytics Dashboard

**Ideation ref**: 1.1  
**Effort**: M (3–4 days)  
**Risk**: Low — SQL aggregations over existing tables

### What
A new "Analytics" tab/page under each project showing: success/failure rate over time, median run duration trend, top-failing steps, most-triggered pipelines.

### Backend changes

**New endpoint**: `GET /projects/:projectId/analytics?range=7d`

**File**: `internal/pipelinerun/handler.go` (add) or new `internal/analytics/handler.go`

Response shape:
```json
{
  "range": "7d",
  "summary": {
    "totalRuns": 42,
    "successRate": 0.857,
    "medianDurationMs": 45000,
    "p95DurationMs": 120000
  },
  "dailyStats": [
    {"date": "2026-04-20", "total": 8, "succeeded": 7, "failed": 1, "medianDurationMs": 43000}
  ],
  "topFailingSteps": [
    {"stepName": "unit-test", "taskRunName": "build", "failureCount": 5, "totalCount": 12, "failureRate": 0.417}
  ],
  "topPipelines": [
    {"pipelineId": "...", "pipelineName": "main-ci", "runCount": 18, "successRate": 0.944}
  ]
}
```

**SQL queries** (all over `pipeline_runs` + `step_executions`):

```sql
-- Daily stats (last 7 days)
SELECT
  DATE(created_at) as date,
  COUNT(*) as total,
  COUNT(*) FILTER (WHERE status = 'succeeded') as succeeded,
  COUNT(*) FILTER (WHERE status = 'failed') as failed,
  PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY duration_ms) as median_duration_ms
FROM pipeline_runs
WHERE project_id = $1 AND created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date;

-- Top failing steps
SELECT step_name, task_run_name,
  COUNT(*) FILTER (WHERE status = 'failed') as failure_count,
  COUNT(*) as total_count
FROM step_executions se
JOIN pipeline_runs pr ON pr.id = se.pipeline_run_id
WHERE pr.project_id = $1 AND se.created_at >= NOW() - INTERVAL '7 days'
GROUP BY step_name, task_run_name
HAVING COUNT(*) FILTER (WHERE status = 'failed') > 0
ORDER BY failure_count DESC
LIMIT 10;
```

**Range parameter**: support `7d`, `30d`, `90d`. Default `7d`.

### Frontend changes

**New page**: `web/src/pages/projects/analytics/AnalyticsPage.tsx`

**Route**: Add to `ProjectLayout.tsx` sidebar nav under a new "Analytics" section.

**Components**:
1. `<SummaryCards>` — 4 stat cards: Total Runs, Success Rate, Median Duration, P95 Duration (use Arco `Statistic`)
2. `<DailyTrendChart>` — line chart with dual Y-axis (run count bars + success rate line). Use Arco `Chart` or recharts.
3. `<TopFailingSteps>` — table: Step Name | Stage | Failures | Total | Failure Rate (progress bar)
4. `<TopPipelines>` — table: Pipeline | Runs | Success Rate

**Range selector**: Arco `Radio.Group` with `7d / 30d / 90d` options at top of page.

### Acceptance criteria
- [ ] Analytics page accessible from project sidebar
- [ ] Summary cards show correct aggregated values
- [ ] Daily trend chart renders with correct date range
- [ ] Top failing steps table populated
- [ ] Range selector changes data
- [ ] Empty state when no runs exist
- [ ] No TypeScript errors

---

## Item 4: Native Slack Integration

**Ideation ref**: 4.1  
**Effort**: M (3–4 days)  
**Risk**: Low — well-documented Slack API

### What
Add `slack` as a notification channel type alongside the existing `webhook`. Sends rich Block Kit messages on build/deploy events.

### Backend changes

**New file**: `internal/notification/slack.go`

```go
package notification

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type SlackSender struct{ httpClient *http.Client }

func NewSlackSender() *SlackSender {
    return &SlackSender{httpClient: &http.Client{Timeout: 10 * time.Second}}
}

// SendBuildNotification sends a Block Kit message to Slack.
// botToken: xoxb-... token; channel: #channel-name or channel ID
func (s *SlackSender) SendBuildNotification(ctx context.Context, botToken, channel string, event BuildEvent) error {
    blocks := buildSlackBlocks(event)
    payload := map[string]any{
        "channel": channel,
        "blocks":  blocks,
        "text":    fmt.Sprintf("[zcid] %s: %s — %s", event.PipelineName, event.Status, event.ProjectName),
    }
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/chat.postMessage", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+botToken)
    req.Header.Set("Content-Type", "application/json")
    resp, err := s.httpClient.Do(req)
    if err != nil { return fmt.Errorf("slack send: %w", err) }
    defer resp.Body.Close()
    var result struct{ OK bool `json:"ok"`; Error string `json:"error"` }
    json.NewDecoder(resp.Body).Decode(&result)
    if !result.OK { return fmt.Errorf("slack API error: %s", result.Error) }
    return nil
}
```

**Block Kit message structure**:
- Header block: `[zcid] Pipeline Run — {status emoji} {status}`
- Section: Project / Pipeline / Branch / Commit SHA (first 8 chars)
- Section: Duration | Triggered by
- Actions block: "View Run" button linking to `{zcid_base_url}/projects/{id}/pipelines/runs/{runId}`
- Color sidebar: green for success, red for failure, yellow for cancelled

**Model change**: `internal/notification/model.go`

```go
type ChannelType string

const (
    ChannelWebhook ChannelType = "webhook"
    ChannelSlack   ChannelType = "slack"
)

type NotificationRule struct {
    // existing fields...
    ChannelType ChannelType `gorm:"column:channel_type"`
    WebhookURL  string      `gorm:"column:webhook_url"`   // used for webhook type
    SlackToken  string      `gorm:"column:slack_token"`   // encrypted, used for slack type
    SlackChannel string     `gorm:"column:slack_channel"` // used for slack type
}
```

**Migration**: `000023_add_notification_channel_type.up.sql`
```sql
ALTER TABLE notification_rules
  ADD COLUMN IF NOT EXISTS channel_type VARCHAR(20) NOT NULL DEFAULT 'webhook',
  ADD COLUMN IF NOT EXISTS slack_token TEXT,
  ADD COLUMN IF NOT EXISTS slack_channel VARCHAR(120);
```

**Service change**: `internal/notification/service.go` — dispatch based on `ChannelType`. Encrypt `SlackToken` using existing `pkg/crypto` AES-256-GCM before storing.

**Config**: Add `notification.slack_base_url` (for the "View Run" link) to `config/config.yaml`.

### Frontend changes

**File**: `web/src/pages/projects/notifications/` (existing notification rule form)

Add channel type selector (Radio: Webhook / Slack):
- Webhook selected → show existing webhook URL field
- Slack selected → show:
  - "Bot Token" field (password input, stored encrypted)
  - "Channel" field (text, e.g. `#ci-alerts`)
  - "Test" button → calls `POST /projects/:id/notification-rules/:rid/test`

### Acceptance criteria
- [ ] Migration applies cleanly
- [ ] Slack token stored encrypted in DB
- [ ] Block Kit message sent on build_success / build_failed events
- [ ] "View Run" link in message navigates to correct run detail page
- [ ] Test button sends a sample message
- [ ] Webhook type still works unchanged
- [ ] No TypeScript errors, no Go build errors

---

## Item 5: Parameterized Pipeline Templates

**Ideation ref**: 2.1  
**Effort**: M (3–4 days)  
**Risk**: Low — `PipelineConfig.Params` and `PipelineTemplate.Params` already exist

### What
Templates already declare `[]ParamConfig` (name, type, defaultValue, description, required). Currently these params are ignored at instantiation time — the template config is copied verbatim. This item:
1. Adds a param-fill step to the template selection flow
2. Substitutes param values into the pipeline config at instantiation
3. Validates required params before creating the pipeline

### Backend changes

**Existing**: `PipelineConfig.Params []ParamConfig` — already in model. `PipelineTemplate.Params []ParamConfig` — already in template.go.

**New endpoint**: `POST /projects/:projectId/pipelines/from-template`

```go
type FromTemplateRequest struct {
    TemplateID string            `json:"templateId" binding:"required"`
    Name       string            `json:"name" binding:"required"`
    Params     map[string]string `json:"params"` // param name → value
}
```

**Service logic** (`internal/pipeline/service.go`):
```go
func (s *Service) CreateFromTemplate(ctx context.Context, projectID string, req FromTemplateRequest) (*Pipeline, error) {
    tmpl := s.templateRegistry.Get(req.TemplateID)
    if tmpl == nil { return nil, ErrTemplateNotFound }

    // Validate required params
    for _, p := range tmpl.Params {
        if p.Required {
            if v, ok := req.Params[p.Name]; !ok || v == "" {
                return nil, fmt.Errorf("required param %q missing", p.Name)
            }
        }
    }

    // Substitute params into config JSON
    config := substituteParams(tmpl.Config, req.Params, tmpl.Params)

    return s.Create(ctx, projectID, CreatePipelineRequest{
        Name:   req.Name,
        Config: config,
    })
}

func substituteParams(config PipelineConfig, values map[string]string, defs []ParamConfig) PipelineConfig {
    // Build substitution map: $(params.NAME) → value (or defaultValue)
    subs := make(map[string]string)
    for _, p := range defs {
        v := p.DefaultValue
        if override, ok := values[p.Name]; ok && override != "" {
            v = override
        }
        subs[fmt.Sprintf("$(params.%s)", p.Name)] = v
    }
    // Marshal config to JSON, do string replacement, unmarshal back
    raw, _ := json.Marshal(config)
    result := string(raw)
    for placeholder, value := range subs {
        result = strings.ReplaceAll(result, placeholder, value)
    }
    var out PipelineConfig
    json.Unmarshal([]byte(result), &out)
    return out
}
```

**Tekton-style `$(params.NAME)` syntax** — consistent with Tekton's own param substitution, familiar to users.

### Frontend changes

**File**: `web/src/pages/projects/pipelines/TemplateSelectPage.tsx`

Add a two-step flow:
1. **Step 1** (existing): Select template from grid
2. **Step 2** (new): Fill params form + set pipeline name → "Create Pipeline"

Step 2 UI:
- Pipeline name field (required)
- For each `ParamConfig` in selected template:
  - `type: string` → text input with placeholder = `defaultValue`
  - `type: secret` → password input (value will be stored as variable reference — future work; for now treat as string)
  - `type: enum` → select dropdown (options from `ParamConfig.Options` — add `Options []string` to `ParamConfig` if not present)
  - Show `description` as helper text
  - Mark required params with asterisk
- "Back" button returns to template grid
- "Create Pipeline" calls `POST /projects/:id/pipelines/from-template`

**State management**: Use existing React `useState` for selected template + param values. No new state library needed.

### Model addition (if needed)

Add `Options []string` to `ParamConfig` for enum support:
```go
type ParamConfig struct {
    Name         string   `json:"name"`
    Type         string   `json:"type"`   // "string" | "secret" | "enum"
    Options      []string `json:"options,omitempty"` // for enum type
    DefaultValue string   `json:"defaultValue,omitempty"`
    Description  string   `json:"description,omitempty"`
    Required     bool     `json:"required"`
}
```

Update built-in templates in `template.go` to use `$(params.NAME)` placeholders and declare params with descriptions.

### Acceptance criteria
- [ ] `POST /projects/:id/pipelines/from-template` creates pipeline with substituted params
- [ ] Required param validation returns 400 with clear error message
- [ ] Template select page shows param form after template selection
- [ ] Param form validates required fields before submit
- [ ] Created pipeline config has substituted values (no `$(params.*)` placeholders remaining)
- [ ] Existing "create blank pipeline" flow unchanged
- [ ] No TypeScript errors, no Go build errors

---

## Shared Implementation Notes

### Conventions to follow
- Go: handler → service → repo layering; no business logic in handlers
- Error wrapping: `fmt.Errorf("context: %w", err)` throughout
- Frontend: Arco Design components; no new UI libraries
- New API endpoints: register in `cmd/server/main.go` route groups
- Encrypted fields: use `pkg/crypto` AES-256-GCM (same as variable encryption)
- All new DB columns: add to GORM model struct with `gorm:"column:..."` tag

### Testing
- Backend: add unit tests in `*_test.go` alongside service files
- Frontend: `npx vitest run` must pass
- TypeScript: `npx tsc --noEmit` must pass with zero errors
- Go build: `go build ./...` must succeed

### Verification checklist (per item)
- [ ] `go build ./...` — zero errors
- [ ] `go test ./pkg/... ./internal/... -count=1` — all pass
- [ ] `cd web && npx tsc --noEmit` — zero errors
- [ ] `cd web && npx vitest run` — all pass
- [ ] Manual smoke test in browser (mock mode)

---

## Dependency Graph

```
Item 1 (Timeline)          ─── independent
Item 2 (Scoped Tokens)     ─── independent
Item 3 (Analytics)         ─── independent (needs stepexec data to exist)
Item 4 (Slack)             ─── independent (needs migration 000023)
Item 5 (Param Templates)   ─── independent
```

All five can be developed in parallel. Suggested merge order: 2 → 1 → 5 → 3 → 4 (simplest to most complex).
