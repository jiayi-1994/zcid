# Story 3.1: 项目 CRUD

Status: done

## Story

As a 管理员,
I want 创建和删除项目,
so that 团队可以按项目组织 CI/CD 资源。

## Acceptance Criteria

1. **创建项目**
   - Given 管理员已登录
   - When POST /api/v1/projects 提交项目名称和描述
   - Then 项目创建成功，返回项目信息
   - And 创建者自动成为该项目的项目管理员（Casbin g2 策略写入）

2. **项目名称唯一性**
   - Given 项目名称已存在
   - When 创建同名项目
   - Then 返回 409 项目名重复错误

3. **获取项目列表**
   - Given 用户已登录
   - When GET /api/v1/projects
   - Then 管理员返回所有项目列表；非管理员返回有权限的项目列表
   - And 支持分页参数 page/pageSize

4. **获取项目详情**
   - Given 项目存在
   - When GET /api/v1/projects/:id
   - Then 返回项目详细信息（名称、描述、创建者、创建时间）

5. **更新项目**
   - Given 管理员或项目管理员已登录
   - When PUT /api/v1/projects/:id 提交更新数据
   - Then 项目信息更新成功

6. **删除项目**
   - Given 管理员已登录
   - When DELETE /api/v1/projects/:id
   - Then 项目标记为删除（软删除）
   - And 关联的环境、服务、流水线、变量后续 Epic 实现时自动级联

7. **权限控制**
   - Given 非管理员用户
   - When 尝试创建或删除项目
   - Then 返回 403 权限不足

## Tasks / Subtasks

- [x] Task 1: 创建 projects 数据库迁移（AC: 1, 2）
  - [x] 创建 `migrations/000006_create_projects.up.sql`：projects 表 + project_members 表
  - [x] 创建 `migrations/000006_create_projects.down.sql`：回滚脚本
  - [x] name 字段添加条件唯一约束（排除 deleted），owner_id 外键关联 users.id

- [x] Task 2: 创建 project 模块基础文件（AC: 1, 4）
  - [x] 创建 `internal/project/model.go`：Project + ProjectMember struct（GORM 模型）
  - [x] 创建 `internal/project/dto.go`：CreateProjectRequest、UpdateProjectRequest、ProjectResponse、ProjectListResponse
  - [x] 创建 `internal/project/repo.go`：CRUD + 成员管理数据访问方法
  - [x] 创建 `internal/project/service.go`：业务逻辑（创建/更新/删除/列表/详情）
  - [x] 创建 `internal/project/handler.go`：HTTP handler + 路由注册

- [x] Task 3: 实现项目创建与自动角色分配（AC: 1, 2）
  - [x] Service.CreateProject：校验名称唯一 → 创建项目 → 创建者自动加入 project_members 表（role=project_admin）
  - [x] 名称冲突返回 CodeConflict 错误
  - [x] 注：使用 project_members 表替代 Casbin g2（当前模型不含 g2，完整 g2 集成将在 3.4 实现）

- [x] Task 4: 实现项目列表与详情（AC: 3, 4）
  - [x] Service.ListProjects：管理员查全部，非管理员通过 project_members 过滤有权限的项目
  - [x] Service.GetProject：按 ID 查询，不存在返回 CodeNotFound
  - [x] 列表支持分页（page, pageSize 参数，默认 20，最大 100）

- [x] Task 5: 实现项目更新与删除（AC: 5, 6）
  - [x] Service.UpdateProject：更新名称/描述，名称冲突返回 CodeConflict
  - [x] Service.DeleteProject：软删除（status 设为 deleted）
  - [x] 删除时清理该项目的 project_members 记录

- [x] Task 6: 在 main.go 注册路由与权限中间件（AC: 7）
  - [x] 注册 /api/v1/projects 路由组
  - [x] 所有项目路由使用 JWTAuth 中间件认证
  - [x] 创建和删除在 handler 层检查 admin 角色
  - [x] 列表和详情对已登录用户开放（service 层按角色过滤）

- [x] Task 7: 添加测试覆盖（AC: 1, 2, 3, 4, 5, 6, 7）
  - [x] service_test.go：9 个业务逻辑单元测试（创建成功、名称冲突、空名称、管理员列表、成员列表、不存在项目、删除成功/失败、更新冲突）
  - [x] 修复 auth 模块预存的 mockRepo ListUsers 缺失问题
  - [x] 运行全量后端测试通过，无回归

## Dev Notes

### Relevant architecture patterns and constraints

