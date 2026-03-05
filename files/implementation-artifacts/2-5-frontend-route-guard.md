# Story 2.5: 前端权限路由守卫

Status: done

## Story

As a 用户,
I want 只看到我有权限的页面和操作,
so that 界面清晰且不会误操作。

## Acceptance Criteria

1. **无权限路由入口隐藏**
   - Given 用户已登录且角色为普通成员
   - When 路由渲染
   - Then 无权限的路由入口不显示（基于权限数据）

2. **无权限直达 URL 返回 403 页面**
   - Given 用户直接访问无权限 URL
   - When 路由守卫检查
   - Then 显示 403 无权限页面

3. **PermissionGate 控制操作按钮**
   - Given PermissionGate 组件包裹操作按钮
   - When 用户无该操作权限
   - Then 按钮不渲染

## Tasks / Subtasks

- [x] Task 1: 路由与守卫实现（AC: 1,2）
  - [x] 新增 `RequirePermission` 组件，基于 `hasPermission` 执行鉴权
  - [x] 新增 `/admin/users` 受限路由并在无权限时跳转 `/403`
  - [x] 新增 `/403` 页面并接入路由

- [x] Task 2: 页面入口与操作按钮权限控制（AC: 1,3）
  - [x] 在 `AppLayout` 中新增导航入口并按权限隐藏“用户管理”入口
  - [x] 在 `DashboardPage` 使用 `PermissionGate` 控制“新建用户”按钮显示

- [x] Task 3: 测试与回归（AC: 1,2,3）
  - [x] 补充路由测试：未登录跳转登录页仍生效
  - [x] 补充路由测试：普通成员直达 `/admin/users` 显示 403
  - [x] 补充 UI 测试：普通成员看不到“用户管理”入口和“新建用户”按钮
  - [x] 执行 `npm test` 与 `npm run build`

## Dev Notes

- 复用 `authStore` 中的 `PermissionKey` 与 `hasPermission` 作为前端统一权限判断入口。
- 路由层通过 `RequireAuth + RequirePermission` 组合实现“先认证、后鉴权”。
- 页面层通过 `PermissionGate` 做细粒度操作级显示控制，避免未授权操作入口暴露。

### Completion Notes List

- 已新增权限路由守卫组件并接入受限路由。
- 已新增 403 页面用于无权限访问反馈。
- 已在侧边导航隐藏普通成员无权访问的“用户管理”入口。
- 已在 Dashboard 对“新建用户”操作按钮做权限可见性控制。
- 已补充并通过路由与权限行为相关测试。

### File List

- files/implementation-artifacts/2-5-frontend-route-guard.md
- files/implementation-artifacts/sprint-status.yaml
- web/src/App.tsx
- web/src/components/common/RequirePermission.tsx
- web/src/components/layout/AppLayout.tsx
- web/src/pages/admin-users/AdminUsersPage.tsx
- web/src/pages/forbidden/ForbiddenPage.tsx
- web/src/pages/dashboard/DashboardPage.tsx
- web/src/pages/login/LoginPage.tsx
- web/src/App.auth.test.tsx
- web/src/services/http.test.ts
