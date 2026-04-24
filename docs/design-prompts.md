# zcid 前端设计词（Claude Design 输入）

> 产品：**zcid**，云原生 CI/CD 平台（基于 Tekton + ArgoCD）。
> 目标风格：shadcn/ui + Linear + Vercel Dashboard 风 — neutral zinc 调色板、Inter 字体、极简阴影、白底、圆角 8–12px，primary 品牌蓝（可 gradient），失败红、成功绿、警告橙。信息密度中等偏高，工程师审美。深色代码块用暗色终端风格。桌面优先（1280+），侧导航 + 顶栏壳层。语言中英混排（英文标题 + 中文副标题）。

---

## 壳层

### 1. `AppLayout` — 全局左侧导航 + 顶栏

左侧 232px 固定侧栏：顶部 `Z` logo + zcid 字样；分组 "Workspace"（Dashboard / 项目管理）与 "System"（用户管理 / 全局变量 / 集成管理 / 审计日志 / 系统设置），导航项带 icon + label，active 态用 primary 浅底 + 左侧强调条。底部用户入口：头像字母 + 用户名 + 角色标签 + 下拉箭头，点击弹出"退出登录"菜单。右侧内容区：顶部细 header（面包屑 `zcid`），主体可滚动。

### 2. `ProjectLayout` — 项目级二级侧栏

在 AppLayout 内嵌套。左侧 220px 项目子栏：顶部项目标识（首字母方块图标 + 项目名），下方菜单：流水线 / 环境 / 部署 / 服务 / 成员 / 变量 / 通知，各带 icon。

---

## 公开页

### 3. `LoginPage` — 登录

两栏布局（桌面左右各 50%）。
- **左栏**：渐变品牌面板，大号 `Z` logo 方块（圆角 + 柔和光晕），`zcid` 标题，副文"云原生 CI/CD 平台 基于 Tekton + ArgoCD 构建"，下方 4 条 feature 带 emoji icon：⚡ 可视化流水线编辑、🐳 Kaniko/BuildKit 容器化构建、🚀 ArgoCD GitOps 自动部署与回滚、🔐 AES-256 加密变量管理 + RBAC。
- **右栏**：垂直居中登录卡片（max-width 360）。标题"欢迎回来"+ 副文"使用你的账号密码登录平台"。字段：用户名 / 密码（Input 高 46，圆角 8）。主按钮 long + primary gradient "登录"。底部小字 "Powered by Tekton + ArgoCD"。

### 4. `ForbiddenPage` — 403

全屏居中：大号"403"数字（渐变文字），标题"无权限访问"，描述"你没有访问当前页面的权限，请联系管理员获取相应权限。"，主按钮"返回 Dashboard"（左箭头 icon）。套 AppLayout 壳。

### 5. `NotFoundPage` — 404

同 403 布局：大号"404"，"页面不存在"+"你访问的页面不存在或已被移除"，主按钮"返回首页"（home icon）。

---

## 全局页

### 6. `DashboardPage` — 工作台首页

顶部 page-header：面包屑 "Cloud-Native Overview"，大标题"早上好/下午好/晚上好，{username}"（基于小时动态），副文"Global infrastructure health and deployment telemetry. 当前 CI/CD 工作台概览"。右上角按钮组：Refresh + primary "New Pipeline"。

**指标区** — 4 列等宽 metric cards：
- Total Projects（apps icon，primary tone）
- Total Pipelines（thunderbolt icon）
- Recent Success（check icon，绿色，副文 `xx% success rate`）
- Recent Failures（close icon，红色，副文 Needs attention / All green）

每 card：左上角 icon 圆角方块 + label（小号全大写）+ 大号数值 + 次要 trend 文本。

**主区** — 两栏不对称布局（左 2fr / 右 1fr）：
- **左 Projects 列表**：section title "Projects" + 右"查看全部 →"文字按钮；下方项目行，每行：圆形首字母 avatar（按最近运行状态上色）+ 名称 + 描述 + 最近 run 状态徽章 + 右箭头。空态"还没有项目"+创建按钮。
- **右 快速操作**：section title "快速操作"+ 纵向 4 张 onboarding-step 卡片：新建项目、创建流水线、集成管理、查看文档。

### 7. `ProjectListPage` — 项目列表

