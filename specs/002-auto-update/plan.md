# Implementation Plan: LocalHosts 检查更新与自动更新

**Branch**: `002-auto-update` | **Date**: 2026-04-13 | **Spec**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/spec.md)  
**Input**: Feature specification from `/specs/002-auto-update/spec.md`

## Summary

增加“检查更新 + 自动更新”能力：支持手动检查更新、启动时自动检查更新（可配置）、发现新版本后展示版本号与更新说明；用户确认后下载与安装最新版本，并在安装阶段触发管理员授权（Windows UAC / macOS 系统权限验证），最后重启应用进入新版本。

设计产物：

- [research.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/research.md)
- [data-model.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/data-model.md)
- [contracts/update-api.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/contracts/update-api.md)
- [quickstart.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/quickstart.md)

## Technical Context

**Language/Version**: Go 1.23+（go.mod 当前 1.23）  
**Primary Dependencies**: Wails v2.12.x（Go 侧绑定），Vue 3 + Vite（前端构建）  
**Storage**: 用户目录配置文件（JSON），更新包临时缓存（文件）  
**Testing**: `go test`（版本比较、资产选择等纯函数单测优先）  
**Target Platform**: macOS + Windows（Linux 更新流程可后续补齐）  
**Project Type**: 桌面应用（Wails + Vue）  
**Performance Goals**: 冷启动 < 1s，内存 < 30MB；更新检查不阻塞 UI  
**Constraints**: 更新过程中不记录/上传敏感信息；安装阶段需要管理员授权；失败可重试且不影响现有版本继续使用  
**Scale/Scope**: Release 元数据（单次请求），安装包（几十 MB 级）

## Constitution Check

*GATE: Phase 0 前必须通过；Phase 1 设计后复查。*

- 技术栈强制：Go + Wails v2；Vue 3 + Vite + JavaScript
- 平台适配：更新资产选择需按 OS/ARCH 匹配（至少覆盖 macOS universal 与 Windows amd64/arm64）
- 安全与权限最小化：安装/替换需要管理员授权；拒绝授权必须安全退出更新流程
- 不记录/上传敏感信息：更新请求与日志不得包含 hosts 内容或用户配置内容
- 轻量目标：不引入常驻后台服务；下载/安装流程仅在用户触发或设置开启时运行

结论：存在原则冲突（constitution 中“纯本地与离线优先”限制网络）。本特性需要联网获取版本与下载更新包；将通过“仅用于更新、默认可关闭、不上传任何用户数据”的约束来降低风险，并在 Complexity Tracking 中记录。

## Project Structure

### Documentation (this feature)

```text
specs/002-auto-update/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── update-api.md
└── tasks.md
```

### Source Code (repository root)

```text
LocalHosts/
├── app.go
├── main.go
├── internal/
│   ├── platform/
│   ├── workspace/
│   └── update/
└── frontend/
    └── src/
        ├── App.vue
        └── wailsjs/
```

**Structure Decision**: 在现有 Wails 单项目结构下，新增 `LocalHosts/internal/update/` 承载更新检查、下载与安装的业务逻辑；前端通过 Wails 绑定方法驱动流程，并通过事件接收下载进度与状态变化。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| 离线优先：不联网 | 检查更新与下载更新包需要网络访问 | 仅提供“打开发布页手动下载”无法满足“自动更新”需求 |
