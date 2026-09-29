# free-kiro SKILL.md bundle

把 `free-kiro` 的能力以 SKILL.md 形态安装到 AI 编码助手的 skills 目录，
AI 在用户谈到"spec / EARS / lint / PRD / wave"等意图时自动调用。

## 安装方式（四选一）

### 1. `free-kiro skill install`（推荐）

```bash
free-kiro skill install                       # 装到所有 detected apps
free-kiro skill install --app claude-code   # 只装 Claude Code
free-kiro skill install --dry-run           # 只打印计划
free-kiro skill install --version 0.7.0     # 指定版本
free-kiro skill update --check              # 检查是否有新版
```

### 2. `curl | bash`

```bash
curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/contrib/skills/install.sh | bash
```

可选环境变量：

| 变量 | 默认 | 说明 |
|---|---|---|
| `FREE_KIRO_SKILL_APP` | `all` | `all` / `claude` / `opencode` / `codex` / `codebuddy` |
| `FREE_KIRO_SKILL_VERSION` | latest | `v0.7.0` 等具体版本 |
| `FREE_KIRO_SKILL_FROM` | （空） | 本地目录或 zip URL（跳过下载） |
| `FREE_KIRO_SKILL_DRY_RUN` | （空） | `1` 只打印不写盘 |
| `FREE_KIRO_HOME` | `$HOME` | 覆盖 home（CI / 测试用） |

例：

```bash
FREE_KIRO_SKILL_APP=claude-code FREE_KIRO_SKILL_VERSION=v0.7.0 \
  curl -fsSL .../install.sh | bash
```

### 3. `npx`

```bash
npx @jingyu525/free-kiro-skill
```

这是 `install.sh` 的薄 Node 包装（不进上游 `add-skill` registry）。

### 4. 手动

从 [GitHub Releases](https://github.com/jingyu525/free-kiro/releases)
下载 `free-kiro-skill_<version>.zip`，解压到 `~/.claude/skills/free-kiro/`
（或其他 app 的 skills 目录）。

## 安装位置

| App | 路径 |
|---|---|
| Claude Code | `~/.claude/skills/free-kiro/` |
| OpenCode | `~/.opencode/skills/free-kiro/` |
| Codex CLI | `~/.codex/skills/free-kiro/` |
| CodeBuddy | `~/.codebuddy/skills/free-kiro/` （experimental） |

## 验证 / 卸载 / 更新

```bash
free-kiro skill show        # 打印每个 app 的安装状态
free-kiro skill version     # bundle + binary 版本
free-kiro skill update      # 拉最新并重装
free-kiro skill uninstall --app all
```

## 软依赖：`bsk`

SKILL.md 暴露的 `--from-browser` 选项（`free-kiro spec new --from-browser <url>`）
依赖 [browser-skill](https://github.com/jingyu525/browser-skill) 的 `bsk`
CLI 在 PATH 上。装本 skill 不强制要求 `bsk`；缺时 `--from-browser` 会给
actionable error，并提示用 `--from-prd`（不需 bsk）作为备选。

## bundle 内容

```
free-kiro/
├── SKILL.md                        # 主入口
├── skill.json                      # 版本 + sha256 + 支持的 app
└── references/
    ├── spec.md                     # spec 生命周期深挖
    ├── prd-fetch.md                # --from-prd / --from-issue / --from-browser
    ├── ops.md                      # watch / serve / report / doctor
    ├── hooks.md                    # event-driven hooks
    └── troubleshooting.md          # 常见错误
```

## 退出码

| Exit | 含义 |
|---|---|
| 0 | 成功 / 已是最新 |
| 1 | sha256 校验失败 / 部分 app 失败 |
| 2 | engine error（GitHub API / 网络） |
| 3 | 参数错 |