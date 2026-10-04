# free-kiro project Makefile
#
# 约定：
#   - 所有 lint 入口统一走 free-kiro（与 CI 一致），本地工具链退化有兜底
#   - 增量构建用 `make build`，CI 用 `make ci`
#   - 所有 target 在 PATH 缺失时打印清晰的安装命令
#
# 配套：.golangci.yml + .kiro/steering/coding-style.md + .github/workflows/ci.yml

SHELL := /bin/bash
GO    ?= go
BIN   := bin/free-kiro
PKG   := ./cmd/free-kiro

# ---------- 工具检测 ----------

# 是否安装了 golangci-lint（推荐 v2.x；Go 1.27 项目需 v2+）
HAS_GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null)
# PATH 上的版本是 v2.x 时直接用；否则回退到 brew / Linuxbrew 安装的 v2。
# v1.x 自带 go1.26，无法解析 go.mod 的 go 1.27 module。
LINT_BIN := $(shell \
  if [ -n "$(HAS_GOLANGCI_LINT)" ] && $(HAS_GOLANGCI_LINT) version 2>/dev/null | head -1 | grep -qE "version v[2-9]\."; then \
    echo "$(HAS_GOLANGCI_LINT)"; \
  else \
    ls /opt/homebrew/bin/golangci-lint /usr/local/bin/golangci-lint /home/linuxbrew/.linuxbrew/bin/golangci-lint 2>/dev/null | head -1; \
  fi)

# 是否安装了 free-kiro（spec/lint 入口）
HAS_FREE_KIRO := $(shell command -v free-kiro 2>/dev/null)

# benchstat? golang.org/x/perf/cmd/benchstat is the standard tool for
# comparing two benchmark output files. `go install` once; after that
# the binary lives on PATH and `make bench` picks it up.
HAS_BENCHSTAT := $(shell command -v benchstat 2>/dev/null)

# dashboard frontend build 用 pnpm（dashboard-frontend-react-vite-fsd 锁定的
# 包管理器；package.json 声明 packageManager: pnpm@9.x，engines.node >= 20）。
# corepack 默认绑定 Node 20+，缺 pnpm 时 corepack enable 一键启用。
HAS_PNPM := $(shell command -v pnpm 2>/dev/null)

# ---------- 默认 target ----------

.PHONY: help
help: ## 列出所有 target
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ---------- 构建 ----------

.PHONY: build
build: ## 编译二进制到 ./bin/free-kiro
	$(GO) build -trimpath -o $(BIN) $(PKG)

.PHONY: install
install: ## go install 到 $(GOBIN)/free-kiro
	$(GO) install $(PKG)

# 改完 internal/ide/templates/*.md 后用这个：
#   1) go install 把新模板编进 binary（否则 PATH 上的 free-kiro 还是旧的）
#   2) steering inject 把 IDE 指令文件同步成新模板
#   3) drift 校验（exit 0 = 无漂移；有漂移时打印 diff stat 供 review，不阻断）
# 详见 memory: free-kiro-template-embed-rebuild
.PHONY: reinstall-templates
reinstall-templates: install ## 改完 templates/*.md 后：重编 binary + 注入 IDE 指令文件 + drift 校验
	@if [ -z "$(HAS_FREE_KIRO)" ]; then \
		echo "free-kiro not on PATH"; exit 3; \
	fi
	@free-kiro steering inject || true
	@git_root=$$(git rev-parse --show-toplevel 2>/dev/null) || git_root=.; \
		if git -C $$git_root diff --exit-code CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md >/dev/null 2>&1; then \
			echo "✓ reinstall-templates: no drift"; \
		else \
			echo "drift (review then commit):"; \
			git -C $$git_root diff --stat CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md; \
		fi

.PHONY: clean
clean: ## 删除 ./bin 与临时构建产物
	rm -rf bin/ coverage.out benchdata/current.txt benchdata/report.txt

# ---------- 测试 ----------

.PHONY: test
test: ## 跑单元测试（带 race + coverage）
	$(GO) test -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -func=coverage.out | tail -1

.PHONY: test-fast
test-fast: ## 快速测试（不带 race）
	$(GO) test ./...

# ---------- Lint ----------

.PHONY: lint
lint: ## spec 门禁（free-kiro lint，CI 与本地一致）
	@if [ -z "$(HAS_FREE_KIRO)" ]; then \
		echo "free-kiro not found on PATH; install: brew install jingyu525/free-kiro/free-kiro"; \
		exit 3; \
	fi
	free-kiro lint

.PHONY: lint-go
lint-go: ## Go 源码 lint（golangci-lint，按 .golangci.yml）
	@if [ -z "$(LINT_BIN)" ]; then \
		echo "golangci-lint v2.x not found; install:"; \
		echo "  brew install golangci-lint"; \
		echo "  # 或"; \
		echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 3; \
	fi
	@if [ "$(LINT_BIN)" != "$(HAS_GOLANGCI_LINT)" ]; then \
		echo "golangci-lint on PATH < v2; using $(LINT_BIN)"; \
	fi
	$(LINT_BIN) run --timeout 5m ./...

