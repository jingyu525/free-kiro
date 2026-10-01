# golangci-v2-shadow-cleanup — Tasks

- [x] #1 修 `internal/cli/demo.go:78` shadow：把内层 `if err := ...` 改为 `if err = ...` 复用外层 err
- [x] #2 修 `internal/cli/spec_from_browser.go:41` shadow：函数 named return 改内层 err 为 `navErr`
- [x] #3 修 `internal/cli/spec_from_browser.go:56` shadow：函数 named return 改内层 err 为 `dumpErr`
- [x] #4 修 `internal/cli/spec_generate.go:32` shadow：先 `var paths []string` 再 `paths, err = eng.GenerateAll(...)` 复用外层 err
- [x] #5 修 `internal/cli/spec_simple.go:36` shadow：先 `var meta *models.SpecMeta` 再 `meta, err = loadMetaViaEngine(...)` 复用外层 err
- [x] #6 修 `internal/cli/status_test.go:196` shadow：测试文件改 `err` → `execErr`
- [x] #7 修 `internal/ide/ide.go:239` shadow：函数 named return 改 `err` → `mkdirErr`
- [x] #8 修 `internal/skill/install.go:55` shadow：内层 `if _, err := ...` 改为 `if _, err = ...`
- [x] #9 修 `internal/skill/install.go:101` shadow：内层 `if err := ...` 改为 `if err = ...`
- [x] #10 修 `internal/skill/skill_test.go:222` shadow：测试文件 `if err = os.WriteFile(...)`
- [x] #11 修 `internal/skill/skill_test.go:226` shadow：测试文件 `if err = os.WriteFile(...)`
- [x] #12 修 `internal/skill/skill_test.go:229` shadow：测试文件 `if err = os.MkdirAll(...)`
- [x] #13 修 `internal/workspace/workspace.go:173` shadow：内层 `if err := ...` 改为 `if err = ...`
- [x] #14 临时改 `.golangci.yml` 的 `govet.enable-all: true`，跑 `golangci-lint run --timeout 5m` 确认 0 issues [deps: #1-#13]
- [x] #15 把 `.golangci.yml` 的 `govet.enable-all: true` 正式启用；`golangci-lint run --timeout 5m` 在 enable-all: true 模式下输出 0 issues [deps: #14]
- [x] #16 跑 `go test ./...` + `go test -race ./...` 全绿；`go vet ./...` 无输出；`free-kiro lint golangci-v2-shadow-cleanup` OK 无 drift [deps: #15]
- [x] #17 启用 enable-all: true 后 lint 暴露新发现：`internal/skill/install.go:57` 嵌套 if `if m, err := LoadManifest(...)` shadow line 45 的 err；修法：上插 `var m *Manifest` + 改 `m, err = LoadManifest(...)` 复用外层 err
- [x] #18 启用 enable-all: true 后 lint 暴露新发现：`internal/skill/skill_test.go:234` `if err := os.WriteFile(filepath.Join(target, "skill.json"), ...)` shadow line 213 的 err；修法：改 `if err = ...` 复用外层 err
