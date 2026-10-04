<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

依赖不能：
  - 自引用（#1 不能依赖 #1）
  - 指向不存在的任务
  - 形成循环

完成后把 [ ] 改成 [x]；wave 视图用 `free-kiro task list <spec>`。
-->

<!-- Wave 1: 创建 steering 文件 + 解决跨文档冲突（独立，无依赖） -->

- [ ] #1 创建 `.kiro/steering/coding-style.md`：从 `docs/CODING_STYLE.md` 搬迁全部内容；在文档顶部加 YAML frontmatter（`mode: always` + `description: Go 编码规范（命名/错误处理/并发/接口/测试/注释/依赖）`）；不改变正文内容顺序
- [ ] #2 创建 `.kiro/steering/agent-rules.md`：从 `docs/AGENT_RULES.md` 搬迁全部内容；在文档顶部加 YAML frontmatter（`mode: always` + `description: AI agent 协作硬性要求与零豁免政策`）；不改变正文内容顺序
- [ ] #3 创建 `.kiro/steering/policy.md`：从 `docs/POLICY.md` 搬迁全部内容；在文档顶部加 YAML frontmatter（`mode: always` + `description: free-kiro 项目策略（覆盖率/TODO/协议/依赖/commit/规模）`）；不改变正文内容顺序
- [ ] #4 解决 5 个跨文档冲突（Wave 1 期间完成）：

  - **testify 冲突**：修改 `.kiro/steering/coding-style.md` §5.1，改为「只用 stdlib `testing`，table-driven 测试优先，不引入 testify/assert/require」
  - **覆盖率阈值冲突**：修改 `tech.md` 与 `coding-style.md`，删除冲突措辞，统一指向 `policy.md`（保留硬阈值）
  - **`//nolint` 上限冲突**：修改 `tech.md`，删去重复措辞，改指向 `agent-rules.md`
  - **`-race` 必跑**：在 `tech.md`「测试」节补一行 `go test -race ./...` 必跑；`coding-style.md` 与 `agent-rules.md` 简化为引用 `tech.md`
  - **依赖升级 commit 模板**：修改 `coding-style.md` §7.2，改为引用 `policy.md`；删除重复内容

  [deps: #1, #2, #3]

<!-- Wave 2: 删除 docs/ 三件套 + 更新 IDE 引用（依赖 Wave 1） -->

- [ ] #5 删除 `docs/CODING_STYLE.md`：[deps: #1, #4]
- [ ] #6 删除 `docs/AGENT_RULES.md`：[deps: #2, #4]
- [ ] #7 删除 `docs/POLICY.md`：[deps: #3, #4]
- [ ] #8 更新 5 份 IDE 指令文件引用：将 `docs/CODING_STYLE.md` / `docs/AGENT_RULES.md` / `docs/POLICY.md` 路径替换为 `.kiro/steering/coding-style.md` / `agent-rules.md` / `policy.md`（涉及文件：CLAUDE.md / AGENTS.md / .cursorrules / `.cursor/rules/free-kiro.md` / `.kiro/AGENTS.md`），[deps: #5, #6, #7]
- [ ] #9 更新 4 份 IDE 模板引用：将 `docs/CODING_STYLE.md` / `docs/AGENT_RULES.md` / `docs/POLICY.md` 路径替换为 `.kiro/steering/coding-style.md` / `agent-rules.md` / `policy.md`（涉及文件：`internal/ide/templates/agents_en.md` / `agents_zh.md` / `instructions_en.md` / `instructions_zh.md`），[deps: #5, #6, #7]

<!-- Wave 3: 更新 spec 文档 + 其他配置文件（依赖 Wave 2） -->

- [ ] #10 更新 11 个 spec 文档的引用路径：`agents-template-coding-standards` / `golang-coding-standards` / `enforce-golang-standards-zero-exemptions` / `enforce-golang-standards-zero-dead-code` / `per-ide-instructions` / `steering-inject-to-ide` / `top1-demo-onboarding` / `performance-benchmarks` / `update-contrib-skill-bundle` 的 requirements.md / design.md / tasks.md 中的 `docs/CODING_STYLE.md` → `.kiro/steering/coding-style.md`、`docs/AGENT_RULES.md` → `agent-rules.md`、`docs/POLICY.md` → `policy.md`，[deps: #8, #9]
- [ ] #11 更新 `.kiro/specs/separate-codestyle-projectspecific/requirements.md`：在顶部 Background 段之前加 superseded 声明（"Superseded by `migrate-docs-triple-to-steering`，本 spec 历史保留不再推进"），[deps: #8, #9]
- [ ] #12 更新其他配置文件引用：`.golangci.yml`（L3, L4）、`Makefile`（L8）、`CONTRIBUTING.md`（L32, L34, L36, L39, L98）、`.github/workflows/ci.yml`（L55）、`internal/ide/ide_test.go`（L45 注释）、`internal/visualize/server_handlers.go`（L327 注释），[deps: #8, #9]

<!-- Wave 4: 重编 binary + steering inject + 验证（依赖 Wave 3） -->

- [ ] #13 `make install` 重编 binary（将更新后的 IDE 模板 embed 进 binary），[deps: #9]
- [ ] #14 `free-kiro steering inject` 重生成 5 份 IDE 指令文件底部 `<!-- free-kiro-managed:start -->` ... `<!-- free-kiro-managed:end -->` 块，确认块内包含 `coding-style.md` / `agent-rules.md` / `policy.md` 的内容，[deps: #13]
- [ ] #15 人工 grep 复核 0 命中 `grep -rE "docs/(CODING_STYLE|AGENT_RULES|POLICY)\.md" .`（排除 `.kiro/specs/separate-codestyle-projectspecific/`），[deps: #10, #11, #12]
- [ ] #16 跑全量门禁：`go vet ./...` 零警告、`go build ./...` 成功、`go test ./...` 全绿、`free-kiro lint migrate-docs-triple-to-steering` 零 ERROR、`make lint-go` 零 ERROR，[deps: #13, #14, #15]
- [ ] #17 `free-kiro spec complete migrate-docs-triple-to-steering`：[deps: #16]
