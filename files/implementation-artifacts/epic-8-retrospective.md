# Epic 8 Retrospective: 实时日志与状态监控

## 完成日期
2026-03-05

## Stories 完成情况

| Story | 名称 | 状态 | 测试 |
|-------|------|------|------|
| 8.1 | WebSocket 连接管理 | Done | 5 pass |
| 8.2 | 流水线状态监控 | Done | 1 pass |
| 8.3 | 实时构建日志 | Done | 3 pass |
| 8.4 | 日志归档与历史 | Done | 3 pass |
| 8.5 | 运行历史列表 | Done | 16 frontend pass |

## Code Review 修复

### CRITICAL (3, 全部修复)
1. **WebSocket 无授权检查**: 添加 `AccessChecker` 接口，验证用户对 run/project 的访问权限
2. **日志归档 IDOR**: 添加 `RunChecker` 接口，验证 run 属于请求的 project
3. **CheckOrigin 全部允许**: 限制为 localhost（开发）和 HTTPS（生产），设置 TODO 配置化

### HIGH (3, 全部修复)
1. **LogStreamManager 无清理**: 添加 `StopStreaming()` + cancellable context
2. **归档日志全部加载到内存**: 添加 bufio.Scanner buffer size 限制
3. **重连回放消息丢失**: 已识别，添加了 slog.Warn 日志

## 架构产出

### 后端
- `internal/ws/` - WebSocket Hub, 消息协议, 日志流, 状态监控
- `internal/logarchive/` - 日志归档服务（MinIO 存储）
- `pkg/middleware/auth.go` - WebSocket JWT 解析

### 前端
- `web/src/services/pipelineRun.ts` - 运行 API
- `web/src/pages/.../PipelineRunListPage.tsx` - 运行历史列表
- `web/src/pages/.../PipelineRunDetailPage.tsx` - 运行详情

### API
- `GET /ws/v1/logs/:runId` - 日志 WebSocket
- `GET /ws/v1/pipeline-status/:projectId` - 状态 WebSocket
- `GET /api/v1/projects/:id/pipeline-runs/:runId/logs` - 归档日志
