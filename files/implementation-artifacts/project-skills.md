# Project Skill Summaries

## Story 1.4 - 结构化日志与脱敏引擎

### Skill: Go slog runtime level control
- Use `slog.LevelVar` as shared runtime switch.
- Keep `Init(level)` tolerant (fallback to info) and `SetLevel(level)` strict (return error on invalid input).
- Expose `CurrentLevel()` for admin API and diagnostics.

### Skill: Safe logging with recursive masking
- Apply masking to both log message and attrs.
- Detect sensitive keys (`password/secret/token/apikey/...`) and replace with `***`.
- Recursively process grouped attrs via `slog.GroupValue(...)` to avoid nested leaks.

### Skill: Request-level observability in Gin
- Put `RequestID` middleware before access logging.
- Emit request logs after `c.Next()` with status + latency + route path.
- Include `requestId` in access and panic recovery logs for correlation.

### Skill: Config override hardening
- Keep defaults in code, YAML for baseline, ENV as highest priority.
- Validate critical env override paths with tests (`SERVER_PORT`, `SERVER_LOG_LEVEL`).

### Skill: Story completion checklist (backend)
- Update story artifact status to `review`.
- Update `files/implementation-artifacts/sprint-status.yaml` story state.
- Run both `go test ./... -v` and `go build ./...` before marking complete.

## Story 1.5 - Redis 连接与缓存基础层

### Skill: Cache wrapper error contracts
- Return typed `ErrCacheMiss` for misses so callers can branch into DB fallback.
- Return wrapped operational errors for Redis/network failures; never panic on cache operations.
- Validate nil client early in each method to keep failure mode explicit.

### Skill: Redis key strategy
- Use a shared prefix strategy (`keyPrefix + key`) to keep namespaces clear.
- Keep TTL policy centralized with `defaultTTL`, and allow per-call override.

### Skill: Availability-aware cache behavior
- Treat cache as best-effort infra: failures should be recoverable by upper layers.
- Keep readiness using direct Redis ping, but application paths should degrade gracefully.

## Story 1.7 - 前端设计系统基础

### Skill: Status style single source of truth
- Define a centralized `STATUS_MAP` to standardize status color/bg/icon/label tokens across pages.
- Keep status definitions typed (`StatusStyle`) so downstream components consume consistent fields.

### Skill: Error boundary retry test pattern
- For React 19 concurrent rendering, use a controlled throw flag in test setup and toggle it before retry click to avoid false unhandled error recovery paths.
- Assert fallback (`role="alert"`) and retry button visibility before triggering retry.

### Skill: Lightweight reusable UI primitives
- Keep generic cross-page components in `components/common` (`PageSkeleton`, `PermissionGate`) with minimal props and predictable behavior.
- Add focused tests for each primitive to lock expected render/no-render contracts.

## Story 1.8 - API 客户端自动生成流水线

### Skill: Stabilize OpenAPI codegen pipeline early
- Validate `make swag` + `npm run codegen` chain as part of baseline engineering flow before real business endpoints are added.
- Keep generated client output fixed at `web/src/services/generated` so import paths remain stable across stories.

### Skill: Tooling compatibility over planned flags
- Prefer CLI flags supported by the installed toolchain; remove unsupported `swag` options to keep CI and local runs deterministic.
- Treat generator invocation as executable contract in Makefile/scripts and verify with a real run after changes.

### Skill: Codegen verification pattern
- After regeneration, assert both generated SDK/types files and frontend build success to catch schema-client drift early.
- Even with empty/seed OpenAPI specs, keep generation artifacts committed to confirm pipeline health for upcoming API stories.

## Story 1.9 - Helm Chart 基础骨架

### Skill: Minimal Helm chart structure
- Start with Chart.yaml + values.yaml + _helpers.tpl + deployment/service/configmap templates.
- Keep _helpers.tpl focused: labels, selectorLabels, fullname — avoid over-abstracting at MVP stage.

