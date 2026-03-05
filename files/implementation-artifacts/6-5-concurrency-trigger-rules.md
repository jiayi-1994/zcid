# Story 6.5: 并发控制与触发规则配置

Status: done

## Story

As a 用户,
I want 在流水线编辑页配置触发方式和并发策略,
so that 流水线可以按手动/Webhook/定时触发，并控制并发执行行为。

## Acceptance Criteria

1. **PipelineSettingsPanel 组件**
   - Given 流水线已加载
   - When 打开设置面板
   - Then 显示表单字段：triggerType（Select）、concurrencyPolicy（Select）、description（TextArea）

2. **triggerType 选项**
   - manual（手动）、webhook（Webhook）、scheduled（定时）

3. **concurrencyPolicy 选项**
   - queue（排队）、cancel_old（取消旧任务）、reject（拒绝）

4. **Settings 按钮与 Drawer**
   - Given 流水线编辑页面（非新建）
   - When 点击 header 中的"设置"按钮
   - Then 以 Drawer 形式打开 PipelineSettingsPanel
   - When 保存设置
   - Then 调用 updatePipeline API 更新 triggerType、concurrencyPolicy、description

## Tasks / Subtasks

- [x] Task 1: 创建 PipelineSettingsPanel.tsx（Form + Drawer）
- [x] Task 2: 在 PipelineEditorPage header 添加 Settings 按钮
- [x] Task 3: 集成 Drawer 与 updatePipeline 调用

## Dev Agent Record

### Agent Model Used
Claude Sonnet (Executor)

### Verification
- `npx tsc --noEmit`: 通过
- `npx vitest run`: 16/16 通过

### File List
- `web/src/components/pipeline/PipelineSettingsPanel.tsx` (新建)
- `web/src/pages/projects/pipelines/PipelineEditorPage.tsx` (修改: Settings 按钮、Drawer 集成)
