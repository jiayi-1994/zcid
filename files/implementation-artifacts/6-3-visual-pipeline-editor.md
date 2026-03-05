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
