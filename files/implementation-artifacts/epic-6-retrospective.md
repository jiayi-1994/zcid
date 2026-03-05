# Epic 6 Retrospective: 流水线可视化编排

## 完成日期
2026-03-05

## Epic 概览
Epic 6 实现了流水线的完整管理功能，涵盖后端 CRUD、模板系统、可视化编排器、JSON 编辑模式、并发/触发配置以及列表增强。

## Stories 完成情况

| Story | 名称 | 状态 | 后端测试 | 前端测试 | CR 问题 |
|-------|------|------|---------|---------|---------|
| 6.1 | 流水线 CRUD 与 JSONB 存储 | Done | 15 pass | - | 1C/2H/5M/4L, all fixed |
| 6.2 | 模板一键创建流水线 | Done | 24 pass | - | 2C/3H/5M, all fixed |
| 6.3 | 可视化流水线编排器 | Done | - | 16 pass | self-review |
| 6.4 | JSON 模式编辑 | Done | - | 16 pass | batch CR with 6.5/6.6 |
| 6.5 | 并发控制与触发规则配置 | Done | - | 16 pass | batch CR with 6.4/6.6 |
| 6.6 | 流水线列表与运行时参数 | Done | - | 16 pass | batch CR with 6.4/6.5 |

## 关键技术决策

### 1. JSONB 存储流水线配置
- **决策**: 使用 PostgreSQL JSONB 存储 `PipelineConfig`，而非拆分成多个关系表
- **原因**: 流水线配置结构灵活且频繁整体读写，JSONB 更适合这种文档型数据
- **效果**: 简化了 CRUD 操作，避免了复杂的多表 JOIN

### 2. 自定义 Scanner/Valuer 接口
- **决策**: 为 `PipelineConfig` 实现 `driver.Valuer` 和 `sql.Scanner`
- **原因**: 实现 Go 结构体与 JSONB 之间的无缝序列化
- **学习**: 需同时处理 `[]byte` 和 `string` 类型输入

### 3. 内存模板注册表
- **决策**: 使用 `TemplateRegistry` + 硬编码模板，而非数据库存储
- **原因**: MVP 阶段模板数量少且固定，内存方案简单高效
- **改进**: 后续可扩展为数据库存储以支持用户自定义模板

### 4. @xyflow/react + dagre 可视化
- **决策**: 采用 ReactFlow v12 + dagre 自动布局
- **原因**: 成熟的流程图库，支持自定义节点类型和自动布局
- **效果**: 实现了 Stage/Step 拖拽编辑的完整体验

### 5. JSON 模式而非 YAML
- **决策**: 使用 JSON.stringify 美化输出代替 YAML 编辑
- **原因**: 避免引入额外的 YAML 库，JSON 对开发者同样友好
- **效果**: 零额外依赖，功能等价

## Code Review 发现的关键问题

### CRITICAL 级别 (已全部修复)
1. **项目作用域未强制执行** (6.1): GetPipeline/UpdatePipeline/DeletePipeline 未验证 project_id，允许跨项目访问
2. **模板参数 JSON 注入** (6.2): 朴素字符串替换可导致 JSON 结构破坏
3. **必选模板参数未验证** (6.2): 缺少 Required 参数检查
4. **模式切换数据丢失** (6.4-6.6): 可视化编辑器未将变更同步到父组件

### HIGH 级别 (已全部修复)
1. **唯一约束阻止软删除后重名** (6.1): 改为 partial unique index
2. **模板列表非确定性排序** (6.2): 添加 sort by ID
3. **nil TemplateParams 导致占位符残留** (6.2): 初始化为空 map
4. **表单验证 Promise 未 catch** (6.5): 添加 .catch() 处理
5. **JSON 解析验证不够严格** (6.4): 增加 stage 结构校验

## 架构产出

### 后端新增
- `internal/pipeline/` 模块 (model, dto, repo, service, handler, template)
- `migrations/000012_create_pipelines` (up/down)
- `pkg/response/codes.go` 新增 402xx 错误码
- 路由: `/api/v1/projects/:id/pipelines/*`、`/api/v1/pipeline-templates/*`

### 前端新增
- `web/src/services/pipeline.ts` - API 服务层
- `web/src/components/pipeline/` - 6 个组件
  - PipelineEditor, StageNode, StepNode, StepConfigPanel
  - YamlEditor, ModeSwitch, PipelineSettingsPanel, RunPipelineModal
- `web/src/lib/pipeline/configYaml.ts` - JSON 转换工具
- `web/src/pages/projects/pipelines/` - 2 个页面
  - PipelineListPage, PipelineEditorPage

### 依赖新增
- `@xyflow/react` ^12 - 流程图编辑器
- `dagre` - 自动布局算法

## What Went Well
1. 后端 CRUD + JSONB 模式非常高效，一次迁移即完成表结构
2. 模板系统设计灵活，参数验证和安全替换机制健壮
3. Code Review 持续发现并修复了安全和数据完整性问题
4. 可视化编排器 UX 体验良好，dagre 自动布局效果不错

## What Could Be Improved
1. 前端测试覆盖率不足，主要依赖 TypeScript 类型检查
2. 状态过滤当前为客户端实现，后端 API 未支持 status 过滤参数
3. RunPipelineModal 为占位符，需 Epic 7 补全
4. 可考虑引入 Monaco Editor 替代 textarea 以提升 JSON 编辑体验

## Action Items for Next Epics
1. Epic 7 需实现流水线执行引擎，补全 RunPipelineModal
2. 考虑后端 List API 支持 status/name 过滤参数
3. 评估是否需要将模板存储迁移到数据库
4. 前端增加 E2E 测试覆盖流水线编排流程
