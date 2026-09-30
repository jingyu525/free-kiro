# free-kiro project Makefile
#
# 约定：
#   - 所有 lint 入口统一走 free-kiro（与 CI 一致），本地工具链退化有兜底
#   - 增量构建用 `make build`，CI 用 `make ci`
#   - 所有 target 在 PATH 缺失时打印清晰的安装命令
#
# 配套：.golangci.yml + docs/CODING_STYLE.md + .github/workflows/ci.yml

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

.PHONY: clean
clean: ## 删除 ./bin 与临时构建产物
	rm -rf bin/ coverage.out

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

# ---------- 组合 ----------

.PHONY: ci
ci: lint lint-go test ## CI 全量（spec 门禁 + Go lint + test）
	@echo "✓ ci passed"

.PHONY: precommit
precommit: fmt lint-go test ## 本地提交前（fmt + Go lint + test，不含 spec lint）
	@echo "✓ precommit passed"
