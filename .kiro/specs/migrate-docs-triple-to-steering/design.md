# migrate-docs-triple-to-steering — Design

## 总体设计

本 spec 把 `docs/` 三件套迁移到 `.kiro/steering/` 下，分 4 个阶段推进：

1. **Wave 1**：创建 3 份 steering 文件（含 frontmatter） + 解决 5 个跨文档冲突
2. **Wave 2**：删除 `docs/` 三件套 + 更新 IDE 指令文件引用 + 更新 IDE 模板引用
3. **Wave 3**：更新 11 个 spec 文档引用 + 更新其他配置文件
4. **Wave 4**：重编 binary + steering inject + 验证

```
仓库根
  ├─ .kiro/steering/
  │    ├─ coding-style.md   (NEW — 从 docs/CODING_STYLE.md 搬迁)
  │    ├─ agent-rules.md    (NEW — 从 docs/AGENT_RULES.md 搬迁)
  │    └─ policy.md         (NEW — 从 docs/POLICY.md 搬迁)
  └─ docs/
       ├─ CODING_STYLE.md   (DELETE — 内容已迁走)
       ├─ AGENT_RULES.md    (DELETE — 内容已迁走)
       └─ POLICY.md          (DELETE — 内容已迁走)
```

不新建 service/usecase 层；不引入新 Go 依赖；不改动 `internal/` 业务逻辑代码（仅注释中的路径引用需更新）。

## 关键决策

| 决策点 | 选择 | 理由 |
|---|---|---|
| 文件命名 | kebab-case：`coding-style.md` / `agent-rules.md` / `policy.md` | 与 `product.md` / `structure.md` / `tech.md` 风格一致 |
| frontmatter schema | `mode: always` + `description` 字段 | 与 `tech.md` 完全对齐 |
| 跨文档冲突解决 | 以 `tech.md` 为准 | tech.md 是已approved的约束文档 |
| 删除 docs/ 三件套 | **不**保留 shim | 用户目标"让 code style 归 steering"，shim 会让结构变复杂 |
| 模板改动 | 必须 `make install` 重编 binary | memory: `free-kiro-template-embed-rebuild` |
| `separate-codestyle-projectspecific` 处置 | 加 superseded 声明，不删除文件 | 保留 git history 与设计讨论记录 |

## 文件变更表

### A. 新建 3 个 steering 文件（从 docs/ 搬迁 + 加 frontmatter）

| 新文件 | 来源 | frontmatter |
|---|---|---|
| `.kiro/steering/coding-style.md` | `docs/CODING_STYLE.md` | `mode: always`, `description: Go 编码规范（命名/错误处理/并发/接口/测试/注释/依赖）` |
| `.kiro/steering/agent-rules.md` | `docs/AGENT_RULES.md` | `mode: always`, `description: AI agent 协作硬性要求与零豁免政策` |
| `.kiro/steering/policy.md` | `docs/POLICY.md` | `mode: always`, `description: free-kiro 项目策略（覆盖率/TODO/协议/依赖/commit/规模）` |

### B. 解决 5 个跨文档冲突

| 冲突点 | 解决方式 |
|---|---|
| **测试框架** | `coding-style.md` §5.1 改为「只用 stdlib `testing`，table-driven 测试优先，不引入 testify/assert/require」；`tech.md` 措辞不变（已为准） |
| **覆盖率阈值** | `policy.md` 保留硬阈值（≥70% 全包 / ≥80% 新增修改）；`tech.md` 与 `coding-style.md` 删除冲突措辞，统一指向 `policy.md` |
| **`//nolint` 上限** | 统一措辞为 `agent-rules.md` 版本（"//nolint 注释总数 ≤5，每条必须附 `//nolint:reason`"）；`tech.md` 删去重复措辞，改指向 `agent-rules.md` |
| **`-race` 必跑** | `tech.md`「测试」节补一行：`go test -race ./...` 必跑；`coding-style.md` 与 `agent-rules.md` 保留但简化为引用 `tech.md` |
| **依赖升级 commit 模板** | 重复内容统一到 `policy.md` §3；`coding-style.md` §7.2 改为引用 `policy.md` |