### Skill: Config-to-values alignment
- Mirror the application config.yaml structure in values.yaml so ConfigMap template renders a valid config file directly.
- Sensitive values (passwords, keys) go through K8s Secret env vars with `optional: true` to avoid hard dependency on Secret existence during development.

### Skill: Helm validation as acceptance gate
- Use `helm lint` (zero errors) + `helm template` (valid manifests) + `helm template --set` (override verification) as the three-step validation for chart stories.

## Story 2.1 - 用户登录与 JWT 双 Token 认证

### Skill: JWT dual-token authentication
- Access Token (30min) + Refresh Token (7 days) stored in Redis.
- Refresh flow must validate: 1) JWT valid and not expired; 2) Redis session exists and matches.
- Use Redis for session revocation (logout, user disable).

### Skill: Password security with bcrypt
- Always use bcrypt for hashing; never store plaintext passwords.
- Reuse a shared `HashPassword` utility across create/update flows.

### Skill: Login failure information disclosure
- Avoid revealing whether a user exists (same message for "user not found" and "wrong password").

## Story 2.2 - 用户账号管理

### Skill: Session invalidation on user disable
- When disabling a user, remove all Refresh Tokens from Redis.
- Ensure disabled users cannot login or refresh.

### Skill: Permission enforcement at backend
- Do not rely on frontend hiding; enforce access control in middleware.
- Non-admin access to admin endpoints must return 403.

## Story 2.3 - 角色与权限管理

### Skill: Casbin RBAC four-tuple model
- Use `(sub, proj, obj, act)` for policy checks.
- Store policies in PostgreSQL; load via Casbin Enforcer.

### Skill: Redis Watcher for policy hot-reload
- Use Redis Watcher so policy changes take effect without restart.
- Write policy changes to PostgreSQL and trigger Watcher.

### Skill: Cross-epic TODO tracking
- Mark cross-epic dependencies explicitly as TODO (e.g., FR5 depends on variable module in Epic 4).

## Story 2.4 - 前端登录页与认证状态管理

### Skill: Frontend state layering
- Server data: TanStack Query. Client/session state: Zustand.
- Do not cache business API data in auth store.

### Skill: Centralized 401 handling
- Use Axios interceptors for 401 and token refresh.
- On 401: refresh → retry; on refresh failure → clear session and redirect to login.

### Skill: Project structure conventions (frontend)
- Pages: `web/src/pages/<domain>/`
- State: `web/src/stores/<domain>.ts`
- Network: `web/src/services/<domain>.ts`
- Route guard: `web/src/components/common/RequireAuth.tsx`

## Story 2.5 - 前端权限路由守卫

### Skill: RequireAuth + RequirePermission flow
- First authenticate (RequireAuth), then check permissions (RequirePermission).
- Use `authStore` `PermissionKey` and `hasPermission` as single source of truth.

### Skill: PermissionGate for operation-level control
- Use PermissionGate around action buttons to hide unauthorized actions.
- Combine route-level and operation-level checks.

## Story 2.6 - 前端用户管理页面

### Skill: JWT payload as source of truth
- Parse role from JWT payload; do not hardcode roles.
- Incorrect role handling can break the permission system.

### Skill: API baseURL configuration
- Align frontend and backend on baseURL; avoid double prefixes like `/api/v1/api/v1`.

### Skill: Database migration in Definition of Done
- Include migration scripts (up + down) for any schema change.
- Test migrations before marking a story done.

### Skill: Form validation and feedback
- Username required; password length >= 6 for create.
- Success: Message (3s); destructive actions: Popconfirm.

## Story 2.7 - 前端顶部个人信息与退出登录

### Skill: Single source of truth for auth state
- Use Zustand `useAuthStore` for `accessToken`, `refreshToken`, `user`, `permissions`.
- Use `clearSession()` for logout; avoid parallel auth stores.

### Skill: Logout robustness
- Call logout API, then always run `clearSession()` and redirect.
- If API fails, still clear local session and redirect.

