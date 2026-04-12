# Data Model: LocalHosts Hosts 管理

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/spec.md)  
**Research**: [research.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/research.md)  
**Created**: 2026-04-12

## 概览

系统将 hosts 内容分为“块外内容（只读保留）”与“LocalHosts 管理块（可编辑与写回）”。应用内以工作区（Workspace）为主进行编辑，并持久化到用户目录。

## Entities

### HostEntry

表示一个可编辑 hosts 条目（一个域名/别名对应一个条目）。

字段：

- `id: string`：稳定标识（用于编辑、排序与去重）
- `ip: string`：IPv4/IPv6 文本
- `domain: string`：域名或别名（单个 token，不含空白）
- `comment: string`：备注（不含换行）
- `enabled: bool`：条目开关（true 输出为非注释；false 输出为整行注释）
- `groupId: string`：所属分组
- `source: string`：来源（例如 `managed-block`）
- `raw: string`：读取时的原始行片段（用于诊断与尽量保留格式，但不承诺 100% 原样）

校验规则：

- `ip` 必须为合法 IPv4/IPv6 表达
- `domain` 必须为非空且不包含空白
- `comment` 不允许包含换行

状态：

- enabled = true：写回时为 `ip domain [# comment]`
- enabled = false：写回时为 `# ip domain [# comment]`

### HostGroup

表示条目分组。

字段：

- `id: string`：稳定标识
- `name: string`：展示名称
- `enabled: bool`：分组开关（仅在 composeEnabled=true 时影响组装）
- `order: int`：分组输出顺序（升序）

校验规则：

- `name` 必须非空
- `order` 在同一工作区内唯一

### WorkspaceState

用户目录中的持久化工作区。

字段：

- `version: int`：数据结构版本号
- `composeEnabled: bool`：是否按分组组装
- `groups: HostGroup[]`
- `entries: HostEntry[]`
- `updatedAt: string`：ISO 时间戳

约束：

- 每个 `HostEntry.groupId` 必须引用存在的分组
- entries 的顺序用于写回输出的稳定顺序（不隐式排序）

### HostDocument（内存模型）

表示一次从系统 hosts 读取到的完整文档结构，用于保存时重建输出。

字段：

- `prefixLines: string[]`：管理块之前的所有原始行（含换行符的逐行文本）
- `managedBlock: string[]`：管理块边界行与内容（逐行文本，保存时将被重建）
- `suffixLines: string[]`：管理块之后的所有原始行（逐行文本）

## Relationships

- WorkspaceState 1..N HostGroup
- WorkspaceState 0..N HostEntry
- HostGroup 1..N HostEntry（通过 groupId）

## 组装与写回

### 组装输入集

- 若 `composeEnabled = true`：仅组装 `groups.enabled = true` 的分组
- 若 `composeEnabled = false`：组装所有分组

### 组装输出

按 `HostGroup.order` 输出分组；分组内按 entries 的稳定顺序输出条目。输出格式见 [hosts-block-format.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/contracts/hosts-block-format.md)。

