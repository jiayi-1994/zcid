# Story 2.4: 前端登录页与认证状态管理

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a 用户,
I want 一个简洁的登录页面和自动 Token 管理,
so that 我可以方便地登录和保持会话。

## Acceptance Criteria

1. **未登录自动跳转登录页**
   - Given 用户未登录
   - When 访问任意页面
   - Then 自动跳转到 `/login` 页面

2. **登录成功后写入认证状态并跳转**
   - Given 用户在登录页输入正确凭证
   - When 点击登录按钮
   - Then 按钮进入 loading 状态，登录成功后跳转到 `/dashboard`
   - And Token 存储到 `authStore`（Zustand）

3. **401 自动刷新与重试**
   - Given Access Token 过期
   - When API 返回 401
   - Then Axios 拦截器自动使用 Refresh Token 刷新
   - And 刷新成功后重试原请求，刷新失败跳转登录页

## Tasks / Subtasks

- [x] Task 1: 路由与页面骨架（AC: 1,2）
  - [x] 在 `web/src/App.tsx` 增加 `/login`、`/dashboard` 路由
  - [x] 增加受保护路由包装：未登录时跳转 `/login`
  - [x] 新建登录页组件与 Dashboard 占位页面

- [x] Task 2: 认证状态管理（AC: 2）
  - [x] 新增 `authStore`（Zustand），保存 `accessToken` / `refreshToken` / `user`
  - [x] 提供 `setSession`、`clearSession`、`isAuthenticated` 等能力
  - [x] 登录成功后写入 store 并执行页面跳转

- [x] Task 3: API 客户端与拦截器（AC: 3）
  - [x] 新增 Axios 实例，统一配置 baseURL
  - [x] 请求拦截器注入 `Authorization: Bearer <accessToken>`
  - [x] 响应拦截器处理 401：调用 refresh、刷新成功后重试原请求
  - [x] 刷新失败时清理会话并跳转 `/login`

- [x] Task 4: 测试与回归（AC: 1,2,3）
  - [x] 登录页交互测试：输入、loading、成功后跳转
  - [x] 受保护路由测试：未登录重定向至 `/login`
  - [x] 401 刷新流程测试：刷新成功重试 / 刷新失败登出
  - [x] 运行 `npm test` 与 `npm run build`

## Dev Notes

- 前端状态分层遵循架构约束：服务端数据用 TanStack Query，客户端 UI/会话状态用 Zustand；本 Story 仅实现认证会话状态，不在 store 中缓存业务 API 数据。
- 前端 401 处理遵循统一错误策略：通过拦截器集中处理，避免在每个页面重复编写刷新逻辑。
- Token 持久化采用浏览器存储；登出或刷新失败时必须清理。
- 登录流程仅覆盖账号密码登录与会话续期，不扩展到权限路由细粒度控制（留给 Story 2.5）。

### Project Structure Notes

- 与架构目录约定对齐：
  - 页面：`web/src/pages/login/`、`web/src/pages/dashboard/`
  - 状态：`web/src/stores/auth.ts`
  - 网络层：`web/src/services/http.ts`、`web/src/services/auth.ts`
  - 路由守卫：`web/src/components/common/RequireAuth.tsx`
- 当前代码库尚未接入真实业务页面，本 Story 将先以 Dashboard 占位页验证登录闭环。

### References

- Story 2.4 原始定义与 AC：[Source: files/planning-artifacts/epics.md#Story 2.4: 前端登录页与认证状态管理]
- 前端状态管理分层与 401 处理规范：[Source: files/planning-artifacts/architecture.md#前端状态管理分层] [Source: files/planning-artifacts/architecture.md#前端错误处理]
- JWT 双 Token 约束（Access/Refresh TTL）：[Source: files/planning-artifacts/architecture.md#Authentication & Security]
- 认证安全 NFR（Token 刷新机制）：[Source: files/planning-artifacts/prd.md#NFR9: 认证安全]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- create-story workflow
- dev-story workflow (in progress)

### Completion Notes List

- 已完成前端登录闭环：路由守卫、登录页、Dashboard 占位页、认证状态持久化
- 已完成 Axios 401 自动刷新与重试逻辑，刷新失败时清会话并跳转 `/login`
- 已补充 401 刷新流程自动化测试（成功重试 / 失败登出）
- 已执行并通过 `npm test` 与 `npm run build`

### File List

- files/implementation-artifacts/2-4-frontend-login.md
- files/implementation-artifacts/sprint-status.yaml
- web/package.json
- web/src/App.tsx
- web/src/components/common/RequireAuth.tsx
- web/src/components/layout/AppLayout.tsx
- web/src/pages/login/LoginPage.tsx
- web/src/pages/dashboard/DashboardPage.tsx
- web/src/services/auth.ts
- web/src/services/http.ts
- web/src/services/http.test.ts
- web/src/stores/auth.ts
- web/src/App.auth.test.tsx
