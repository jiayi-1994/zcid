# Story 6.6: 流水线列表与运行时参数

Status: done

## Story

As a 用户,
I want 在流水线列表中按状态筛选、并可通过"运行"按钮触发流水线,
so that 我可以快速找到目标流水线并启动运行（执行功能在 Epic 7 实现）。

## Acceptance Criteria

1. **状态筛选**
   - Given 流水线列表页
   - When 选择状态筛选（draft/active/disabled）
   - Then 列表显示符合该状态的流水线（前端筛选）

2. **RunPipelineModal 组件**
   - Given 用户点击"运行"按钮
   - When 打开 RunPipelineModal
   - Then 显示流水线名称和"功能开发中"占位文案（Epic 7 实现实际执行）

3. **运行按钮**
   - Given 流水线列表
   - When 每行操作列
   - Then 显示"运行"按钮，点击后打开 RunPipelineModal

## Tasks / Subtasks

- [x] Task 1: 在 PipelineListPage 添加状态筛选（Select）
- [x] Task 2: 创建 RunPipelineModal.tsx（占位文案）
- [x] Task 3: 在列表操作列添加"运行"按钮

## Dev Agent Record

### Agent Model Used
Claude Sonnet (Executor)

### Verification
- `npx tsc --noEmit`: 通过
- `npx vitest run`: 16/16 通过

### Notes
- 状态筛选为前端 client-side 筛选（当前页数据），后端 List API 暂不支持 status 参数
- RunPipelineModal 为占位实现，实际执行逻辑在 Epic 7

### File List
- `web/src/components/pipeline/RunPipelineModal.tsx` (新建)
- `web/src/pages/projects/pipelines/PipelineListPage.tsx` (修改: 状态筛选、运行按钮、RunPipelineModal)
