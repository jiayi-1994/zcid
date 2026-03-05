# Story 5.4: 前端 Git 集成管理页面

Status: done

## Story

As a 管理员,
I want 在界面上配置和查看 Git 仓库连接,
so that 可以直观管理集成状态。

## Acceptance Criteria

1. **集成管理页面**
   - Given 管理员访问集成页面
   - When 访问 /admin/integrations
   - Then 显示已配置的 Git 连接列表及状态

2. **添加 Git 连接**
   - Given 管理员点击"添加连接"
   - When 填写名称、Provider类型、Server URL、Access Token
   - Then 连接添加成功，列表刷新显示新连接

3. **连接状态显示**
   - Given 连接状态异常
   - When 页面渲染
   - Then StatusBadge 显示对应状态色（绿=已连接, 红=已断开, 黄=Token过期）

4. **测试连接**
   - Given 管理员点击"测试连接"
   - When 调用 test API
   - Then 显示测试结果（成功/失败原因）

5. **Webhook Secret 复制**
   - Given 管理员查看连接详情
   - When 点击"复制 Webhook Secret"
   - Then Secret 复制到剪贴板，显示复制成功提示

6. **权限控制**
   - Given 非管理员用户
   - When 尝试访问 /admin/integrations
   - Then 路由守卫拦截，重定向到 403 页面

## Tasks / Subtasks

- [x] Task 1: 创建 services/integration.ts API 服务层 (AC: #1-#5)
- [x] Task 2: 创建集成管理页面 IntegrationsPage.tsx (AC: #1, #3)
- [x] Task 3: 创建添加/编辑连接 Modal (AC: #2)
- [x] Task 4: 实现测试连接和 Webhook Secret 复制功能 (AC: #4, #5)
- [x] Task 5: 在 App.tsx 注册路由，添加权限守卫 (AC: #6)
- [x] Task 6: 在 AppLayout 侧边栏添加"集成管理"导航 (AC: #1)
- [x] Task 7: 权限扩展 stores/auth.ts (AC: #6)

## Dev Notes

### 代码模式参考
- 页面结构参考 `pages/admin/variables/AdminVariablePage.tsx`
- API 服务层参考 `services/variable.ts`
- Modal 参考 `pages/projects/variables/VariableFormModal.tsx`
- StatusBadge 使用 STATUS_MAP 中的状态颜色
- 路由注册参考 `/admin/variables` 模式

### References
- [Source: files/planning-artifacts/ux-design-specification.md#集成状态 Dashboard] — 集成管理 wireframe
- [Source: files/planning-artifacts/epics.md#Story 5.4] — 前端 Git 集成管理页面 AC
- [Source: web/src/pages/admin/variables/AdminVariablePage.tsx] — 管理员页面参考

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Code Review (2026-03-05)
- [M5] 编辑模式 description 空字符串处理统一使用 ?? undefined
- [M9] Table 组件启用客户端分页（pageSize=20）
- [M10] 提取共享 ApiResponse 类型到 services/types.ts，消除跨文件重复定义

### Debug Log References

### Completion Notes List
- 创建 services/integration.ts API 层，封装 CRUD + 测试连接 + 获取 Webhook Secret
- IntegrationsPage 实现连接列表、状态 Badge（绿=已连接、红=已断开、黄=Token过期）
- ConnectionFormModal 支持创建（填写 Provider/URL/Token）和编辑模式
- 测试连接按钮带 loading 状态，显示成功/失败结果
- Webhook Secret 一键复制到剪贴板（navigator.clipboard API）
- App.tsx 注册 /admin/integrations 路由，RequirePermission 守卫
- AppLayout 侧边栏添加"集成管理"导航入口（IconLink 图标），仅 admin 可见
- stores/auth.ts 扩展 route:admin-integrations:view 权限
- 全部 16 个前端测试通过，零回归

### File List
- `web/src/services/integration.ts` (新建)
- `web/src/pages/admin/integrations/IntegrationsPage.tsx` (新建)
- `web/src/pages/admin/integrations/ConnectionFormModal.tsx` (新建)
- `web/src/App.tsx` (修改 — 添加 /admin/integrations 路由)
- `web/src/components/layout/AppLayout.tsx` (修改 — 添加侧边栏集成管理入口)
- `web/src/stores/auth.ts` (修改 — 添加 route:admin-integrations:view 权限)