Page-header：面包屑 "Project Directory"，标题"项目管理"，副文"管理所有项目及其 CI/CD 配置"。右侧搜索框（220px，search icon）+ primary "新建项目"（admin 可见）。

**2 列指标卡**：Total Projects / Active。

**项目卡片网格**（`auto-fill` 280px 最小宽）：每张 card — 顶部首字母渐变色方块 + 名称 + 状态徽章（运行中绿/未激活灰/已归档）；中部 2 行描述；底部分隔线上方 创建日期 + 右下操作（删除文本按钮 + outline "进入"箭头按钮）。整卡 hover 可点击进入。空态：大号标题+描述+创建按钮。底部分页（>12 显示）。带新建弹窗。

### 8. `AdminUsersPage` — 用户管理（admin）

Page-header：面包屑 "System · Access Control"，标题"用户管理"，副文"管理系统用户账号与角色"。右侧 primary "新建用户"。

**表格卡**：列 用户名 / 角色（badge：管理员蓝 / 项目管理员绿 / 普通成员灰）/ 状态（启用绿 / 禁用灰）/ 创建时间 / 操作（编辑 + 启用/禁用 Popconfirm）。空态"暂无用户数据"。新建/编辑走 UserFormModal。

### 9. `AdminVariablePage` — 全局变量（admin）

Page-header：面包屑 "System › Variables"，标题"Global Variables"，副文"System-wide variable management. 管理跨项目共享的全局变量和密钥"。右侧 primary "+ Add Variable"。

**表格卡**：列 变量名 / 值（secret 类型显示 ****** 灰色）/ 类型（badge：Secret 红 / Variable 蓝）/ 描述 / 创建时间 / 操作（编辑+删除）。弹窗表单 key、value、类型、description。

### 10. `IntegrationsPage` — 集成管理（admin）

Page-header：面包屑 "Settings › Integrations"，标题"Integration Management"，副文"Connect and manage your external CI/CD toolchain..."。右侧 Refresh + primary "Connect New Service"。

**3 列指标卡**：Total Integrations / Connected / Needs Attention。

**集成卡片网格**：每卡顶部 emoji provider icon（🐙 GitHub / 🦊 GitLab）+ 名称 + "Provider · serverUrl" 副文；中部键值行 Token（mono 脱敏）/ Description / Created；底部分割线下状态点（connected 绿呼吸点 / token_expired 黄 / disconnected 红）+ 右侧 icon 动作组（测试连接 ▶ / 复制 Webhook Secret / 编辑 / 删除）。空态+创建按钮。带连接表单弹窗。

### 11. `AuditLogPage` — 审计日志（admin）

Page-header：面包屑 "System · Compliance"，标题"审计日志"，副文"全量 API 操作记录与合规追溯"。右侧过滤条：方法 Select + 用户 Input + 时间范围 DateRangePicker。

**表格卡**：列 时间（小字变体色）/ 用户（admin#xxx 缩写）/ 方法（badge：GET 蓝 POST 绿 PUT 橙 DELETE 红）/ 接口（mono code，超长省略，Tooltip 原文）/ 资源类型 / 资源 ID（mono 8 位截断）/ 结果（成功绿/失败红 badge）/ IP（mono 小字）。分页 showTotal。

### 12. `SystemSettingsPage` — 系统设置（admin）

Page-header：面包屑 "System › Settings"，标题"System Settings"，副文"Platform configuration & health. 平台级配置、健康状态与集成监控"。

最大宽 900。纵向 3 张卡：
- **平台配置**：Form 三字段 K8s API Server 地址 / 默认镜像仓库 / ArgoCD 地址，底部"保存配置"按钮。
- **健康状态**：头部 extra "刷新"按钮；总体状态 Badge（绿 ok / 黄 degraded / 红 fail）+ Descriptions 列出各项 check 子项状态。
- **集成状态**：每行 name + Badge + detail（mono 小字描述）。

---

## 项目子页（嵌 ProjectLayout）

### 13. `EnvironmentListPage` — 环境

Page-header：面包屑 "Environments"，标题"Environment Health"，副文"Deployment status & rollback. 监控和管理部署环境状态"。右侧 primary "New Environment"（admin）。

**3 列指标**：Total Environments / Active Services / Health Rate（百分比）。

