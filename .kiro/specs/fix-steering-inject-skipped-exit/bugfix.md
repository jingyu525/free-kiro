# fix-steering-inject-skipped-exit — Bug Fix

## Current Behavior (Defect)

`free-kiro steering inject` 在有 ≥ 1 个 skipped 目标时，进程以退出码 1..5 退出（n = skipped 数，最多 5）。这与 spec `steering-inject-to-ide` AC-3 设计的语义一致，但**实际语义不合理**——skipped 是 warning 不是 error：

- skipped 原因包括"目标文件不存在"（如 `.continue/rules/free-kiro.md` 在没装 Continue 的项目上）——这是常见情况，不是真错
- 调用方（如 `.githooks/pre-commit`）需要用 `stderr ^error:` 前缀检测才能区分 warning 与 error，等于把退出码语义复杂化
- CI 调用 `make drift-check` 时，drift-check 用 `git diff --exit-code` 检测，inject exit code 1 不会让它失败——但混用两套退出码机制让 spec AC-3 与现实 drift-check 设计互相对抗

更糟的是：用户跑 `free-kiro steering inject` 后，shell `$?` 显示非零，会误以为 inject 失败；但 stderr 实际只说 "warning: file ...: no such file or directory"，文件没写成功。

## Expected Behavior (Correct)

- WHEN 用户执行 `free-kiro steering inject` 且 ≥ 1 个目标被 skipped THE SYSTEM SHALL 在 stderr 打印 1 行 per-skipped-file 的 `warning: file <relpath>: <reason>` 信息并继续处理其他目标，最终进程以退出码 0 结束。
- WHEN `free-kiro steering inject` 跳过 ≥ 1 个目标且全部 other 目标成功写入 THE SYSTEM SHALL 让调用方（shell / pre-commit hook / CI）通过 `$? == 0` 判断"全部目标按可达性处理完毕"，无需检查 stderr。
- WHEN `free-kiro steering inject` 跳过 ≥ 1 个目标但部分 written 失败 THE SYSTEM SHALL 让进程以退出码 1 结束（仍属于"完全失败"语义，区别于 skipped）。
- THE SYSTEM SHALL 保留 DocCount==0 → exit 3 与 glob 非法 → exit 4 两个语义入口不变。

## Unchanged Behavior (Regression Prevention)

- WHEN `free-kiro steering inject --dry-run` THE SYSTEM SHALL CONTINUE TO 把将注入的内容打印到 stdout，且不修改任何文件，进程以退出码 0 结束。
- WHEN 全部 5 个目标文件成功写入 THE SYSTEM SHALL CONTINUE TO 让进程以退出码 0 结束。
- WHEN `.kiro/steering/` 不存在或目录内没有 `mode: always` 文档 THE SYSTEM SHALL CONTINUE TO 报错退出，stderr 输出 1 行错误消息（含路径），进程以退出码 3 结束。
- WHEN `--only` 的 glob 语法非法 THE SYSTEM SHALL CONTINUE TO 报错退出，stderr 输出 1 行错误消息（含 pattern），进程以退出码 4 结束。
- THE SYSTEM SHALL CONTINUE TO 让 pre-commit hook（`.githooks/pre-commit`）通过 stderr `^error:` 前缀区分 warning 与真错——这是冗余防御，不依赖 exit code 语义；移除本 bugfix 不应破坏 pre-commit 工作。

## Root Cause

`internal/cli/steering_inject.go` 的 `runInject` 函数在 line 98–103 把 `len(res.Skipped)` 当作 exit code：

```go
if n := len(res.Skipped); n > 0 {
    if n > injectExitMaxSkipped {
        n = injectExitMaxSkipped
    }
    os.Exit(n)
}
```

该逻辑源自 spec `steering-inject-to-ide` AC-3 设计意图（让调用方能看出"跳了几个"），但与：
- pre-commit hook 工作流（要求 inject exit 0 才能 commit 继续）
- CI drift-check 设计（drift 检测靠 git diff，不是 inject exit code）

均冲突。

修复方法：删除该 `os.Exit(n)` 块；skipped 仅在 stderr warning；退出码 0 表示"全部目标处理完毕（含 skipped）"。
