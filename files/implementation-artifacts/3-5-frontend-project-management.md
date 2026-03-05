# Story 3.5: 前端项目管理页面

Status: done

## Story

As a 用户,
I want 在界面上管理项目、环境、服务和成员,
so that 不需要调用 API 即可完成管理操作。

## Acceptance Criteria

1. **项目列表页**
   - Given 用户已登录
   - When 访问 /projects
   - Then 显示有权限的项目列表

2. **创建项目**
   - Given 管理员已登录
   - When 点击"新建项目"
   - Then 弹出表单填写名称/描述，创建成功后刷新列表

3. **项目详情布局**
   - Given 用户进入项目
   - When 访问 /projects/:id
   - Then 显示项目级布局（侧边栏 + Outlet）
   - And 侧边栏包含环境、服务、成员导航

4. **环境/服务/成员管理子页面**
   - Given 用户进入项目内
   - When 访问各子路由
   - Then 显示对应的列表和 CRUD 操作

5. **项目隔离**
   - Given 不同项目
   - When 用户在项目 A 中操作
   - Then 看不到项目 B 的任何资源（FR10）

## Tasks / Subtasks

- [x] Task 1: 创建项目列表页面 ProjectListPage
- [x] Task 2: 创建项目表单 Modal（ProjectFormModal）
- [x] Task 3: 创建项目级布局 ProjectLayout（侧边栏 + Outlet）
- [x] Task 4: 创建环境管理子页面（EnvironmentListPage）
- [x] Task 5: 创建服务管理子页面（ServiceListPage）
- [x] Task 6: 创建成员管理子页面（MemberListPage）
- [x] Task 7: 在 App.tsx 注册路由（/projects, /projects/:id/*）
- [x] Task 8: 在 AppLayout 侧边栏添加"项目管理"导航入口
- [x] Task 9: 权限扩展（route:projects:view）

## Dev Notes

### Source tree components to touch
- `web/src/pages/projects/ProjectListPage.tsx` (新建)
- `web/src/pages/projects/ProjectFormModal.tsx` (新建)
- `web/src/pages/projects/ProjectLayout.tsx` (新建)
- `web/src/pages/projects/environments/EnvironmentListPage.tsx` (新建)
- `web/src/pages/projects/services/ServiceListPage.tsx` (新建)
- `web/src/pages/projects/members/MemberListPage.tsx` (新建)
- `web/src/App.tsx` (修改)
- `web/src/components/layout/AppLayout.tsx` (修改)
- `web/src/stores/auth.ts` (修改 - 添加项目相关权限)

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 新建 services/project.ts API 层，封装所有项目/环境/服务/成员 API
- ProjectListPage 支持分页、新建（admin）、删除（admin）、点击进入项目
- ProjectLayout 提供项目级侧边栏导航（环境/服务/成员）
- 环境/服务/成员子页面均支持列表查看和 CRUD 操作
- 成员角色可内联下拉修改
- AppLayout 侧边栏添加"项目管理"入口，对所有角色可见
- stores/auth.ts 扩展 route:projects:view 权限
- 全部 16 个前端测试通过，无回归

### File List
- `web/src/services/project.ts` - API 服务层
- `web/src/pages/projects/ProjectListPage.tsx` - 项目列表页
- `web/src/pages/projects/ProjectFormModal.tsx` - 项目创建弹窗
- `web/src/pages/projects/ProjectLayout.tsx` - 项目级布局
- `web/src/pages/projects/environments/EnvironmentListPage.tsx` - 环境管理页
- `web/src/pages/projects/services/ServiceListPage.tsx` - 服务管理页
- `web/src/pages/projects/members/MemberListPage.tsx` - 成员管理页
- `web/src/App.tsx` - 路由注册
- `web/src/components/layout/AppLayout.tsx` - 侧边栏入口
- `web/src/stores/auth.ts` - 权限扩展
