# Story 4.4: 前端变量管理页面

Status: done

## Story

As a 项目管理员,
I want 在界面上管理变量和密钥,
so that 不需要命令行即可配置变量。

## Acceptance Criteria

1. **项目变量列表**
   - Given 用户进入项目
   - When 访问 /projects/:id/variables
   - Then 显示变量列表，密钥类型值显示为 ******

2. **创建变量**
   - Given 项目管理员或管理员
   - When 点击"新建变量"
   - Then 弹出表单，可选类型（普通/密钥），创建成功后刷新列表

3. **编辑变量**
   - Given 普通变量存在
   - When 编辑变量值
   - Then 变量更新成功，页面显示 Message 成功提示

4. **密钥变量不可编辑值**
   - Given 密钥变量存在
   - When 查看密钥变量
   - Then 值显示为 ******，可更新描述但需重新输入值

5. **普通成员视角 (FR5)**
   - Given 普通成员进入变量页面
   - When 列表返回
   - Then 密钥变量不显示（后端已过滤）

6. **全局变量管理 (admin)**
   - Given 管理员已登录
   - When 访问 /admin/variables
   - Then 显示全局变量列表，支持 CRUD

## Tasks / Subtasks

- [x] Task 1: 创建 services/variable.ts API 服务层
- [x] Task 2: 创建项目变量管理页面 (VariableListPage)
- [x] Task 3: 创建变量表单 Modal (VariableFormModal)
- [x] Task 4: 创建全局变量管理页面 (AdminVariablePage)
- [x] Task 5: 在 App.tsx 注册路由
- [x] Task 6: 在 ProjectLayout 侧边栏添加"变量"导航
- [x] Task 7: 在 AppLayout 侧边栏添加"全局变量"导航 (admin)
- [x] Task 8: 权限扩展 (route:admin-variables:view)

## Dev Notes

### Source tree components to touch
- `web/src/services/variable.ts` (新建)
- `web/src/pages/projects/variables/VariableListPage.tsx` (新建)
- `web/src/pages/projects/variables/VariableFormModal.tsx` (新建)
- `web/src/pages/admin/variables/AdminVariablePage.tsx` (新建)
- `web/src/App.tsx` (修改)
- `web/src/pages/projects/ProjectLayout.tsx` (修改)
- `web/src/components/layout/AppLayout.tsx` (修改)
- `web/src/stores/auth.ts` (修改 - 权限扩展)

## Dev Agent Record
### Agent Model Used
Claude Opus 4.6
### Completion Notes List
- 新建 services/variable.ts API 层，封装项目/全局变量 API
- VariableListPage 支持列表查看、新建、编辑、删除
- VariableFormModal 支持普通/密钥类型选择
- AdminVariablePage 全局变量管理（仅 admin 可见）
- ProjectLayout 侧边栏添加"变量"入口
- AppLayout 侧边栏添加"全局变量"入口（admin only）
- stores/auth.ts 扩展 route:admin-variables:view 权限
- 密钥变量值在前端显示为 ******
- 全部 16 个前端测试通过，无回归

### File List
- `web/src/services/variable.ts`
- `web/src/pages/projects/variables/VariableListPage.tsx`
- `web/src/pages/projects/variables/VariableFormModal.tsx`
- `web/src/pages/admin/variables/AdminVariablePage.tsx`
- `web/src/App.tsx`
- `web/src/pages/projects/ProjectLayout.tsx`
- `web/src/components/layout/AppLayout.tsx`
- `web/src/stores/auth.ts`
