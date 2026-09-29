# ops — 运维命令（watch / serve / report / doctor）

## `free-kiro watch`

`.kiro/` 文件变化时跑命令。fsnotify 后端，默认 200ms 防抖。

```bash
free-kiro watch                                       # 默认：free-kiro lint
free-kiro watch --preset lint                         # 同上
free-kiro watch --preset status                       # 变化后打印 spec status
free-kiro watch --preset reactive                     # lint + status
free-kiro watch --preset full                         # lint + status + report
free-kiro watch --command "go test ./..."             # 自定义
free-kiro watch --command "go build" --command "go test"
free-kiro watch --root /path/to/project
free-kiro watch --debounce 500ms
free-kiro watch --verbose
```

适用：编辑器里编辑 `.kiro/specs/foo/requirements.md` 时自动 lint，编辑反馈即时。

## `free-kiro serve`

本地 web 看板（默认 `127.0.0.1:7373`），SSE 实时推送。

```bash
free-kiro serve                       # 默认 127.0.0.1:7373
free-kiro serve --bind 0.0.0.0        # 局域网可访问
free-kiro serve --port 8080
free-kiro serve --open                # 启动后调系统浏览器打开
```

适用：跟同事演示 / 自己盯多 spec 进度 / 写报告时截图。

## `free-kiro doctor`

单次自检：

- free-kiro 版本
- `.kiro/` 工作区存在 + 结构
- 已装 IDE hook 数量
- bsk 可用性（仅 `--verbose` 模式）
- 当前 spec 状态（若 `.kiro/.current` 存在）

```bash
free-kiro doctor
free-kiro doctor --strict             # warn 也算 fail（exit 非零）
```

## `free-kiro report`

把当前 workspace 状态写成 `.kiro/REPORT.md`：

- 所有 specs 列表 + 阶段 + 漂移
- 所有 hooks 列表
- 当前 steering 清单
- doctor 结果摘要

```bash
free-kiro report                      # 写到 .kiro/REPORT.md
```

适用：项目交接、CI artifact、周报素材。

## 组合用法

```bash
# 在 CI 里：写完 spec 后自动 lint + 出报告
free-kiro lint && free-kiro report

# 本地开发：watch + serve 一起
free-kiro watch --preset full &
free-kiro serve --port 7373 &
```

## 与 IDE hook 的分工

- `watch` 是"持续监听 + 主动跑命令"
- IDE hook（`PreToolUse` 等）是"工具调用前后拦截"
- 两者互补：hook 在每次 Edit/Write 之前即时拦截；watch 是后台长跑

推荐 hook 模式 + watch 模式同时开（hook 拦截 + watch 兜底）。