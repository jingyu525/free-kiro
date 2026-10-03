# Steering — 项目约定文档

> Steering 是项目的"持久上下文"。每个 spec 工作流生成（agent 写代码）时，会按 mode
> 规则自动注入对应的文档。让 agent 不需要反复重述"项目的整体约定"。

## 概念

Steering 是 `.kiro/steering/*.md` 下的 markdown 文件（外加 `AGENTS.md`），每个文件
带 YAML frontmatter 声明 `mode`（注入时机）。

```
.kiro/
├── steering/
│   ├── product.md       # always — 产品目标
│   ├── structure.md     # always — 代码组织
│   └── tech.md          # auto   — 技术栈
└── AGENTS.md            # always — 给 agent 的总体指令
```

## 两种作用域

| 作用域 | 路径 | 用途 |
|---|---|---|
| **workspace** | `<root>/.kiro/steering/*.md` | 项目级，提交到 git |
| **global** | `~/.kiro/steering/*.md` | 用户级，跨项目共享 |

**workspace 覆盖 global**（同名文档 workspace 优先）。这是为了让单个项目可以"局部
推翻"用户的全局约定。

## 四种 inclusion mode

### `always` — 每次都注入

最常用。适合**项目级硬性约定**（编码风格、产品边界、技术栈）。

```markdown
---
mode: always
description: 这个产品的目标、核心能力与边界
---
# Product

用 1-3 段说明产品是什么、不是什么。
```

### `auto` — prompt 关键词命中才注入

适合**领域知识**（API 规范、设计模式）。prompt 关键词与 `description` 重叠时拉起。

```markdown
---
mode: auto
description: REST API design patterns. Use when creating or modifying API endpoints.
---
# API 设计
- 用 nouns 表示资源，verbs 表示动作
- 状态码遵循 RFC 7231
- ...
```

### `manual` — 显式 #name 引用才注入

适合**专用工具 / 罕见工作流**（调试某个子系统、跑迁移脚本）。

```markdown
---
mode: manual
---
# 数据库迁移工具

只在跑数据库迁移时显式引用。
```

agent 调用 `free-kiro steering context --manual migration` 才会拉起。

### `filematch` — 编辑特定文件时才注入

适合**文件级约定**（React 组件规范、SQL 风格、文档格式）。

```markdown
---
mode: filematch
fileMatchPattern: "**/*.tsx"
---
# React 组件规范
- 用 functional component + hooks
- props 用 TS interface 定义
- ...
```

agent 在编辑 `components/Button.tsx` 之前，调用：

```bash
free-kiro steering context --file components/Button.tsx
```

会自动拉起这份文档。

## 官方 `inclusion` 键 vs 克隆 `mode` 键

free-kiro 接受两种键名，等价：

```markdown
---
mode: always          # 克隆风格
inclusion: always     # Kiro 官方风格
---
```

`mode:` 与 `inclusion:` 同时存在时，`inclusion:` 优先（Kiro 官方优先）。

## fileMatchPattern 语法

支持 `**` / `*` / `?` 三种 glob：

| 模式 | 匹配 |
|---|---|
| `**/*.tsx` | 任意目录下的 .tsx |
| `src/**/*.go` | src/ 下任意深度的 .go |
| `*.md` | 仅根目录的 .md |
| `?single.md` | 一个字符 + single.md |

也接受 YAML 列表格式：

```markdown
fileMatchPattern: ["**/*.ts", "**/*.tsx"]
```

## CLI 命令

```bash
# 列出全部（workspace + global 合并）
free-kiro steering list

# 打印单个
free-kiro steering show product

# 组装上下文（agent 用入口）
free-kiro steering context \
  [--file <path>] \
  [--prompt <text>]
```

`context` 输出所有匹配的文档拼成的 block，可直接拼到生成 prompt 前缀。

## 在 agent 里集成

每次生成前调用 `steering context`，把结果拼到 prompt 头部：

```bash
CTX=$(free-kiro steering context --file "$CURRENT_FILE" --prompt "$USER_PROMPT")
PROMPT="$CTX

---

$USER_PROMPT"
```

`filematch` + `auto` 会被自动筛选，只有真正相关的文档会被拉起，避免污染上下文。

## 自动注入到 IDE 指令文件

