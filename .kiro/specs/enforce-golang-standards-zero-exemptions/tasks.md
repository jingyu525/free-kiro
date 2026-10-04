# enforce-golang-standards-zero-exemptions — Tasks

- [x] #1 迁移 `.golangci.yml` 从 v1 到 v2 schema：加 `version: "2"`；`gofmt`/`goimports` 从 `linters.enable` 移到 `formatters.enable`；`output.formats` 由 slice 改 map；`goimports.local-prefixes` 移到 `formatters.settings`
- [x] #2 移除 `.golangci.yml` 中 `path: 'internal/(spec|lint|steering|hooks|skill|upgrade|workspace|ide)/.*\.go$'` 的 exclude-rule [deps: #1]
- [x] #3 `.github/workflows/ci.yml` 的 `lint-go` job 已无 `continue-on-error`（CI 默认硬门禁）；更新 job 注释"历史欠账已按目录豁免"为"全代码库零豁免，硬门禁生效" [deps: #1]
- [x] #4 执行 `gofmt -s -w .` 一键格式化全部 .go 文件，清掉 86 个 gofmt 告警 [deps: #2]
- [x] #5 `internal/cli/*.go` 中 cobra handler 的 `args []string` 未用参数改名 `_`（19 处 unused-parameter）[deps: #4]
- [x] #6 `internal/cli/spec_from_prd.go` 重命名本地 `min` 函数（redefines-builtin-id）[deps: #4]
- [x] #7 `internal/cli/spec_from_prd_test.go` 改写 QF1001 De Morgan 比较 [deps: #4]
- [x] #8 `internal/cli/upgrade.go` 移除 SA4023 always-true 比较 [deps: #4]
- [x] #9 `internal/cli` 抽出 `writeOut/printOK/printErr` helper 并把 ~150 处 `fmt.Fprintf`/`Fprintln` 调用替换为 helper（模式 A/B）[deps: #5]
- [x] #10 `internal/visualize` 全部 `fmt.Fprintf`/`Fprintln` 加 `_ =` 前缀（22+15=37 处）[deps: #4]
- [x] #11 `internal/skill` `os.RemoveAll`/`src.Close`/`dst.Close` 等 errcheck 加 `_ =`（~15 处）；`internal/upgrade` `resp.Body.Close` defer 改闭包（2 处）；`internal/watch` `fs.Close`/`fmt.Fprintf` 加 `_ =`（~8 处）；`internal/lint` `eng.NewSpec`/`Close` 加 `_ =`（~5 处）；`internal/spec` `eng.NewSpec`/`Close` 加 `_ =`（~6 处）；`internal/errors`/`models`/`hooks`/`ide`/`steering`/`taskgraph`/`workspace`/`frontmatter`/`cmd` 剩余 errcheck ~30 处 [deps: #4]
- [x] #12 全仓库 `revive exported` 缺注释补单行 doc（~20 处 const/func/var），重点 `internal/cli/doctor.go` `GitHubRepo` 等 const [deps: #4]
- [x] #13 `.kiro/steering/coding-style.md` 第 8 章追加"零豁免 + //nolint 必须附理由 + PR reviewer 签字"条款 [deps: #12]
- [x] #14 端到端验证：`make lint-go` + `golangci-lint run ./...` 均 exit 0 且无 issue；`make test` 仍 pass；统计全仓库 `//nolint:` 注释 ≤ 5 处 [deps: #3,#5,#6,#7,#8,#9,#10,#11,#12,#13]
