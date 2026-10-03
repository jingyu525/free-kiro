# ci-steering-inject-drift-check — Design

## Architecture

3 层防御，按"近用户 → 远端"顺序：

```
┌──────────────────────────────────────────────────────────┐
│ Layer 1: pre-commit hook（.githooks/pre-commit）        │
│   - 仅本地，开发体验最好                                │
│   - 只在 .kiro/steering/*.md / templates 改动时触发      │
│   - 自动 git add 4 个 IDE 指令文件                       │
└──────────────────────────────────────────────────────────┘
                          │ 失败 abort commit
                          ▼
┌──────────────────────────────────────────────────────────┐
│ Layer 2: Makefile drift-check（任何时机手动跑）          │
│   - make drift-check                                    │
│   - inject + git diff --exit-code 检查                   │
│   - 本地开发期反复用                                    │
└──────────────────────────────────────────────────────────┘
                          │ 失败 exit 1
                          ▼
┌──────────────────────────────────────────────────────────┐
│ Layer 3: CI workflow step（远端兜底）                   │
│   - .github/workflows/ci.yml 加 drift-check-steering    │
│   - 跑 make drift-check                                 │
│   - drift 时 fail PR                                    │
└──────────────────────────────────────────────────────────┘
```

3 层互不依赖；任一层有效都能挡住 drift。pre-commit 让本地 commit 时
就已同步；CI 兜底防止有人 `git commit --no-verify` 绕过。

## Components

| Component | Responsibility | Path |
|---|---|---|
| pre-commit hook | git commit 前自动跑 inject + git add | `.githooks/pre-commit`（新文件） |
| Makefile target `install-hooks` | `git config core.hooksPath .githooks` + chmod | `Makefile`（追加） |
| Makefile target `drift-check` | inject + git diff --exit-code 检查 | `Makefile`（追加） |
| CI step | 跑 `make drift-check`，drift 拒收 | `.github/workflows/ci.yml`（追加） |

## Pre-commit hook 设计

**文件** `.githooks/pre-commit`（bash，~25 行）：

```bash
#!/usr/bin/env bash
# .githooks/pre-commit — auto-run free-kiro steering inject on commit.
# 用法：
#   make install-hooks    # 一次性配置 git core.hooksPath
#   git commit            # 自动触发本 hook

set -euo pipefail

# 仅在 .kiro/steering/*.md / internal/ide/templates/*.md 改动时跑
# （避免每次 commit 都跑 inject，~50ms 的开销可忽略但仍可省）
if ! git diff --cached --name-only -- \
    .kiro/steering/ internal/ide/templates/ 2>/dev/null \
    | grep -q .; then
    exit 0
fi

if ! command -v free-kiro >/dev/null 2>&1; then
    echo "free-kiro not on PATH; install: brew install jingyu525/free-kiro/free-kiro" >&2
    echo "or: make install && export PATH=\$HOME/go/bin:\$PATH" >&2
    exit 1
fi

if ! free-kiro steering inject; then
    echo "free-kiro steering inject failed; commit aborted" >&2
    exit 1
fi

# inject 退出码可能为 1（部分目标文件 skipped，如 .continue/rules/...），
# 上面已 set -e，只在 inject 真正报错时 abort；下面 git add 无条件跑。
git add CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md 2>/dev/null || true
```

## Makefile 追加

在现有 `precommit` target 之后、`precommit` 之前插入 2 个新 target：

```makefile
.PHONY: install-hooks
install-hooks: ## 配置 git core.hooksPath 指向 .githooks（一次性）
	git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit
	@echo "✓ git hooks installed; pre-commit will run \`free-kiro steering inject\`"

.PHONY: drift-check
drift-check: ## 检查 IDE 指令文件与 .kiro/steering/ 是否 drift
	@if [ -z "$(HAS_FREE_KIRO)" ]; then \
		echo "free-kiro not on PATH"; exit 3; \
	fi
	@free-kiro steering inject || true   # 容忍 skipped 退出码
	@if git diff --exit-code CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md >/dev/null 2>&1; then \
		echo "✓ steering inject drift check passed"; \
	else \
		echo "drift: $(git diff --name-only CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md | tr '\n' ' ')"; \
		echo "→ 跑 'free-kiro steering inject' 同步 IDE 指令文件"; \
		exit 1; \
	fi
```

并修改现有 `precommit` target：把 `drift-check` 加进去。

## CI workflow step

在 `.github/workflows/ci.yml` 的现有 `lint-go` step 之后插入：

```yaml
      - name: Drift check (steering inject)
        run: make drift-check
```

CI 容器里已 `go install` 出 free-kiro + make 都可用。失败 → job exit 1 → PR 拒收。

## 关键决策记录

- **为什么 pre-commit hook 直接调 `free-kiro` 而非 `make drift-check`**：
  - Makefile 在 CI 容器里 100% 有，但开发者本地可能没装 make（罕见但有）
  - pre-commit hook 要"开箱即用"，少 1 层依赖 = 更可靠
  - `free-kiro` 二进制在 init 后已经存在（homebrew / install.sh 装的）
  - hook 25 行内，避免复杂 shell 逻辑

- **为什么 4 个目标文件不用 `DefaultInjectTargets` 常量**：hook 是
  纯 bash，不调 Go。硬编码 4 个路径简单可靠；如果将来 init 改了
  默认目标列表，hook 需要同步更新——这是已知的耦合点，由 spec
  `steering-inject-to-ide` 维护者负责同步通知。

- **为什么只跑在 .kiro/steering/ 或 templates 改动时**：避免每次
  commit 都 ~50ms 开销；更重要的是保持 commit 噪音最小（没改
  steering 时不应该看到 inject 输出）。

- **为什么不调 `make drift-check` 在 hook 里**：避免 make 依赖 +
  让 hook 跨平台；hook 单独跑 inject 已经够用，drift 检测靠 CI。

- **为什么不进 `precommit` Makefile target**：precommit 是给开发者
  本地手跑的入口（`make precommit`），包含 fmt + lint-go + test +
  drift-check；CI 走 `make ci`（包含 lint + lint-go + test）独立加
  drift-check step。`make precommit` + `make install-hooks` 是两套
  本地防御。

## 验证策略

1. **Makefile target 行为**：本地手跑 `make install-hooks` 后
   `git config core.hooksPath` 应输出 `.githooks`；跑
   `make drift-check` 在干净状态下应 exit 0、改 steering 后应 exit 1。
2. **pre-commit hook 行为**：手动修改 `.kiro/steering/product.md` +
   `git add` + `git commit`，观察 inject 是否自动跑、4 个 IDE 指令
   文件是否自动被 add。
3. **CI 行为**：本地模拟——在另一个分支故意不跑 inject 就 commit，
   push 后看 CI 是否报 drift 错误（这一步靠 PR review，不强制自动化测）。
4. **跨平台**：在 macOS / Linux 各跑一次 `make install-hooks`。

## 风险与回滚

- 风险 1：开发者本地没装 `free-kiro` → pre-commit abort commit，
  报错提示安装命令。**不是 bug**，是有意的"fail-fast"。
- 风险 2：CI drift 检查误报（inject 写入空内容触发 diff）→ 用
  `git diff --exit-code` 配合 `--` 分隔符限定 4 个目标文件，已规避。
- 回滚：删 `.githooks/pre-commit` + Makefile 2 行 + CI 1 step 即可。
