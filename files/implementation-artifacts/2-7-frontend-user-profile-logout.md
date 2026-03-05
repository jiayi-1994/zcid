# Story 2.7: 前端顶部个人信息与退出登录

Status: done

## Story

As a 已登录用户,
I want 在页面右上角看到我的个人信息并可一键退出登录,
so that 我可以确认当前登录身份并安全结束会话。

## Acceptance Criteria

1. **顶部个人信息展示**
   - Given 用户已登录并进入受保护页面（如 /dashboard）
   - When 页面顶部 Header 渲染
   - Then 右上角显示当前用户信息入口（至少包含用户名与角色）
   - And 用户信息来自登录态（authStore）而不是硬编码

2. **用户菜单与退出登录入口**
   - Given 用户点击右上角个人信息入口
   - When 展开下拉菜单
   - Then 菜单中包含“退出登录”操作
   - And 菜单交互符合当前 Arco Design 组件风格

3. **退出登录行为一致性**
   - Given 用户点击“退出登录”
   - When 退出流程执行
   - Then 调用后端登出接口 `POST /api/v1/auth/logout`（若接口失败也不阻塞本地退出）
   - And 清理本地会话（accessToken、refreshToken、user、permissions）
   - And 跳转到 `/login`

4. **登出后访问受保护页面受限**
   - Given 用户已执行退出登录
   - When 用户访问受保护路由
   - Then 路由守卫判定为未认证并跳转登录页

5. **权限与可见性不回归**
   - Given 不同角色用户（admin / project_admin / member）
   - When 登录并进入应用
   - Then 侧边栏权限控制逻辑保持不变
   - And 新增顶部用户信息与退出入口对所有已登录角色可用

## Tasks / Subtasks

- [x] Task 1: 扩展 AppLayout 顶部区域并展示当前用户信息（AC: 1）
  - [x] 在 `web/src/components/layout/AppLayout.tsx` 的 Header 右侧增加用户信息区域
  - [x] 从 `useAuthStore` 读取 `user.username` 与 `user.role`
  - [x] 角色显示使用统一中文映射（管理员/项目管理员/普通成员）

- [x] Task 2: 实现个人信息下拉菜单（AC: 2）
  - [x] 使用 Arco Dropdown/Trigger + Menu 实现用户菜单
  - [x] 菜单项包含“退出登录”
  - [x] 保持与现有布局样式一致（不破坏响应式与侧边栏行为）

- [x] Task 3: 实现退出登录流程（AC: 3, 4）
  - [x] 在 `web/src/services/auth.ts` 增加/复用 `logout` API 调用
  - [x] 触发退出时执行：请求登出接口 → `clearSession()` → `navigate('/login', { replace: true })`
  - [x] 接口失败场景下仍执行本地清理与跳转（保证用户可退出）

- [x] Task 4: 回归权限和路由守卫（AC: 4, 5）
  - [x] 验证 `RequireAuth` 对未登录状态的跳转行为不变
  - [x] 验证 admin 与非 admin 菜单可见性不回归

- [x] Task 5: 增加测试覆盖（AC: 1, 2, 3, 4, 5）
  - [x] 为 AppLayout 新增/更新测试：展示用户名、角色、点击退出登录
  - [x] 为认证流程测试补充：退出后受保护路由不可访问
  - [x] 运行现有认证相关测试并确保通过

## Dev Notes

### Relevant architecture patterns and constraints

- 认证状态由 Zustand `useAuthStore` 统一管理，包含 `accessToken` / `refreshToken` / `user` / `permissions`，退出登录应通过 `clearSession()` 清理。[Source: web/src/stores/auth.ts#AuthState]
- HTTP 层已有 401 刷新与失效跳转逻辑；主动退出应与该机制兼容，不引入冲突状态。[Source: web/src/services/http.ts#interceptors]
- 受保护页面统一通过 `AppLayout` 承载，顶部用户信息和退出入口应放在该布局 Header，避免页面级重复实现。[Source: web/src/components/layout/AppLayout.tsx#AppLayout]
- Epic 2 已完成登录、权限守卫、用户管理能力，本故事只补齐“当前身份可见 + 主动退出入口”的 UX 缺口，不新增鉴权模型。[Source: files/planning-artifacts/epics.md#Epic 2]

### Source tree components to touch

- `web/src/components/layout/AppLayout.tsx`（Header 右上角用户信息 + 下拉菜单 + 退出触发）
- `web/src/services/auth.ts`（登出 API）
- `web/src/stores/auth.ts`（复用 clearSession，不改动会话结构）
- `web/src/App.auth.test.tsx` 或相关测试文件（补充退出/展示/权限回归测试）

### Testing standards summary

- 优先覆盖用户可见行为：
  - 已登录时显示用户名/角色
  - 点击菜单可触发退出
  - 退出后跳转登录页且受保护路由不可访问
- 验证回归面：
  - admin 仍可见“用户管理”菜单
  - member/project_admin 不应新增越权导航

### Project Structure Notes

- 维持现有 layout + pages 分层：公共导航能力放在 `components/layout`，业务页面不重复实现。
- 维持现有 auth store 单一事实源，不新增并行会话状态容器。

### References

- Epic 2 认证与权限范围：[Source: files/planning-artifacts/epics.md#Epic 2: 用户认证与权限管理]
- Story 2.4 登录与认证状态：[Source: files/implementation-artifacts/2-4-frontend-login.md]
- Story 2.5 路由守卫：[Source: files/implementation-artifacts/2-5-frontend-route-guard.md]
- Story 2.6 用户管理页（现有布局集成）：[Source: files/implementation-artifacts/2-6-frontend-user-management.md]
- 当前布局入口与 Header：`web/src/components/layout/AppLayout.tsx:14`
- 当前认证状态结构：`web/src/stores/auth.ts:34`
- 当前 token 刷新与失效处理：`web/src/services/http.ts:58`
- 当前登录页跳转行为：`web/src/pages/login/LoginPage.tsx:15`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- create-story workflow continuation in compacted session

### Completion Notes List

- 在 `AppLayout` 顶部右侧新增用户信息入口，展示 `username` 与角色中文映射（管理员/项目管理员/普通成员）
- 新增下拉菜单“退出登录”，退出时调用 `logout(refreshToken)`；接口失败时仍执行 `clearSession()` 并跳转 `/login`
- 保持原有侧边栏权限控制逻辑不变（admin 可见用户管理，member/project_admin 不越权）
- 已补充并通过认证回归测试：`npm --prefix "F:\other code\zcid\web" test -- src/App.auth.test.tsx`（6 passed）

### File List

- web/src/components/layout/AppLayout.tsx
- web/src/styles/global.css
- web/src/App.auth.test.tsx
- files/implementation-artifacts/2-7-frontend-user-profile-logout.md
- files/implementation-artifacts/sprint-status.yaml
