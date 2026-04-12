# Implementation Plan: LocalHosts Hosts 管理（跨平台 + 分组组装 + 管理块写回）

**Branch**: `001-hosts-manager` | **Date**: 2026-04-12 | **Spec**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/spec.md)  
**Input**: Feature specification from `/specs/001-hosts-manager/spec.md`

## Summary

实现一个轻量级 Hosts 管理工具：启动时读取系统 hosts，展示“块外只读 + 管理块可编辑”的统一视图；支持条目增删改、启用/禁用；支持分组与“是否按分组组装”开关；保存时仅更新 LocalHosts 标记块并通过管理员授权写入系统 hosts，块外内容原样保留。

设计产物：

- [research.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/research.md)
- [data-model.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/data-model.md)
- [contracts/hosts-block-format.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/contracts/hosts-block-format.md)
- [contracts/wails-api.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/contracts/wails-api.md)
- [quickstart.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/quickstart.md)

## Technical Context

**Language/Version**: Go 1.23+（go.mod 当前 1.23）  
**Primary Dependencies**: Wails v2.12.x（Go 侧绑定），Vue 3 + Vite（前端构建）  
**Storage**: 用户目录配置文件（JSON），不使用数据库  
**Testing**: `go test`（解析/组装/块定位等纯函数单测优先）  
**Target Platform**: macOS + Windows + Linux  
**Project Type**: 桌面应用（Wails + Vue）  
**Performance Goals**: 冷启动 < 1s，内存 < 30MB  
**Constraints**: 纯本地、无网络、无数据库；写回系统 hosts 需管理员授权；块外内容原样保留  
**Scale/Scope**: 典型 hosts 文件（百行级），支持多分组与百级条目

## Constitution Check

*GATE: Phase 0 前必须通过；Phase 1 设计后复查。*

- 纯本地/离线：不引入网络能力，不依赖任何在线服务
- 技术栈强制：Go + Wails v2；Vue 3 + Vite + JavaScript
- 平台适配：macOS/Windows/Linux 的 hosts 路径与权限模型均纳入
- 路径不硬编码：系统 hosts 路径与用户目录路径通过统一解析层获得
- 权限最小化：读取普通权限；写入必须管理员授权；失败/取消不修改系统文件
- 可回滚：写入前备份，写入失败恢复
- 代码标准：无 TODO、无冗余注释、纯函数优先、单一出口、错误不忽略
- 轻量目标：不引入常驻重依赖与后台服务

结论：通过（本计划所有决策均与 [constitution.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/.specify/memory/constitution.md) 对齐）。

## Project Structure

### Documentation (this feature)

```text
specs/001-hosts-manager/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── hosts-block-format.md
│   └── wails-api.md
└── tasks.md
```

### Source Code (repository root)

```text
LocalHosts/
├── app.go
├── main.go
├── internal/
│   ├── hosts/
│   │   ├── model.go
│   │   ├── parse.go
│   │   ├── render.go
│   │   └── block.go
│   ├── platform/
│   │   ├── paths.go
│   │   ├── elevate_darwin.go
│   │   ├── elevate_linux.go
│   │   └── elevate_windows.go
│   └── workspace/
│       ├── model.go
│       └── store.go
└── frontend/
    └── src/
        ├── App.vue
        ├── main.js
        ├── components/
        └── services/
            └── hosts.js
```

**Structure Decision**: 采用现有 Wails 单项目结构，在 `LocalHosts/internal/` 下按域划分（hosts/、workspace/、platform/），前端通过 `wailsjs` 调用 Go 绑定方法。
