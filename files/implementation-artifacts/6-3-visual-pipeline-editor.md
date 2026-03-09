# Story 6.3: 可视化流水线编排器

Status: done

## Story

As a 用户,
I want 通过可视化画布拖拽编排流水线 Stage 和 Step,
so that 我可以直观地设计 CI/CD 流程。

## Acceptance Criteria

1. **可视化画布渲染**
   - Given 流水线配置已加载
   - When 进入编排页面
   - Then 使用 @xyflow/react 渲染 Stage 和 Step 节点，Stage 为分组节点，Step 为子节点

2. **Stage 和 Step 管理**
   - Given 可视化编排器已渲染
   - When 用户添加/删除 Stage 和 Step
   - Then 节点自动通过 dagre 布局算法重新排列

3. **Step 配置面板**
   - Given 用户点击某个 Step 节点
   - When 侧边面板打开
   - Then 显示 Step 的配置项（name、type、image、command 等）

4. **保存流水线配置**
   - Given 用户编辑完流水线
   - When 点击保存
   - Then 将画布状态转换为 JSONB 配置并调用 PUT API

## Tasks / Subtasks

- [x] Task 1: 创建 pipeline 前端 API 服务
- [x] Task 2: 创建 PipelineEditor 核心组件
- [x] Task 3: 创建 StageNode / StepNode 自定义节点
- [x] Task 4: 创建 StepConfigPanel 配置面板
- [x] Task 5: 创建 PipelineEditorPage 页面
- [x] Task 6: 创建 PipelineListPage 页面
- [x] Task 7: 注册路由

### UX Enhancements (2026-03-09)

- [x] Stage/Step 上下移动功能 - 添加上移/下移按钮和键盘支持
- [x] 全局键盘快捷键 - Ctrl+S 保存, Ctrl+Z 撤销, Ctrl+Y 重做, Del 删除选中, Esc 关闭面板
- [x] 实时校验提示 - 保存按钮显示验证状态，错误时禁用并显示提示
- [x] 模板选择页面 - 创建流水线时显示模板选择页面，支持从模板创建或空白开始
- [x] 修复 Form.Item 布局问题

### Review Follow-ups (AI)
- [x] [AI-Review][CRITICAL] PipelineEditor: useNodesState/useEdgesState 只使用初始值，画布添加/删除 Stage/Step 后不刷新 — 需要添加 useEffect 同步或改用受控模式
- [x] [AI-Review][CRITICAL] PipelineEditor: handleSave 在异步 API 调用完成前就弹出 Message.success，且父组件也有 success toast 导致双重弹窗
- [x] [AI-Review][HIGH] StepConfigPanel: 缺少 env（环境变量）和 config（Step 类型特定配置如 repoUrl/branch/imageName）字段的编辑能力，模板创建的流水线核心配置不可视化编辑
- [x] [AI-Review][HIGH] PipelineEditor: useEffect onChange 依赖数组缺少 config 和 onChange，存在 stale closure 风险
- [x] [AI-Review][HIGH] PipelineEditorPage: 无未保存变更提醒，用户导航离开会丢失所有编辑工作，需添加 beforeunload + useBlocker
- [x] [AI-Review][MEDIUM] PipelineEditorPage: Form.Item 在 Form 上下文外使用，名称字段没有表单级校验反馈
- [x] [AI-Review][MEDIUM] PipelineEditorPage: description 字段在页面头部和设置面板中重复编辑，可能产生不一致
- [x] [AI-Review][LOW] StageNode/StepNode: 内联样式 hover 效果不支持键盘 :focus 状态
- [x] [AI-Review][LOW] PipelineEditor: genId 使用模块级计数器，HMR/StrictMode 下可能产生意外行为

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Debug Log References
- TypeScript: http 模块导入修复（default export → named export）
- tsc --noEmit: 通过
- vitest run: 16/16 通过

### Code Review (2026-03-05)
**Self-review** (前端 Story，组件级审查)
- @xyflow/react + dagre 布局算法正确使用
- 节点 ID 使用唯一生成器避免冲突
- StepConfigPanel 使用 Drawer + Form 规范化
- 路由使用 lazy loading 减少初始包体积
- PipelineEditorPage 区分新建/编辑两种模式
- ProjectLayout 菜单新增"流水线"入口

### Completion Notes List
- 安装 @xyflow/react + dagre + @types/dagre
- 创建 pipeline API 服务层（CRUD + 模板 + 复制）
- 创建可视化编排器核心组件（PipelineEditor + StageNode + StepNode + StepConfigPanel）
- 创建流水线列表页和编辑器页
- 路由注册（/projects/:id/pipelines, /projects/:id/pipelines/:pipelineId）
- ProjectLayout 侧边栏新增"流水线"菜单

### File List
- `web/src/services/pipeline.ts` (新建)
- `web/src/components/pipeline/PipelineEditor.tsx` (新建)
- `web/src/components/pipeline/StageNode.tsx` (新建)
- `web/src/components/pipeline/StepNode.tsx` (新建)
- `web/src/components/pipeline/StepConfigPanel.tsx` (新建)
- `web/src/pages/projects/pipelines/PipelineListPage.tsx` (新建)
- `web/src/pages/projects/pipelines/PipelineEditorPage.tsx` (新建)
- `web/src/App.tsx` (修改: 添加流水线路由)
- `web/src/pages/projects/ProjectLayout.tsx` (修改: 添加流水线菜单)
- `web/package.json` (修改: 添加 @xyflow/react + dagre 依赖)
