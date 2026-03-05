# Story 1.8: API 客户端自动生成流水线

Status: done

## Story

As a 前端开发者,
I want 后端 API 变更后自动生成 TypeScript 客户端代码,
so that 前后端类型始终同步，不手写 API 调用。

## Acceptance Criteria (BDD)

1. **Given** 后端执行 `make swag` 生成 OpenAPI 文档 **When** 前端执行 `npm run codegen` **Then** `@hey-api/openapi-ts` 从 `docs/swagger.json` 生成 `web/src/services/generated/` 客户端代码，包含请求 SDK 与类型定义
2. **Given** 后端新增或修改 API **When** 重新执行 swag + codegen **Then** 前端生成类型同步更新，不兼容变更可由 TypeScript 编译暴露

## Tasks / Subtasks

- [x] Task 1: 校准代码生成工具链 (AC: #1)
  - [x] 1.1 检查 `web/openapi-ts.config.ts` 输入输出配置
  - [x] 1.2 验证 `web/package.json` `codegen` 脚本可直接执行
  - [x] 1.3 修复 `Makefile` 中 `swag` 命令参数兼容性（移除无效 `-v2`）

- [x] Task 2: 执行 OpenAPI 文档与客户端生成 (AC: #1)
  - [x] 2.1 执行 `make swag` 生成 `docs/swagger.json|yaml|docs.go`
  - [x] 2.2 执行 `npm run codegen` 生成 `web/src/services/generated/*`
  - [x] 2.3 确认生成物包含 SDK 与类型文件（`sdk.gen.ts` / `types.gen.ts` 等）

- [x] Task 3: 验证生成链路可集成到前端构建 (AC: #2)
  - [x] 3.1 执行 `npm run build`，确保 codegen 产物不破坏构建
  - [x] 3.2 记录当前后端尚无 swag 注解接口，生成客户端为基础骨架（后续 Story 扩展）

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `make -C "/f/other code/zcid" swag` (PASS)
- `npm --prefix "/f/other code/zcid/web" run codegen` (PASS)
- `npm --prefix "/f/other code/zcid/web" run build` (PASS)

### Completion Notes List

- 修复 `Makefile` 的 `swag` 命令参数，避免 `flag provided but not defined: -v2` 导致流水线中断。
- 生成并落地 OpenAPI 文档文件：`docs/docs.go`、`docs/swagger.json`、`docs/swagger.yaml`。
- 生成前端 API 客户端目录 `web/src/services/generated/`，包含 `index.ts`、`sdk.gen.ts`、`types.gen.ts` 与 `core/client` 相关生成文件。
- 通过前端构建验证 codegen 产物可被项目消费。

### Change Log

- 2026-03-02: 完成 Story 1.8，状态更新为 `review`。
- 2026-03-02: Code review (AI) — fixed pre-existing swag dependency missing from go.mod (go get github.com/swaggo/swag). Moved to `done`.

### File List

- `Makefile`
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`
- `web/src/services/generated/index.ts`
- `web/src/services/generated/sdk.gen.ts`
- `web/src/services/generated/types.gen.ts`
- `web/src/services/generated/client.gen.ts`
- `web/src/services/generated/core/auth.gen.ts`
- `web/src/services/generated/core/bodySerializer.gen.ts`
- `web/src/services/generated/core/params.gen.ts`
- `web/src/services/generated/core/pathSerializer.gen.ts`
- `web/src/services/generated/core/types.gen.ts`
- `web/src/services/generated/client/client.gen.ts`
- `web/src/services/generated/client/index.ts`
- `web/src/services/generated/client/types.gen.ts`
- `web/src/services/generated/client/utils.gen.ts`
- `files/implementation-artifacts/1-8-api-client-codegen.md`
- `files/implementation-artifacts/sprint-status.yaml`
- `files/implementation-artifacts/project-skills.md`