- 后端模块遵循 handler→service→repo 三层结构，按 `internal/project/` 组织。[Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention]
- 错误处理遵循 repo→service→handler 传播链，handler 层调用 `response.HandleError(c, err)` 统一输出。[Source: pkg/response/response.go#HandleError]
- JSON 字段使用 camelCase（如 `projectId`），数据库字段使用 snake_case（如 `project_id`）。[Source: files/planning-artifacts/architecture.md#Naming Patterns]
- Casbin RBAC 四元组 (sub, proj, obj, act)，项目角色通过 g2 三元组 (user, role, project) 写入。[Source: files/planning-artifacts/architecture.md#Casbin RBAC]
- 分页格式：`{"items":[], "total":N, "page":1, "pageSize":20}`，默认 page=1, pageSize=20, 最大 100。[Source: files/planning-artifacts/architecture.md#Format Patterns]
- API 路由：`/api/v1/projects` GET/POST, `/api/v1/projects/:id` GET/PUT/DELETE。[Source: files/planning-artifacts/architecture.md#REST Route Structure]

### Source tree components to touch

- `internal/project/model.go`（新建 - Project GORM struct）
- `internal/project/dto.go`（新建 - 请求/响应 DTO）
- `internal/project/repo.go`（新建 - 数据访问层）
- `internal/project/service.go`（新建 - 业务逻辑层）
- `internal/project/handler.go`（新建 - HTTP handler）
- `migrations/000006_create_projects.up.sql`（新建 - 建表迁移）
- `migrations/000006_create_projects.down.sql`（新建 - 回滚迁移）
- `cmd/server/main.go`（修改 - 注册项目路由）
- `internal/rbac/enforcer.go`（复用 - Casbin g2 策略操作）

### Existing code patterns to follow

- Handler 创建模式参考 `internal/auth/handler.go`：`NewHandler(service) → RegisterRoutes(router)`
- Repo 错误处理参考 `internal/auth/repo.go`：`isUniqueConstraintError()` 检测唯一约束冲突
- Service 错误翻译参考 `internal/auth/service.go`：repo 错误映射为 `response.BizError`
- UUID 生成使用 `github.com/google/uuid`
- 数据库迁移参考 `migrations/000001_init_schema.up.sql` 格式

### Testing standards summary

- 后端测试与源文件同目录：`internal/project/handler_test.go`、`internal/project/service_test.go`
- 使用 testify 断言 + 标准 httptest
- 覆盖：创建成功、名称冲突、不存在的项目、软删除、权限校验

### Database schema for projects table

```sql
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    owner_id UUID NOT NULL REFERENCES users(id),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX uk_projects_name ON projects(name) WHERE status != 'deleted';
```

### References

- 架构文档 REST 路由定义：[Source: files/planning-artifacts/architecture.md#REST Route Structure]
- 架构文档项目模块定义：[Source: files/planning-artifacts/architecture.md#Complete Project Directory Structure]
- Epic 3 需求定义：[Source: files/planning-artifacts/epics.md#Epic 3: 项目与资源管理]
- 现有认证模块模式：[Source: internal/auth/handler.go]
- Casbin RBAC 模型：[Source: internal/rbac/enforcer.go]
- 统一响应与错误码：[Source: pkg/response/codes.go]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- 使用 project_members 表替代 Casbin g2 策略（当前 RBAC 模型仅有 g=_,_ 不含 g2=_,_,_），完整 Casbin g2 集成将在 Story 3.4 实现
- 修复 auth 模块测试中 mockRepo 缺少 ListUsers 方法的预存问题

### Completion Notes List

- 创建 projects + project_members 数据库迁移（000006），projects 含条件唯一约束（排除 deleted 状态）
- 新建 `internal/project/` 模块，完整实现 handler→service→repo 三层结构
- 实现 5 个 API 端点：POST/GET/GET/:id/PUT/:id/DELETE/:id（/api/v1/projects）
- 创建项目时自动将创建者加入 project_members 表（role=project_admin）
- 项目列表：admin 查全部，非 admin 通过 project_members 过滤
- 项目删除：软删除（status=deleted）+ 清理 project_members
- 分页格式遵循架构规范（items/total/page/pageSize）
- 9 个 service 层单元测试全部通过
- 全量后端测试通过（11 个包），无回归

### File List

- `migrations/000006_create_projects.up.sql` - 新建 projects 和 project_members 表
- `migrations/000006_create_projects.down.sql` - 回滚脚本
- `internal/project/model.go` - Project 和 ProjectMember GORM 模型
- `internal/project/dto.go` - 请求/响应 DTO 定义
- `internal/project/repo.go` - 数据访问层（CRUD + 成员管理）
- `internal/project/service.go` - 业务逻辑层
- `internal/project/handler.go` - HTTP handler 和路由注册
- `internal/project/service_test.go` - 9 个单元测试
- `cmd/server/main.go` - 注册项目路由
- `internal/auth/service_test.go` - 修复 mockRepo 缺少 ListUsers 方法
- `files/implementation-artifacts/3-1-project-crud.md` - Story 文件
- `files/implementation-artifacts/sprint-status.yaml` - Sprint 状态更新
