# upgrade-shasums-filename

修 `free-kiro upgrade` 的 SHA256SUMS 校验 bug：当前实现把 `p.Binary`
（字符串 `"free-kiro"`）当 key 去 SHA256SUMS 找条目，但 SHA256SUMS 实际
只列 tarball 全名（如 `free-kiro_0.8.0_darwin_arm64.tar.gz`），导致
升级永远报 `no SHA256 entry found for free-kiro in SHA256SUMS` 而失败。

边界：只修 `internal/upgrade/` 这一条 lookup 路径，不动
`internal/skill/release.go`（那里已经正确），不改 release 资产命名约定。

## User Stories

- 作为 free-kiro 用户（开发者），我希望执行 `free-kiro upgrade` 时能从
  release SHA256SUMS 自动匹配到当前 GOOS/GOARCH 的 tarball 条目并校验通过，
  以便一行命令把本地 dev 版 binary 升到 latest release 而不用手动下载替换。

## Acceptance Criteria

- [AC-1] WHEN 用户执行 `free-kiro upgrade` 且当前 binary 版本落后于 latest release，
  THE SYSTEM SHALL 从 SHA256SUMS 中按完整 tarball 文件名
  `free-kiro_<version>_<GOOS>_<GOARCH>.tar.gz` 匹配到对应的 sha256 条目。
- [AC-2] WHILE `free-kiro upgrade.Apply` 在校验 tarball，
  THE SYSTEM SHALL 把 `p.Tarball`（而非 `p.Binary`）作为 SHA256SUMS 的查找 key。
- [AC-3] THE SYSTEM SHALL 删除 `LookupSHA256` 内的 `name == "free-kiro"` fallback
  分支，因为 SHA256SUMS 实际不列 tarball 内部二进制条目，该分支是死代码。
- [AC-4] IF 当前 GOOS/GOARCH 在 release assets 中不存在匹配的 tarball 条目，
  THEN THE SYSTEM SHALL 返回 `upgrade.apply` 错误并停止流程，不写入任何文件。
- [AC-5] WHERE 调用方传 `p.Tarball = "free-kiro_0.8.0_darwin_arm64.tar.gz"` 且 SHA256SUMS 内含对应行，THE SYSTEM SHALL 返回该行首字段的 hex sha256 字符串。
- [AC-6] THE SYSTEM SHALL 新增单元测试 `TestApplyLookupSHA256UsesTarballNotBinary`，
  覆盖"传 `p.Binary` 当 key 会失败、传 `p.Tarball` 当 key 能成功"这一回归路径。

## Out of Scope

- 不修改 `internal/skill/release.go` 的 `SHA256SUMSURL`（已正确）。
- 不改 release 资产命名约定（保持 GoReleaser 默认
  `free-kiro_<version>_<os>_<arch>.tar.gz` + `free-kiro_<version>_SHA256SUMS`）。
- 不引入 force 重试、回滚、断点续传等新机制。
- 不动 `internal/upgrade/upgrade_test.go` 中已有且仍然通过的测试。
