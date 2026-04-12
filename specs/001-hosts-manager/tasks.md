---

description: "LocalHosts Hosts 管理特性任务清单（跨平台 + 分组组装 + 管理块写回）"
---

# Tasks: LocalHosts Hosts 管理（跨平台 + 分组组装 + 管理块写回）

**Input**: 设计文档来自 `/specs/001-hosts-manager/`  
**Prerequisites**: plan.md（必需）, spec.md（必需）, research.md, data-model.md, contracts/, quickstart.md  
**Tests**: 未在 spec 中要求强制测试，本清单不包含测试任务（如后续需要可再补充）  

**Organization**: 任务按用户故事分组，确保每个用户故事都可独立完成与验证。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件、无依赖冲突）
- **[Story]**: 对应用户故事（US1/US2/US3）
- 描述中必须包含准确文件路径

---

## Phase 1: Setup（共享基础结构）

- [x] T001 创建后端目录结构：LocalHosts/internal/{hosts,platform,workspace}/
- [x] T002 [P] 新建数据模型：LocalHosts/internal/hosts/model.go（HostEntry/HostGroup/WorkspaceState/HostDocument）
- [x] T003 [P] 新建工作区模型：LocalHosts/internal/workspace/model.go（与 contracts/wails-api.md 对齐的 HostWorkspace 视图）
- [x] T004 [P] 新建平台路径解析：LocalHosts/internal/platform/paths.go（获取系统 hosts 路径与用户配置目录路径）
- [x] T005 在 specs/001-hosts-manager/contracts/wails-api.md 中确认对外方法集合（GetHosts/SaveHosts/ApplyHosts）并以其为实现契约

---

## Phase 2: Foundational（阻塞性基础能力）

**Checkpoint**: 完成后，US1/US2/US3 的实现可并行推进。

- [x] T006 实现管理块定位：LocalHosts/internal/hosts/block.go（识别 START/END；处理缺失/损坏策略；输出 prefixLines/suffixLines）
- [x] T007 实现 hosts 解析：LocalHosts/internal/hosts/parse.go（解析合法条目；区分禁用条目 vs 纯注释行；保留空行/不可解析行到块外或原始行集合）
- [x] T008 实现 hosts 渲染：LocalHosts/internal/hosts/render.go（按“一域名一行”输出管理块；禁用条目整行注释；块外原样拼接）
- [x] T009 实现工作区持久化：LocalHosts/internal/workspace/store.go（读写用户目录 JSON；包含 groups/entries/composeEnabled/version）
- [x] T010 实现组装规则：LocalHosts/internal/hosts/render.go（根据 composeEnabled + 分组 enabled/order 生成最终管理块内容）

---

## Phase 3: User Story 1 - 查看与快速切换 Hosts（Priority: P1）🎯 MVP

**Goal**: 启动加载系统 hosts，展示块外只读与块内可编辑条目列表，支持条目开关切换（仅影响工作区，未写系统文件）。

**Independent Test**: 使用 quickstart.md 的最小用例 1/2/3/4，在不点击“写入系统 hosts”的情况下完成加载与开关切换验证。

- [x] T011 [P] 实现读取系统 hosts：LocalHosts/internal/platform/paths.go + LocalHosts/internal/hosts/parse.go（按当前平台路径读取并解析）
- [x] T012 实现 GetHosts 绑定：LocalHosts/app.go（返回 HostWorkspace：groups/entries/composeEnabled/readonlySystemLines）
- [x] T013 [P] 前端封装 API 调用：LocalHosts/frontend/src/services/hosts.js（封装 wailsjs 调用 GetHosts/SaveHosts/ApplyHosts）
- [x] T014 [P] 前端页面骨架：LocalHosts/frontend/src/App.vue（分区展示：可编辑条目列表 + 块外只读内容）
- [x] T015 [US1] 前端加载流程：LocalHosts/frontend/src/App.vue（启动调用 GetHosts；loading/错误态展示）
- [x] T016 [US1] 前端条目开关：LocalHosts/frontend/src/App.vue（切换 enabled 后调用 SaveHosts 持久化工作区；不触发 ApplyHosts）

**Checkpoint**: US1 完成后，用户可打开应用看到数据并开关条目，退出再进可恢复工作区状态。

---

## Phase 4: User Story 2 - 管理条目与分组（Priority: P2）

**Goal**: 支持新增/编辑/删除条目；支持分组管理、分组排序、分组启用/禁用；支持“是否按分组组装”开关；均持久化到用户目录。

**Independent Test**: 不写系统 hosts 的前提下，完成新增/编辑/删除与分组开关/排序变更，退出再进仍能恢复。

- [x] T017 [P] 后端工作区合并策略：LocalHosts/internal/workspace/store.go（将系统 hosts 的块外只读行与用户目录工作区合并成 HostWorkspace）
- [x] T018 [P] 扩展 HostWorkspace 结构：LocalHosts/internal/workspace/model.go（加入 groups/entries/composeEnabled/readonlySystemLines）
- [x] T019 [US2] 实现 SaveHosts 绑定：LocalHosts/app.go（接收 HostWorkspace 并写入用户目录；包含版本与更新时间）
- [x] T020 [P] 前端分组 UI：LocalHosts/frontend/src/App.vue（分组列表：启用/禁用开关、排序按钮/拖拽占位、选择分组）
- [x] T021 [US2] 前端条目 CRUD：LocalHosts/frontend/src/App.vue（新增/编辑表单：ip/domain/comment/group；删除确认）
- [x] T022 [US2] 前端“是否组装”开关：LocalHosts/frontend/src/App.vue（composeEnabled 切换并调用 SaveHosts）
- [x] T023 [US2] 前端多域名规则体现：LocalHosts/frontend/src/App.vue（domain 为单 token；提示“一个域名一条”）

