# Story 4.2: 密钥变量加密与安全

Status: done

## Story

As a 项目管理员,
I want 创建加密存储的密钥变量,
so that 敏感信息安全存储且不可回显。

## Acceptance Criteria

1. **密钥加密存储**
   - Given 创建密钥类型变量 (var_type=secret)
   - When 存储到数据库
   - Then 值使用 AES-256-GCM 加密，密钥来自环境变量 ZCID_ENCRYPTION_KEY

2. **密钥不可回显**
   - Given 查询变量列表
   - When 返回密钥类型变量
   - Then 值显示为 `******`，不可回显原文

3. **普通成员不可见密钥变量 (FR5)**
   - Given 普通成员查询项目变量
   - When 列表返回
   - Then 密钥类型变量完全不可见（从列表中过滤掉）

4. **密钥启动校验**
   - Given 环境变量 ZCID_ENCRYPTION_KEY 未设置或长度不足
   - When 服务启动
   - Then 服务正常启动但记录警告日志（不 panic，因为非密钥功能不受影响）

5. **更新密钥变量**
   - Given 密钥变量已存在
   - When 更新密钥变量
   - Then 新值重新加密存储

## Tasks / Subtasks

- [x] Task 1: 创建 pkg/crypto/aes.go (AES-256-GCM 加解密)
- [x] Task 2: 添加 ZCID_ENCRYPTION_KEY 环境变量支持
- [x] Task 3: 在 variable service 中集成加密/解密逻辑
- [x] Task 4: 实现密钥值 masking (列表/详情返回 ******)
- [x] Task 5: 实现 FR5 普通成员密钥变量不可见过滤
- [x] Task 6: 添加加密相关测试 (6 个 AES 单元测试)

## Dev Notes

### Source tree components to touch
- `pkg/crypto/aes.go` (新建)
- `pkg/crypto/aes_test.go` (新建)
- `internal/variable/service.go` (修改 - 集成加密)
- `internal/variable/handler.go` (修改 - 角色过滤)
- `config/config.go` (修改 - 添加加密密钥配置)

### 加密规范
- 算法: AES-256-GCM
- 密钥来源: ZCID_ENCRYPTION_KEY 环境变量 (32 bytes)
- 密文格式: nonce(12字节) + ciphertext + tag, base64 编码存储
- 密钥轮换: MVP 不支持，需重启服务

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 新建 pkg/crypto/aes.go，实现 AES-256-GCM 加解密
- 密文格式：nonce(12字节) + ciphertext + tag, base64 编码
- 环境变量 ZCID_ENCRYPTION_KEY 支持 (32 bytes)
- 未设置密钥时服务正常启动但密钥变量功能不可用
- variable service 集成加密：创建和更新时自动加密
- ToVariableResponse 自动 mask 密钥值为 ******
- FilterForRole 实现 FR5：普通成员看不到密钥变量
- Code review 修复：total 计数在 FR5 过滤后与 items 一致
- 6 个 AES 单元测试通过

### File List
- `pkg/crypto/aes.go`
- `pkg/crypto/aes_test.go`
- `config/config.go` - 添加 EncryptionConfig
- `internal/variable/service.go` - 集成加密
- `internal/variable/handler.go` - 角色过滤 + total 修复
- `internal/variable/dto.go` - MaskedValue + ToVariableResponse
