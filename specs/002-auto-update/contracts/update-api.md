# Contract: Go ↔ Vue（Update）

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/spec.md)  
**Created**: 2026-04-13

## 数据类型

### UpdateSettings

| 字段 | 类型 | 说明 |
|------|------|------|
| autoCheckOnStartup | boolean | 启动时是否自动检查更新 |

### UpdateInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| currentVersion | string | 当前运行版本号（语义化版本） |
| latestVersion | string | 最新版本号（语义化版本） |
| hasUpdate | boolean | 是否存在更新 |
| releaseNotes | string | 更新说明 |
| publishedAt | string | 发布时间（ISO） |
| assetName | string | 匹配到的安装包名称 |
| assetURL | string | 安装包下载地址 |
| assetSize | number | 安装包大小（字节） |

### UpdateProgress

| 字段 | 类型 | 说明 |
|------|------|------|
| stage | string | idle/checking/available/downloading/downloaded/installing/done/error |
| downloadedBytes | number | 已下载字节 |
| totalBytes | number | 总字节（未知时为 0） |
| percent | number | 0..100 |
| message | string | 用户可读提示 |

## 事件

### update:progress → UpdateProgress

用于在下载/安装过程中向前端推送进度与阶段变更。

## 方法

### GetUpdateSettings() → UpdateSettings

读取用户目录持久化的更新设置。

### SaveUpdateSettings(settings: UpdateSettings) → void

保存更新设置到用户目录。

### GetAppVersion() → string

返回当前运行版本号（语义化版本）。

### CheckForUpdate() → UpdateInfo

检查是否有新版本可用，并返回最新版本信息与更新说明。

错误语义（示例）：

- 网络不可用/超时：返回错误
- 解析失败：返回错误
- 未找到匹配安装包：返回错误

### DownloadUpdate(update: UpdateInfo) → string

下载对应安装包到本机临时目录，并返回下载后的本地文件路径。下载过程中通过 `update:progress` 推送进度。

### InstallUpdate(installerPath: string) → void

触发安装流程：

- Windows：运行安装程序并退出当前应用
- macOS：根据目标路径触发管理员授权，完成替换后重启应用

错误语义：

- 用户拒绝授权：返回错误且不影响现有版本
- 安装启动失败：返回错误且不影响现有版本
