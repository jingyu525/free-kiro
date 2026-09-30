# separate-codestyle-projectspecific — Design

## Architecture

把 docs/ 下的规范类文档分成三层，每层只承载一类内容：

```
┌─────────────────────────────────────────────────────────────┐
│ 入口层                                                       │
│  CLAUDE.md / .kiro/AGENTS.md                                 │
│    └─ "AI 必读"摘录：指向下面两层文档                          │
└─────────────────────────────────────────────────────────────┘
              │                              │
              ▼                              ▼
┌──────────────────────────┐  ┌──────────────────────────────┐
│ docs/CODING_STYLE.md     │  │ docs/POLICY.md                │
│ 通用 Go 编码规范          │  │ free-kiro 项目特定策略          │
│  1. 命名                 │  │  · 覆盖率门槛                  │
│  2. 错误处理              │  │  · TODO 注释带 owner           │
│  3. 并发                  │  │  · 协议合规                    │
│  4. 接口设计              │  │  · 内部依赖路径                │
│  5. 测试（通用部分）       │  │  · commit message 中文         │
│  6. 注释与文档（通用部分） │  │  · 函数 ≤ 50 行、文件 ≤ 500 行 │
│  7. 依赖管理（通用部分）   │  │  · 每个 PR 解决 1 个 spec       │
└──────────────────────────┘  └──────────────────────────────┘
              │
              ▼
┌──────────────────────────────────────────────────────────────┐
│ docs/AGENT_RULES.md                                            │
│  AI agent 协作硬性要求 + 项目级 zero-policy 政策               │
│   8.1 硬性要求（5 条）                                         │
│   8.3 上下文注入（SessionStart 引导）                           │
│   8.4 失败处置                                                 │
│   8.5 由工具强制                                               │
│   8.6 零豁免政策（zero-exemption）                              │
│   8.7 零死代码政策（zero-dead-code）                            │
└──────────────────────────────────────────────────────────────┘
```

### 关键决策

1. **拆成 2 个文档而不是 1 个**：AI 协作政策（AGENT_RULES.md）与项目策略
   （POLICY.md）的受众与变更频率不同——AI 政策随 free-kiro 工作流调整；
   项目策略随团队约定调整。合并会让两套不同演进节奏的规则互相干扰。

2. **CODING_STYLE.md 不彻底重写，只"剥离"**：第 1–7 章的通用部分保留
   ✅/❌ 示例，只删除其中混入的项目特定条款（如 5.1 覆盖率硬阈值）。
   读者可以放心把它当作 Go 风格参考。

3. **AGENT_RULES.md 单独建文件而非合并进 .kiro/AGENTS.md**：.kiro/AGENTS.md
   是 free-kiro 工具自身的 agent 上下文注入入口，与"AI 必须遵守的硬性要求"
   是不同的关注点；后者应该独立成可被多个 IDE 复用的规范文档。

4. **CLAUDE.md 不删除，只改硬约束摘录的指向**：CLAUDE.md 自身是项目特定
   的 AI 编码 agent 入口（顶部已写明），保持它在仓库根目录；但硬约束摘录
   段不再混入 Go 通用规范，转而指向新两个文档。

5. **不引入文档完整性校验工具**：当前 spec 范围是拆分，过度工具化会让 review
   成本上升；由 PR review + `free-kiro lint` 守门即可。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `docs/CODING_STYLE.md`（改） | 仅承载 7 章通用 Go 规范；顶部声明指向新两个文档 | N/A |
| `docs/AGENT_RULES.md`（新建） | AI agent 协作硬性要求 + 零豁免/零死代码政策 | N/A |
| `docs/POLICY.md`（新建） | free-kiro 项目特定策略（覆盖率、TODO owner、协议、内部依赖、commit、行数上限、PR 范围） | N/A |
| `CLAUDE.md`（改） | 硬约束摘录段改为指向 AGENT_RULES.md + POLICY.md | N/A |
| `.golangci.yml` / `Makefile` / `CONTRIBUTING.md`（检查） | 若有指向第 8 章或被删条款的链接/引用，同步更新 | N/A |
| `.kiro/AGENTS.md`（检查） | 若有指向 docs/CODING_STYLE.md 第 8 章的指引，同步更新 | N/A |

## Data Model

无数据结构。这是文档结构重组项目。

## Error Handling

- 文档改动不涉及运行时错误处理。
- `free-kiro lint separate-codestyle-projectspecific` 必须在文档改动完成后
  通过退出码 0，否则 PR 拒收（CI lint-go job 不出现新 ERROR）。
- 链接悬空检查：人工在 PR review 时 grep `docs/CODING_STYLE.md#8` /
  `docs/CODING_STYLE.md` 第 8 章相关锚点，验证无残留引用。

## Testing Strategy

- **人工 review**：拆分后三份文档由 ≥ 1 名维护者 sign-off，确认：
  1. docs/CODING_STYLE.md 仅含通用 Go 规范，无项目特定条款残留
  2. docs/AGENT_RULES.md 完整覆盖原第 8 章的硬性要求与 zero-policy
  3. docs/POLICY.md 完整覆盖散落的 6 条以上项目策略
  4. CLAUDE.md 顶部摘录准确指向新两个文档
- **链接核查**：在 PR 描述里附 `grep -rn "docs/CODING_STYLE.md#8" .` 与
  `grep -rn "覆盖率" docs/` 的输出，确认无悬空引用与遗漏条款。
- **CI 复跑**：`free-kiro lint` + `make lint-go` 必须保持绿色。

## Migration / Rollout

纯文档结构变更，单 PR 落地，不需要分阶段迁移：

1. **第一波（同 PR）**：先新建 docs/AGENT_RULES.md 与 docs/POLICY.md，
   确保所有项目特定条款都有归宿。
2. **第二波（同 PR）**：再修改 docs/CODING_STYLE.md 删除第 8 章与第 5/6/7
   章散落项目特定条款；顶部声明更新。
3. **第三波（同 PR）**：修改 CLAUDE.md 硬约束摘录改为指向新两个文档。
4. **第四波（同 PR）**：检查 .golangci.yml / Makefile / CONTRIBUTING.md /
   .kiro/AGENTS.md 中有无悬空引用并修复。

回滚策略：纯文档，PR revert 即回滚，无运行时风险。
