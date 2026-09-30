# ops — 运维命令（watch / serve / report / doctor / status / upgrade / demo）

## `free-kiro watch`

`.kiro/` 文件变化时跑命令。fsnotify 后端，默认 **500ms** 防抖。

### Preset 全集

| Preset | 等价命令集 |
|---|---|
| `default`（无 `--preset`） | `free-kiro lint` |
| `lint` | `free-kiro lint` |
| `status` | `free-kiro status` |
| `reactive` | `free-kiro lint` + `free-kiro status` |
| `full` | `free-kiro lint` + `free-kiro status` + `free-kiro report` |

### Flag

| Flag | 说明 | 默认 |
|---|---|---|
| `--preset` | 上述 5 个 preset 之一 | `default` |
| `--command` | 自定义命令（可重复；覆盖 `--preset`） | — |
| `--debounce` | 安静期（最后一次改动到触发的间隔） | `500ms` |
| `--root` | 监听目录（可重复；默认 `.kiro`） | `.kiro` |
| `--verbose` | 打印每个 fs event | `false` |

```bash
free-kiro watch                                       # = --preset default
free-kiro watch --preset reactive                     # lint + status
free-kiro watch --preset full                         # lint + status + report
free-kiro watch --command "free-kiro lint"
free-kiro watch --command "make" --command "go test ./..."
free-kiro watch --debounce 1s                         # 长安静期
free-kiro watch --root /path/to/project --verbose
```

适用：编辑器里编辑 `.kiro/specs/foo/requirements.md` 时自动 lint，编辑反馈即时。
退出 / 信号：SIGINT/SIGTERM 干净退出（Ctrl+C 不留 leak）。

## `free-kiro serve`

本地 web 看板（默认 `127.0.0.1:7373`），SPA（vanilla JS，无构建）+ 5 秒
轮询 `/api/summary` 刷新。

### Flag

| Flag | 说明 | 默认 |
|---|---|---|
| `--bind` | bind 地址 | `127.0.0.1`（loopback；`0.0.0.0` 暴露到 LAN） |
| `--port` | 端口（`0` 表示随机空闲端口） | `7373` |
| `--open` | 启动后尝试用系统浏览器打开 URL | `false` |

```bash
free-kiro serve                       # 127.0.0.1:7373
free-kiro serve --bind 0.0.0.0        # 局域网可访问
free-kiro serve --port 8080
free-kiro serve --port 0              # 随机端口（启动时打印）
free-kiro serve --open                # macOS / Linux desktop 自动开浏览器
```

### 端点

| Method | Path | 返回 |
|---|---|---|
| GET | `/` | 单页 SPA（vanilla JS，无构建） |
| GET | `/api/summary` | JSON：所有 spec 状态 + drift |
| GET | `/api/specs` | JSON：spec 列表 |
| GET | `/api/spec/<n>` | JSON：单个 spec 状态 |

适用：跟同事演示 / 自己盯多 spec 进度 / 写报告时截图。

## `free-kiro doctor`

单次自检。**5 项**检查：

1. free-kiro binary 位置 + 版本
2. PATH 配置（`~/.local/bin` 是否在 PATH 中）
3. 当前目录 `.kiro/` 工作区状态
4. 本机已安装的 IDE（Claude Code / CodeBuddy / OpenCode / Codex / Cursor / Continue）
5. GitHub latest release 版本（可选）

```bash
free-kiro doctor
free-kiro doctor --strict             # warn 也算 fail（exit 非零）
free-kiro doctor --verbose           # 列出全部受支持的 IDE，即使本机未装
```

## `free-kiro report`

把当前 workspace 状态聚合到 1 份 markdown（含 mermaid 图）：

- Summary 表（spec / phase / drift / tasks）
- Drift alerts 表 + 修复指引
- 项目级 Mermaid 图（所有 spec 概览）
- 每个 spec 的 phase / workflow / wave Mermaid 图

```bash
free-kiro report                      # 写到 .kiro/REPORT.md
free-kiro report --stdout            # 输出到 stdout（CI artifact / 管道）
free-kiro report --output ./REPORT.md # 自定义路径
```

适用：项目交接、CI artifact、贴 GitHub PR description / Notion / Confluence。

## `free-kiro status`

聚合所有 spec 状态、drift、wave 进度。

```bash
free-kiro status                      # 人类可读表格（默认）
free-kiro status --human              # 同上（显式）
free-kiro status --json               # 与 serve /api/summary 同源结构
```

`--json` 输出适合脚本 / dashboard 复用；`--human` 适合终端巡检。

## `free-kiro upgrade`

检查 GitHub releases/latest，下载并替换为最新版本。

```bash
free-kiro upgrade                     # 检查 + 升级（如有新版本）
free-kiro upgrade --check             # 只检查，不下载
free-kiro upgrade --force             # 强制重装当前版本（修复用）
```

升级流程：

1. 调 GitHub API 拉 latest release
2. 匹配当前平台（darwin/linux/windows + amd64/arm64）的 tarball
3. 下载 + SHA256 校验
4. 解压到临时路径 + 原子 rename 到当前 binary 位置
5. re-exec 替换当前进程（POSIX；Windows 提示手动重启）

退出码：`0` 已是最新（或升级成功）/ `1` 网络 / SHA256 / IO 失败 /
`2` 当前 binary 路径无法解析。

## `free-kiro demo`

5 分钟端到端 onboarding 入口。会：

1. 校验当前目录必须是 free-kiro 仓库根（含 `examples/todo-app/`）
2. 写时间戳 marker 到 `.kiro/.demostart`（30 秒内重复执行会去重）
3. 打印 5 行中文步骤摘要，告诉用户下一步该跑哪些命令
4. `--ide <name>` 不为 `none` 时，额外打印对应 IDE 的 hook 配置片段

```bash
free-kiro demo                        # 走完打印默认步骤
free-kiro demo --ide claude-code     # 额外打印 Claude Code hook 片段
free-kiro demo --no-color             # 禁 ANSI（CI / tee 友好）
```

不修改用户项目，不写任何 `.kiro` 业务文件之外的内容。

## 组合用法

```bash
# 在 CI 里：写完 spec 后自动 lint + 出报告
free-kiro lint && free-kiro report --output ./REPORT.md

# 本地开发：watch + serve 一起
free-kiro watch --preset full &
free-kiro serve --port 7373 &

# 升级 + 重装 skill bundle
free-kiro upgrade && free-kiro skill update
```

## 与 IDE hook 的分工

- `watch` 是"持续监听 + 主动跑命令"
- IDE hook（`PreToolUse` 等）是"工具调用前后拦截"
- 两者互补：hook 在每次 Edit/Write 之前即时拦截；watch 是后台长跑

推荐 hook 模式 + watch 模式同时开（hook 拦截 + watch 兜底）。