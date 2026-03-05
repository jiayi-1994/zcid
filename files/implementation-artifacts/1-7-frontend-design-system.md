# Story 1.7: 前端设计系统基础

Status: done

## Story

As a 前端开发者,
I want 构建可复用的前端设计系统基础组件,
so that 后续页面开发可以保持一致性与可维护性。

## Acceptance Criteria (BDD)

1. **Given** 状态展示组件需要统一视觉 **When** 引用 `STATUS_MAP` **Then** 提供 success/running/failed/warning/pending/cancelled/timeout 七种状态的统一 color/bg/icon/label 映射
2. **Given** 页面渲染发生异常 **When** 触发 React 错误边界 **Then** 显示降级 UI 并提供“重试”操作
3. **Given** 页面数据仍在加载 **When** 渲染骨架屏 **Then** `PageSkeleton` 提供可复用占位布局
4. **Given** 页面存在权限控制区域 **When** 用户无权限 **Then** `PermissionGate` 不渲染受保护内容

## Tasks / Subtasks

- [x] Task 1: 建立全局状态样式字典 (AC: #1)
  - [x] 1.1 创建 `web/src/constants/statusMap.ts`
  - [x] 1.2 定义七种状态的 `color/bg/icon/label` 映射
  - [x] 1.3 添加 `statusMap` 单元测试覆盖字段完整性

- [x] Task 2: 实现错误边界组件 (AC: #2)
  - [x] 2.1 创建 `web/src/components/common/ErrorBoundary.tsx`
  - [x] 2.2 提供降级 UI（alert + 重试按钮）
  - [x] 2.3 在 `web/src/App.tsx` 挂载 `ErrorBoundary`
  - [x] 2.4 添加 `ErrorBoundary` 单元测试覆盖“出错 -> 重试 -> 恢复”流程

- [x] Task 3: 实现基础通用组件 (AC: #3, #4)
  - [x] 3.1 创建 `PageSkeleton` 骨架屏组件
  - [x] 3.2 创建 `PermissionGate` 权限门禁组件
  - [x] 3.3 分别添加对应单元测试

- [x] Task 4: 验证 (AC: #1, #2, #3, #4)
  - [x] 4.1 执行 `npm run test`
  - [x] 4.2 执行 `npm run build`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `npm --prefix "/f/other code/zcid/web" run test` (PASS)
- `npm --prefix "/f/other code/zcid/web" run build` (PASS)

### Completion Notes List

- 新增 `STATUS_MAP` 作为全局状态样式单一来源，覆盖七类运行状态。
- 新增 `ErrorBoundary` 并在应用根节点接入，支持异常降级与重试恢复。
- 新增 `PageSkeleton` 与 `PermissionGate` 两个可复用通用组件。
- 为 `statusMap`、`ErrorBoundary`、`PageSkeleton`、`PermissionGate` 增加单测并全部通过。

### Change Log

- 2026-03-02: 完成 Story 1.7，状态更新为 `review`。
- 2026-03-02: Code review (AI) — no issues found, moved to `done`.

### File List

- `web/src/App.tsx`
- `web/src/constants/statusMap.ts`
- `web/src/components/common/ErrorBoundary.tsx`
- `web/src/components/common/PageSkeleton.tsx`
- `web/src/components/common/PermissionGate.tsx`
- `web/src/components/common/statusMap.test.ts`
- `web/src/components/common/ErrorBoundary.test.tsx`
- `web/src/components/common/PageSkeleton.test.tsx`
- `web/src/components/common/PermissionGate.test.tsx`
- `files/implementation-artifacts/1-7-frontend-design-system.md`
- `files/implementation-artifacts/sprint-status.yaml`
- `files/implementation-artifacts/project-skills.md`
