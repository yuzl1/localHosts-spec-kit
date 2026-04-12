<!--
Sync Impact Report
Version: 1.3.0 -> 1.3.1
Modified principles: 4 (敏感信息处理口径澄清：禁止记录/上传，UI 本地展示允许)
Added sections: None
Removed sections: None
Templates requiring updates:
- ✅ .specify/templates/plan-template.md (verified; no change)
- ✅ .specify/templates/spec-template.md (verified; no change)
- ✅ .specify/templates/tasks-template.md (verified; no change)
Deferred items: None
-->

# LocalHosts Constitution

## Core Principles

### 1. 纯本地与离线优先（Non-Negotiable）
产品必须纯本地运行：不联网、无数据库、无云同步；所有数据与状态均在本机文件与内存中管理。
用户配置与分组数据必须存储在用户目录中，用于最终组装生成 hosts 文件内容。

### 2. 技术栈与平台边界（强制）
- 后端必须使用 Go 1.23+（以 go.mod 为准）与 Wails v2
- 前端必须使用 Vue 3 + Vite + JavaScript
- 目标平台为 macOS、Windows 与 Linux，必须适配各自 hosts 文件路径与权限模型
  - macOS: /etc/hosts
  - Windows: C:\Windows\System32\drivers\etc\hosts
  - Linux: /etc/hosts

### 3. 代码标准：无冗余注释、无 TODO、纯函数优先、单一出口
- 代码中禁止无意义注释与 TODO
- Go 的导出结构体/方法必须有简短文档注释（为公开 API 与静态检查服务）
- 逻辑优先以纯函数表达；副作用集中在边界层（文件系统、Wails 运行时、UI 事件）
- 每个函数保持单一出口；所有错误必须显式处理，不允许静默忽略
- 路径不得硬编码：统一通过变量/配置/常量集中管理

### 4. 安全与权限最小化
- 读写系统 hosts 文件必须走明确的权限路径：读取可在普通权限下完成；保存写入必须通过管理员授权流程（macOS/Linux 为 sudo）
- 不记录或上传任何敏感信息（包括 hosts 中可能包含的内网域名与注释内容）；UI 展示仅限本机用户主动查看
- 任何涉及系统文件写入的操作必须具备可回滚策略（例如保存前备份与失败时恢复）

### 5. 轻量与交付目标（强约束）
- 运行内存占用目标小于 30MB
- 冷启动目标小于 1 秒
- 交付形态为单文件应用（单一 app）

## 功能范围

- 读取系统 hosts 文件（macOS、Windows、Linux 的目录都要适配）并解析为可编辑条目
- 新增/编辑/删除 hosts 条目
- 切换条目启用/禁用（通过注释/取消注释实现）
- 保存到系统 hosts 文件（macOS/Linux 需要 sudo；Windows 需要管理员权限）
- 不引入网络功能、不引入数据库、纯本地运行
- 需要在用户目录存储配置与分组，并支持按分组最终组装生成 hosts 文件内容

## 工程规范

### Go
- 结构体首字母必须大写
- 方法必须有简短文档注释
- 所有错误必须处理，不允许忽略
- 路径不硬编码，使用变量统一管理

### Vue 3
- 使用 Composition API
- 组件名使用 PascalCase
- 样式必须 scoped
- 与 Go 通信必须使用 Wails runtime

## Governance
- Constitution 对本项目所有开发活动具有最高约束力；与其冲突的约定以本文件为准。
- 任何新增/修改原则与边界必须同步更新本文件，并按语义化版本（MAJOR.MINOR.PATCH）递增：
  - MAJOR：删除/重定义原则或边界（破坏性治理变化）
  - MINOR：新增原则/新增章节/显著扩展约束
  - PATCH：澄清措辞与小幅补充（不改变治理含义）
- 每次变更必须说明影响范围与迁移方式（若适用），并在评审中显式检查对技术栈、权限流程与性能目标的影响。

**Version**: 1.3.1 | **Ratified**: 2026-04-12 | **Last Amended**: 2026-04-12
