# Story 6.2: 模板一键创建流水线

Status: done

## Story

As a 用户,
I want 从预置模板一键创建流水线,
so that 不需要从零配置即可快速开始。

## Acceptance Criteria

1. **查询可用模板列表**
   - Given 用户在创建流水线页面
   - When GET /api/v1/pipeline-templates
   - Then 返回预置模板列表（Go 微服务/Java Maven/前端 Node/通用 Docker）

2. **查询模板详情**
   - Given 模板列表已加载
   - When GET /api/v1/pipeline-templates/:templateId
   - Then 返回模板的完整配置（Stage/Step 结构）和需要填写的参数列表

3. **基于模板创建流水线**
   - Given 用户选择模板并填写参数（仓库、分支、镜像名等）
   - When POST /api/v1/projects/:projectId/pipelines with templateId and params
   - Then 基于模板 JSON 生成完整流水线配置并保存（FR21）
   - And 模板参数替换到 config 中

4. **模板结构包含参数定义**
   - Given 模板配置
   - When 解析模板
   - Then 每个模板包含 id、name、description、category、params（需要用户填写的参数列表）、config（模板配置）

## Tasks / Subtasks

- [x] Task 1: 创建模板注册表和内置模板 (AC: #1, #2, #4)
  - [x] 1.1 创建 internal/pipeline/template.go — 模板类型定义 + 内置模板注册
  - [x] 1.2 定义 4 个内置模板 JSON（Go 微服务/Java Maven/前端 Node/通用 Docker）
- [x] Task 2: 扩展 service.go — 模板相关方法 (AC: #1, #2, #3)
  - [x] 2.1 ListTemplates() 返回模板摘要列表（排序）
  - [x] 2.2 GetTemplate(templateId) 返回模板完整配置
  - [x] 2.3 CreatePipeline 集成 templateId + templateParams（含验证和默认值合并）
- [x] Task 3: 扩展 handler.go — 模板路由 (AC: #1, #2, #3)
  - [x] 3.1 GET /pipeline-templates — 列出模板
  - [x] 3.2 GET /pipeline-templates/:templateId — 获取模板详情
  - [x] 3.3 CreatePipeline 支持 templateId + templateParams
- [x] Task 4: 注册模板路由到 main.go (AC: #1, #2)
  - [x] 4.1 模板路由挂载到 /api/v1/pipeline-templates，使用 JWTAuth 中间件
- [x] Task 5: 编写单元测试 (AC: #1-#4)
  - [x] 5.1 24 个测试覆盖模板列表/详情/基于模板创建/缺失参数/默认值/非法字符/排序

## Dev Notes

### 架构决策

**模板存储方式：** MVP 阶段使用内存注册表（硬编码模板），不存数据库。理由：
- 模板数量固定（4 个），不需要动态管理
- 后续可扩展为从数据库或配置文件加载
- 降低首版复杂度

**4 个内置模板：**

1. **go-microservice** — Go 微服务构建：git clone → go build → docker build → push
2. **java-maven** — Java Maven 构建：git clone → mvn package → docker build → push
3. **frontend-node** — 前端 Node 构建：git clone → npm install → npm build → docker build → push
4. **generic-docker** — 通用 Docker 构建：git clone → docker build → push

**模板参数替换：** 使用 `{{.ParamName}}` 占位符，创建时用用户提供的参数替换。

**路由设计：** 模板列表和详情是公共 API（不属于特定项目），路由挂在 `/api/v1/pipeline-templates`。从模板创建流水线仍使用现有 `/api/v1/projects/:id/pipelines` 端点，通过 `templateId` 字段区分。

### 代码模式参考

复用 Story 6.1 的 pipeline 模块，扩展 service 和 handler。

### Project Structure Notes

```
zcid/
├── internal/pipeline/
│   ├── template.go          # 新增：模板定义 + 注册表
│   ├── service.go           # 修改：新增模板方法
│   ├── handler.go           # 修改：新增模板路由
│   ├── dto.go               # 修改：新增模板 DTO
│   └── service_test.go      # 修改：新增模板测试
└── cmd/server/main.go       # 修改：注册模板路由
```

### References

- [Source: files/planning-artifacts/epics.md#Story 6.2] — 模板一键创建 AC
- [Source: files/planning-artifacts/architecture.md#流水线编排与执行] — CRD 翻译层使用统一内部数据模型
- [Source: internal/pipeline/] — Story 6.1 代码基础

### Previous Story Intelligence

**Story 6.1 关键学习：**
- Code Review 发现跨项目访问漏洞，所有操作需通过 projectID 过滤
- 使用 partial unique index 解决软删除后名称重用问题
- 枚举值必须在 service 层验证

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Debug Log References
- Build: 一次通过
- Tests: 24/24 通过，全量回归零失败

### Code Review (2026-03-05)
**Reviewer**: Adversarial AI Code Review
**Issues Found**: 2 CRITICAL, 3 HIGH, 3 MEDIUM, 2 LOW
**All CRITICAL, HIGH and MEDIUM issues fixed.**
**Key Fixes Applied**:
- [CRITICAL] 模板参数替换改为结构化遍历（不再对 JSON 字符串做全局替换），防止 JSON 注入
- [CRITICAL] 新增 required 参数校验 + DefaultValue 自动合并
- [HIGH] 参数值包含非法字符（引号/反斜杠/换行）时返回校验错误
- [HIGH] 模板列表返回排序结果（按 ID 字母序）
- [HIGH] nil TemplateParams 时初始化为空 map 再校验
- [M1] applyTemplateParams 改为 type-safe 结构化处理
- [M2] 新增 7 个测试覆盖参数校验/注入/默认值/排序

### Completion Notes List
- 创建 TemplateRegistry 内存注册表，包含 4 个内置流水线模板
- 模板参数使用 {{.paramName}} 占位符，通过结构化遍历安全替换
- 模板参数验证：必填检查 + 默认值合并 + 非法字符检测
- 模板列表返回稳定排序结果
- 路由 /api/v1/pipeline-templates（JWT 认证，无项目作用域）
- CreatePipeline 扩展支持 templateId + templateParams 字段

### File List
- `internal/pipeline/template.go` (新建)
- `internal/pipeline/service.go` (修改: 新增模板方法 + CreatePipeline 集成)
- `internal/pipeline/handler.go` (修改: 新增模板路由)
- `internal/pipeline/dto.go` (修改: CreatePipelineRequest 增加 templateId/templateParams)
- `internal/pipeline/service_test.go` (修改: 新增 9 个模板测试)
- `cmd/server/main.go` (修改: 注册模板路由)
