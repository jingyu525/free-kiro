# 与 kiro-clone / 官方 Kiro 的兼容性

> free-kiro 设计目标：**完全兼容 kiro-clone 的 spec 工作流**，**完全兼容 Kiro 官方
> v1 hook 信封**。你可以无痛切换工具，spec 文档 / 配置 / hook 文件在三者之间
> 直接迁移。

## 兼容性矩阵

| 维度 | kiro-clone (Python) | 官方 Kiro IDE | free-kiro (Go) | 兼容性 |
|---|---|---|---|---|
| `.kiro/` 目录结构 | ✅ | ✅ | ✅ | 100% |
| `specs/<name>/{requirements\|design\|tasks\|bugfix}.md` | ✅ | ✅ | ✅ | 100% |
| `steering/*.md` | ✅ | ✅ | ✅ | 100% |
| `hooks/*.json` | ✅（自家 flat） | ✅（v1 envelope） | ✅（两种都接受） | 100% |
| EARS 5 种句式 + 无条件基线 | ✅ | ✅ | ✅ | 100% |
| Bugfix 三段式 | ✅ | ✅ | ✅ | 100% |
| Spec 状态机 | ✅ | ✅ | ✅ | 100% |
| Wave 并行调度（依赖图拓扑） | ✅ | ✅ | ✅ | 100% |
| Drift baseline 锁定 | ✅ | ✅ | ✅ | 100% |
| Analyze advisory | ✅ | ✅ | ✅ | 100% |
| 退出码契约（0/1/2/3） | ✅ | n/a | ✅ | n/a |
| IDE PreToolUse hook 拦截 | n/a | ✅ | ✅（同样的 `\|\| exit 2`） | n/a |

## 不兼容的部分

### free-kiro 不做，kiro-clone 也不做

- ❌ AI 模型调用（spec 文档是骨架，由 agent 写内容）
- ❌ IDE 集成 / 编辑器插件（hook 系统是 IDE 的职责）
- ❌ TUI 交互（CLI only）

### free-kiro 比 kiro-clone 改进的地方

| 维度 | kiro-clone | free-kiro |
|---|---|---|
| 启动时间 | ~150ms（Python 解释器） | ~3ms（Go 编译产物） |
| 分发 | `pip install -e .` 或 `PYTHONPATH=… python -m kiro.cli` | 单二进制 `curl\|bash` 一行安装 |
| 跨平台 | Python runtime 依赖 | darwin/linux/windows × amd64/arm64 原生二进制 |
| Hook 信封 | 只支持自家 flat shape | 同时接受自家 + Kiro v1 envelope |
| 退出码 | KiroError=2，lint=1 | 同样 + 区分 UsageError=3 |

### free-kiro 暂时比官方 Kiro 少的功能

| 功能 | 状态 |
|---|---|
| IDE 内嵌 UI | 不做（CLI only） |
| 自动驱动 hook（agent 自动响应事件） | 不做（保持被动） |
| 模型驱动的 spec 生成（自动写 requirements.md 内容） | 不做（保持 offline / template only） |
| Homebrew tap 自动发布 | v0.1.0 暂时没配；需要单独的 homebrew-tap repo |

## spec 文档格式：完全兼容

kiro-clone 生成的 spec 目录，free-kiro 可以直接读：

```bash
# 用 kiro-clone 生成
kiro init
kiro spec new demo --prompt "..."
kiro spec generate demo --phase all

# 用 free-kiro 接管
cd same-project
free-kiro lint demo                    # 同样的 lint 报告
free-kiro spec approve demo            # 同样的 approve 流程
free-kiro spec status demo             # JSON 输出格式相同
```

反之亦然——free-kiro 生成的 spec 用 kiro-clone 也能管理。

## hook 文件迁移

### free-kiro → 官方 Kiro

free-kiro 写出的 hook 自动符合 Kiro v1 信封。直接复制 `.kiro/hooks/<id>.json` 到
官方 Kiro IDE 的 hooks 目录即可加载。

