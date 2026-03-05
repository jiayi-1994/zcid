# Story 3.4: 项目成员与角色管理

Status: done

## Story

As a 项目管理员,
I want 将用户添加到项目并分配项目内角色,
so that 团队成员可以按角色协作。

## Acceptance Criteria

1. **添加项目成员**
   - Given 项目管理员或管理员已登录
   - When POST /api/v1/projects/:id/members 添加用户
   - Then 用户加入项目并获得指定角色

2. **修改成员角色**
   - Given 成员已存在于项目中
   - When PUT /api/v1/projects/:id/members/:uid
   - Then 成员角色更新成功

3. **移除项目成员**
   - Given 管理员或项目管理员
   - When DELETE /api/v1/projects/:id/members/:uid
   - Then 成员从项目中移除

4. **获取项目成员列表**
   - Given 项目存在
   - When GET /api/v1/projects/:id/members
   - Then 返回该项目下的成员列表（含用户名和角色）

5. **重复添加处理**
   - Given 用户已是项目成员
   - When 再次添加
   - Then 返回 409 冲突错误

## Tasks / Subtasks

- [x] Task 1: 在 internal/project/ 扩展成员管理 handler 路由
- [x] Task 2: 实现 AddMember/UpdateMemberRole/RemoveMember/ListMembers 业务逻辑
- [x] Task 3: 在 main.go 注册成员路由（嵌套在 /projects/:id/members 下）
- [x] Task 4: 添加测试覆盖（7 个成员管理单元测试）

## Dev Notes

### Source tree components to touch
- `internal/project/handler.go` (修改 - 添加成员路由)
- `internal/project/service.go` (修改 - 添加成员业务逻辑)
- `internal/project/repo.go` (已有 AddMember 等方法，需扩展)
- `internal/project/dto.go` (修改 - 添加成员 DTO)
- `cmd/server/main.go` (修改)

### 技术说明
- project_members 表已在 Story 3.1 的 migration 000006 中创建
- 复用已有的 Repo.AddMember/RemoveMembersByProject/GetUserProjectIDs 方法

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 复用 Story 3.1 已创建的 project_members 表
- 扩展 project 模块 Repository 接口，新增 ListMembers/RemoveMember/UpdateMemberRole 方法
- ListMembers 使用 JOIN users 查询关联用户名
- AddMember 返回 ErrMemberExists 用于冲突检测（CreateProject 中忽略此错误）
- 角色验证支持 project_admin 和 member 两种角色
- 7 个成员管理单元测试通过

### File List
- `internal/project/dto.go` - 添加 AddMemberRequest, UpdateMemberRoleRequest, MemberResponse, MemberListResponse, MemberWithUsername
- `internal/project/repo.go` - 添加 ListMembers, RemoveMember, UpdateMemberRole 方法；修改 AddMember 返回 ErrMemberExists
- `internal/project/service.go` - 添加 AddMember, RemoveMember, UpdateMemberRole, ListMembers 业务方法
- `internal/project/handler.go` - 添加 RegisterMemberRoutes 和 4 个成员 handler
- `internal/project/service_test.go` - 添加 7 个成员管理测试
- `cmd/server/main.go` - 注册 /projects/:id/members 路由组