## Epic 2 Retrospective Skills

### Skill: API Contract First workflow
- Write OpenAPI spec before implementation.
- Backend implements from spec; frontend generates client from spec.

### Skill: E2E testing for integration
- Unit tests alone miss integration issues.
- Add Playwright E2E for flows: login → permission check → user management → logout.

### Skill: Early frontend-backend integration
- Integrate early to catch contract and behavior mismatches.

## Story 3.1 - 项目 CRUD

### Skill: Domain module pattern (handler → service → repo)
- New domain: `internal/<domain>/` with model, dto, repo, service, handler.
- Handler: `NewHandler(service) → RegisterRoutes(router)`.

### Skill: Soft delete with conditional uniqueness
- Use `status = 'deleted'` instead of physical delete.
- Unique index: `WHERE status != 'deleted'` so deleted names can be reused.

### Skill: Repo error handling chain
- Use `isUniqueConstraintError()` for unique constraint violations.
- Map repo errors to `response.BizError` in service.

### Skill: Naming conventions (JSON/DB)
- JSON: camelCase (e.g., `projectId`). DB: snake_case (e.g., `project_id`).

### Skill: Pagination format
- `{"items":[], "total":N, "page":1, "pageSize":20}`.
- Default page=1, pageSize=20; max pageSize=100.

### Skill: project_members for project roles
- Use `project_members` table for project-level roles when Casbin g2 is not yet in use.
- Full Casbin g2 integration can be added later.

## Story 3.2 - 环境管理与 Namespace 映射

### Skill: Global uniqueness for external resources
- Namespace must be globally unique; use unique index with `WHERE status != 'deleted'`.
- Project-scoped uniqueness: `(project_id, name)` unique within project.

### Skill: Nested resource routing
- Route pattern: `/api/v1/projects/:id/<resource>`.
- Register nested routes under parent resource group.

## Story 3.3 - 服务管理

### Skill: Avoiding Go naming conflicts
- Use `internal/svcdef/` instead of `internal/service/` to avoid clash with Go "service" layer.
- Choose module names that do not conflict with common Go terms.

### Skill: Project-scoped uniqueness
- Unique constraint: `(project_id, name) WHERE status != 'deleted'`.
- Return 409 for duplicate names within a project.

## Story 3.4 - 项目成员与角色管理

### Skill: Extending existing modules
- Add member routes to existing `internal/project/` instead of a new module.
- Reuse `project_members` table and existing repo methods.

### Skill: Conflict error handling for relationships
- `AddMember` returns `ErrMemberExists` for duplicate members.
- Ignore `ErrMemberExists` in `CreateProject` to handle "creator already member" case.

### Skill: JOIN for related data
- Use JOIN with `users` in `ListMembers` to include usernames.
- Avoid N+1 queries when listing members with user info.

## Story 3.5 - 前端项目管理页面

### Skill: Domain-specific API service layer
- Create `services/<domain>.ts` for all related APIs.
- Centralize API calls per domain.

### Skill: Project-level layout with nested routes
- `ProjectLayout` with sidebar + Outlet for nested routes.
- Sidebar navigation: environments, services, members.

### Skill: Permission extension pattern
- Add new `PermissionKey` in `authStore` when adding new features.
- Extend `ROLE_PERMISSIONS` mapping for all applicable roles.

### Skill: Inline editing for simple updates
- Use inline dropdown for member role changes.
- Reduces need for separate edit modals for simple single-field updates.

## Cross-cutting Skills (Epic 2 & 3)

### Skill: Error propagation chain
- Repo → service → handler. Handler uses `response.HandleError(c, err)` for consistent responses.

### Skill: UUID for primary keys
- Use `github.com/google/uuid` for IDs.

### Skill: Migration file format
- Follow `migrations/NNNNNN_description.up.sql` and `.down.sql`.
- Include rollback scripts for every migration.

