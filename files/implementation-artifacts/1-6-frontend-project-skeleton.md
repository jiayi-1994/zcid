# Story 1.6: 前端项目骨架

Status: done

## Story

As a 前端开发者,
I want React+TS+Vite+Arco 前端项目骨架可运行,
so that 我可以开始开发页面。

## Acceptance Criteria (BDD)

1. **Given** 开发者进入 web/ 目录 **When** 执行 `npm install && npm run dev` **Then** Vite dev server 启动，浏览器访问 localhost:5173 显示空白布局页
2. **Given** 前端项目初始化 **When** 查看目录结构 **Then** 包含 pages/、components/common/、components/layout/、hooks/、stores/、lib/ws/、theme/、utils/、constants/、styles/ 目录 **And** 符合三层组件架构目录规范
3. **Given** Arco Design 主题配置 **When** 应用加载 **Then** 使用 Apple 风格蓝白色调 Token（primary #1677FF、bg #FFFFFF/#F7F8FA、大圆角）**And** CSS Variables 响应式断点（768/1024/1280/1440px）已定义
4. **Given** 浏览器窗口宽度 < 1280px **When** AppLayout 渲染 **Then** Sidebar 自动折叠为 icon-only 模式（64px）

## Tasks / Subtasks

- [x] Task 1: 初始化 web 前端工程 (AC: #1)
  - [x] 1.1 创建 `web/package.json`，配置 `dev/build/test/lint/codegen` scripts
  - [x] 1.2 创建 `vite.config.ts`、`tsconfig.json`、`vitest.config.ts`、`openapi-ts.config.ts`
  - [x] 1.3 创建 `index.html` 与 `src/main.tsx` 入口

- [x] Task 2: 建立前端目录骨架 (AC: #2)
  - [x] 2.1 创建 `pages/`、`components/common/`、`components/layout/`
  - [x] 2.2 创建 `hooks/`、`stores/`、`lib/ws/`
  - [x] 2.3 创建 `theme/`、`utils/`、`constants/`、`styles/`

- [x] Task 3: 提供基础空白布局页与主题 token (AC: #1, #3, #4)
  - [x] 3.1 创建 `AppLayout` 和 `BlankLayoutPage`
  - [x] 3.2 添加 Arco 主题 token（primary/bg/radius）
  - [x] 3.3 添加 CSS Variables 断点与全局样式
  - [x] 3.4 实现 `<1280px` 时 Sidebar 自动折叠逻辑（64px）

- [x] Task 4: 验证 (AC: #1, #4)
  - [x] 4.1 执行 `npm install`
  - [x] 4.2 执行 `npm run build`
  - [x] 4.3 执行 `npm run test`
  - [x] 4.4 执行 `npm run dev` 并验证 `http://localhost:5173` 启动

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `cd web && npm install` (PASS)
- `cd web && npm run build` (PASS)
- `cd web && npm run test` (PASS)
- `cd web && npm run dev` (PASS, Local: `http://localhost:5173/`)

### Completion Notes List

- 新建 `web/` React + TypeScript + Vite 前端骨架，并接入 Arco Design。
- 落地 Story 1.6 要求的目录结构，覆盖 pages/components/common/components/layout/hooks/stores/lib/ws/theme/utils/constants/styles。
- 实现空白布局页 `BlankLayoutPage` + `AppLayout`，并在窗口宽度 `<1280px` 时自动折叠侧边栏为 64px。
- 增加 Apple 风格蓝白色调主题 token 与 CSS Variables 响应式断点。
- 通过 `build/test/dev` 验证前端骨架可运行。

### Change Log

- 2026-03-02: 完成 Story 1.6，状态更新为 `review`。
- 2026-03-02: Code review (AI) — M4 fix: installed react-router-dom@7, wrapped App.tsx with BrowserRouter/Routes for Epic 2 route guard readiness. Moved to `done`.

### File List

- `web/package.json`
- `web/index.html`
- `web/tsconfig.json`
- `web/tsconfig.node.json`
- `web/vite.config.ts`
- `web/vitest.config.ts`
- `web/openapi-ts.config.ts`
- `web/src/main.tsx`
- `web/src/App.tsx`
- `web/src/pages/BlankLayoutPage.tsx`
- `web/src/components/layout/AppLayout.tsx`
- `web/src/components/common/Placeholder.tsx`
- `web/src/hooks/useSidebarCollapsed.ts`
- `web/src/hooks/useSidebarCollapsed.test.tsx`
- `web/src/stores/ui.ts`
- `web/src/lib/ws/manager.ts`
- `web/src/theme/tokens.ts`
- `web/src/constants/breakpoints.ts`
- `web/src/utils/index.ts`
- `web/src/styles/global.css`
- `web/src/test/setup.ts`
- `files/implementation-artifacts/1-6-frontend-project-skeleton.md`
- `files/implementation-artifacts/sprint-status.yaml`
- `files/implementation-artifacts/project-skills.md`
