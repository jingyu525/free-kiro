# docs-template-marker-explanation

本仓库的 IDE 指令文件模板（`internal/ide/templates/{instructions,agents}_{zh,en}.md`）
的"项目上下文（自动注入的 steering）"段只解释了底部
`<!-- free-kiro-managed:start/end -->` marker 由 `steering inject` 注入，
**没有**解释顶部由 `prependMarker`（`internal/ide/ide.go:531-545`）运行时
注入的 `# free-kiro-managed:` 文件级 marker。

用户/agent 阅读 `CLAUDE.md` / `AGENTS.md` 等实际生成文件时，看到顶部 marker
紧挨 H1 标题，容易误判为"重复 / bug / 早期 inject bug 残留"。本次 spec
不修改运行时行为（顶部 marker 是有意设计），只优化文档让两套 marker 的
注入源和用途**自描述**。

## User Stories

- As a 用户 / agent 阅读 CLAUDE.md / AGENTS.md 等 IDE 指令文件 I want 顶部 `# free-kiro-managed:` 和底部 `<!-- free-kiro-managed:start/end -->` 两套 marker 在模板正文里都被命名 + 标注注入源 + 标注用途 so that 我不会把顶部 marker 误诊为 bug 或冗余。
- As a free-kiro 维护者 I want `internal/ide/ide_test.go` 新增 `TestInstructionTemplate_ExplainsBothMarkers` 回归测试 so that 未来模板被改回"只解释底部 marker"的回归能被 `go test ./internal/ide/...` 立刻捕获。
- As a 用户查阅项目级文档 I want `docs/STEERING.md` §"自动注入到 IDE 指令文件"段同步说明两套 marker so that 我能从 docs 路径查到这两套 marker 的区别而不必 Read Go 源码。

## Acceptance Criteria

[AC-1] WHEN 用户阅读 4 个 IDE 指令文件模板（`instructions_zh.md` / `instructions_en.md` / `agents_zh.md` / `agents_en.md`）的"项目上下文（自动注入的 steering）"段 THE SYSTEM SHALL 同时命名至少 2 套 marker（顶部 `# free-kiro-managed:` 文件级 + 底部 `<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->` 区域级），并标注两者各自的注入源（`prependMarker` / `steering inject`）与用途（识别 ownership / 划定注入区）。

[AC-2] WHILE 模板处于 zh / en 任一语言状态 THE SYSTEM SHALL 在 4 个模板中至少 2 个使用中文、至少 2 个使用英文解释两套 marker，且 marker 字符串原文（`# free-kiro-managed:` / `<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`）保留不译。

[AC-3] IF 用户手动删除了顶部 `# free-kiro-managed:` 行 THEN `free-kiro init` / `init --ide auto --overwrite-instructions` 二次运行 THE SYSTEM SHALL 在 1 次调用中重新注入顶部 marker 行到 YAML frontmatter 闭合 `---` 之后，且不会重复插入 2 行及以上。

[AC-4] IF 4 个模板中任意 1 个的"项目上下文"段未同时包含 `# free-kiro-managed:` 字符串与 `<!-- free-kiro-managed:start -->` 字符串 THEN `go test ./internal/ide/...` THE SYSTEM SHALL 报告 `TestInstructionTemplate_ExplainsBothMarkers` 失败，且失败信息 SHALL 至少包含 1 处未命中模板的文件路径。

[AC-5] IF 4 个模板中任意 1 个的"项目上下文"段未出现 `prependMarker` 或等效中文术语（`init 注入` / `运行时注入`）THEN `TestInstructionTemplate_ExplainsBothMarkers` THE SYSTEM SHALL 报告测试失败，失败信息 SHALL 至少 1 处明确指出顶部 marker 来源未被提及。

[AC-6] THE SYSTEM SHALL 保留 `internal/ide/ide.go` 的 `prependMarker` 函数（行 517-545，共 29 行）与 `IsFreeKiroInstruction` 函数（行 418-440，共 23 行）的现有签名、行为契约不变 —— 两套 marker 的运行时区分是有意设计，本次只优化文档说明。

[AC-7] WHEN 用户查阅 `docs/STEERING.md` §"自动注入到 IDE 指令文件"段 THE SYSTEM SHALL 在该段 markdown 中至少包含 4 个 `free-kiro-managed` 字面量引用（顶部 marker 1 处 + 底部 marker 1 处 + `prependMarker` 1 处 + `steering inject` 1 处），覆盖两套 marker 的全部要素。

## Out of Scope

- 不修改 `prependMarker` / `IsFreeKiroInstruction` 函数行为
- 不修改 `steering inject` 的 marker 匹配逻辑（属于 `fix-inject-marker-match` spec 范围，已 done）
- 不引入新的 marker 字符串或命名空间
- 不修改 IDE settings（hook 配置）文件中的 hook marker（`FreeKiroHookPrefix`）
- 不修改 `WriteSingleInstruction` / `WriteAgentsMD` 调用流程
- 不为 4 个模板增加新的 section 或重排现有 section 顺序（仅改"项目上下文"段）
- 不动 `internal/steering/` 包任何代码