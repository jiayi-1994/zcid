# Story 6.4: JSON 模式编辑

Status: done

## Story

As a 用户,
I want 在流水线编辑器中切换可视化和 JSON 两种编辑模式,
so that 我可以选择使用可视化拖拽或直接编辑 JSON 配置。

## Acceptance Criteria

1. **YamlEditor 组件**
   - Given 流水线配置已加载
   - When 切换到 JSON 模式
   - Then 显示使用等宽字体的文本区域，支持编辑 PipelineConfig 的 JSON 格式

2. **ModeSwitch 组件**
   - Given 流水线编辑页面
   - When 用户在两种模式间切换
   - Then 可在"可视化"和"JSON 模式"之间切换

3. **configYaml 工具**
   - Given PipelineConfig 对象
   - When 调用 configToJson / jsonToConfig
   - Then 正确转换为 JSON 字符串并解析回对象（使用 JSON 作为中间格式）

4. **PipelineEditorPage 集成**
   - Given 编辑页面
   - When 切换模式
   - Then 条件渲染 PipelineEditor 或 YamlEditor，两种模式共享同一 config 状态

## Tasks / Subtasks

- [x] Task 1: 创建 configYaml.ts 工具（configToJson、jsonToConfig）
- [x] Task 2: 创建 YamlEditor.tsx（TextArea + 等宽字体 + 校验反馈）
- [x] Task 3: 创建 ModeSwitch.tsx（Radio.Group 切换）
- [x] Task 4: 更新 PipelineEditorPage 集成模式切换与条件渲染

### Review Follow-ups (AI)
- [ ] [AI-Review][MEDIUM] YamlEditor: 每次击键触发 JSON.parse + onChange 无防抖，大型配置下会导致输入延迟 [web/src/components/pipeline/YamlEditor.tsx:28-39]
- [ ] [AI-Review][MEDIUM] configYaml.ts 文件名暗示 YAML 但实际只处理 JSON，应重命名为 configJson.ts 以消除歧义 [web/src/lib/pipeline/configYaml.ts]

## Dev Agent Record

### Agent Model Used
Claude Sonnet (Executor)

### Verification
- `npx tsc --noEmit`: 通过
- `npx vitest run`: 16/16 通过

### File List
- `web/src/lib/pipeline/configYaml.ts` (新建)
- `web/src/components/pipeline/YamlEditor.tsx` (新建)
- `web/src/components/pipeline/ModeSwitch.tsx` (新建)
- `web/src/pages/projects/pipelines/PipelineEditorPage.tsx` (修改: 模式切换、YamlEditor 集成)