**环境卡片网格**：按 health 类型染色边框与状态徽标：Healthy ✓ 绿 / Syncing ↻ 蓝 / Degraded ! 橙 / Down ✕ 红。内容：环境名 + 状态徽 + 键值 meta（Namespace 用 mono tag 圆角 pill / Description / Created 日期）。admin 显示删除按钮。空态+创建按钮。新建弹窗：环境名、Namespace、描述。

### 14. `DeploymentListPage` — 部署列表

Page-header：面包屑 "Project · Delivery"，标题"部署管理"，副文"触发与追踪 ArgoCD 部署同步状态"。右侧 primary "触发部署"。

**表格卡**：列 镜像（mono code）/ 环境 / 状态 badge（待部署灰 / 同步中蓝 / 健康绿 / 异常橙 / 失败红 / 已回滚灰）/ 同步状态 / 健康状态 / 部署人 / 时间 / 详情。分页。触发部署弹窗：环境 Select、镜像 Input、流水线运行 ID（可选）。

### 15. `DeploymentDetailPage` — 部署详情

顶部"返回列表"。Page-header：标题"部署详情"，副文镜像 mono code。右侧按钮组：刷新状态 / 重新同步 / 回滚（warning 色，带 Popconfirm）。

**基本信息卡**：头部 section title + 右侧状态大 badge。Descriptions 单列：ID / 镜像 / 环境 ID / 同步状态 / 健康状态 / ArgoCD 应用 / 部署人 / (错误信息) / 开始时间 / 完成时间 / 创建时间。

### 16. `ServiceListPage` — 服务

Page-header：面包屑 "Project · Services"，标题"服务管理"，副文"管理项目下的微服务与源码仓库绑定"。admin 可见 primary "新建服务"。

**表格卡**：列 服务名 / 描述 / 仓库地址（mono code）/ 创建时间 / 删除。分页。新建弹窗：服务名、描述、仓库地址。

### 17. `MemberListPage` — 成员

Page-header：面包屑 "Project · Access"，标题"成员管理"，副文"分配项目成员与角色权限"。admin 可见 primary "添加成员"。

**表格卡**：列 用户名 / 角色（admin 显示内联 Select 可直接切换；非 admin 显示徽标）/ 加入时间 / 移除。添加弹窗：用户 ID + 角色 Select。

### 18. `VariableListPage` — 项目变量

Page-header：面包屑 "Project › Variables"，标题"Project Variables"，副文"Secure variable management for the project. 项目级环境变量与密钥管理"。管理员可见 "+ Add Variable"。表格结构同全局变量页。

### 19. `NotificationRulesPage` — 通知规则

Page-header：面包屑 "Project · Signals"，标题"通知规则"，副文"配置构建与部署事件的 Webhook 推送"。右侧 primary "创建规则"。

**表格卡**：列 名称 / 事件类型 badge（构建成功绿 / 构建失败红 / 部署成功蓝 / 部署失败橙）/ Webhook URL（mono 省略）/ 启用 Switch / 创建时间 / 操作（编辑 + 删除）。空态。表单弹窗：名称 + 事件类型 Select + Webhook URL（校验 https://）+ 启用 Switch。

---

## 流水线模块

### 20. `PipelineListPage` — 流水线列表

Page-header：面包屑 "Pipelines"，标题"Automated Pipelines"，副文"Real-time status of your deployment infrastructure..."。右侧 primary "Create Pipeline"。

**4 列可点指标卡**（点击切换状态筛选）：Total Pipelines / Active / Draft / Disabled。

**搜索过滤栏**：容器浅灰背景，左搜索 input（search icon 内嵌，focus 发光描边）；右 pill 形 Tag 组："全部/已启用/草稿/已停用"，active 项 primary gradient 填充。

**流水线行列表**（纵向 gap 8）：每行 — 左 状态大圆 icon（✓ / ⏸ / 📝 上色）+ 中间 名称 + 次要 "触发类型 · 描述 · 更新时间" + 右侧状态小徽章带彩色圆点 + 操作组（运行 primary icon / 运行历史 outline icon / 更多下拉 [编辑/复制] / 删除 icon）。空态+创建按钮。带 RunPipelineModal 触发弹窗。

### 21. `TemplateSelectPage` — 新建流水线向导

最大宽 960 居中。顶部返回文本按钮。Page-header：面包屑 "Create Pipeline"，标题"Architect New Pipeline"，副文"Real-time topology preview. 从模板快速创建流水线，或从空白开始构建"。

**步骤条**：2 步 — Template Selection / Configuration，active 高亮编号圆，完成步变✓。

