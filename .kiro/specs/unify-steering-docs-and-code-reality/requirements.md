# unify-steering-docs-and-code-reality

把 `.kiro/steering/{product,structure,tech}.md` 三份 always-mode 文档与
代码实现的实际状态对齐，消除 21 处文档漂移（9 条必修 + 5 条建议 + 7 条可选）。

零代码变更（全部走 A 选项：改文档承认现状），分 2 个 wave：
- **Wave 1**（必修 9 + 建议 5）：product.md + tech.md 字段级修订
- **Wave 2**（可选 7）：structure.md 升级走 β 方案（按职责分组 + 代表性文件）

## User Stories

- As a free-kiro 用户 I want `.kiro/steering/*.md` 里的状态机形态、EARS 数量、
  advisory 类型、watch presets、GoReleaser 平台数等数字描述与代码 `internal/`
  实际行为一一对应 so that 我对照文档做集成时不会因为"文档说一套、代码做一套"
  走错路径。
- As a free-kiro 用户 I want product.md 显式声明 `--from-issue` 依赖 `gh` CLI、
  `--from-browser` 依赖 `bsk` CLI so that 我在没装这些依赖时不会被冷不丁地
  报一个 `exec: gh: not found` 才知道有外部前置。
- As a free-kiro 维护者 I want structure.md 反映 80+ 个 .go 源文件的真实组织
  （按职责分组 + 代表性文件），而不是只列顶层包 so that 新人 onramp 不需要
  跑 `ls internal/cli/` 才能知道有哪些子命令。

## Acceptance Criteria

[AC-1] WHEN 工程师 `grep -E 'draft.*→.*requirements.*→.*design' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 在 product.md 内匹配至少 0 行线性箭头串（允许包含 `→` 但不得出现 `draft → requirements → design → tasks` 这种 4-段线性描述），且 product.md 必须包含至少 1 处显式提到 "planning 阶段可互转" 或 "APPROVED → TASKS" 的字符串。

[AC-2] WHEN 工程师 `grep -E '5 种 EARS' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 命中至少 1 行包含 "6 类正则分支" 的字符串（表示已把 ubiquitous 基线 + 5 种触发句式合并表述），且同一行 200 字符内必须出现 "ubiquitous" 至少 1 次。

[AC-3] WHEN 工程师 `grep -E 'vague|duplicate-acceptance|uncovered-acceptance|tasks-without-requirements' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 至少命中 3 个不同的 advisory code 字符串（`vague-language` / `duplicate-acceptance-criteria` / `uncovered-acceptance-criteria` / `tasks-without-requirements` 中至少 3 个出现在 product.md 的 advisory 行）。

[AC-4] WHERE `free-kiro spec new --from-issue <url>` 子命令被文档化时 THE SYSTEM SHALL 让 product.md 在该行 200 字符内至少出现 1 次 "gh CLI" 与 1 次 "auth login" 字符串，且行长度在 80 ~ 250 字符之间。

[AC-5] WHERE `free-kiro spec new --from-browser <url>` 子命令被文档化时 THE SYSTEM SHALL 让 product.md 在该行 200 字符内至少出现 1 次 "bsk" 与 1 次 "browser-skill" 字符串；UNLESS `bsk` CLI 不在 PATH 上，否则 product.md 必须显式声明该前置依赖。

[AC-6] WHEN 工程师 `grep -cE '(default|lint|status|reactive|full).*preset' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 计数 ≥ 5（即 watch 5 个 preset 全部出现在 product.md 的 watch 行附近）。

