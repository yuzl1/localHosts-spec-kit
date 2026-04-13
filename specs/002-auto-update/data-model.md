# Data Model: LocalHosts 检查更新与自动更新

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/spec.md)  
**Research**: [research.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/research.md)  
**Created**: 2026-04-13

## 概览

更新能力由三部分组成：更新设置（是否启动时检查）、更新检查结果（最新版本与说明）、更新会话状态（下载与安装的进度与阶段）。

## Entities

### UpdateSettings

用户目录持久化的更新设置。

字段：

- `autoCheckOnStartup: bool`：启动时是否自动检查更新

### UpdateInfo

一次检查更新得到的“最新发布信息”。

字段：

- `currentVersion: string`：当前运行版本号（语义化版本）
- `latestVersion: string`：远端最新版本号（语义化版本）
- `hasUpdate: bool`：是否存在更新
- `releaseNotes: string`：更新说明（用于展示）
- `publishedAt: string`：发布时间（ISO 时间戳）
- `assetName: string`：匹配到的安装包名称
- `assetURL: string`：安装包下载地址
- `assetSize: number`：安装包大小（字节）

### UpdateProgress

下载/安装过程的进度与阶段状态（用于 UI 展示）。

字段：

- `stage: string`：`idle` | `checking` | `available` | `downloading` | `downloaded` | `installing` | `done` | `error`
- `downloadedBytes: number`
- `totalBytes: number`
- `percent: number`
- `message: string`：面向用户的简短状态信息

## Relationships

- UpdateSettings 影响启动时是否触发一次检查更新。
- UpdateInfo 是“检查更新”的输出，也是“下载/安装”的输入来源。
