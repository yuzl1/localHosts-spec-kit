# Feature Specification: Auto Update

**Feature Branch**: `002-auto-update`  
**Created**: 2026-04-13  
**Status**: Draft  
**Input**: User description: "增加功能 检查更新 并且自动更新"

## Clarifications

### Session 2026-04-13

- Q: 更新过程中的系统权限问题：Windows和macOS在覆盖应用本身时是否需要额外请求管理员权限，或当前权限已足够？ → A: 需要请求管理员权限（Windows 触发 UAC，macOS 触发系统权限验证，如 Touch ID/密码）。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 手动检查更新 (Priority: P1)

用户可以通过点击界面上的“检查更新”按钮，主动向服务器查询是否有新版本的应用可用。

**Why this priority**: 提供给用户主动获取最新版本的能力，是更新功能的基础。

**Independent Test**: Can be fully tested by clicking the check update button and seeing a dialog indicating either "Up to date" or "New version available".

**Acceptance Scenarios**:

1. **Given** 当前版本为最新, **When** 用户点击“检查更新”, **Then** 系统提示“当前已是最新版本”。
2. **Given** 存在更高版本, **When** 用户点击“检查更新”, **Then** 系统弹窗提示发现新版本，并展示版本号和更新日志。

---

### User Story 2 - 启动时自动检查更新 (Priority: P2)

每次启动应用时，系统会在后台静默检查是否有新版本。如果发现新版本，则在界面上通过非侵入式的方式（如弹窗或通知条）提醒用户。

**Why this priority**: 确保用户能够及时知道有修复或新功能发布，提升应用的活跃度和稳定性。

**Independent Test**: Can be fully tested by launching an older version of the app and verifying that an update notification appears automatically.

**Acceptance Scenarios**:

1. **Given** 存在更高版本, **When** 用户打开应用, **Then** 系统在后台检查并在主界面弹出发现新版本的提示。

---

### User Story 3 - 下载并自动安装更新 (Priority: P1)

当用户确认更新后，系统自动下载最新的安装包，并在下载完成后自动执行安装和重启应用的流程。

**Why this priority**: 满足“自动更新”的核心诉求，免去用户跳转浏览器手动下载和安装的繁琐步骤。

**Independent Test**: Can be fully tested by accepting an update, watching the download progress, and verifying that the app restarts into the new version.

**Acceptance Scenarios**:

1. **Given** 发现新版本并提示, **When** 用户点击“立即更新”, **Then** 系统开始下载更新包，并在界面展示下载进度。
2. **Given** 更新包下载完成, **When** 系统准备安装, **Then** 系统自动退出当前应用，执行更新安装，并重新启动应用到新版本。

### Edge Cases

- 网络断开或超时：在检查更新或下载过程中网络中断时，系统应提示网络错误并允许重试，而不应崩溃。
- API 速率限制：如果检查更新的接口（如 GitHub API）触发了限流，应妥善处理并提示稍后重试。
- 权限不足：在 macOS 或 Windows 上覆盖安装时，如果需要管理员权限，系统应正确请求权限或引导用户。
- 空间不足：下载更新包时若磁盘空间不足，系统应提示用户清理空间。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 能够对比当前运行版本与远程发布的最新版本号（语义化版本号）。
- **FR-002**: 系统 MUST 在界面提供一个明显的“检查更新”入口。
- **FR-003**: 系统 MUST 能够解析远程版本的更新日志（Release Notes）并展示给用户。
- **FR-004**: 系统 MUST 支持在后台下载更新文件，并实时反馈下载进度（百分比/大小）。
- **FR-005**: 系统 MUST 能够校验下载文件的完整性，防止安装损坏的包。
- **FR-006**: 系统 MUST 能够在下载完成后自动执行更新替换，并重启应用。
- **FR-007**: 系统 MUST 在执行“安装/替换当前应用”步骤前请求并获得管理员权限；若用户拒绝授权，系统 MUST 中止更新且不影响现有版本继续使用。

### Key Entities

- **UpdateInfo**: 包含最新版本号、发布时间、更新日志、下载链接等信息的实体。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 用户能在 3 次点击以内完成从收到更新提示到开始下载更新的操作。
- **SC-002**: 检查更新的响应时间在正常网络下不超过 3 秒。
- **SC-003**: 更新下载过程中的进度显示延迟不超过 1 秒，确保用户能明确感知下载状态。
- **SC-004**: 自动更新安装成功率达到 95% 以上，且更新后原有配置（如 hosts 数据、主题设置）不丢失。

## Assumptions

- 远程版本信息托管在 GitHub Releases 上，并可通过 GitHub API 获取。
- 更新包为完整的安装程序或可执行文件（macOS 为 .app 或 .dmg，Windows 为 .exe 安装包）。
- 用户网络环境能够正常访问 GitHub。
- "自动更新" 指的是自动下载和安装，但仍会先询问用户是否要开始更新，而不是在后台完全静默强行覆盖（以防打断用户工作）。
