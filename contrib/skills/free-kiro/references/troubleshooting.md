# troubleshooting — 常见错误

## 退出码

| Exit | 含义 | 典型场景 |
|---|---|---|
| 0 | OK | 一切正常 |
| 1 | lint ERROR（CI / PreToolUse 拦截信号） | requirements 缺 EARS、tasks 依赖环 |
| 2 | engine error（不可恢复） | .kiro/ 损坏、spec state 文件不合法 |
| 3 | usage error（参数错） | 未知 flag、互斥 flag 同时给 |

## lint 错误

### `no-ears`

requirements.md 没用 EARS 句式。修复：每条 AC 用 `WHEN/WHILE/WHERE/UNLESS/IF … THE SYSTEM SHALL …`。

### `dependency cycle / dangling / self-reference`

tasks.md 依赖写错：
- 只能引用更早的编号
- 不能环（A 依赖 B，B 依赖 A）
- 不能自引用
- 编号必须连续无跳号

### `illegal phase transition`

状态机不可回退。改完 spec 不要 `spec generate` 重生成已经填了真实内容的文档 —— 就地编辑 + `spec sync`。

### `cannot approve: tasks.md must be generated first`

approve 之前必须先 `spec generate --phase tasks`。

## 命令问题

### `free-kiro: command not found`

```bash
# 看是否装到 ~/.local/bin
ls ~/.local/bin/free-kiro

# 缺则重装：
curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/install.sh | bash
```

### `--from-browser` 报 `bsk not found`

```bash
# 装 browser-skill（一次性）
# 路径见 https://github.com/jingyu525/browser-skill
# 验证：
which bsk
bsk --help
```

备选：非 JS 页用 `--from-prd` 即可，不需 bsk。

### `--from-prd` 抓不到内容

- URL 必须 HTTP(S)，不支持 file:// / data: / ftp://
- 1 MB 上限，超大页会截断
- 30 s 超时，慢站点会 fail
- JS 渲染的页面不会执行 JS（用 `--from-browser`）

### `doctor` 报 `workspace missing`

```bash
free-kiro init   # 在项目根目录跑
```

### `skill show` 报 `not installed`

```bash
free-kiro skill install --app all
# 或：
curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/contrib/skills/install.sh | bash
```

## IDE hook 问题

### PreToolUse 不拦截

- 看 `~/.claude/settings.json` 里 free-kiro-managed 的命令是否还在
- 退出码契约：`free-kiro lint || exit 2`（exit 2 才是 PreToolUse 拦截信号；exit 1 是普通 fail）
- 重跑 `free-kiro init --ide claude-code` 恢复默认

### SessionStart 不引导

- `.kiro/.current` 文件不存在或指向不存在的 spec
- 跑 `free-kiro spec list` 看当前有哪些 spec
- 手动设：`echo "<spec-name>" > .kiro/.current`

## 上报 bug

GitHub Issues: <https://github.com/jingyu525/free-kiro/issues>

报告时附：
- `free-kiro --version` 输出
- `free-kiro doctor` 输出
- 复现命令 + 完整错误信息
- `.kiro/` 结构（`ls -R .kiro/`）