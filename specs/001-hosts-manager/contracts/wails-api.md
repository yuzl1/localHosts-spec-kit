# Contract: Go ↔ Vue（Wails Bindings）

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/spec.md)  
**Created**: 2026-04-12

## 数据类型

### HostEntry

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 稳定标识 |
| ip | string | IPv4/IPv6 文本 |
| domain | string | 单个域名/别名 token |
| comment | string | 备注 |
| enabled | boolean | true=启用，false=禁用（整行注释） |
| groupId | string | 所属分组 |
| raw | string | 原始行片段（用于诊断/尽量保留） |

### HostGroup

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 稳定标识 |
| name | string | 分组名称 |
| enabled | boolean | 分组开关（仅 composeEnabled=true 时影响组装） |
| order | number | 输出顺序 |

### HostWorkspace

| 字段 | 类型 | 说明 |
|------|------|------|
| composeEnabled | boolean | 是否按分组组装 |
| groups | HostGroup[] | 分组列表 |
| entries | HostEntry[] | 条目列表 |
| readonlySystemLines | string[] | 系统 hosts 中管理块外的只读行（逐行文本） |
| systemHostsLines | string[] | 系统 hosts 的完整内容（逐行文本，仅用于“全部”视图只读展示） |

## 方法

### GetHosts() → HostWorkspace

读取系统 hosts 与用户目录工作区，并返回合并后的工作区视图：

- `readonlySystemLines` 仅用于展示（只读）
- `groups/entries/composeEnabled` 用于编辑（可写）

### CanWriteHosts() → boolean

检查当前进程是否可直接写入系统 hosts（不触发授权弹窗）。用于启动时的权限预检。

### EnsureWriteAccess() → void

触发管理员授权流程（平台相关）。用于启动时在无写入权限时请求授权。

### SaveHosts(workspace: HostWorkspace) → void

将工作区持久化到用户目录（不写系统 hosts）。用于：

- 新增/编辑/删除/开关后的即时保存
- 下次启动恢复编辑状态

### ApplyHosts(workspace: HostWorkspace) → void

将工作区组装为管理块并写入系统 hosts（需要管理员授权）。规则见：

- [hosts-block-format.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/contracts/hosts-block-format.md)

错误语义：

- 授权失败/取消：返回错误且系统 hosts 不变
- 写入失败：返回错误且尝试回滚恢复原文件
