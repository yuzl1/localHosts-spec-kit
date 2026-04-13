# Research: LocalHosts 检查更新与自动更新

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/spec.md)  
**Created**: 2026-04-13

## 结论概览

- 更新源：以 GitHub Releases 为默认发布源（latest release）
- 版本判断：以语义化版本号比较（忽略 tag 前缀 `v`）
- 资产选择：按 OS/ARCH 选择对应安装包（macOS 使用 universal；Windows 分 amd64/arm64）
- 交互：提供手动“检查更新”；启动时自动检查作为可配置项
- 安装：下载完成后进入安装阶段并触发管理员授权；拒绝授权则中止更新且不影响现有版本

## 更新源与版本获取

### 决策

- 使用“发布源的最新版本元数据”作为更新判断依据，并展示更新说明（release notes）。

### 理由

- 便于与现有发布流程对齐（CI 产出安装包并上传到 Releases）。

## 版本号比较策略

### 决策

- 将“当前版本”与“最新版本”都归一化为 `MAJOR.MINOR.PATCH` 形式进行比较：
  - 支持 tag/version 前缀 `v`
  - 非法或缺失字段视为不可比较，并返回可读错误

### 理由

- 避免因字符串比较导致的错误（例如 `0.0.10` 与 `0.0.2`）。

## 资产（安装包）选择规则

### 决策

- 根据运行环境选择单个“最佳匹配资产”：
  - macOS：优先 `LocalHosts-macos-universal.zip`，次选 `LocalHosts-macos-universal.dmg`
  - Windows：`LocalHosts-windows-amd64-setup.exe` 或 `LocalHosts-windows-arm64-setup.exe`

### 理由

- 与当前 CI 命名保持一致，前端无需额外配置即可完成选择与下载。

## 下载与进度反馈

### 决策

- 下载过程提供实时进度（已下载字节、总字节、百分比），并允许失败后重试。

### 理由

- 安装包体积较大时，必须提供可见反馈，避免用户误以为卡死。

## 安装与管理员授权

### 决策

- 安装阶段触发管理员授权：
  - Windows：运行安装程序并让系统弹出 UAC
  - macOS：若更新目标路径需要写权限（例如 `/Applications`），触发系统权限验证（如 Touch ID/密码）

### 理由

- 与现有“写入系统 hosts 需要管理员授权”的权限策略一致，且符合用户对系统级变更的预期。

## 隐私与安全

### 决策

- 更新请求与日志不包含任何用户 hosts 内容、工作区数据或系统路径细节（仅记录必要的错误码与简要信息）。
- 更新失败不应破坏现有版本：下载与安装分离，安装失败不覆盖现有可用版本。

### 理由

- 满足“敏感信息不记录/上传”的约束，并确保更新过程可恢复与可重试。
