# Story 4.3: 运行时密钥注入与清理

Status: done

## Story

As a 系统,
I want 提供变量合并解析能力，并为流水线运行时的密钥注入预留接口,
so that 后续流水线执行 Epic 可直接调用。

## Acceptance Criteria

1. **变量合并解析**
   - Given 全局变量 DB_HOST=global-db，项目变量 DB_HOST=project-db
   - When 调用 ResolveVariables(projectID)
   - Then 返回 DB_HOST=project-db（项目级覆盖全局级）

2. **密钥解密用于注入**
   - Given 解析出的变量包含密钥类型
   - When 调用 ResolveVariables(projectID) with decrypt=true
   - Then 密钥变量值被解密为明文（仅供内部注入使用，不对外暴露）

3. **K8s Secret 接口定义**
   - Given 系统定义了 SecretInjector 接口
   - When 流水线执行层调用
   - Then 可创建临时 K8s Secret 并在结束后清理（接口已定义，实现留至 Epic 7）

## Tasks / Subtasks

- [x] Task 1: 实现 variable service 的 ResolveVariables 方法
- [x] Task 2: 定义 SecretInjector 接口 (pkg/k8s/secret.go)
- [x] Task 3: 添加变量解析测试 (TestResolveVariables_DecryptsSecrets)

## Dev Notes

### Source tree components to touch
- `internal/variable/service.go` (修改 - 添加 ResolveVariables)
- `pkg/k8s/secret.go` (新建 - SecretInjector 接口定义)

### 说明
- K8s Secret 的实际创建和清理将在 Epic 7 (流水线执行) 中实现
- 本 story 聚焦于变量合并逻辑和接口定义
- 日志脱敏已在 Story 1.4 中实现（slog handler 自动脱敏 sensitive keys）

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- ResolveVariables 方法：合并全局+项目变量后解密密钥值（仅供内部使用）
- SecretInjector 接口定义：CreateSecret/DeleteSecret（实现留至 Epic 7）
- TestResolveVariables_DecryptsSecrets 验证解密逻辑

### File List
- `internal/variable/service.go` - ResolveVariables 方法
- `pkg/k8s/secret.go` - SecretInjector 接口
- `internal/variable/service_test.go` - 解析测试
