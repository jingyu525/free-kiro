# upgrade-shasums-filename — Design

## 根因

`internal/upgrade/upgrade.go` 里 `Apply()` 把 `p.Binary`（字符串
`"free-kiro"`）传给 `LookupSHA256`，但 GoReleaser 产出的 SHA256SUMS
只列**外部**资产（tarball 全名 / zip 全名），**不列** tarball 内部
的二进制名。`LookupSHA256` 的 fallback 分支 `name == "free-kiro"`
在 SHA256SUMS 数据上永远是死代码，导致任何 `free-kiro upgrade`
调用都因找不到条目而失败。

## Architecture

改动局限在 `internal/upgrade/upgrade.go` + `internal/upgrade/upgrade_test.go`，
跨包影响为零：

- `internal/skill/release.go` 的 `SHA256SUMSURL` 已经用正确的
  `free-kiro_<version>_SHA256SUMS` 模板，不动。
- `findAssets()` 已经返回正确的 tarball 全名 URL，不动。
- 只改 **一个调用点**（Apply 传错参数）+ **一处死代码**（fallback 分支）。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `upgrade.Apply` | 下载 sums + tarball，校验 sha256，写入目标 | `Apply(ctx, *Plan, force bool) error` |
| `upgrade.LookupSHA256` | 从 SHA256SUMS 文本中查某条目的 sha256 | `LookupSHA256(sums, tarball string) (string, error)` |
| `upgrade.verifyTarballSHA256` | 抽出 Apply 的 basename + lookup 集成路径，可单测 | `verifyTarballSHA256(sumsFile []byte, tarballURL string) (string, error)` |

## Data Model

`Plan` 字段不动（保持公开 API 稳定）。修复通过抽取一个未导出的辅助函数
`verifyTarballSHA256` 完成，让 `Apply` 调用点的关键路径（`path.Base` +
`LookupSHA256` 集成）可以脱离 HTTP `Download` 单独单测。

```go
// Plan 现有结构（不动）
type Plan struct {
    Download string  // full URL of the binary tarball, e.g. "https://.../free-kiro_0.8.0_darwin_arm64.tar.gz"
    Binary   string  // basename of the tarball inside the archive, e.g. "free-kiro"
    Checksum string  // full URL of the SHA256SUMS file
    // ...
}

// 抽出的未导出辅助函数（紧挨 LookupSHA256）
// verifyTarballSHA256 提取 tarballURL 的 basename 并查 SHA256SUMS。
// 让 Apply 的关键校验路径脱离 HTTP 可单测。
func verifyTarballSHA256(sumsFile []byte, tarballURL string) (string, error) {
    return LookupSHA256(string(sumsFile), path.Base(tarballURL))
}

// Apply 内调用点（一行替换）
expected, err := verifyTarballSHA256(sumsFile, p.Download)
```

为什么不直接用 `path.Base(p.Download)` 内联在 Apply 里：
- `Apply` 整体涉及 `Download(...)` HTTP 调用 + `extractBinary` + `reexec`，
  无法在单测中触发；不抽函数则 `path.Base(p.Download)` 这行永远拿不到覆盖
- 抽函数后，`verifyTarballSHA256` 是纯函数，可达 100% 行覆盖，
  等价覆盖 Apply 调用点的全部逻辑

为什么不新增 `Plan.Tarball string` 字段：
- 改动最小（`Plan` 公开 API 不动），不破坏任何已依赖 `Plan` 字段的 caller
- `Download` 已经携带完整 URL 信息，basename 是纯函数变换，不重复存储

## Error Handling

- 错误码保持 `upgrade.apply` 不变，仅错误信息里 "for free-kiro" 自动
  变成 "for free-kiro_0.8.0_darwin_arm64.tar.gz"，更可观测。
- LookupSHA256 删 fallback 后，找不到时仍然返回同一错误码，行为兼容。

## Testing Strategy

- **单元测试**：新增 `TestApplyLookupSHA256UsesTarballNotBinary`，断言
  - 传 `p.Binary` 作 key 找不到条目（旧 bug 触发条件）
  - 传 `p.Tarball` 作 key 能正确返回 hex sha256
- **既有测试**：`internal/upgrade/upgrade_test.go` 里所有 `LookupSHA256`
  相关 case 重跑一遍，确保签名兼容、行为不退化。
- **手动 e2e**（可选，dev 环境）：用 `go build` 产出新 binary，手动
  `cp` 覆盖 `~/.local/bin/free-kiro`，执行 `free-kiro upgrade --check`
  确认 latest 版本能正确报告。

## Migration / Rollout

不需要 expand-contract / feature flag：

- 改动 < 5 行 Go 代码 + 测试
- 行为修正而非行为变更（之前 100% 失败，现在 100% 成功）
- 随下一个 free-kiro binary release 一起发布即可生效
- 老用户升级后即获得修复，没有"半新半旧"中间态
