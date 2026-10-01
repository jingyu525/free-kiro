# upgrade-shasums-filename — Tasks

- [x] #1 抽未导出辅助函数 `verifyTarballSHA256(sumsFile []byte, tarballURL string) (string, error)`（紧挨 `LookupSHA256` 之后），把 `Apply` 调用点改为 `verifyTarballSHA256(sumsFile, p.Download)`；import 加 `"path"`
- [x] #2 删 `LookupSHA256` 内 `name == "free-kiro"` fallback 分支（死代码），更新函数注释去掉"Falls back to free-kiro for the binary inside a tarball"一句；同步删 `upgrade_test.go` 里 `TestLookupSHA256_FallbackToBinaryName`（覆盖的就是被删的 fallback 路径，测试本身也变死代码）
- [x] #3 新增单元测试 `TestVerifyTarballSHA256`：覆盖 happy path（URL basename 命中 sha256）+ regression guard（传 "free-kiro" 必须失败，证明 fallback 死代码已删）；同步删旧的 `TestApply_LookupSHA256UsesTarballBasename`（名字误导，实际没调 Apply） [deps: #1,#2]
- [x] #4 跑 `gofmt -l` + `go vet ./...` + `go test -cover ./internal/upgrade/...` 全绿；`verifyTarballSHA256` 行覆盖率 ≥ 80%；如失败按错误信息就地修，不新开 spec [deps: #3]
- [x] #5 跑 `free-kiro lint upgrade-shasums-filename` 全绿；`free-kiro spec status upgrade-shasums-filename` 显示无 drift [deps: #4]
- [x] #6 手动 e2e：`go build -o /tmp/free-kiro-fix ./cmd/free-kiro` 产出新 binary → 跑 `/tmp/free-kiro-fix upgrade` 完整流程。证据：第一次输出 `current: vdev → latest: v0.8.0 → status: update available`，第二次输出 `current: v0.8.0 → status: already on the latest version`，证明 Apply 成功 Download + verifyTarballSHA256 校验 + 替换 binary + reexec。如果 fix 未生效，Apply 会卡在 "no SHA256 entry found" 退出，binary 不会被替换 [deps: #5]
