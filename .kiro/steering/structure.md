---
mode: always
description: free-kiro 仓库目录结构与代码组织约定
---

# Structure

## 顶层布局

```
free-kiro/
├── AGENTS.md / CLAUDE.md  # 注入到 Claude Code / Cursor 的项目级指令
├── CONTRIBUTING.md        # 贡献流程
├── LICENSE                # 开源协议
├── Makefile               # 常用开发命令（build / test / lint / verify）
├── README.md
├── cmd/free-kiro/            # main 入口（Cobra root 命令）
├── docs/                     # 面向用户的文档
├── examples/                 # 示例 fixture（如 todo-app）
├── testdata/                 # 测试 fixture（含 bsk-stub / inject 子目录）
├── install.sh                # curl|bash 安装脚本
├── contrib/                  # 第三方分发物料（当前仅 skill bundle 子目录；Homebrew formula 改由 release.yml 推送）
├── .goreleaser.yaml          # 跨平台构建配置（5 平台，见 tech.md）
├── .golangci.yml             # golangci-lint v2 配置（启用 staticcheck / errcheck / revive）
├── go.mod / go.sum
└── internal/                 # 全部业务逻辑（私有包）
```

## internal/ 顶层包

| 包 | 职责 | 代表性文件 |
|---|---|---|
| `cli/` | Cobra 子命令注册与 flag 绑定 | ~22 个 cobra 命令文件（每个子命令一文件：spec*、init、lint、watch、upgrade、serve、doctor、demo、skill*、status、hook、task 等）；完整命令清单见 `cmd/free-kiro/main.go` 的 `rootCmd.AddCommand(...)` 注册 |
| `ide/` | IDE hook 信封生成（Claude Code / CodeBuddy 等） | `init.go`、`templates/`（实际是 `agents_*.md` / `instructions_*.md` 模板，不是 `.tmpl`） |
| `lint/` | EARS + 结构 + 依赖图 lint 引擎 | `ears.go`（EARS 正则）/ `bugfix.go`（三段式）/ `requirements.go` & `tasks.go`（结构 lint）/ `linter.go`（总入口）/ `baseline.go`（基线规则）/ `quality.go`（质量指标 / semantic gates） |
| `spec/` | spec 文档读写 + 状态机推进 + drift 检测 | `engine.go` / `engine_io.go` / `engine_state.go`（状态机）/ `generator.go`（生成器）/ `analyze.go`（advisory）/ `drift.go`（baseline 锁）/ `warn.go`（vague / 重复 AC / 追溯）/ `templates/`（4 个 .tmpl） |
| `steering/` | steering 注入与作用域解析 | `frontmatter.go` / `assemble.go`（4 种 inclusion 模式选择）/ `store.go`（两级作用域） |
| `taskgraph/` | tasks.md 依赖图 → wave 拓扑 | `waves.go`（拓扑排序 + 并行 wave）/ `cycle.go`（DetectCycle） |
| `hooks/` | hook JSON 信封序列化 | `envelope.go`（Kiro 兼容 event-keyed）/ `dispatch.go` / `registry.go` / `glob.go` |
| `watch/` | fsnotify 文件监听 | `watcher.go`（fallback 2s mtime 轮询） |
| `workspace/` | `.kiro/` 目录发现与对账 | `workspace.go`（含 `.current` 标记） |
| `models/` | 跨包共享的数据结构 | `spec.go` / `hook.go` / `phase.go` / `task.go` / `steering.go` |
| `visualize/` | ASCII tree / Mermaid / web dashboard | `server.go`（路由 + 生命周期）/ `server_handlers.go`（HTML 渲染 + REST）/ `server_sse.go`（SSE + fsnotify）/ `etag.go`（缓存协商）/ `io.go`（文件 IO 适配）/ `middleware.go`（logging / CORS / recovery）/ `mermaid.go`（Mermaid LR 渲染）/ `report.go` / `visualize.go` / `static/`（前端） |
| `visualize/static/` | **完整的 Vite + FSD 前端工程** | `index.html`（embed.FS 入口）/ `vite.config.ts` / `package.json` / `pnpm-lock.yaml` / `src/`（FSD: app/ entities/ features/ pages/ shared/ widgets/）/ `dist/`（build 产物）/ `scripts/`（check-fsd-layers.mjs、e2e-dashboard.mjs）/ `node_modules/` |
| `skill/` | self-install 的 skill bundle 读写 | `app.go` / `install.go` / `manifest.go` / `paths.go` / `release.go` / `show.go`（与 `contrib/skills/` 无软链 / 拷贝关系，独立脚本化分发） |
| `upgrade/` | 自升级下载 + SHA256 校验 | `upgrade.go`（Check/Apply/reexec；POSIX `os.Exec`，Windows 不 re-exec）/ `io.go` |
| `frontmatter/` | YAML frontmatter 解析（spec/steering 共用） | `frontmatter.go` / `schema.go` |
| `text/` | 文本处理工具（normalize / dedent 等） | `text.go` |
| `errors/` | 包级 sentinel 错误 + wrap helper | `errors.go` |

## 包约定

- **包名**：小写单词，不复数，不缩写（`spec` 而非 `specs`，`lint` 而非 `l`）
- **导出符号**：每个 exported 类型/函数/常量必须有 godoc 注释
- **IO 隔离**：所有外部 IO（文件、网络、子进程）必须放在显式 adapter 后，
  便于测试时 mock
- **依赖方向**：`cli → models + 各领域包`，领域包之间禁止互相 import
  内部细节（只允许通过 `models` 共享类型）
- **错误处理**：跨包错误用 sentinel（在 `internal/errors`）+ `fmt.Errorf %w`
  wrap；不要把 string repr 当 error identity
- **测试**：单元测试与被测文件同包同目录（`*_test.go`），不开 mirror package

## 模板与生成器分离

- `*/templates/*.md.tmpl` 只放模板字符串
- 生成器是纯函数（输入结构 → 输出字符串），不做 IO
- 模板里不嵌入业务逻辑，逻辑全部前置到生成器

## dashboard 后端拆分约定

`internal/visualize/` 把 dashboard 服务拆成多个职责单一的文件：

- `server.go`：路由注册 + 生命周期
- `server_handlers.go`：HTML 渲染 + REST handler
- `server_sse.go`：SSE 长连接 + fsnotify 订阅
- `etag.go`：缓存协商逻辑（与 handler 解耦，便于单测）
- `io.go`：文件系统读取
- `middleware.go`：logging / CORS / recovery

新增 dashboard 端点时，先确认要落在哪个文件；横向功能（如指标采集）开新
文件，不要往 `server_handlers.go` 里堆。

## 文档漂移修正说明

本节记录 structure.md 与实际仓库**已确认**的差异，避免未来误读：

- **`contrib/` 当前不含 Homebrew formula**：Homebrew 分发改由 `.github/workflows/release.yml` 推送（避开 GoReleaser v2 的 brews 段已知 bug），不再落在 `contrib/` 下
- **`internal/skill/` 与 `contrib/skills/` 无软链 / 拷贝关系**：skill bundle 是独立脚本化分发（`contrib/skills/scripts/`），源码与 bundle 是两套产物
- **`ide/templates/` 实际是 `agents_*.md` / `instructions_*.md` 模板**：不是 `.tmpl` 文件，是 IDE 指令文件直接复制粘贴的目标
