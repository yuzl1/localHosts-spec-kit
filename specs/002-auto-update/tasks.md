---

description: "Tasks for 002-auto-update"
---

# Tasks: LocalHosts 检查更新与自动更新

**Input**: Design documents from `/specs/002-auto-update/`  
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/update-api.md

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: 建立更新模块与持久化字段，确保前后端可调用

- [x] T001 Create update package skeleton in LocalHosts/internal/update/ (model.go, semver.go, github.go)
- [x] T002 Add update settings field to workspace model in LocalHosts/internal/hosts/model.go and persistence in LocalHosts/internal/workspace/store.go
- [x] T003 Wire update settings through existing save/load flows in LocalHosts/app.go (SaveHosts/GetHosts)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 核心能力（版本比较、Release 解析、资产选择、错误语义）完成后，用户故事才能开始

- [x] T004 Implement semantic version parsing/compare in LocalHosts/internal/update/semver.go
- [x] T005 Implement GitHub release fetch + JSON decoding in LocalHosts/internal/update/github.go
- [x] T006 Implement platform asset selection in LocalHosts/internal/update/selector.go (macOS universal, Windows amd64/arm64)
- [x] T007 Define UpdateInfo/UpdateProgress contracts in LocalHosts/internal/update/model.go to match specs/002-auto-update/contracts/update-api.md
- [x] T008 Add Wails runtime event emission helper in LocalHosts/internal/update/service.go (update:progress)

**Checkpoint**: 基础能力就绪（可在 Go 侧独立调用 CheckForUpdate 并得到稳定结果）

---

## Phase 3: User Story 1 - 手动检查更新 (Priority: P1) 🎯 MVP

**Goal**: 用户点击“检查更新”即可获得“已是最新/发现新版本 + 更新说明”的反馈

**Independent Test**: 启动应用 → 点击“检查更新” → 能看到结果弹窗（包含当前版本/最新版本/更新说明或错误提示）

### Implementation for User Story 1

- [x] T009 [US1] Expose GetAppVersion() in LocalHosts/app.go
- [x] T010 [US1] Expose CheckForUpdate() in LocalHosts/app.go (returns UpdateInfo)
- [x] T011 [US1] Add update UI entry + result dialog in LocalHosts/frontend/src/App.vue
- [x] T012 [US1] Handle “up to date / update available / error” UI states in LocalHosts/frontend/src/App.vue

**Checkpoint**: US1 可独立演示（不要求下载/安装）

---

## Phase 4: User Story 3 - 下载并自动安装更新 (Priority: P1)

**Goal**: 用户确认更新后，自动下载对应安装包并进入安装流程，安装阶段触发管理员授权

**Independent Test**: 在“发现新版本”弹窗点击“立即更新” → 看到下载进度 → 下载完成后进入安装阶段（Windows 启动安装程序 / macOS 执行替换并重启或进入可安装状态）

### Implementation for User Story 3

- [x] T013 [US3] Implement DownloadUpdate(update) in LocalHosts/internal/update/downloader.go (writes to temp dir, emits update:progress)
- [x] T014 [US3] Expose DownloadUpdate(update) in LocalHosts/app.go (returns installerPath)
- [x] T015 [US3] Implement InstallUpdate(installerPath) for Windows in LocalHosts/internal/update/installer_windows.go
- [x] T016 [US3] Implement InstallUpdate(installerPath) for macOS in LocalHosts/internal/update/installer_darwin.go (admin auth + replace + relaunch)
- [x] T017 [US3] Expose InstallUpdate(installerPath) in LocalHosts/app.go
- [x] T018 [US3] Implement download/install progress UI in LocalHosts/frontend/src/App.vue (subscribe to update:progress)

**Checkpoint**: US3 完成后可端到端更新到新版本（失败/拒绝授权不影响现有版本）

---

## Phase 5: User Story 2 - 启动时自动检查更新 (Priority: P2)

**Goal**: 启动时在后台检查更新，发现新版本后用非侵入方式提示

**Independent Test**: 打开应用后自动触发检查；如有新版本，出现提示条/角标并可打开更新弹窗

### Implementation for User Story 2

- [x] T019 [US2] Add UpdateSettings bindings (GetUpdateSettings/SaveUpdateSettings) in LocalHosts/app.go
- [x] T020 [US2] Add “启动时检查更新” UI toggle in LocalHosts/frontend/src/App.vue (persist to workspace)
- [x] T021 [US2] Trigger background CheckForUpdate on startup in LocalHosts/frontend/src/App.vue (respect setting)
- [x] T022 [US2] Add non-intrusive update notification UI in LocalHosts/frontend/src/App.vue (e.g., banner)

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 可靠性、安全性与可维护性增强

- [ ] T023 [P] Normalize error codes/messages for update operations in LocalHosts/internal/update/service.go
- [x] T024 Add minimal Go unit tests for semver compare and asset selection in LocalHosts/internal/update/semver_test.go and LocalHosts/internal/update/selector_test.go
- [ ] T025 Run quickstart validation steps from specs/002-auto-update/quickstart.md

---

## Dependencies & Execution Order

- Phase 1 → Phase 2 → US1（MVP）→ US3 → US2 → Polish
- US2 依赖 US1 的基础 UI/接口（复用 CheckForUpdate 的结果展示）
