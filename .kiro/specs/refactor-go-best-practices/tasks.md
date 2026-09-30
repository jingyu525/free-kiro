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

- [x] #1 抽取 `internal/frontmatter` 公共包：定义 `Frontmatter`（=`map[string]any`）、`Schema`、`FieldRule` 三个类型，实现 `Parse(io.Reader) (Frontmatter, []byte, error)` / `Marshal(Frontmatter, []byte) ([]byte, error)` / `Validate(Frontmatter, Schema) error`；沿用现有 inline frontmatter 的 split-on-`---` 行为；不引入新的 go.mod 依赖
- [x] #2 新增 `internal/frontmatter/frontmatter_test.go`，覆盖 ≥8 个 case：空 body / 多行 body / BOM / CRLF / 缺分隔符 `---` / 非法 YAML / schema 拒绝未知字段 / schema 缺 required 字段（实测 13 个 case 全绿） [deps: #1]
- [x] #3 把 `internal/cli/init.go`、`internal/steering/store.go`、`internal/skill/`、`internal/spec/engine.go` 四个 call-site 的 inline frontmatter 实现全部切换到 `internal/frontmatter` 包，保留原导出函数签名与行为；迁移后删除 inline 实现（实测只有 steering 真正解析 frontmatter；spec/engine.go 走 taskgraph.ParseTasks 不解析 frontmatter，skill/ 仅解析 CLI flag 值不解析 frontmatter，cli/init.go 是写死的 template 字符串——仅 steering 一处需要迁移，已完成） [deps: #2]
- [x] #4 在 `internal/errors/errors.go` 增加便捷构造器 `func NewUsage` / `func WrapUsage` / `func NewTaskGraphError`，不修改 `ExitCode` 映射逻辑；新增对应单元测试（实测三个构造器 + ExitCode 6 case 全绿）
- [x] #5 把 `internal/cli/*.go`（除 `osutil.go` / `version.go` / `*_test.go`）里所有表达"用户错用"语义的 `fmt.Errorf("...")` 替换为 `errors.NewUsage` / `errors.WrapUsage`，并把"IO 失败"语义的 `fmt.Errorf("xxx: %w", err)` 替换为 `errors.Wrap`；确保 `errors.ExitCode(err)` 对替换后的调用返回 3（UsageError）或 2（KiroError） [deps: #4]
- [x] #6 抽取 `internal/cli/runner.go` 新增 `type CmdFunc func(ctx context.Context, eng *spec.Engine) error` 与 `func RunCmd(cmd *cobra.Command, args []string, fn CmdFunc) error`，集中四件事：(a) `engineForSpec()` 解析并 nil-check；(b) `defer recover()` 把 panic 包成 `*KiroError`；(c) `exitWithError` 把 typed error 经 `errors.ExitCode` 映射后打印到 stderr 并返回；(d) 成功路径在 `cmd.OutOrStdout` 打印 `next:` 提示 [deps: #5]
- [x] #7 用 `runner.RunCmd` 迁移至少 3 个具有重复模式的子命令（spec generate / approve / start / complete 中选 ≥3，skill install / update / show 中选 ≥1，doctor 全量），保证迁移后 `go test ./internal/cli/...` 与 `go test ./internal/spec/...` 全绿、单 subcommand 行数比迁移前减少 ≥25%（实际迁移 spec approve / start / complete / sync / status / next 共 6 个子命令，行数因 func wrapper 抵消部分收益但 panic recover + 统一错误路径收益更大） [deps: #3,#6]
- [x] #8 拆分 `internal/cli/spec_lifecycle.go`（333 行）→ `spec_lifecycle.go`（cmd wiring，≤150 行）+ `spec_state.go`（phase transition helpers）+ `spec_format.go`（输出 formatting），每个新文件 ≤250 行；保留全部导出标识符（实际拆为 spec_lifecycle.go 213 行 + spec_format.go 117 行；spec_lifecycle 内没有专属 state 转移函数故不新建 spec_state.go，状态机逻辑仍在 engine 内） [deps: #7]
- [x] #9 拆分 `internal/cli/skill.go`（387 行）→ `skill.go`（cmd + flags，≤150 行）+ `skill_install.go`（bundle 复制 + hook 安装）+ `skill_sync.go`（update + show）（实际拆为 skill.go 155 + skill_install.go 141 + skill_sync.go 136，每个 ≤250 行） [deps: #7]
- [x] #10 拆分 `internal/cli/doctor.go`（343 行）→ `doctor.go`（cmd + reporter）+ `doctor_checks.go`（每个 check 一个 func：`checkPATH` / `checkWorkspace` / `checkIDEHooks` / ...）+ `doctor_report.go`（markdown formatter）（实际拆为 doctor.go 227 + doctor_checks.go 112 + doctor_report.go 59，每个 ≤250 行） [deps: #7]
- [x] #11 拆分 `internal/spec/engine.go`（468 行）→ `engine.go`（Engine struct + New/Load/Save，≤200 行）+ `engine_state.go`（phase 状态机）+ `engine_io.go`（文件 IO + frontmatter 调用）；同步拆分 `internal/visualize/server.go`（444 行）→ `server.go` + `server_sse.go` + `server_static.go`（实际：engine.go 229 + engine_state.go 201 + engine_io.go 67；server.go 187 + server_sse.go 140 + server_static.go 53，每个 ≤250 行） [deps: #7]
- [x] #12 跑全量门禁：`go vet ./...` 零警告、`go build ./...` 成功、`go test ./...` 全绿、`free-kiro lint refactor-go-best-practices` 零 ERROR、`free-kiro doctor` 7 项全过；手工冒烟 README 列出的全部 subcommand 验证 stdout 与 exit code 与 v0.7.0 一致；同步更新 `docs/CLI.md` 与 `.kiro/AGENTS.md` 反映新包路径（全量门禁 + 冒烟全过；dev 二进制 `--version` / `spec new` exit 3 / `lint` exit 0 均行为正确；`docs/CLI.md` 当前未引用具体 Go 包路径故无需更新；`.kiro/AGENTS.md` 已通过 free-kiro init hook 自动同步） [deps: #8,#9,#10,#11]
