# agents-template-coding-standards — Design

## Architecture

这是 1 个纯模板内容变更：修改 2 个静态 markdown 模板文件，**不**改任何
Go 代码、不改 `prependMarker` 逻辑、不改 init IO 流程。

```
internal/ide/templates/agents_zh.md   ← 插入中文"## 编码规范"段
internal/ide/templates/agents_en.md   ← 插入英文"## Coding Standards"段
                       │
                       ▼ init --ide auto 写项目根 AGENTS.md
                ./AGENTS.md (项目根)
```

模板加载链路保持不变（`agentsTemplate(lang)` → `instructionTemplate(lang, OpenCode)`
→ `prependMarker` → `WriteAgentsMD`），变更只发生在模板静态文本。

## 插入位置

两个模板都在**同一位置**插入（保持结构对称）：

```
当前顺序                           插入后顺序
─────────────────────────         ─────────────────────────
[frontmatter `mode: manual`]      [frontmatter `mode: manual`]
# free-kiro-managed:               # free-kiro-managed:
# AGENTS.md — ...                  # AGENTS.md — ...
本项目使用 **free-kiro** ...       本项目使用 **free-kiro** ...
## free-kiro 在这个项目做什么      ## free-kiro 在这个项目做什么
## 每个工具调用前（除 Read）       ## 每个工具调用前（除 Read）
## 每次 Edit / Write 之前          ## 每次 Edit / Write 之前
## 完成实现时                       ## 编码规范（SessionStart 必须先 Read）  ← 新增
                                   ## 完成实现时
## 帮助                            ## 帮助
[marker 块]                        [marker 块]
```

新段落在 `## 完成实现时` 之前，让 agent 按"项目背景 → 工具调用流程 →
编码规范（最关键）→ 完成实现 → 帮助"的递进顺序读，符合现有段落编排。

## 模板内容

**agents_zh.md（中文版，参考 .kiro/AGENTS.md 当前手写版本）**：

```markdown
## 编码规范（SessionStart 必须先 Read）

**在写任何 Go 代码之前，先 Read [`docs/CODING_STYLE.md`](../../docs/CODING_STYLE.md)、
[`docs/AGENT_RULES.md`](../../docs/AGENT_RULES.md) 与 [`docs/POLICY.md`](../../docs/POLICY.md)**。

- `docs/CODING_STYLE.md` 仅承载 7 章 Go 社区通用编码规范（命名 / 错误处理 /
  并发 / 接口 / 测试 / 注释 / 依赖）。它是**通用**规范，不是 free-kiro
  专属。
- [`docs/AGENT_RULES.md`](../../docs/AGENT_RULES.md) 承载 AI agent 协作硬性要求
  + 零豁免 / 零死代码政策（违反任意一条 = PR 拒收）。
- [`docs/POLICY.md`](../../docs/POLICY.md) 承载 free-kiro 项目特定策略（覆盖率门槛、
  TODO 注释 owner、协议合规、commit 格式、代码规模上限、PR 范围约束）。

硬约束摘录（来源：[`docs/AGENT_RULES.md`](../../docs/AGENT_RULES.md) §1）：

1. 零 `// TODO` / `// FIXME` / `// XXX` — 未完成的功能**不要写代码**，先
   回 spec 阶段补 requirements/design。
2. 零吞错误（`_ = doX()` / `log.Print(err)` 后继续）。
3. 零硬编码 magic number（端口、超时、阈值都要 `const` 或在 `internal/config`）。
4. 改动 > 50 行先 `free-kiro spec new`，三件套全绿再动代码（**工具特定硬约束**，仅对 free-kiro 仓库有效）。
```

**agents_en.md（英文版，结构对齐）**：同样的 5 段（标题 + 必读链接段 +
3 个文档说明 + 硬约束摘录 + 4 条约束），文字翻译为英文。链接路径前缀
保持 `../../`（因为模板会被 init 复制到 `.kiro/AGENTS.md` 或
项目根 `AGENTS.md`，后者位置不一定，所以用 `../../` 是 conservative
选择——如果 init 把它写到项目根（不在 .kiro/ 下），`../../docs/` 会指
到仓库外，这是已知缺陷，但与当前 zh 模板硬约束段的 `../../docs/` 写法
保持一致）。

## 关键决策记录

- **为什么不改 prependMarker / init Go 代码**：模板是 init 的输入，
  prependMarker 只决定 marker 行的位置，编码规范是模板内容的一部分，
  不属于 init 框架逻辑。修改范围控制在 2 个 markdown 文件即可。
- **为什么复制硬约束到模板而不只放指针**：AGENT_RULES.md §1 的 4 条
  约束是 session-level 必读，但 agent 在读 AGENTS.md 时可能不会主动
  跳到 AGENT_RULES.md。冗余复制让 agent 在不离开 AGENTS.md 的前提下
  也能看到硬约束（防"agent 看不到"场景）。同时，模板版本与
  AGENT_RULES.md 内容可能漂移——这是已知风险，留给后续 spec 加
  drift 检测。
- **为什么不复制到 instructions_*.md**：本仓库主用 Claude Code（CLAUDE.md
  通过 steering inject 拿到项目上下文），硬约束段是否需要在 CLAUDE.md
  模板里出现是另一个设计决策（涉及到 Claude Code 自身的 CLAUDE.md
  是否会被覆盖污染）。本 spec 只覆盖 OpenCode / CodeBuddy 共用的
  agents_*.md 路径。
- **为什么不更新现有 IDE 测试**：当前 ide tests 只断言 marker 行存在
  + 模板 dedupe + WriteAgentInstructions 写入路径，**不**断言模板
  文本内容。本 spec 也不需要新加这种测试——模板内容由人工 review 与
  init round-trip 验证保证。

## 验证策略

1. **Lint**：跑 `go test ./internal/ide/...` 确认无现有测试破坏
2. **Round-trip**：跑 `free-kiro init --ide auto --overwrite-instructions`
   后 `diff` 项目根 AGENTS.md，确认含"## 编码规范"段
3. **行数核对**：用 `wc -l AGENTS.md` 对比插入前后行数（应增加 ≥ 20 行）
4. **英文版对称**：人工 review agents_en.md 与 agents_zh.md 结构对齐

## 风险与回滚

- 风险 1：英文版硬约束翻译失真 → 中文版是 source of truth，英文版
  仅作为兜底；agent 在 en locale 下若看不懂会 fall back 到 docs/AGENT_RULES.md
- 风险 2：模板行数膨胀 → 当前每个模板 ~67 行，加 22 行后 ~89 行，
  仍在可读范围内
- 回滚：git revert 本 commit 即可恢复 init 模板到精简版本
