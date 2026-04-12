# Quickstart: LocalHosts 开发与验证

**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/spec.md)  
**Created**: 2026-04-12

## 运行开发环境

在 Wails 项目目录执行：

```bash
cd LocalHosts
wails dev
```

## 构建

```bash
cd LocalHosts
wails build
```

## 权限验证要点

### macOS / Linux

- 读取 `/etc/hosts`：无需管理员权限
- 写入 `/etc/hosts`：需要管理员授权流程；验证点：
  - 取消授权：系统 hosts 不变
  - 授权成功：仅 LocalHosts 管理块被更新，块外内容保持原样

### Windows

- 读取 `C:\Windows\System32\drivers\etc\hosts`：通常无需管理员权限
- 写入该文件：需要管理员权限；验证点同上

## 数据落盘（用户目录）

工作区（分组、条目、组装开关）需要写入用户目录以便恢复编辑状态。验证点：

- 不写系统 hosts 的情况下退出再打开，工作区仍能恢复
- 变更分组顺序、开关与条目状态后，组装结果可预测

## 本地验证用例（最小集）

1. 系统 hosts 无管理块：应用保存后文件末尾出现完整管理块，块外内容不变
2. 纯注释行保持：块外 `#` 注释行与空行在保存前后完全一致
3. 禁用条目识别：`# 127.0.0.1 example.com` 被识别为禁用条目并可开关；`# some note` 不可开关
4. 多域名拆分：`127.0.0.1 a.local b.local` 在列表中拆为两条并能分别启用/禁用，保存后写为两行

