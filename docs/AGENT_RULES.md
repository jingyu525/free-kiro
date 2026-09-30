# free-kiro AI Agent 协作规则

> 适用对象：所有对本仓库（`github.com/jingyu525/free-kiro`）自动或半自动
> 修改 Go 代码 / 文档的 AI agent（Claude Code、CodeBuddy、OpenCode 等）。
> 本文档**不是** Go 编码规范——通用 Go 风格请见
> [`CODING_STYLE.md`](./CODING_STYLE.md)；本仓库的项目策略（非 AI 协作、
> 团队约定层面）请见 [`POLICY.md`](./POLICY.md)。
>
> **配套**：`docs/CODING_STYLE.md` 第 8 章已整体迁移至此文档；
> `.golangci.yml` 聚合的 linter 与 `free-kiro spec` / `free-kiro lint`
> 是本文档的硬门禁实现。

## 目录

1. [硬性要求](#1-硬性要求)
2. [上下文注入](#2-上下文注入)
3. [失败处置](#3-失败处置)
4. [由工具强制](#4-由工具强制)
5. [零豁免政策（zero-exemption）](#5-零豁免政策zero-exemption)
6. [零死代码政策（zero-dead-code）](#6-零死代码政策zero-dead-code)

---

## 1. 硬性要求

> 违反任意一条 = PR 拒收。

1. **先 spec 后代码**。任何涉及 > 50 行新增 / 改动的 Go 代码，必须先
   有 `.kiro/specs/<name>/{requirements,design,tasks}.md` 三件套，且
   `free-kiro lint <name>` 全绿。参见 `.kiro/AGENTS.md`。
2. **零 `// TODO` / `// FIXME` / `// XXX`**。AI 生成代码不允许留任何形式
   的占位符注释。如果某功能未完成，**不要写代码**，先回 spec 阶段补
   requirements/design。
3. **零吞错误**。AI 不允许写 `_ = doX()`、`if err != nil { /* ignore */ }`、
   `log.Print(err)` 后继续。错误必须按 `CODING_STYLE.md` 第 2 章处理或 wrap。
4. **零硬编码 magic number**。常量必须有 `const` 或具名变量；端口、超时、
   阈值都要可配置或位于 `internal/config`。
5. **必须跑 `go vet ./...` 与 `gofmt -l`** 后再交付。CI 会再跑一遍。

## 2. 上下文注入

- SessionStart hook 会自动跑 `free-kiro spec next`，告诉 AI 当前活跃
  spec 的下一步动作。
- AI 在动笔前应主动 `Read`：
  1. `.kiro/specs/<current>/requirements.md`
  2. `.kiro/specs/<current>/design.md`
  3. `.kiro/specs/<current>/tasks.md`
  4. `docs/CODING_STYLE.md`（通用 Go 风格）
  5. `docs/AGENT_RULES.md`（本文档）
  6. `docs/POLICY.md`（项目策略）

## 3. 失败处置

如果 lint / test 失败：

1. AI 不应"瞎改到通过"——先读错误，理解根因。
2. 如果是 spec 不全 → 回 spec 阶段补 requirements/design，不要改代码绕。
3. 如果是规范冲突 → 在 PR 描述里说"违反第 N 章规则，原因是 X，请评审
   是否豁免"，**不要静默 `//nolint`**。

## 4. 由工具强制

- `free-kiro lint`（spec EARS + 结构）
- `golangci-lint run`（Go 源码规范）
- `go test -race ./...`（并发正确性）
- `free-kiro spec complete <name>`（任务未全部完成不允许 complete）

---

## 5. 零豁免政策（zero-exemption）

> 由 `enforce-golang-standards-zero-exemptions` spec 落地（2026-Q4）。

**核心条款**：

1. `.golangci.yml` 的 `issues.exclude-rules` 不允许包含任何按
   `path: 'internal/...'` 的目录级豁免。新代码 + 旧代码一视同仁。
2. 全仓库 `//nolint:<linter>` 注释总数 **≤ 5**（任何一行都不算豁免）。
3. 任何新增 `//nolint` 必须紧跟 `//nolint:reason <一句话解释>`，
   解释为什么这条规则在该处不适用。**不带 reason 的 `//nolint`
   视为违规，PR reviewer 必须拒收**。
4. CI lint job（`.github/workflows/ci.yml` 的 `lint-go`）不允许使用
   `continue-on-error: true` 兜底；任何 lint ERROR 直接阻断 merge。

**新增 `//nolint` 的审批流程**：

1. 在 PR 描述里写明：`//nolint: <linter> at <file:line>; reason: <X>`
2. reviewer 在 PR 上签字确认（GitHub PR review approval）
3. 维护者在合并前更新本节统计数字（`//nolint` 总数）

**当前 `//nolint` 总数**：0。

---

## 6. 零死代码政策（zero-dead-code）

> 由 `enforce-golang-standards-zero-dead-code` spec 落地（2026-Q4）。
> 与第 5 章零 `//nolint` 豁免并列适用。

**核心条款**：

1. `staticcheck U1000`（unused）必须被 `.golangci.yml` 的
   `linters-settings.staticcheck.checks: ["unused"]` 显式开启——CI 命中
   即视为硬门禁违规，与 errcheck / revive / gofmt 同等待遇。
2. 任何新增未引用的 package-level 函数、常量、类型、变量，都必须在
   PR 阶段就清理掉，不允许"先写后删"的过渡状态进入 main 分支。
3. 删除前必须 `grep -rn '<符号名>' .` 全仓库确认无 caller；删除后
   再跑一遍 `go test ./...` + `golangci-lint run ./...` 确认无回归。
4. build tag 隔离的文件（`//go:build !xxx`）默认不视为死代码；但如果
   该文件最终被实际 build tag 包含进入二进制，则其内部所有符号必须被
   引用，否则视为死代码清理对象。

**当前死代码（U1000）总数**：0。

---

> 文档结束。变更请联系 `.kiro/specs/golang-coding-standards/` 的维护者，
> 任何修改需同步更新 `.golangci.yml`、`.github/workflows/ci.yml`、
> `Makefile`、`CONTRIBUTING.md`、`.kiro/AGENTS.md`、`docs/STEERING.md`。