.PHONY: fmt
fmt: ## gofmt + goimports 格式化
	$(GO) fmt ./...
	@if [ -n "$(LINT_BIN)" ]; then \
		$(LINT_BIN) run --no-config --disable-all -E goimports --fix ./...; \
	fi

# ---------- Dashboard frontend ----------

# dashboard 前端产物路径；internal/visualize/server.go 的
# `//go:embed all:static/dist/assets/*` 依赖该目录存在 + 含 index.html
# + assets/ 子目录至少 4 个 chunk（index / index.css / react-vendor /
# query-vendor；hash 由 vite 决定）。
DASHBOARD_DIR  := internal/visualize/static
DASHBOARD_DIST := $(DASHBOARD_DIR)/dist

.PHONY: dashboard-dist
dashboard-dist: ## build dashboard frontend（pnpm install + pnpm build，//go:embed 依赖）
	@if [ -z "$(HAS_PNPM)" ]; then \
		echo "pnpm not on PATH; install one of:"; \
		echo "  corepack enable pnpm"; \
		echo "  npm i -g pnpm@9"; \
		exit 3; \
	fi
	cd $(DASHBOARD_DIR) && pnpm install --frozen-lockfile
	cd $(DASHBOARD_DIR) && pnpm build
	@if [ ! -f $(DASHBOARD_DIST)/index.html ]; then \
		echo "FAIL: $(DASHBOARD_DIST)/index.html missing after pnpm build"; \
		exit 1; \
	fi
	@dist_files=$$(ls $(DASHBOARD_DIST) $(DASHBOARD_DIST)/assets/ 2>/dev/null | wc -l | tr -d ' '); \
		if [ "$$dist_files" -lt 5 ]; then \
			echo "FAIL: $(DASHBOARD_DIST) has $$dist_files entries (expected >= 5: index.html + assets/*)"; \
			exit 1; \
		fi
	@echo "✓ dashboard-dist: $(DASHBOARD_DIST) populated ($$(ls $(DASHBOARD_DIST) $(DASHBOARD_DIST)/assets/ | wc -l | tr -d ' ') entries)"

# ---------- Benchmark ----------

# performance-benchmarks spec 落地：6 个关键包 / 16 个 Benchmark 函数 /
# 41 个子 bench。第一次跑 `make bench` 会初始化 benchdata/baseline.txt；
# 后续每次跑与 baseline 对比，回归 ≥ 10% 报 exit 1。
BENCH_TARGETS := \
  ./internal/lint/... \
  ./internal/spec/... \
  ./internal/taskgraph/... \
  ./internal/visualize/...

.PHONY: bench
bench: ## 跑全部 benchmark + benchstat 对比 baseline（回归 ≥ 10% 报错）
	@mkdir -p benchdata
	@echo "##### running benchmarks (-benchtime=1s) #####"
	$(GO) test -run='^' -bench=. -benchmem -benchtime=1s $(BENCH_TARGETS) > benchdata/current.txt 2>&1 || true
	@if [ ! -f benchdata/baseline.txt ]; then \
		echo "##### no baseline.txt yet — initialising from current #####"; \
		cp benchdata/current.txt benchdata/baseline.txt; \
		echo "baseline.txt initialised (next 'make bench' will compare)"; \
		exit 0; \
	fi
	@if [ -z "$(HAS_BENCHSTAT)" ]; then \
		echo "benchstat not on PATH; install:"; \
		echo "  go install golang.org/x/perf/cmd/benchstat@latest"; \
		exit 4; \
	fi
	@echo "##### benchstat baseline → current #####"
	benchstat -alpha=0.10 benchdata/baseline.txt benchdata/current.txt | tee benchdata/report.txt
	@benchstat -alpha=0.10 benchdata/baseline.txt benchdata/current.txt > /dev/null; \
		status=$$?; \
		if [ $$status -ne 0 ]; then \
			echo "FAIL: benchmark regression detected (see benchdata/report.txt)"; \
			exit $$status; \
		fi; \
		echo "✓ no regression"

.PHONY: bench-init
bench-init: ## 把当前 current.txt 复制为 baseline.txt（首次 / 主动 reset）
	@if [ ! -f benchdata/current.txt ]; then \
		echo "no benchdata/current.txt — run 'make bench' first"; \
		exit 1; \
	fi
	cp benchdata/current.txt benchdata/baseline.txt
	@echo "✓ baseline.txt updated from current.txt"

# ---------- 组合 ----------

.PHONY: ci
ci: dashboard-dist lint lint-go test build ## CI 全量（dashboard frontend + spec 门禁 + Go lint + test + Go build）
	@echo "✓ ci passed"

.PHONY: precommit
precommit: fmt lint-go test drift-check ## 本地提交前（fmt + Go lint + test + drift check，不含 spec lint）
	@echo "✓ precommit passed"

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
	@free-kiro steering inject || true   # 容忍 skipped 退出码（如 .continue/rules/ 缺失）
	@if git diff --exit-code CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md >/dev/null 2>&1; then \
		echo "✓ steering inject drift check passed"; \
	else \
		echo "drift: $(git diff --name-only CLAUDE.md AGENTS.md .cursorrules .cursor/rules/free-kiro.md | tr '\n' ' ')"; \
		echo "→ 跑 'free-kiro steering inject' 同步 IDE 指令文件"; \
		exit 1; \
	fi