### C. 删除 `docs/` 三件套

| 删除文件 | 理由 |
|---|---|
| `docs/CODING_STYLE.md` | 内容已迁 `.kiro/steering/coding-style.md` |
| `docs/AGENT_RULES.md` | 内容已迁 `.kiro/steering/agent-rules.md` |
| `docs/POLICY.md` | 内容已迁 `.kiro/steering/policy.md` |

### D. 引用更新清单（36 个文件 / ~155 命中行）

#### D.1 IDE 指令文件（5 份）

| 文件 | 引用更新 |
|---|---|
| `CLAUDE.md` | L60, L61, L63, L66, L68, L71 |
| `AGENTS.md` | L56, L57, L59, L62, L64, L67 |
| `.cursorrules` | L60, L61, L63, L66, L68, L71 |
| `.cursor/rules/free-kiro.md` | L60, L61, L63, L66, L68, L71 |
| `.kiro/AGENTS.md` | L56, L57, L59, L62, L64, L67 |

**更新模式**：`docs/CODING_STYLE.md` → `.kiro/steering/coding-style.md`；
`docs/AGENT_RULES.md` → `.kiro/steering/agent-rules.md`；
`docs/POLICY.md` → `.kiro/steering/policy.md`。

#### D.2 IDE 模板（4 份，必须 `make install` 重编）

| 文件 | 引用更新 |
|---|---|
| `internal/ide/templates/agents_en.md` | L60, L61, L63, L65, L67, L70 |
| `internal/ide/templates/agents_zh.md` | L55, L56, L58, L61, L63, L66 |
| `internal/ide/templates/instructions_en.md` | L66 |
| `internal/ide/templates/instructions_zh.md` | L59, L60, L62, L65, L67, L70 |

#### D.3 Spec 文档（11 个 spec）

| spec | 处理 |
|---|---|
| `separate-codestyle-projectspecific` | 加 superseded 声明（不更新路径引用） |
| `agents-template-coding-standards` | 批量更新引用路径 |
| `golang-coding-standards` | 批量更新引用路径 |
| `enforce-golang-standards-zero-exemptions` | 批量更新引用路径 |
| `enforce-golang-standards-zero-dead-code` | 批量更新引用路径 |
| `per-ide-instructions` | 批量更新引用路径 |
| `steering-inject-to-ide` | 批量更新引用路径 |
| `top1-demo-onboarding` | 批量更新引用路径 |
| `performance-benchmarks` | 批量更新引用路径 |
| `update-contrib-skill-bundle` | 批量更新引用路径 |

#### D.4 其他配置文件 / 文档（6 个文件）

| 文件 | 引用更新 |
|---|---|
| `.golangci.yml` | L3, L4 |
| `Makefile` | L8 |
| `CONTRIBUTING.md` | L32, L34, L36, L39, L98 |
| `.github/workflows/ci.yml` | L55 |
| `internal/ide/ide_test.go` | L45（Go 源码注释） |
| `internal/visualize/server_handlers.go` | L327（Go 源码注释） |

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| IDE 模板改动未触发 binary 重编，inject 写出旧模板 | 显式 `make install` 重编（memory: `free-kiro-template-embed-rebuild`） |
| IDE 指令文件 hand-written 段与 inject 段顺序错乱 | `free-kiro steering inject` 幂等；CI 注入 + 人工 `git diff` 复核 |
| 跨文档冲突口径统一后，旧代码未对齐 | Out of scope（spec 阶段不回头改历史代码）；新代码按新口径 |
| 11 个 spec 引用路径批量更新遗漏 | `grep -rE` 复核；CI 跑 lint 时若 spec 引用 `docs/` 三件套则报错 |
| `free-kiro` 二进制自身用 `docs/` 三件套做 lint 锚点 | 复核 `internal/lint/` 各文件；调研未发现硬编码引用（仅 spec 文档内部引用 `docs/`） |
