# MeterSphere（私有化）通过 OpenAPI 创建/导入功能用例（Test Case）

> 适用实例：`http://172.20.18.75:8081`
>
> 已验证：`GET /v3/api-docs` 可访问；`GET /v3/api-docs/swagger-config` 返回 `url: /v3/api-docs`。

## 1. 文档入口（无需 UI 也能用）

- OpenAPI JSON：
  - `GET http://172.20.18.75:8081/v3/api-docs`
- Swagger 配置：
  - `GET http://172.20.18.75:8081/v3/api-docs/swagger-config`

> 说明：你当前的 `swagger-ui/index.html` 里默认指向 petstore，需要改成读取 `swagger-config` 才会显示 MeterSphere 接口；但即便 UI 不可用，OpenAPI JSON 也足够用于对接。

## 2. 与“功能用例”相关的关键接口（来自你们 /v3/api-docs）

### 2.1 登录

- `POST /signin`（tag：`login-controller`）
  - `Content-Type: application/json`
  - Body schema：`LoginRequest { username, password, authenticate }`

> 注：前端会用 `/isLogin` 返回的 publicKey 对用户名/密码做 RSA 加密，但接口本身接受明文（用错误用户名密码会返回 `Incorrect password or username`，说明可直连）。

补充：我用 Node.js 的 `fetch` 请求同样的错误账号密码时，返回的 message 文案不同（`Not Support Key: password_is_incorrect`）。这不影响判断“接口可直连”，但说明错误信息可能受网关/国际化/异常处理链影响。

### 2.2 创建/保存用例（推荐用 JSON 方式：/test/case/save）

- `POST /test/case/save`（tag：`test-case-controller`）
  - `Content-Type: application/json`
  - Header：`CSRF-TOKEN`（OpenAPI 标注 required）
  - Request schema：`EditTestCaseRequest`（关键字段举例）
    - `projectId` (string)
    - `nodeId` (string)
    - `name` (string)
    - `priority` (string)
    - `status` (string)
    - `tags` (string)
    - `maintainer` (string)
    - `prerequisite` (string)
    - `remark` (string)
    - `steps` (string)
    - `expectedResult` (string)
    - `stepDescription` (string)
    - `customFields` (string)
  - 200 响应：`TestCaseWithBLOBs`（含 `id/projectId/nodeId/name/...`）

### 2.3 创建用例（multipart 方式：/test/case/add）

- `POST /test/case/add`
  - `Content-Type: multipart/form-data`
  - Header：`CSRF-TOKEN`
  - Form fields：
    - `request`：`EditTestCaseRequest`（通常是 JSON 字符串）
    - `file`：array[binary]（可选）

### 2.4 导入用例（/test/case/import）

- `POST /test/case/import`
  - Header：`CSRF-TOKEN`
  - Request schema（OpenAPI 显示为对象，required：`file` + `request`）：
    - `file`：binary
    - `request`：`TestCaseImportRequest`
      - `projectId`
      - `userId`
      - `importType`
      - `versionId`
      - `ignore` / `userIds` / `testCaseNames` / `customFields` / ...

## 3. curl 示例（按“浏览器会话”方式最稳）

> 下面示例假设：你用账号密码登录（会返回/设置 `MS_SESSION_ID` cookie），并沿用同一个 cookie 做后续接口。

### 3.1 登录拿到会话 cookie

```bash
curl -sS -c ms.cookie -X POST "http://172.20.18.75:8081/signin" \
  -H "Content-Type: application/json" \
  --data '{"username":"<USER>","password":"<PASS>","authenticate":"LOCAL"}'
```

如果你们实例要求 RSA 加密（与前端一致），需要先拿 publicKey：

```bash
curl -sS "http://172.20.18.75:8081/isLogin"
```

其中 `message` 字段就是 publicKey（base64 DER）。你可以用前端同样的 RSA 公钥加密方式加密 username/password，再调用 `/signin`。

### 3.2 保存/创建用例（/test/case/save）

```bash
curl -sS -b ms.cookie -X POST "http://172.20.18.75:8081/test/case/save" \
  -H "Content-Type: application/json" \
  -H "CSRF-TOKEN: 1" \
  --data '{
    "projectId": "<PROJECT_ID>",
    "nodeId": "<MODULE_NODE_ID>",
    "name": "登录-正常登录",
    "priority": "P0",
    "status": "Underway",
    "tags": "login,smoke",
    "maintainer": "<USER_ID>",
    "prerequisite": "已存在有效用户",
    "steps": "1. 打开登录页\\n2. 输入用户名密码\\n3. 点击登录",
    "expectedResult": "登录成功进入首页"
  }'
```

> `CSRF-TOKEN`：OpenAPI 标注为必填 header，但并未在 OpenAPI 中给出获取方式。实际场景里它可能不校验值，或者校验与 cookie/session 绑定。
> 如果你碰到 403/CSRF 校验失败：需要从页面/接口返回的 header/cookie 中提取真正 token（可通过抓包或在浏览器 Network 看请求头）。

## 4. 对接建议（落地路线）

1) 用 `/signin` 登录，保留 `MS_SESSION_ID` cookie。
2) 用你们项目的 `projectId` 和用例模块 `nodeId`，走 `/test/case/save` 写入用例（最简单）。
3) 若要批量导入历史用例，走 `/test/case/import`（需要平台支持的文件格式/模板）。
4) 最后把这套封装成脚本/CI Job：把生成的用例结构化（title/steps/expected/tags/priority），自动同步到 MeterSphere。

## 5. 如何拿到 projectId 与 nodeId（来自 OpenAPI）

### 5.1 获取项目列表（projectId）

- `GET /project/listAll`（tag：`project-controller`）

> 你们这份 OpenAPI 定义里，该接口 **直接返回 `ProjectDTO[]`**（不是 ResultHolder 包一层）。

> 该接口需要登录态（否则会 302 跳转 /login）。拿到 cookie 后再调用。

### 5.2 获取用例模块树（nodeId）

- `GET /case/node/list/{projectId}`（tag：`test-case-node-controller`）
  - Header：`CSRF-TOKEN`
  - 返回：`TestCaseNodeDTO[]`，其中包含 `id/name/parentId/children/caseNum...`

示例：

```bash
curl -sS -b ms.cookie "http://172.20.18.75:8081/case/node/list/<PROJECT_ID>" \
  -H "CSRF-TOKEN: 1"
```
