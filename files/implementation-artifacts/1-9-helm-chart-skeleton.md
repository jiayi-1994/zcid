# Story 1.9: Helm Chart 基础骨架

Status: done

## Story

As a 运维人员,
I want Helm Chart 骨架就绪,
so that 后续可以通过 `helm install` 部署到 K8s。

## Acceptance Criteria (BDD)

1. **Given** charts/zcid/ 目录存在 **When** 执行 `helm template zcid ./charts/zcid` **Then** 生成有效的 K8s Deployment、Service、ConfigMap 资源清单
2. **Given** values.yaml 参数化 **When** 修改 values.yaml 中的数据库连接、Redis 地址等 **Then** 渲染出的资源清单反映修改后的值
3. **Given** Helm Chart 就绪 **When** 执行 `helm lint ./charts/zcid` **Then** 无 error 级别问题

## Tasks / Subtasks

- [x] Task 1: 创建 Helm Chart 基础结构 (AC: #1, #3)
  - [x] 1.1 创建 `charts/zcid/Chart.yaml`
  - [x] 1.2 创建 `charts/zcid/.helmignore`
  - [x] 1.3 创建 `charts/zcid/values.yaml`（参数化数据库、Redis、MinIO、服务端口等）

- [x] Task 2: 创建 K8s 资源模板 (AC: #1, #2)
  - [x] 2.1 创建 `templates/_helpers.tpl`（通用模板函数）
  - [x] 2.2 创建 `templates/deployment.yaml`（zcid 后端 Deployment）
  - [x] 2.3 创建 `templates/service.yaml`（ClusterIP Service）
  - [x] 2.4 创建 `templates/configmap.yaml`（应用配置 ConfigMap）

- [x] Task 3: 验证 Helm Chart (AC: #1, #2, #3)
  - [x] 3.1 执行 `helm lint ./charts/zcid` 无 error
  - [x] 3.2 执行 `helm template zcid ./charts/zcid` 生成有效资源清单
  - [x] 3.3 验证修改 values.yaml 参数后渲染结果正确反映变更

## Dev Notes

- 架构要求：ARCH-11 Helm Chart 部署（charts/zcid/）
- 参考 config/config.yaml 中的配置项进行 values.yaml 参数化
- 敏感配置（密码、密钥）通过环境变量注入，不写入 ConfigMap
- 健康检查端点：/healthz (liveness)、/readyz (readiness)
- 服务端口：8080
- MVP 阶段保持简单 Chart 结构

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `helm lint ./charts/zcid` (PASS — 0 charts failed, 1 INFO)
- `helm template zcid ./charts/zcid` (PASS — Deployment/Service/ConfigMap rendered)
- `helm template --set database.host=custom-db` (PASS — values override confirmed)

### Completion Notes List

- 创建完整 Helm Chart 骨架：Chart.yaml、.helmignore、values.yaml
- 模板包含 Deployment（含 liveness/readiness probe、Secret 环境变量注入、ConfigMap 挂载）、Service（ClusterIP）、ConfigMap（映射 config/config.yaml 结构）
- values.yaml 参数化数据库、Redis、MinIO 连接信息，敏感配置通过 K8s Secret 环境变量注入
- 通用模板函数 _helpers.tpl 提供 labels、selectorLabels、fullname

### Change Log

- 2026-03-02: 完成 Story 1.9，状态更新为 `review`。
- 2026-03-02: Code review (AI) — H1 fix: removed `| quote` from configmap.yaml server.port; H2 fix: added ENCRYPT_KEY secret env to deployment.yaml; M3 fix: service.yaml targetPort uses values with default. Moved to `done`.

### File List

- `charts/zcid/Chart.yaml`
- `charts/zcid/.helmignore`
- `charts/zcid/values.yaml`
- `charts/zcid/templates/_helpers.tpl`
- `charts/zcid/templates/deployment.yaml`
- `charts/zcid/templates/service.yaml`
- `charts/zcid/templates/configmap.yaml`
- `files/implementation-artifacts/1-9-helm-chart-skeleton.md`
- `files/implementation-artifacts/sprint-status.yaml`
