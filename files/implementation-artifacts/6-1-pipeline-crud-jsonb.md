# Story 6.1: 流水线 CRUD 与 JSONB 存储

Status: done

## Story

As a 项目管理员,
I want 创建、编辑、删除流水线配置,
so that 我可以管理项目的 CI/CD 流程。

## Acceptance Criteria

1. **创建流水线**
   - Given 项目管理员已登录且在某项目下
   - When POST /api/v1/projects/:projectId/pipelines
   - Then 流水线配置以 JSONB 存储，包含 schemaVersion 字段
   - And 返回流水线 ID 和基本信息

2. **查询流水线列表**
   - Given 项目内存在流水线
   - When GET /api/v1/projects/:projectId/pipelines
   - Then 返回分页列表，每行包含 name/status/triggerType/updatedAt
   - And 支持 page/pageSize 分页参数

3. **查询单条流水线**
   - Given 流水线已存在
   - When GET /api/v1/projects/:projectId/pipelines/:pipelineId
   - Then 返回完整的流水线配置（包含 JSONB config）

4. **更新流水线**
   - Given 流水线已存在
   - When PUT /api/v1/projects/:projectId/pipelines/:pipelineId
   - Then 配置更新成功，高频字段（name/status）同步更新独立列

5. **删除流水线**
   - Given 流水线已存在
   - When DELETE /api/v1/projects/:projectId/pipelines/:pipelineId
   - Then 流水线软删除（status 设为 deleted）

6. **复制流水线**
   - Given 流水线已存在
   - When POST /api/v1/projects/:projectId/pipelines/:pipelineId/copy
   - Then 创建新流水线，配置复制自原流水线，名称加 "-copy" 后缀（FR29）

7. **流水线名称唯一性**
   - Given 同项目内已存在同名流水线
   - When 创建或更新流水线
   - Then 返回 40202 名称重复错误

8. **JSONB 配置 Schema**
   - Given 流水线配置为 JSONB
   - When 存储和读取
   - Then config 字段包含 schemaVersion、stages（Stage→Step 模型）、params、metadata

## Tasks / Subtasks