或者把整个 `.kiro/hooks/` 目录拷过去。

### 官方 Kiro → free-kiro

官方 Kiro 的 hook 文件通常长这样：

```json
{
  "version": "v1",
  "hooks": [
    {
      "name": "kiro-lint",
      "trigger": "PostFileSave",
      "matcher": "\\.tsx$",
      "action": {
        "type": "command",
        "command": "npx eslint --fix"
      }
    }
  ]
}
```

`free-kiro hook list` 会自动归一化读取，显示：

```
ID                   EVENT            FILTER        TYPE    TO    ENABLED
kiro-lint            PostFileSave     \.tsx$ (re)   shell   -     true
```

注意 Kiro 风格的 `matcher`（regex）和 `type: "command"` 都被正确归一化。

### kiro-clone flat → free-kiro

kiro-clone 的 flat shape：

```json
{
  "id": "lint",
  "event": "file.save",
  "glob": "*.go",
  "action_type": "shell",
  "action": "gofmt -w"
}
```

free-kiro 也接受。两种 schema 共存。

## steering 文档：完全兼容

kiro-clone 和 free-kiro 都接受 `mode:` 键。

官方 Kiro 用 `inclusion:` 键——free-kiro 也接受，且 `inclusion:` 优先于 `mode:`。

两种 frontmatter 都可以直接读取。

## 退出码契约（与 kiro-clone 一致）

| 退出码 | kiro-clone | free-kiro |
|---|---|---|
| 0 | 成功 | 成功 |
| 1 | lint ERROR | lint ERROR |
| 2 | KiroError | KiroError |
| 3 | （未区分） | UsageError（额外细化） |

IDE hook 配置可以无缝迁移：

```bash
# 这个命令在 kiro-clone 和 free-kiro 都产生 exit 1/2：
kiro lint      # → exit 1（ERROR）或 exit 2（workspace 缺失）
free-kiro lint  # → exit 1（ERROR）或 exit 2（workspace 缺失）
```

## 升级 / 迁移建议

### 从 kiro-clone 迁到 free-kiro

1. 安装 free-kiro：`curl -fsSL … | bash`
2. 在项目目录跑 `free-kiro init`（如果之前已经有 `.kiro/`，是幂等的）
3. 现有 spec 文档、steering、hook 全部不动——free-kiro 直接接管
4. CI hook 把 `kiro` 命令替换成 `free-kiro` 即可（退出码契约相同）

### 与官方 Kiro IDE 共存

free-kiro 不与官方 Kiro IDE 冲突——它们只是共享 hook JSON 格式：

```bash
# 项目里有 .kiro/，可以被两者同时打开
# spec 文档两边都能读
# hook 文件格式相同，复制即可
```

## 已知差异（小细节）

| 细节 | kiro-clone | free-kiro | 影响 |
|---|---|---|---|
| 默认 generator | `"template"` | `"template"` | 无差异 |
| `prompt` 元数据存储 | ✅ | ✅ | 无差异 |
| `created_at` 时间格式 | RFC3339 UTC | RFC3339 UTC | 无差异 |
| `.meta.json` 缩进 | 2 空格 | 2 空格 | 无差异 |
| EARS 大小写 | 大小写不敏感 | 大小写不敏感 | 无差异 |
| `tasks.md` 行格式 | `- [ ] #N Title [deps: #N1,#N2]` | 同 | 无差异 |

## 总结

free-kiro 是 kiro-clone 的 1:1 Go 重写 + 跨平台二进制化增强。spec 工作流的所有
概念（状态机 / EARS / 漂移 / wave / steering / hook）在三者之间完全可移植。

唯一的"改进"是性能 + 分发：单二进制、毫秒启动、`curl|bash` 安装。

---

参考：[WORKFLOW.md](WORKFLOW.md)（状态机）/ [EARS.md](EARS.md)（句式）/ [STEERING.md](STEERING.md)（约定）/ [HOOKS.md](HOOKS.md)（信封）