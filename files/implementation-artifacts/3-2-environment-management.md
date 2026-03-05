# Story 3.2: 环境管理与 Namespace 映射

Status: done

## Story

As a 项目管理员,
I want 在项目内创建环境并映射到 K8s Namespace,
so that 不同环境（dev/staging/prod）隔离部署。

## Acceptance Criteria

1. **创建环境**
   - Given 项目管理员或管理员已登录
   - When POST /api/v1/projects/:id/environments
   - Then 创建环境并关联 K8s Namespace

2. **Namespace 唯一性**
   - Given Namespace 已被其他环境占用
   - When 尝试映射
   - Then 返回 40302 Namespace 已占用错误

3. **编辑环境**
   - Given 环境存在
   - When PUT /api/v1/projects/:id/environments/:eid
   - Then 环境信息更新成功

4. **获取环境列表**
   - Given 项目存在
   - When GET /api/v1/projects/:id/environments
   - Then 返回该项目下的环境列表

5. **删除环境**
   - Given 管理员或项目管理员
   - When DELETE /api/v1/projects/:id/environments/:eid
   - Then 环境标记为删除（软删除）

## Tasks / Subtasks

- [x] Task 1: 创建 environments 数据库迁移
- [x] Task 2: 创建 internal/environment/ 模块（model, dto, repo, service, handler）
- [x] Task 3: 实现环境 CRUD 和 Namespace 唯一性校验
- [x] Task 4: 在 main.go 注册路由（嵌套在 /projects/:id/environments 下）
- [x] Task 5: 添加测试覆盖（10 个 service 层单元测试）

## Dev Notes

### Source tree components to touch
- `internal/environment/` (新建)
- `migrations/000007_create_environments.up/down.sql` (新建)
- `cmd/server/main.go` (修改)

### Database schema
```sql
CREATE TABLE environments (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    name VARCHAR(100) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX uk_environments_namespace ON environments(namespace) WHERE status != 'deleted';
CREATE UNIQUE INDEX uk_environments_project_name ON environments(project_id, name) WHERE status != 'deleted';
```

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 创建 environments 数据库迁移 (000007)，含 namespace 全局唯一约束和项目内名称唯一约束
- 新建 internal/environment/ 模块，完整实现 handler->service->repo 三层
- 5 个 API 端点：POST/GET/GET/:eid/PUT/:eid/DELETE/:eid
- 权限检查允许 admin 和 project_admin 创建/删除环境
- 10 个 service 层单元测试通过

### File List
- `migrations/000007_create_environments.up.sql`
- `migrations/000007_create_environments.down.sql`
- `internal/environment/model.go`
- `internal/environment/dto.go`
- `internal/environment/repo.go`
- `internal/environment/service.go`
- `internal/environment/handler.go`
- `internal/environment/service_test.go`
- `cmd/server/main.go`