- [x] Task 1: 创建数据库迁移 — pipelines 表 (AC: #1, #2, #8)
  - [x] 1.1 创建 000012_create_pipelines.up.sql（JSONB config + 独立列索引）
  - [x] 1.2 创建 000012_create_pipelines.down.sql
- [x] Task 2: 创建 internal/pipeline 模块 — model.go (AC: #8)
  - [x] 2.1 Pipeline GORM 模型 + PipelineStatus 枚举 + PipelineConfig JSON 结构体
- [x] Task 3: 创建 internal/pipeline 模块 — dto.go (AC: #1-#6)
  - [x] 3.1 CreatePipelineRequest / UpdatePipelineRequest / PipelineResponse / PipelineListResponse
  - [x] 3.2 CopyPipelineResponse（使用 PipelineResponse）
  - [x] 3.3 ToPipelineResponse / ToPipelineSummaryResponse 转换函数
- [x] Task 4: 创建 internal/pipeline 模块 — repo.go (AC: #1-#7)
  - [x] 4.1 Repository 接口定义
  - [x] 4.2 Create / GetByID / List / Update / SoftDelete / ExistsByNameAndProject 实现
- [x] Task 5: 创建 internal/pipeline 模块 — service.go (AC: #1-#7)
  - [x] 5.1 CreatePipeline（含名称唯一校验 + schemaVersion 初始化）
  - [x] 5.2 GetPipeline
  - [x] 5.3 ListPipelines（分页）
  - [x] 5.4 UpdatePipeline（高频字段同步 + 名称冲突检查）
  - [x] 5.5 DeletePipeline（软删除）
  - [x] 5.6 CopyPipeline（深拷贝 config + 名称加 "-copy" 后缀 + 自动递增编号）
- [x] Task 6: 创建 internal/pipeline 模块 — handler.go (AC: #1-#6)
  - [x] 6.1 RegisterRoutes 注册 CRUD + copy 路由
  - [x] 6.2 CreatePipeline / GetPipeline / ListPipelines / UpdatePipeline / DeletePipeline / CopyPipeline handlers
- [x] Task 7: 注册路由到 main.go (AC: #1-#6)
  - [x] 7.1 在 cmd/server/main.go 添加 pipeline 路由注册
  - [x] 7.2 路由挂载到 /api/v1/projects/:id/pipelines，使用 RequireProjectScope 中间件
- [x] Task 8: 注册错误码 (AC: #7)
  - [x] 8.1 在 pkg/response/codes.go 注册 402xx 段流水线错误码（40201-40204）
- [x] Task 9: 编写单元测试 (AC: #1-#7)
  - [x] 9.1 service_test.go — 10 个测试用例全部通过

## Dev Notes

### 架构决策

**JSONB 存储策略（架构文档核心决策）：**

采用"胖 JSONB + 独立列索引"策略：
- 流水线配置整体存为一个 JSONB 字段 `config`，包含 `schemaVersion` 用于数据迁移
- 高频查询字段（name、status、project_id、trigger_type、created_at）保持独立列 + B-tree 索引
- JSONB 只做存储，不做 GIN 索引查询
- 理由：管理平台 QPS 不高，独立列索引覆盖所有列表查询场景

**数据库表设计：**

```sql
CREATE TABLE pipelines (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id      UUID NOT NULL REFERENCES projects(id),
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft',
    config          JSONB NOT NULL DEFAULT '{}',
    trigger_type    VARCHAR(20) NOT NULL DEFAULT 'manual',
    concurrency_policy VARCHAR(20) NOT NULL DEFAULT 'queue',
    created_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_pipelines_project_name UNIQUE (project_id, name)
);

CREATE INDEX idx_pipelines_project_id ON pipelines(project_id);
CREATE INDEX idx_pipelines_status ON pipelines(status);
CREATE INDEX idx_pipelines_trigger_type ON pipelines(trigger_type);
```

**JSONB config 结构体设计：**

```go
type PipelineConfig struct {
    SchemaVersion string        `json:"schemaVersion"`
    Stages        []StageConfig `json:"stages"`
    Params        []ParamConfig `json:"params,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
}

type StageConfig struct {
    ID    string       `json:"id"`
    Name  string       `json:"name"`
    Steps []StepConfig `json:"steps"`
}

type StepConfig struct {
    ID       string         `json:"id"`
    Name     string         `json:"name"`
    Type     string         `json:"type"`
    Image    string         `json:"image,omitempty"`
    Command  []string       `json:"command,omitempty"`
    Args     []string       `json:"args,omitempty"`
    Env      map[string]string `json:"env,omitempty"`
    Config   map[string]any `json:"config,omitempty"`
}

type ParamConfig struct {
    Name         string `json:"name"`
    Type         string `json:"type"`
    DefaultValue string `json:"defaultValue,omitempty"`
    Description  string `json:"description,omitempty"`
    Required     bool   `json:"required"`
}
```

**Pipeline 状态枚举：**
- `draft` — 草稿（刚创建，尚未运行过）
- `active` — 活跃（可被触发运行）
- `disabled` — 已禁用（不可被触发）
- `deleted` — 已删除（软删除）

**触发类型枚举：**
- `manual` — 手动触发
- `webhook` — Webhook 触发
- `scheduled` — 定时触发（P1 预留）

**并发策略枚举：**
- `queue` — 排队等待
- `cancel_old` — 取消旧构建
- `reject` — 拒绝并通知

**错误码分配（架构文档 402xx 段）：**

| 错误码 | 说明 |
|--------|------|
| 40201 | 流水线不存在 |
| 40202 | 流水线名称重复（项目内唯一） |
| 40203 | CRD 翻译失败（Story 7.1 使用） |
| 40204 | 并发数超限（Story 7.2 使用） |

### 代码模式参考

**项目作用域路由注册 — 参考 main.go registerProjectRoutes：**

流水线是项目级资源，路由需要挂载到 `projectScope` 下：

```go
pipelineGroup := projectScope.Group("/pipelines")
pipelineHandler.RegisterRoutes(pipelineGroup)
```

**handler→service→repo 三层结构 — 参考 internal/project/ 模块：**

- handler.go: RegisterRoutes 注册路由，使用 projectScope 中间件提供的 projectId
- service.go: 业务逻辑，name 唯一校验、config JSONB 序列化
- repo.go: 数据访问，GORM + JSONB 存储

**项目 ID 获取 — 从 URL 参数提取：**

```go
projectID := c.Param("id")  // 来自 /projects/:id/pipelines
```

注意：projectID 来自 `RequireProjectScope` 中间件已验证过的 `:id` 参数，不需要再次验证项目存在性。

**JSONB 在 GORM 中的处理：**

使用 `datatypes.JSON` 或自定义 JSON scanner：

```go
type Pipeline struct {
    // ...
    Config datatypes.JSON `gorm:"column:config;type:jsonb;not null;default:'{}'"`
}
```

或者更好的做法 — 使用自定义类型实现 `database/sql` 的 `Scanner` 和 `driver.Valuer` 接口：

```go
type PipelineConfig struct { ... }

func (c PipelineConfig) Value() (driver.Value, error) {
    return json.Marshal(c)
}

func (c *PipelineConfig) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok { return errors.New("type assertion to []byte failed") }
    return json.Unmarshal(bytes, c)
}
```

### 技术约束

1. **JSONB 不使用 GIN 索引**：按照架构决策，列表查询全部通过独立列索引完成
2. **schemaVersion 必须存在**：创建时默认设为 "1.0"，后续数据迁移依赖此字段
3. **名称唯一性约束**：通过 DB 层 UNIQUE CONSTRAINT (project_id, name) 保证，repo 层捕获并翻译错误
4. **软删除**：使用 status 字段而非 GORM DeletedAt，与项目模块保持一致
5. **复制流水线**：深拷贝 config JSON，生成新 UUID，名称添加 "-copy" 后缀
6. **config 字段不可为 null**：使用 `DEFAULT '{}'` 确保总是有合法的 JSON

### Project Structure Notes

新增文件遵循 internal/ 模块化结构：

```
zcid/
├── internal/pipeline/          # 新增模块
│   ├── handler.go
│   ├── service.go
│   ├── service_test.go
│   ├── repo.go
│   ├── model.go
│   └── dto.go
├── migrations/
│   ├── 000012_create_pipelines.up.sql    # 新增
│   └── 000012_create_pipelines.down.sql  # 新增
├── cmd/server/main.go           # 修改：添加 pipeline 路由注册
└── pkg/response/codes.go        # 修改：添加 402xx 错误码
```

### References

- [Source: files/planning-artifacts/architecture.md#Data Architecture] — JSONB 存储策略
- [Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention] — handler→service→repo 结构
- [Source: files/planning-artifacts/architecture.md#API & Communication Patterns] — REST 路由 /api/v1/projects/:id/pipelines
- [Source: files/planning-artifacts/architecture.md#错误码段分配] — 402xx 段流水线
- [Source: files/planning-artifacts/architecture.md#Naming Patterns] — 数据库命名和 JSON 字段命名
- [Source: files/planning-artifacts/epics.md#Story 6.1] — 流水线 CRUD AC
- [Source: internal/project/] — 同结构项目模块实现参考
- [Source: cmd/server/main.go#registerProjectRoutes] — 项目作用域路由注册模式

### Previous Story Intelligence

**Epic 5 完成情况：**
- 4 个 Story 全部完成 + Code Review 修复 14 个问题
- handler→service→repo 三层结构已成熟，本 Story 严格复用
- 错误码段使用了 4041x（Git），本次使用 402xx（Pipeline），按照架构文档分配
- Code Review 发现的典型问题：全表扫描（应该用 DB 查询）、掩码基于密文而非明文、接口封装违规——本 Story 需要从一开始避免

**关键学习：**
- GORM 中使用 `isUniqueConstraintError` 辅助函数检测唯一约束冲突
- 软删除使用 `status != 'deleted'` 过滤，不使用 GORM DeletedAt
- DTO 层 Update 请求使用指针类型 `*string` 表示可选字段
- 路由参数 projectID 来自 RequireProjectScope 中间件，无需重复验证

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References
- Build: 一次通过，无编译错误
- Tests: 15/15 通过（Code Review 后增加 5 个测试），全量回归零失败

### Code Review (2026-03-05)
**Reviewer**: Adversarial AI Code Review
**Issues Found**: 1 CRITICAL, 2 HIGH, 5 MEDIUM, 4 LOW
**All CRITICAL, HIGH and MEDIUM issues fixed**, verified by full regression tests.
**Key Fixes Applied**:
- [CRITICAL] 所有 Get/Update/Delete/Copy 操作现在强制校验 project_id，防止跨项目访问
- [HIGH] 唯一约束改为 partial unique index（WHERE status != 'deleted'），允许软删除后名称重用
- [HIGH] Repo 的 Update/SoftDelete/GetByID 全部增加 projectID 参数过滤
- [M1] 新增 status/triggerType/concurrencyPolicy 枚举验证
- [M2] Handler 统一从 URL 提取 projectID 并传递到 service
- [M3] CopyPipeline 通过 GetByIDAndProject 自动校验项目归属
- [M4] mockRepo.List 添加排序，匹配真实 DB 行为
- [M5] PipelineConfig.Scan 支持 string 和 []byte 两种类型
**新增测试用例**: CrossProjectDenied (Get/Copy)、InvalidTriggerType、InvalidStatus

### Completion Notes List
- 创建 pipelines 数据库迁移（UUID 主键，project_id+name 唯一约束，status/trigger_type 索引）
- Pipeline GORM 模型实现自定义 PipelineConfig JSON Scanner/Valuer（database/sql 接口）
- JSONB config 包含 schemaVersion（默认 "1.0"）、stages（Stage→Step 模型）、params、metadata
- 完整 CRUD + 复制（深拷贝 config JSON，名称自动递增 -copy/-copy-2 等）
- 更新时预先检查名称唯一性（ExistsByNameAndProject 排除自身 ID）
- 列表返回 Summary（不含 config JSONB 减少传输量），详情返回完整 config
- 402xx 错误码段注册（40201 不存在、40202 名称重复、40203 CRD 翻译失败、40204 并发超限）
- 路由挂载到 /api/v1/projects/:id/pipelines，使用 RequireProjectScope 中间件
- 10 个单元测试覆盖：创建/重名/查询/复制/复制重名递增/更新/更新冲突/删除/删除不存在/分页

### File List
- `migrations/000012_create_pipelines.up.sql` (新建)
- `migrations/000012_create_pipelines.down.sql` (新建)
- `internal/pipeline/model.go` (新建)
- `internal/pipeline/dto.go` (新建)
- `internal/pipeline/repo.go` (新建)
- `internal/pipeline/service.go` (新建)
- `internal/pipeline/handler.go` (新建)
- `internal/pipeline/service_test.go` (新建)
- `cmd/server/main.go` (修改: 添加 pipeline 路由注册)
- `pkg/response/codes.go` (修改: 添加 402xx 错误码)
