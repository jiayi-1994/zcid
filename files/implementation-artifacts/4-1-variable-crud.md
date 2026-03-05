# Story 4.1: 多层级变量 CRUD

Status: done

## Story

As a 项目管理员,
I want 在全局和项目两个层级创建和管理变量,
so that 变量可以按层级覆盖，减少重复配置。

## Acceptance Criteria

1. **创建全局变量**
   - Given 管理员已登录
   - When POST /api/v1/admin/variables 提交变量名和值
   - Then 全局变量创建成功

2. **创建项目变量**
   - Given 项目管理员或管理员已登录
   - When POST /api/v1/projects/:id/variables 提交变量名和值
   - Then 项目级变量创建成功

3. **变量名作用域内唯一**
   - Given 同作用域下变量名已存在
   - When 创建同名变量
   - Then 返回 40501 变量名重复错误

4. **获取变量列表**
   - Given 项目存在
   - When GET /api/v1/projects/:id/variables
   - Then 返回该项目下的变量列表（不含密钥值）

5. **编辑变量**
   - Given 变量存在
   - When PUT /api/v1/projects/:id/variables/:vid 或全局路由
   - Then 变量更新成功

6. **删除变量**
   - Given 管理员或项目管理员
   - When DELETE /api/v1/projects/:id/variables/:vid
   - Then 变量标记为删除（软删除）

7. **变量合并查询**
   - Given 项目存在
   - When GET /api/v1/projects/:id/variables/merged
   - Then 返回全局+项目级合并后的变量列表（项目级覆盖全局级）

## Tasks / Subtasks

- [x] Task 1: 创建 variables 数据库迁移 (000009)
- [x] Task 2: 创建 internal/variable/ 模块 (model, dto, repo, service, handler)
- [x] Task 3: 实现变量 CRUD 和名称唯一性校验
- [x] Task 4: 实现变量合并查询逻辑
- [x] Task 5: 在 main.go 注册路由（项目级 + 全局级）
- [x] Task 6: 添加测试覆盖 (11 个 service 层单元测试)

## Dev Notes

### Source tree components to touch
- `internal/variable/` (新建)
- `migrations/000009_create_variables.up/down.sql` (新建)
- `cmd/server/main.go` (修改)
- `pkg/response/codes.go` (修改 - 添加变量相关错误码)

### Database schema
```sql
CREATE TABLE variables (
    id VARCHAR(255) PRIMARY KEY,
    scope VARCHAR(20) NOT NULL DEFAULT 'project',
    project_id VARCHAR(255) REFERENCES projects(id),
    key VARCHAR(200) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    var_type VARCHAR(20) NOT NULL DEFAULT 'plain',
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_variables_global_key ON variables(key) WHERE scope = 'global' AND status != 'deleted';
CREATE UNIQUE INDEX uk_variables_project_key ON variables(project_id, key) WHERE scope = 'project' AND status != 'deleted';
```

### API Endpoints
- `POST /api/v1/admin/variables` - 创建全局变量 (admin only)
- `GET /api/v1/admin/variables` - 获取全局变量列表 (admin only)
- `PUT /api/v1/admin/variables/:vid` - 更新全局变量 (admin only)
- `DELETE /api/v1/admin/variables/:vid` - 删除全局变量 (admin only)
- `POST /api/v1/projects/:id/variables` - 创建项目变量
- `GET /api/v1/projects/:id/variables` - 获取项目变量列表
- `GET /api/v1/projects/:id/variables/merged` - 合并变量列表
- `PUT /api/v1/projects/:id/variables/:vid` - 更新项目变量
- `DELETE /api/v1/projects/:id/variables/:vid` - 删除项目变量

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 创建 variables 数据库迁移 (000009)，含全局和项目级两个 partial unique index
- 新建 internal/variable/ 模块，完整实现 handler->service->repo 三层
- 9 个 API 端点：全局 CRUD (admin) + 项目 CRUD + 合并查询
- 变量合并逻辑：项目级覆盖全局级
- 权限检查：全局变量仅 admin，项目变量允许 admin 和 project_admin
- 11 个 service 层单元测试通过

### File List
- `migrations/000009_create_variables.up.sql`
- `migrations/000009_create_variables.down.sql`
- `internal/variable/model.go`
- `internal/variable/dto.go`
- `internal/variable/repo.go`
- `internal/variable/service.go`
- `internal/variable/handler.go`
- `internal/variable/service_test.go`
- `cmd/server/main.go`
- `pkg/response/codes.go`