### Skill: Code review catches AC misalignment
- Always verify AC permission requirements before coding; copy-paste from other handlers causes role restriction bugs.
- Use `isAdminOrProjectAdmin` helper when AC specifies "管理员或项目管理员".

### Skill: Type safety for database timestamps
- Use `time.Time` for Go structs mapping to DB TIMESTAMP columns.
- Format to string at the API response layer, not at the DTO level.

## Story 4.1 - 多层级变量 CRUD

### Skill: Partial unique index for scoped uniqueness
- Use PostgreSQL `CREATE UNIQUE INDEX ... WHERE scope = 'global' AND status != 'deleted'` for global scope.
- Use `CREATE UNIQUE INDEX ... ON (project_id, key) WHERE scope = 'project' AND status != 'deleted'` for project scope.
- Partial indexes elegantly handle soft delete + uniqueness without business layer checks.

### Skill: Multi-scope variable API design
- Separate global routes (admin-only: `/api/v1/admin/variables`) from project routes (`/api/v1/projects/:id/variables`).
- Variable merge query (`/merged`) returns global + project variables with project-level override.

### Skill: Error code centralization
- All business error codes must be defined in `pkg/response/codes.go`.
- Business modules import and reference codes; never define duplicate constants locally.

## Story 4.2 - 密钥变量加密与安全

### Skill: AES-256-GCM encryption pattern
- Key from environment variable (`ZCID_ENCRYPTION_KEY`, 32 bytes).
- Ciphertext format: nonce (12 bytes) + ciphertext + GCM tag, base64 encoded.
- Each encryption generates a unique nonce via `crypto/rand`, producing different ciphertext for same plaintext.

### Skill: Graceful degradation for optional security features
- If encryption key is not set, service starts normally but secret variable creation fails gracefully.
- Log a warning at startup; do not panic for missing optional config.

### Skill: Role-based data filtering (FR5)
- Implement `FilterForRole` at the service layer to strip secret variables for non-admin/non-project_admin users.
- After filtering, recalculate `total` to match `len(items)` for pagination consistency.

### Skill: Value masking for secrets
- `ToVariableResponse` always replaces secret values with `******` (MaskedValue constant).
- Never expose raw secret values in any API response.

## Story 4.3 - 运行时密钥注入与清理

### Skill: Interface-first stub design
- Define `SecretInjector` interface with `CreateSecret` and `DeleteSecret` in `pkg/k8s/secret.go`.
- Actual K8s client implementation deferred to later Epic; stub enables testing and compilation now.

### Skill: Variable resolution with decryption
- `ResolveVariables(projectID)` merges global + project variables and decrypts secret values.
- Only for internal use (pipeline runtime injection); never exposed via API.

## Story 4.4 - 前端变量管理页面

### Skill: Reusable form modal pattern
- `VariableFormModal` handles both create and edit modes via optional `editingItem` prop.
- Password input for secret type variables; auto-clear value field when switching types.

### Skill: Permission-aware navigation
- Add sidebar entries conditionally based on role permissions.
- Use `canViewAdminVariables` check in AppLayout for admin-only navigation items.

## Epic 4 Cross-cutting Skills

### Skill: Project scope middleware for sub-resource protection
- `RequireProjectScope(db)` middleware checks `project_members` table for non-admin users.
- Admin users bypass the check; non-members get 403.
- Apply middleware to a `/:id` route group that all sub-resource groups inherit from.

### Skill: Route registration splitting
- Split handler route registration into `RegisterCollectionRoutes` (list/create, no :id) and `RegisterResourceRoutes` (get/update/delete, under :id group).
- Enables per-scope middleware application without duplicating auth logic.

### Skill: P0 fix must block epic completion
- Security issues (unauthorized access) must be fixed before marking an epic as done.
- Retrospective-driven P0 fixes are reactive; prefer proactive security review during code review.

### Skill: Pagination with post-filter total correction
- When filtering results after database query (e.g., FR5 role filtering), always recalculate `total = len(filteredItems)`.
- Never return the pre-filter total, as it creates client-side pagination inconsistencies.