[AC-7] WHEN 工程师 `grep -E 'stat-card|sparkline-cell|phase-badge|waves-progress' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 至少命中 3 个 widget 名字字符串，且命中位置距 "4 widget" 字样不超过 200 字符。

[AC-8] WHEN 工程师 `grep -iE 'windows.*(手动|重启)|(手动|重启).*windows' .kiro/steering/product.md` 时 THE SYSTEM SHALL 让 grep 至少命中 1 行（确认 Windows re-exec 限制被声明）。

[AC-9] IF fsnotify 在容器 / 网络 FS 上不可用 THEN THE SYSTEM SHALL 让 product.md 与 tech.md 在 "可视化 Web" / "watch" 相关段落 200 字符内至少出现 1 次 "fallback" 或 "降级" 与 1 个时间单位（如 "2s"）字符串。

[AC-10] WHEN 工程师 `grep -E '\.prev' .kiro/steering/tech.md` 时 THE SYSTEM SHALL 让 grep 命中 0 行（确认 tech.md 不再声明 `~/.prev` 回滚路径）；同时 `grep -E 'os\.Rename' internal/upgrade/upgrade.go` 仍命中 ≥ 1 行（确认代码现状不变）。

[AC-11] WHEN 工程师 `grep -E 'CI 当前仅打印|CI 仅打印|不强制阈值' .kiro/steering/tech.md` 时 THE SYSTEM SHALL 让 grep 至少命中 1 行（确认覆盖率阈值已降级为"建议目标"且声明 CI 行为）。

[AC-12] WHEN 工程师 `grep -E '待基准验证|无.*基准|无.*常量' .kiro/steering/tech.md` 时 THE SYSTEM SHALL 让 grep 至少命中 1 行（确认性能预算已声明"待基准验证"与"无常量、无基准"）。

[AC-13] WHEN 工程师 `grep -E '5 个平台' .kiro/steering/tech.md` 时 THE SYSTEM SHALL 让 grep 至少命中 1 行；UNLESS `.goreleaser.yaml` 修改后 windows+arm64 被启用，否则 product.md 必须显式标注 "windows+arm64 当前忽略"。

[AC-14] IF engineer 在 CI 上跑 `free-kiro lint` THEN THE SYSTEM SHALL 让 `.kiro/steering/*.md` 文件路径不在 lint 扫描范围内（当前默认行为不变），且 `go test ./...` 在 Wave 1 全部任务完成后通过、退出码 0。

[AC-15] WHEN 工程师 `cat .kiro/steering/structure.md` 时 THE SYSTEM SHALL 让该文档 (a) 保留顶层布局小节；(b) `internal/cli/` 段提到 "cobra 命令文件" 与 "~22 个"；(c) `internal/lint/` 段按职责列出 6 个代表性文件（`ears.go` / `bugfix.go` / `requirements.go` / `tasks.go` / `linter.go` / `baseline.go` / `quality.go`）；(d) `internal/visualize/static/` 段提到 "Vite" 或 "FSD" 与 "src/ dist/ scripts/"；(e) 顶层布局段列出 ≥ 6 个根目录文件（`LICENSE` / `CONTRIBUTING.md` / `Makefile` / `AGENTS.md` / `CLAUDE.md` / `examples/`）。

[AC-16] THE SYSTEM SHALL 让所有 15 条文档类 AC 在 Wave 1 + Wave 2 任务完成后通过 `make verify-docs` 类的端到端校验：跑 1 次 `go build ./...`（退出 0）+ 1 次 `go test ./...`（退出 0）+ 1 次 `free-kiro lint`（退出 0）+ 1 次文档字段 grep 校验（按 AC-1 ~ AC-15 的 grep 模式逐条断言），全部 PASS 才允许 spec 进入 done。

## Out of Scope

- 不修改任何 `.go` 源代码文件（除结构文档里"可追溯引用"中提到的代码行号作为验证锚点）
- 不实现自升级 `~/.prev` 回滚（T-01 选 A：承认当前"无回滚"）
- 不实现覆盖率 CI gate（T-02 选 A：文档承认不强制）
- 不实现性能基准测试（T-03 选 A：文档承认无常量、无基准）
- 不启用 windows+arm64（T-04 选 A：文档承认跳过）
- 不修改 `README.md` / `CONTRIBUTING.md` / `AGENTS.md` / `CLAUDE.md` / `docs/*.md`
- 不修改 `Makefile` / `.github/workflows/*.yml` / `.goreleaser.yaml` / `.golangci.yml`
- 不修改 `free-kiro lint` 引擎自身
- 不为本次改动新建 `go test` 测试（无逻辑变更）
- 不动 `contrib/skills/free-kiro/skill.json` 的 SHA256（文档类改动不影响 skill bundle）
- 不为 `internal/visualize/static/` 引入新依赖或修改前端构建配置