**Checkpoint**: US2 完成后，用户可完成条目与分组管理，且工作区可可靠持久化与恢复。

---

## Phase 5: User Story 3 - 授权写回系统 Hosts（Priority: P3）

**Goal**: 将工作区组装为管理块并写入系统 hosts（管理员授权）；块外内容原样保留；写入前备份并可回滚；失败/取消不修改系统文件。

**Independent Test**: 在三个平台之一验证：写入成功仅更新管理块；取消授权/写入失败时系统 hosts 不变。

- [x] T024 [P] 后端生成最终文本：LocalHosts/internal/hosts/render.go（输入 HostDocument + HostWorkspace，输出完整 hosts 文本）
- [x] T025 [P] 后端备份与回滚：LocalHosts/internal/platform/paths.go（定义备份路径规则）+ LocalHosts/internal/platform/elevate_*.go（写入失败恢复）
- [x] T026 实现跨平台提权写入接口：LocalHosts/internal/platform/elevate_darwin.go（sudo）、LocalHosts/internal/platform/elevate_linux.go（sudo）、LocalHosts/internal/platform/elevate_windows.go（管理员）
- [x] T027 [US3] 实现 ApplyHosts 绑定：LocalHosts/app.go（组装→备份→提权写入；错误返回）
- [x] T028 [P] 前端“写入系统 hosts”按钮：LocalHosts/frontend/src/App.vue（调用 ApplyHosts；显示进度与结果）
- [x] T029 [US3] 前端失败/取消提示：LocalHosts/frontend/src/App.vue（区分用户取消/权限失败/写入失败；提示“系统文件未修改”）
- [x] T030 [US3] 写入后刷新：LocalHosts/frontend/src/App.vue（ApplyHosts 成功后重新调用 GetHosts 以验证一致性）

**Checkpoint**: US3 完成后，用户可在授权后写入系统 hosts，且块外内容保持不变并具备回滚。

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T031 统一错误结构与用户提示文案：LocalHosts/app.go + LocalHosts/frontend/src/App.vue
- [x] T032 确保无网络请求且日志不包含 hosts 具体内容：LocalHosts/app.go + LocalHosts/main.go + LocalHosts/frontend/src/App.vue
- [x] T033 性能与轻量检查：避免不必要的全量重渲染与大对象拷贝（LocalHosts/frontend/src/App.vue）
- [x] T034 校验与输入约束：IP/域名/备注输入校验与错误提示（LocalHosts/frontend/src/App.vue）
- [x] T035 整理对外契约一致性：spec.md 与 contracts/wails-api.md 对齐（specs/001-hosts-manager/spec.md）
- [x] T036 明确“块外原样保留”的实现口径：保留空白字符与原文件换行风格（LocalHosts/internal/hosts/block.go + LocalHosts/internal/hosts/render.go）
- [x] T037 明确启动/内存目标的验收口径并记录测量方式（specs/001-hosts-manager/spec.md + LocalHosts/frontend/src/App.vue）

---

## Dependencies & Execution Order

- **Phase 1（Setup）** → **Phase 2（Foundational）**：必须先完成
- **US1（P1）**：依赖 Phase 2
- **US2（P2）**：依赖 Phase 2；可在 US1 完成后或并行推进 UI 细节
- **US3（P3）**：依赖 Phase 2；建议在 US1/US2 稳定后实现（风险更高）

## Parallel Example: US1

```bash
Task: "实现 GetHosts 绑定：LocalHosts/app.go"
Task: "前端封装 API 调用：LocalHosts/frontend/src/services/hosts.js"
Task: "前端页面骨架：LocalHosts/frontend/src/App.vue"
```

## Implementation Strategy

1. 完成 Phase 1/2，确保解析、管理块定位、渲染、持久化都可用
2. 先交付 US1（MVP）：加载 + 展示 + 开关 + 工作区持久化
3. 再交付 US2：条目 CRUD + 分组 + 组装开关
4. 最后交付 US3：提权写入 + 备份回滚 + 写后刷新验证

## FR/SC → Task 映射

| Key | Task IDs |
|-----|----------|
| FR-001 | T004, T011, T012 |
| FR-002 | T012, T014, T015 |
| FR-003 | T021 |
| FR-004 | T021 |
| FR-005 | T021 |
| FR-006 | T016, T008, T010 |
| FR-007 | T026, T027, T029 |
| FR-008 | T007 |
| FR-009 | T007, T008, T023 |
| FR-010 | T010, T020, T022 |
| FR-011 | T006, T008, T027, T035 |
| FR-012 | T006, T008, T027, T035 |
| FR-013 | T007, T012, T014, T015 |
| FR-014 | T009, T019 |
| FR-015 | T032 |
| FR-016 | T007, T008 |
| FR-017 | T008, T024 |
| SC-001 | T037 |
| SC-002 | T009, T016, T019 |
| SC-003 | T006, T008, T027, T035 |
| SC-004 | T026, T027, T029 |
| SC-005 | T037 |
