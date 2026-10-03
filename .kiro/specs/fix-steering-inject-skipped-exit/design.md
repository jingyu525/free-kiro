# fix-steering-inject-skipped-exit — Design

## Architecture

单点修复：删除 `internal/cli/steering_inject.go` 里 `runInject` 函数末尾的
"exit code = skipped count" 分支。skipped 仍走 stderr warning 输出（行为不
变），但进程退出码固定为 0。

不修改：
- `internal/steering/inject.go` —— `Store.InjectAll` 已返回 `Skipped` 字段
- `.githooks/pre-commit` —— 已经通过 `stderr ^error:` 区分 warning/error，
  本 bugfix 让其逻辑更简单（不必看 inject exit code）

## Components

| Component | Responsibility | Change |
|---|---|---|
| `internal/cli/steering_inject.go` | 调 `Store.InjectAll` + 决定 exit code | 删除 `os.Exit(n)` 分支；删除未用常量 `injectExitMaxSkipped` |
| `internal/cli/steering_inject_test.go` | RunE 行为测试（如果存在） | 补 1 个 case：5 文件 4 个 marker 缺失时 exit 0 |
| `internal/steering/inject_test.go` | InjectAll 返回值测试 | 已有 `Skipped` 字段断言；不需要改 |
| `.kiro/specs/steering-inject-to-ide/requirements.md` AC-3 | "最终进程退出码等于被跳过的文件数（最多 5）" | 改为"skipped 不影响退出码，stderr warning 即可" |

## 修改前后对比

**Before（删除分支前）**：

```go
for _, sk := range res.Skipped {
    fmt.Fprintf(cmd.ErrOrStderr(),
        "warning: file %s: %s\n", sk.Path, sk.Reason)
}
if n := len(res.Skipped); n > 0 {
    if n > injectExitMaxSkipped {
        n = injectExitMaxSkipped
    }
    os.Exit(n)   // ← 删
}
```

**After**：

```go
for _, sk := range res.Skipped {
    fmt.Fprintf(cmd.ErrOrStderr(),
        "warning: file %s: %s\n", sk.Path, sk.Reason)
}
// skipped 仅 stderr warning，不影响 exit code。
// 真正的失败（DocCount=0 / glob 非法）已在函数开头 os.Exit。
```

## Spec sync 计划

1. 改完 `internal/cli/steering_inject.go` + 补单测
2. 改 `.kiro/specs/steering-inject-to-ide/requirements.md` AC-3 文字
   与新行为对齐
3. 跑 `free-kiro spec sync steering-inject-to-ide` 捕获新 baseline
4. 跑 `free-kiro spec complete fix-steering-inject-skipped-exit` 标记完成

注：`spec sync` 不需要 `spec approve`（已在 done 状态）；sync 只重 baseline
不重置 phase。

## 关键决策记录

- **为什么不改 spec AC-3 为"0 表示 skipped 非零表示真错"**：这就是本 bugfix
  的目标，让 spec 与现实对齐
- **为什么 pre-commit hook 仍保留 `stderr ^error:` 检测**：冗余防御，避免
  inject 引入新的 stderr 模式时 hook 误判；不依赖于此防御，但保留无害
- **为什么不动 `injectExitMaxSkipped` 常量以外的常量**：`injectExitOK` /
  `injectExitNoDocs` / `injectExitBadGlob` 都是真错的语义入口，不动
- **为什么不删除 `InjectResult.Skipped` 字段**：drift-check 设计可能要
  后续用到（统计 skipped 数、warning 严重性分级）；暂留待未来 spec 决定

## 验证策略

1. **单测**：在 `internal/cli/steering_inject_test.go` 加 case 跑
   `steeringInjectCmd()` + 5 目标 4 缺 marker，断言 exit 0
2. **手工**：在仓库跑 `free-kiro steering inject`，`echo $?` 应为 0
3. **pre-commit 兼容**：故意让 .continue/rules/... 缺失 → 跑
   `make install-hooks` + 改 steering + commit，commit 应成功
4. **CI drift-check 兼容**：`make drift-check` 干净状态仍 exit 0
