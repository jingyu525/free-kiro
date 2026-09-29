---
mode: auto
description: Go 技术栈、依赖、构建、测试与发布规范
---

# Tech

## 语言与运行时

- **Go 1.22+**（README badge；go.mod 当前 go 1.27.0，往下兼容到 1.22）
- `CGO_ENABLED=0`：纯静态二进制，无 glibc / libc++ 运行时依赖
- 单文件 `cmd/free-kiro/main.go` 入口

## 直接依赖（3 个）

| 包 | 用途 |
|---|---|
| `github.com/spf13/cobra` | CLI 命令树 + flag 解析（间接拉 `pflag`） |
| `github.com/fsnotify/fsnotify` | `free-kiro watch` 文件监听 |
| `golang.org/x/net/html` | `--from-prd` HTML 解析 |

**不引入**：`testify`、`gomock`、`mockery`、`logrus`、`zap` 等。新依赖必须
写明理由并 review。

## CLI 框架

- Cobra 子命令注册集中在 `internal/cli`
- flag 命名：小写 kebab，bool 用 `--xxx`/`--no-xxx` 对
- 帮助文本：中文为主，关键命令（spec / init / lint）必须有 1-3 行示例

## 测试

- **只用** `go test ./...` + 标准库 `testing`
- 不引入 testify / gomock / mockery
- table-driven 测试优先
- fixture 放 `testdata/`，不要在测试里 inline 大段 mock 数据
- 目标覆盖率：核心 lint 引擎 ≥ 80%，其他 ≥ 60%

## 错误处理

- `errors.Is` / `errors.As` / `fmt.Errorf("...: %w", err)`
- 包级 sentinel 错误集中在 `internal/errors`
- 用户面向的错误信息：中文，1 行说"出了什么事"+ 1 行说"怎么修"

## 格式化与静态检查

- `gofmt` + `goimports`（CI 必跑）
- `go vet ./...`（GoReleaser before hook 跑）
- **不引入** `golangangci-lint`（保持工具链极简，规则用 vet + 评审兜底）

## 构建与发布

- **GoReleaser v2**（`.goreleaser.yaml`）
- 跨 6 个平台：`darwin` / `linux` / `windows` × `amd64` / `arm64`
- 二进制名固定 `free-kiro`，产物在 `dist/`
- release 流程：`git tag vX.Y.Z && git push --tags` → CI 自动跑
  - `go mod tidy` + `go test ./...`
  - `compute-skill-sha.sh` 重算 skill bundle SHA256
  - GoReleaser build + GitHub Release
  - SHA256 自动写到 `contrib/skills/free-kiro/skill.json`

## 自升级协议

- `free-kiro upgrade` 从 GitHub Releases API 拉最新 tag
- 下载 + `sha256sum -c` 校验 → `os.Exec` 重启自己
- 失败回滚路径：保留上一份二进制到 `~/.local/share/free-kiro/.prev`

## 性能预算

- CLI 冷启动 < 100ms（不要在 init 路径上做重 IO / 重计算）
- `free-kiro lint` 在 10 个 spec 的 `.kiro/` 下 < 500ms
- web dashboard SSE 推送延迟 < 200ms

## 命名规范

- 包名：小写单词，不复数
- 导出符号：PascalCase；私有：camelCase
- 常量：MAX_RETRIES / defaultTimeout 这种全大写下划线只用于"真常量"
- 文件名：小写下划线（`spec_loader.go` 而非 `specLoader.go`）