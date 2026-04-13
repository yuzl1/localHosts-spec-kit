# Quickstart: LocalHosts 更新功能验证

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/002-auto-update/spec.md)  
**Created**: 2026-04-13

## 运行开发环境

在 Wails 项目目录执行：

```bash
cd LocalHosts
wails dev
```

## 手动检查更新（最小验证）

1. 启动应用，点击“检查更新”
2. 验证：
   - 已是最新版本：提示“当前已是最新版本”
   - 存在新版本：展示最新版本号与更新说明，并提供“立即更新”

## 下载进度验证

触发下载后验证：

- 进度条随下载推进更新（百分比、已下载/总大小）
- 下载失败（断网/超时）能够提示并允许重试

## 权限与安装验证要点

### Windows

- 点击“立即更新”后会启动安装程序并触发 UAC（如需要）
- 用户拒绝 UAC：更新流程结束，应用保持可用

### macOS

- 当更新目标路径需要写权限（例如 `/Applications`）时，会触发系统权限验证（Touch ID/密码）
- 用户拒绝授权：更新流程结束，应用保持可用

## 发布与资产命名约定（与 CI 对齐）

更新资产按以下命名上传到 Release：

- `LocalHosts-macos-universal.zip`（优先）
- `LocalHosts-macos-universal.dmg`（备选）
- `LocalHosts-windows-amd64-setup.exe`
- `LocalHosts-windows-arm64-setup.exe`
