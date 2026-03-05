# Story 3.3: 服务管理

Status: done

## Story

As a 项目管理员,
I want 在项目内管理服务,
so that 流水线和部署可以关联到具体服务。

## Acceptance Criteria

1. **创建服务**
   - Given 项目管理员或管理员已登录
   - When POST /api/v1/projects/:id/services
   - Then 服务创建成功

2. **服务名称项目内唯一**
   - Given 同项目下服务名已存在
   - When 创建同名服务
   - Then 返回 409 冲突错误

3. **编辑服务**
   - Given 服务存在
   - When PUT /api/v1/projects/:id/services/:sid
   - Then 服务信息更新成功

4. **获取服务列表**
   - Given 项目存在
   - When GET /api/v1/projects/:id/services
   - Then 返回该项目下的服务列表

5. **删除服务**
   - Given 管理员或项目管理员
   - When DELETE /api/v1/projects/:id/services/:sid
   - Then 服务标记为删除（软删除）

## Tasks / Subtasks

- [x] Task 1: 创建 services 数据库迁移
- [x] Task 2: 创建 internal/svcdef/ 模块（model, dto, repo, service, handler）
- [x] Task 3: 实现服务 CRUD 和名称唯一性校验
- [x] Task 4: 在 main.go 注册路由
- [x] Task 5: 添加测试覆盖（7 个 service 层单元测试）

## Dev Notes

### Source tree components to touch
- `internal/svcdef/` (新建 - 避免与 Go service 层命名冲突)
- `migrations/000008_create_services.up/down.sql` (新建)
- `cmd/server/main.go` (修改)

### Database schema
```sql
CREATE TABLE services (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    repo_url VARCHAR(500) DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uk_services_project_name ON services(project_id, name) WHERE status != 'deleted';
```

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 创建 services 数据库迁移 (000008)，含项目内名称唯一约束
- 新建 internal/svcdef/ 模块，避免与 Go service 层命名冲突
- 5 个 API 端点：POST/GET/GET/:sid/PUT/:sid/DELETE/:sid
- 支持 repoUrl 字段关联 Git 仓库
- 权限检查允许 admin 和 project_admin 创建/删除服务
- 7 个 service 层单元测试通过

### File List
- `migrations/000008_create_services.up.sql`
- `migrations/000008_create_services.down.sql`
- `internal/svcdef/model.go`
- `internal/svcdef/dto.go`
- `internal/svcdef/repo.go`
- `internal/svcdef/service.go`
- `internal/svcdef/handler.go`
- `internal/svcdef/service_test.go`
- `cmd/server/main.go`
