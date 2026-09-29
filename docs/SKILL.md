# `free-kiro skill` — SKILL.md bundle 管理

> 把 free-kiro 的能力以 SKILL.md 形式发布到 Claude Code / OpenCode /
> Codex CLI / CodeBuddy 的 skills 目录，AI 在用户谈到"spec / EARS /
> lint / PRD / wave"等意图时自动调用。

## 顶层结构

```
free-kiro
└── skill                          管理可安装的 SKILL.md bundle
    ├── install [--app ...]        下载 + 写入 ~/.{app}/skills/free-kiro/
    ├── uninstall --app ...        移除已装 bundle
    ├── update [--check]           检查 / 应用更新
    ├── show                       打印每个 app 的安装状态
    ├── path --app <id>            打印指定 app 的 skills 目录
    └── version                    bundle 版本 + binary 版本
```

## `skill install`

下载 `free-kiro-skill_<version>.zip`（默认 latest），校验 sha256，写入
目标 app 的 skills 目录。

```bash
free-kiro skill install                              # 装到全部 detected apps
free-kiro skill install --app claude-code           # 只装 Claude Code
free-kiro skill install --app all                   # 全部（含未 detected）
free-kiro skill install --dry-run                   # 只打印，不写盘
free-kiro skill install --version 0.7.0             # 指定版本
free-kiro skill install --from /path/to/dir         # 本地已解压目录
free-kiro skill install --from https://.../x.zip    # 远程 zip URL
free-kiro skill install --force                     # 覆盖已装
```

源解析优先级：

1. `--from <path>`：本地已解压目录（dev / 测试用）
2. `--from <url>`：远程 zip URL（自托管 / 镜像）
3. `--version` 指定：拉对应 tag 的 `free-kiro-skill_<v>.zip`
4. 默认：调 GitHub API `releases/latest`，拉对应 zip

下载后用 `free-kiro_<v>_SHA256SUMS`（同 release 内的另一个 asset）做
sha256 校验；不一致则拒绝安装。

## `skill uninstall`

```bash
free-kiro skill uninstall --app claude-code
free-kiro skill uninstall --app all       # 默认 all
```

幂等：未装则 no-op。

## `skill update`

```bash
free-kiro skill update               # 检查 + 更新
free-kiro skill update --check       # 只检查：exit 0 = 最新, 1 = 有新版
free-kiro skill update --force       # 即便已是 latest 也重装
```

比对 `~/.{app}/skills/free-kiro/skill.json` 的 `version` 与
`https://api.github.com/repos/jingyu525/free-kiro/releases/latest` 的
`tag_name`，不同则重装。

## `skill show`

```bash
free-kiro skill show
```

输出示例：

```
binary:  0.7.0 (abc1234, 2026-09-29)

APP           STATUS       VERSION      PATH
Claude Code  installed    0.7.0        /Users/x/.claude/skills/free-kiro
OpenCode     not installed -            /Users/x/.opencode/skills/free-kiro
Codex CLI    not installed -            /Users/x/.codex/skills/free-kiro
CodeBuddy    not installed -            /Users/x/.codebuddy/skills/free-kiro (experimental)
```

## `skill path`

```bash
free-kiro skill path --app claude-code
# → /Users/x/.claude/skills/free-kiro
```

## `skill version`

```bash
free-kiro skill version
```

```
binary:   0.7.0
bundle:   0.7.0 (Claude Code)
  warn:   bundle requires free-kiro >= 0.7.0  # 当 binary < min_version 时
```

dev build（`buildVersion == "dev"`）跳过版本告警。

## Bundle format (`skill.json`)

每个 release 的 zip 根目录包含：

```
free-kiro/
├── SKILL.md                 # 主入口（YAML frontmatter + body）
├── skill.json               # manifest
└── references/
    ├── spec.md
    ├── prd-fetch.md
    ├── ops.md
    ├── hooks.md
    └── troubleshooting.md
```

`skill.json` 字段：

| 字段 | 说明 |
|---|---|
| `name` | bundle 名（与 app skills/ 下的子目录同名） |
| `version` | bundle 版本（与 binary 版本绑定） |
| `free_kiro_min_version` | 要求的最低 free-kiro binary 版本 |
| `apps[]` | 支持的 app 列表（`id` / `skills_dir` / `detected_by` / `experimental`） |
| `files[]` | bundle 内文件清单（`path` / `required`） |
| `sha256{}` | 每个文件 sha256（在 release 时由 `compute-skill-sha.sh` 填入） |

## App paths

| App ID | skills 目录 | 状态 |
|---|---|---|
| `claude-code` | `~/.claude/skills/free-kiro/` | stable |
| `opencode` | `~/.opencode/skills/free-kiro/` | stable |
| `codex` | `~/.codex/skills/free-kiro/` | stable |
| `codebuddy` | `~/.codebuddy/skills/free-kiro/` | experimental（loader 约定待实测） |

## Troubleshooting

| 症状 | 处置 |
|---|---|
| `GitHub API returned HTTP 429` | 触发限流；等几分钟重跑，或设 `FREE_KIRO_SKILL_FROM` 走镜像 |
| `sha256 mismatch` | bundle 损坏或 release 流程出问题；用 `--from` 指向本地副本，再开 issue |
| `skill show` 报 `not installed` 但目录存在 | 目录里缺 `skill.json`；rm 后重装 |
| `--app codebuddy` 警告 `experimental` | skill 已写入但 CodeBuddy 的 skills loader 约定未实测；手动验证是否生效 |
| 版本不匹配告警（binary < bundle 要求） | 跑 `free-kiro upgrade` 把 binary 升到 ≥ bundle 要求 |

## 与 `free-kiro upgrade` 的关系

- `free-kiro upgrade`：升级 **binary 本身**
- `free-kiro skill update`：升级 **SKILL.md bundle**

两者独立。binary 升级后建议跑一次 `free-kiro skill update` 同步 bundle。
bundle 升级不影响 binary（向后兼容）。

## 与 IDE hook 的关系

- `free-kiro init` 写 IDE hook（`~/.claude/settings.json` 等），由 IDE 触发调 `free-kiro`
- `free-kiro skill install` 写 SKILL.md，由 AI 模型主动识别用户意图调 `free-kiro`

两者互补：hook 是"工具调用前后拦截"（被 IDE 触发），skill 是"AI 主动调用"
（被模型触发）。推荐两者同时配。

## 退出码

| Exit | 含义 |
|---|---|
| 0 | 成功 / 已是最新 |
| 1 | sha256 校验失败 / 部分 app 失败 |
| 2 | engine error（GitHub API / 网络） |
| 3 | 参数错 |