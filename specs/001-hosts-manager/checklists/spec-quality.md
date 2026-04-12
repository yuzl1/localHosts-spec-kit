# Spec Quality Checklist: LocalHosts Hosts 管理

**Purpose**: 用“单元测试”方式评估 spec 的完整性、清晰度、一致性与可验证性（评审用）  
**Created**: 2026-04-12  
**Feature**: [spec.md](file:///Users/yuzl6/Documents/trae_projects/localHosts-spec-kit/specs/001-hosts-manager/spec.md)

## Requirement Completeness

- [x] CHK001 是否明确区分“管理块内可编辑”与“管理块外只读”，并对块外内容的展示边界写清楚？[Completeness, Spec §US1/§FR-002/§FR-011]
- [x] CHK002 是否明确规定当系统 hosts 中不存在管理块时的写回策略（追加位置、空行处理、说明行是否需要）？[Completeness, Spec §Edge Cases/§FR-011]
- [x] CHK003 是否明确规定当管理块边界缺失/损坏时的处理策略（如何避免覆盖块外内容）？[Completeness, Spec §Edge Cases/§FR-011]
- [x] CHK004 是否明确规定“用户目录持久化”的最小必存字段集合（分组、条目、组装开关、顺序、版本号）？[Completeness, Spec §FR-013]
- [x] CHK005 是否明确规定“分组开关”和“条目开关”在 composeEnabled=true/false 两种模式下各自的作用范围？[Completeness, Spec §FR-010]
- [x] CHK006 是否明确规定“多域名拆分后”保存时的输出形态（是否允许回写为一行多域名或必须一域名一行）？[Completeness, Spec §FR-009]
- [x] CHK007 是否明确规定“备注 comment”在 hosts 行中的表达方式（是否允许空、是否允许包含 `#`、如何转义/保留）？[Completeness, Spec §FR-004/§Edge Cases]
- [x] CHK008 是否明确规定对“不可解析行”的处理边界（保留到块外、是否只读展示、是否允许迁移进管理块）？[Completeness, Spec §Edge Cases/§FR-011]

## Requirement Clarity

- [x] CHK009 “合法条目”的判定标准是否写成可验证规则（IPv4/IPv6 + 域名 token 约束），避免依赖“看起来像”之类描述？[Clarity, Spec §FR-008/§FR-009]
- [x] CHK010 “纯注释行”与“禁用条目”的区分规则是否足够精确（去掉 `#` 与空白后的解析规则、支持 tab/多空格）？[Clarity, Spec §FR-008/§Edge Cases]
- [x] CHK011 “块外内容原样保留”是否定义了“原样”的判定维度（内容、相对顺序、空白字符、行尾换行）？[Clarity, Spec §FR-011/§SC-003]
- [x] CHK012 “管理员授权失败/取消”是否定义为可观察的结果（错误信息分类、用户可理解提示）而不是实现细节？[Clarity, Spec §US3/§FR-007/§SC-004]
- [x] CHK013 “跨平台一致体验”是否明确哪些行为必须一致（条目模型、分组语义、组装规则），哪些可因平台差异而不同（授权方式）？[Clarity, Spec §FR-013/§Clarifications]
- [x] CHK014 “即时”类描述是否可被客观判定（例如以“无明显卡顿”的手工准则，或明确阈值）？[Clarity, Spec §SC-001]

## Requirement Consistency

- [x] CHK015 “保存”在 spec 中是否被一致地区分为“保存工作区（用户目录）”与“应用到系统 hosts（提权写回）”，避免同名歧义？[Consistency, Spec §US2/§US3/§FR-007/§FR-013]
- [x] CHK016 “不修改注释行/保留空行”的约束与“仅管理标记块”的写回策略是否一致且无冲突？[Consistency, Spec §FR-011/§FR-008/§SC-003]
- [x] CHK017 “多域名拆分为多条”与“启用/禁用通过整行注释”是否存在冲突点（例如原始一行多域名被拆分后如何表达单个域名禁用）并已在需求中解释？[Consistency, Spec §FR-006/§FR-009]
- [x] CHK018 目标平台在 Clarifications 中包含 Linux，但 Constitution/Plan/Spec 的目标平台陈述是否一致？[Consistency, Spec §Clarifications + Plan §Technical Context + Constitution]

## Acceptance Criteria Quality

- [x] CHK019 是否为“管理块外内容保持一致”定义了可核验口径（对比范围、忽略项、差异展示方式）？[Measurability, Spec §SC-003]
- [x] CHK020 是否为“退出再进恢复工作区”定义了恢复边界（是否包含分组顺序、开关、未应用到系统的变更）？[Measurability, Spec §SC-002]
- [x] CHK021 是否为“授权失败/取消不修改系统文件”定义了可核验口径（如何判定系统文件未变）？[Measurability, Spec §SC-004/§US3]

## Scenario Coverage

- [x] CHK022 是否覆盖首次使用场景：系统 hosts 无管理块 → 应用仍可展示块外只读内容并允许创建管理块？[Coverage, Spec §Edge Cases/§FR-011]
- [x] CHK023 是否覆盖迁移场景：用户希望将块外某条目纳入管理块进行编辑，需求是否明确“支持/不支持/后续版本”？[Gap, Spec §FR-001-§FR-013]
- [x] CHK024 是否覆盖并发/冲突场景：应用打开期间系统 hosts 被外部程序修改，需求是否明确再加载/保存时的冲突处理策略？[Gap, Spec §Edge Cases]
- [x] CHK025 是否覆盖恢复场景：写入失败后的回滚行为与用户提示要求是否明确？[Coverage, Spec §US3/§FR-007]

## Edge Case Coverage

- [x] CHK026 对“hosts 文件不存在/不可读/权限不足”的要求是否明确包含用户可理解的提示与可恢复路径？[Coverage, Spec §Edge Cases]
- [x] CHK027 对“不可解析行”的保留要求是否明确，且不会被误判为可编辑条目？[Coverage, Spec §Edge Cases/§FR-008]
- [x] CHK028 对“tab/多空格/尾随空白/不同换行符”的处理要求是否明确（解析与回写的一致性）？[Gap, Spec §Edge Cases/§FR-011]
- [x] CHK029 对“重复域名或重复 IP+域名”的展示与写回要求是否明确（允许重复、去重规则、冲突提示）？[Gap, Spec §Edge Cases]

## Non-Functional Requirements

- [x] CHK030 是否将“冷启动 < 1s、内存 < 30MB”从 Constitution 显式映射到 spec 的可测成功标准或约束描述中？[Gap, Constitution + Spec §Success Criteria]
- [x] CHK031 是否定义了隐私/敏感信息处理要求（例如不在 UI/日志中泄露 hosts 中的内网信息）并与 Constitution 一致？[Coverage, Constitution + Spec §FR-011]

## Dependencies & Assumptions

- [x] CHK032 是否明确“用户目录”在三平台上的概念边界（必须使用系统推荐目录而非硬编码）并体现为需求而非实现？[Clarity, Spec §FR-013]
- [x] CHK033 是否明确哪些能力依赖管理员授权（仅写系统 hosts）以及哪些能力不依赖（读取/编辑/保存工作区）？[Clarity, Spec §FR-007/§FR-013]

## Ambiguities & Conflicts

- [x] CHK034 是否存在术语歧义（例如“保存”“组装”“管理块”“只读内容”）且已在 spec 中固定一致用词？[Ambiguity, Spec §US1-§US3/§FR-010-§FR-013]
- [x] CHK035 是否明确“块外内容只读视图”的展示粒度与脱敏要求（逐行展示/摘要展示/隐藏敏感行）？[Gap, Spec §US1/§FR-002]

## Notes

- 该清单用于评审 spec 写作质量，不用于验证实现正确性
- 完成时将 `[ ]` 勾选为 `[x]`，并在条目后追加发现说明
