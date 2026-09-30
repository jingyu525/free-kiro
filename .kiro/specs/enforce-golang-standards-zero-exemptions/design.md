# enforce-golang-standards-zero-exemptions — Design

## Architecture

按"配置先行 → 自动可修 → 人工按波次清理 → CI 切硬门禁"四阶段推进。每阶段
都是"可验证产出"，便于 review 时逐段 sign-off。

```
   阶段 1                 阶段 2                  阶段 3                  阶段 4
   .golangci.yml    →    自动修复        →       人工清理        →       CI 切换
   v1 → v2 schema        (gofmt -s -w)          (errcheck/revive/       (硬门禁)
                         + 删 exclude-rule       staticcheck 逐包)
                         (临时软门禁仍 WARN)
```

## Components

| Component | Responsibility | 改动量 |
|---|---|---|
| `.golangci.yml` | v1→v2 schema 迁移；删除 `internal/(...)/.*\\.go$` exclude-rule；`formatters` 拆分；`output.formats` 改 map | 1 文件 ~20 行 |
| `cmd/free-kiro/main.go` 等 86 个 .go 文件 | `gofmt -s -w` 一键格式化 | 86 文件（自动） |
| `internal/cli/*.go`（cobra handlers） | `unused-parameter`：`args []string` → `_ []string`；19 处 `fmt.Fprintf`/`Fprintln` 加 `_ =` 前缀或抽 `printOK/printErr` helper | ~150 处 |
| `internal/visualize/*.go` | 22 处 `fmt.Fprintln`/`Fprintf` 加 `_ =`；6 个文件 gofmt | ~50 处 |
| `internal/skill/*.go` | `dst.Close`/`src.Close`/`os.RemoveAll` 加 `_ =`；7 个文件 gofmt；1 个 `exported` 补注释 | ~30 处 |
| `internal/lint/*.go` | 8 个文件 gofmt；2 个 `eng.NewSpec` 加 `_ =` | ~10 处 |
| `internal/errors/*.go` | 1 处 `exported` 补注释；8 处 errcheck | ~10 处 |
| `internal/watch/*.go`、`internal/upgrade/*.go` 等 | 小量 errcheck/gofmt | 各 < 10 处 |
| `internal/cli/spec_from_prd.go` | 重命名本地 `min` 函数（redefines-builtin-id） | 1 处 |
| `internal/cli/spec_from_issue_test.go` | QF1001 De Morgan 改写 | 1 处 |
| `internal/cli/upgrade.go` | SA4023 always-true 比较移除 | 1 处 |
| `docs/CODING_STYLE.md` | 第 8 章追加"零豁免 + //nolint 必须附理由"条款 | 1 段 ~10 行 |
| `.github/workflows/ci.yml` | `lint-go` job 移除 `continue-on-error: true` | 1 处 |

## Data Model

无。这是代码风格规约工程，无数据结构变更。

## Error Handling — lint 问题处理策略（统一约定）

为了避免每个 case 重复决策，本 spec 强制使用以下三种模式：

### 模式 A：`_ = <call>`

适用于：
- `fmt.Fprintf`/`fmt.Fprintln`/`fmt.Fprint`/`io.Writer.Write` 写到
  cobra command 的 stdout/stderr（写失败无法恢复，CLI 用户能感知）
- `os.RemoveAll(path)`（best-effort 清理）

模式：`_ = fmt.Fprintf(cmd.OutOrStdout(), "...")`

### 模式 B：helper 包装（仅当某类调用 > 5 处时抽 helper）

仅 `internal/cli` 的 cobra 命令 print 集中：抽 1 个 helper
`func writeOut(w io.Writer, format string, a ...any)` 内部 `_ = fmt.Fprintf(w, format, a...)`。
避免在 150+ 处机械加 `_ =` 让 diff 噪音过大。

### 模式 C：defer 闭包

适用于：`defer resp.Body.Close()` 这类必须在 defer 调用的资源释放。

```go
defer func() { _ = resp.Body.Close() }()
```

不接受的"模式"：
- ❌ `//nolint:errcheck` 不带理由（违反规范第 2 章"零吞错误"）
- ❌ 把 `err` 接住后 `log.Println` 吞掉（同样是吞错误）
- ❌ 改函数签名增加 `error` 返回（cobra `RunE` 已支持，本项目部分 command
  仍是 `Run`；本次不强制改签名，超出 scope）

### unused-parameter 处理

cobra `func(cmd *cobra.Command, args []string)` 中 `args` 未用 → 改为
`func(cmd *cobra.Command, _ []string)`。**不改名为 `_ cmd`**（cobra 签名
习惯 + IDE 跳转需要保留 cmd）。

### exported 缺注释

按 `revive` `exported` 规则补 `// FuncName ...` 单行 doc，对齐
`docs/CODING_STYLE.md` 第 6 章"注释与文档"。

## Testing Strategy

- **机械验证**：`make lint-go` 必须在每个任务完成后 exit 0。
- **CI 模拟**：本地用 `act`（如已装）或直接 `golangci-lint run ./...`
  验证。CI 上 `lint-go` job 不再 `continue-on-error: true`，失败即
  阻断 PR。
- **regression**：`make test`（已存在的 `go test ./...`）继续通过；
  本次不引入新测试用例（纯风格清理不应改变行为）。
- **golden 文件**：无需。所有改动可由 lint 输出驱动机械定位。
- **人工 review sign-off**：每 wave 完成后由维护者在 PR 上确认
  "无新增 `//nolint`"。

## Migration / Rollout

按 wave 顺序执行，单 PR 落地（清理量适中 ~327 处，单 PR 可 review）：

```
Wave 0 (config)        #1 .golangci.yml v1→v2 + 删 exclude-rule
                       #2 .github/workflows/ci.yml 切硬门禁（同步）
Wave 1 (auto)          #3 gofmt -s -w . （86 个文件，一键）
Wave 2 (mechanical)    #4 internal/cli unused-parameter 19 处
                       #5 internal/cli redefines-builtin-id min 重命名
                       #6 internal/cli 测试文件 unused-parameter r → _
Wave 3 (errcheck big)  #7 internal/cli print helper + 替换 ~150 处
                       #8 internal/visualize errcheck 替换
                       #9 internal/skill/upgrade/watch errcheck 替换
Wave 4 (revive misc)   #10 全仓库 exported 补注释 ~20 处
Wave 5 (staticcheck)   #11 QF1001/SA4023 改 2 处
Wave 6 (docs)          #12 docs/CODING_STYLE.md 追加零豁免条款
Final                  #13 端到端：make lint-go + golangci-lint run ./... 双 0
```

不采用 feature flag：lint 规则一开就硬门禁，分阶段提交是"PR review 友好"
而非"运行时灰度"。