`free-kiro steering context` 适用于**自定义 agent runner**（CLI / IDE
插件 / 测试 harness），把 context 拼到生成 prompt 前缀。但 Claude Code /
CodeBuddy / Cursor / Continue / OpenCode 这些 IDE 内置的 agent loader
**不会**调 free-kiro CLI，它们只在会话开始时一次性读项目根的指令文件
（`CLAUDE.md` / `AGENTS.md` / `.cursorrules` / `.cursor/rules/*.md` /
`.continue/rules/*.md`），见 `docs/HOOKS.md` §"Agent instructions per
IDE"。

为了让 steering 内容**自动**进入这些 IDE 的 agent 上下文，新增
`free-kiro steering inject` 子命令：

```bash
free-kiro steering inject [--dry-run] [--only <glob>]
```

行为：

- 读取 `.kiro/steering/*.md` 中 `mode: always` 的文档
- 按文件名字母序拼成 1 个 markdown 块
- 写入 5 个 IDE 指令文件（`CLAUDE.md` / `AGENTS.md` / `.cursorrules` /
  `.cursor/rules/free-kiro.md` / `.continue/rules/free-kiro.md`）的
  `<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`
  marker 区域内
- marker **之外**的所有内容（包括首行 `# free-kiro-managed:` 注释与
  用户手写段落）原样保留

### 两套 marker 的区分

IDE 指令文件里实际有 **2 套** free-kiro 维护的标记，名字都带
`free-kiro-managed`，但来源、用途完全不同，别混淆：

| 标记 | 注入源 | 用途 | 出现位置 |
|---|---|---|---|
| 首行 `# free-kiro-managed:` | `internal/ide/ide.go` 的 `prependMarker` 函数（行 517-545），由 `free-kiro init` 运行时调用 | `IsFreeKiroInstruction` 读取首行识别"这文件是 free-kiro 生成的"，防止后续 `init --overwrite-instructions` 覆盖用户在同路径手写的内容 | YAML frontmatter 闭合 `---` 之后的第一行 |
| `<!-- free-kiro-managed:start -->` ... `<!-- free-kiro-managed:end -->` | `free-kiro steering inject` 在每次 `init` / `inject` 运行时写入 | 划定 `mode: always` steering 文档注入区域的边界 | 文件末尾 |

修改路径：

- 想改**顶部 marker** 的语义、位置或文案 → 改
  `internal/ide/ide.go` 的 `prependMarker` 函数，然后重跑
  `free-kiro init --ide auto --overwrite-instructions`
- 想改**底部 marker 区域**的注入内容 → 改
  `.kiro/steering/<name>.md` 后跑 `free-kiro steering inject`
- 想改 IDE 指令模板里"项目上下文与文件标记"那段说明 → 改
  `internal/ide/templates/{instructions,agents}_{zh,en}.md` 后重跑
  `free-kiro init --ide auto --overwrite-instructions`

退出码（与 spec `.kiro/specs/steering-inject-to-ide/` 对齐）：

| 码 | 含义 |
|---|---|
| 0 | 全部目标文件写入成功 |
| 3 | `.kiro/steering/` 中没有 `mode: always` 文档 |
| 4 | `--only` glob 语法非法（包含 `**` 或未闭合字符类） |
| 1–5 | 这么多目标文件被跳过（缺 marker / IO 错），上限 5 |

`free-kiro init` 写入的 4 个模板（`instructions_zh/en.md` 与
`agents_zh/en.md`）已自带 marker 块，因此 init 之后立即跑一次
`steering inject` 即可生效。`auto` / `manual` / `filematch` 模式的文档
**不**会被注入（它们需要 prompt 关键词或显式引用，不适合 always 注入）。

> 注意：`steering inject` 写的是 IDE 指令文件，不是 steering store
> 读取的源文件。`.kiro/steering/*.md` 仍然是 source of truth；修改
> 文档后重跑 inject 即可同步到 IDE 指令文件。

## 完整示例：.kiro/steering/

`free-kiro init` 默认生成 3 个示例文档：

**product.md**（always）— 一两段说清产品目标  
**structure.md**（always）— 代码组织原则  
**tech.md**（auto）— 技术栈与开发规范（"Go projects" / "python projects" 等）

可以继续添加：

- `api.md`（auto）— REST API 设计
- `frontend.md`（filematch, `**/*.tsx`）— React 组件规范
- `db-migration.md`（manual）— 数据库迁移脚本指南

---

参考：[CLI.md](CLI.md)（steering 命令）/ [HOOKS.md](HOOKS.md)（用 hook 自动触发 context 组装）