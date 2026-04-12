# LocalHosts 🌐

轻量级的跨平台 Hosts 管理桌面应用，支持 macOS、Windows 和 Linux。核心功能类似于 SwitchHosts，但在资源占用和启动速度上更加极致。

## ✨ 特性 (Features)

- **跨平台支持**：全面兼容 macOS、Windows、Linux。
- **分组管理**：支持将 Hosts 记录进行分组，可单独开启或关闭整个分组，支持拖拽排序（通过上下箭头）。
- **纯文本极速编辑**：抛弃繁琐的表格表单，直接通过纯文本编辑器修改 Hosts，支持语法高亮与智能解析。
- **无损修改系统 Hosts**：
  - 仅管理特定区块（`#######分组名开始####` 与 `#######分组名结束####`），**绝对不破坏**、不修改系统 hosts 文件中原有的内容、注释或空行。
  - 自动将文本解析合并，并在写入时严格保证“一域名一行”方便独立开关。
  - 空分组自动忽略，不会在系统 hosts 中产生冗余标签。
- **一次授权，全局复用**：写入系统文件时自动请求管理员权限（macOS `osascript` / Windows `UAC` / Linux `pkexec`），单次应用运行期间只需授权一次，后续保存自动复用提权会话，告别频繁弹窗。
- **实时自动保存**：编辑内容防抖自动同步至用户配置目录工作区，并可无缝应用到系统 hosts。
- **极低资源占用**：纯本地应用，无网络请求、无数据库依赖。典型启动时间 < 1s，空闲常驻内存 < 30MB。

## 🛠 技术栈 (Tech Stack)

- **后端**：[Go](https://golang.org/) (1.23+) + [Wails v2](https://wails.io/)
- **前端**：[Vue 3](https://vuejs.org/) (Composition API) + [Vite](https://vitejs.dev/) + 原生 CSS

## 🚀 快速开始 (Quick Start)

### 环境依赖
- Go 1.23 及以上版本
- Node.js & npm (用于构建前端)
- Wails CLI

安装 Wails CLI：
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 开发模式 (Development)
```bash
# 进入应用主目录
cd LocalHosts

# 启动开发服务器（包含前端热重载与 Go 后端服务）
wails dev
```

### 构建打包 (Build)
```bash
# 进入应用主目录
cd LocalHosts

# 构建对应平台的独立可执行程序
wails build
```
构建产物将输出在 `LocalHosts/build/bin/` 目录下。

## 📁 目录结构 (Project Structure)

```text
.
├── LocalHosts/             # Wails 桌面应用主目录
│   ├── app.go              # Go 后端入口与前端 API 绑定接口
│   ├── frontend/           # Vue 3 前端源码目录
│   ├── internal/
│   │   ├── hosts/          # Hosts 文本解析、区块合并与生成核心逻辑
│   │   ├── platform/       # 跨平台路径获取与原生提权 (Elevate) 会话机制
│   │   └── workspace/      # 本地工作区状态 (JSON) 的持久化管理
│   └── main.go             # Wails 程序启动入口
└── specs/                  # 基于 Spec Kit 的需求分析与特性设计文档
```

## 📝 核心安全机制 (Core Mechanisms)

1. **会话提权驻留 (Elevated Session)**：
   应用首次请求写入系统 hosts 时，会通过各平台的提权指令拉起一个临时的 `helper` 进程。该进程监听本地随机端口，并使用基于随机 Token 的验证机制进行安全通信。主应用后续的系统 hosts 写入均通过与 Helper 通信完成，实现“运行期间只需授权一次”。
2. **防覆盖策略 (Safe Rollback)**：
   每次执行覆盖写入前，都会先将系统最新的 hosts 文件备份为 `.localhosts.bak`。同时写入操作会首先读取当前最新 hosts 文件内容，仅针对 `LocalHosts START` 和 `END` 管理块内的文本进行定点替换，最大程度防止其他程序对系统 hosts 造成的外部修改被意外覆盖。

## 📄 License

MIT License