**Step 1 模板网格**（4 列 span 6，高 180）：首卡"Custom ＋ 从零开始配置"直跳 /pipelines/blank。其他卡：Go 🔵 / Java Maven ☕ / Node.js 🟢 / Docker 🐳 / Java JAR ☕。

**Step 2 双栏**（左 14/24，右 10/24）：
- 左 config-panel "Configuration Parameters"：流水线名称（必填）+ 描述；分隔线 + "Template Parameters" 段，双列 Form 填模板参数（repoUrl 跨 24 span）。底部"创建流水线"primary + "返回"。
- 右 config-panel "Real-time Stage Preview"：横向节点链（编号圆 + 名 + step 数 + 中间连线），下方 mono 黑底 JSON 配置预览。

### 22. `PipelineEditorPage` — 流水线编辑器（全屏）

全屏 fixed inset 0，不走 AppLayout。顶部 56px 玻璃态 header（backdrop-blur）：返回箭头 + 竖分隔 + 32 小渐变 icon 方块（code icon）+ 流水线名称 Input（无边框，display 字体 17px bold，内联可编辑）+ 右侧描述小字；中间 `ModeSwitch` 可视化/JSON 双档；右侧"设置"文本按钮（打开右抽屉 PipelineSettingsPanel，配置 triggerType / concurrencyPolicy / 描述）+ JSON 模式下的 primary "保存"。

**主画布区**：
- **可视化模式**：横向 DAG 编辑器。画布灰底栅格，stage 节点横排，stage 间连线箭头；每 stage 内纵向堆叠 step 节点（git-clone / shell / kaniko-build / buildkit-build，各 icon+名）；stage 头部双击重命名；支持添加 stage / step，step 点击右侧打开配置面板（shell 使用暗色代码编辑器）；底部浮动保存按钮。
- **JSON 模式**：Monaco 风 YAML/JSON 编辑器占满，校验错误阻止保存。

切到别页前 dirty 警告 confirm。

### 23. `PipelineRunListPage` — 运行历史

Page-header：左侧返回箭头 + "xxx - 运行历史"标题。右侧 ListFilters：Commit SHA 搜索 / 状态 Select（待执行/排队中/运行中/成功/失败/已取消）/ 触发方式 Select（手动/Webhook/定时）+ primary "触发运行"。

**表格**（stripe + hover，圆角 8）：列 #（runNumber）/ 状态（pill 彩点徽：pending 灰 / queued 蓝 / running 青 / succeeded 绿 / failed 红 / cancelled 橙）/ 触发方式 / 触发人 / 分支（branch icon + 名）/ 耗时（格式化 m s）/ 时间 / 操作（详情 outline + 进行中可取消 danger outline）。分页 showTotal。

### 24. `PipelineRunDetailPage` — 运行详情

最大宽 1000。顶部返回 icon + 面包屑 "Build Observation" + 大标题 "Build #{runNumber}" + 副文 "triggeredBy · duration" + 右侧大状态徽章。进行中可见"取消运行"按钮。

**Stage 进度条**：横向 4 阶段 Source Checkout / Build Artifacts / Image Push / K8s Deployment，每阶段 dot + label + 状态文本 Completed/Active/Pending，active 发光 pulse。

**4 列信息小卡**：TRIGGER / BRANCH / TRIGGERED BY / COMMIT（mono）。

**Run Details config-panel**：2 列 key-value — 开始时间 / 结束时间 / Tekton Name（mono）/ Namespace（mono）；运行参数用 pill tag 展示 `k=v`（mono）；错误信息用 error-container 红背景 mono 代码块。

**Build Output 终端块**：头部黑色条 "Build Output" + macOS 三色圆点；主体深色终端：每行 行号 + 日志内容（level error 红 / info 蓝），等宽字体。无数据时居中 `>_` + "Waiting for build output..."。

**Build Artifacts 区**（有才显示）：pill tag 展示工件名 + KB 大小，可点击下载。

---

## 通用组件

### 25. 表单弹窗（公共模式）

用于新建/编辑 项目 / 用户 / 变量 / 连接 / 环境 / 服务 / 成员 / 通知规则 / 触发部署。Arco Modal：标题 + Form 垂直布局 + 字段圆角 8 Input / TextArea / Select / Switch / Password，底部"确定"primary + "取消"text。validation 显示行内红字。
