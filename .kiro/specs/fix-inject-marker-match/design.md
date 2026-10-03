# fix-inject-marker-match — Design

<!--
决策记录（30 天后回看"为什么这么做"）：
  - 用严格行匹配（TrimSpace == marker）而非正则：① 与 Kiro IDE marker 格式
    一致（marker 行只有 HTML 注释，无 prose），② 比正则易 review，③
    不引入新的依赖或正则风险
  - 不动 marker 字符串本身：避免与既有 docs/HOOKS.md + .githooks/
    pre-commit + init template 文档冲突
  - 加 TestInject_IgnoresMarkerInProse 而非广覆盖 fuzz：单 test 用一个
    真实 init 输出（CLAUDE.md / AGENTS.md）的 prose 段就够暴露 bug
-->

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│ 注入路径（before）                                    │
│   injectToFile(file):                                      │
│     lines = file.Split("\n")                              │
│     for line in lines:                                    │
│       if strings.Contains(line, markerStart): ← 宽松命中  │
│         startIdx = i                                       │
│       if startIdx != -1 and Contains(line, markerEnd):   │
│         endIdx = i                                         │
│     ↑ 解释段里 \`<!-- free-kiro-managed:start -->\`         │
│       也含 marker 子串，被误命中 → 写到错误位置           │
└──────────────────────────────────────────────────────────────┘
                              ↓
┌──────────────────────────────────────────────────────────────┐
│ 注入路径（after）                                     │
│   injectToFile(file):                                      │
│     lines = file.Split("\n")                              │
│     for line in lines:                                    │
│       if startIdx == -1 and                                │
│          strings.TrimSpace(line) == markerStart: ← 严格  │
│         startIdx = i                                       │
│       if startIdx != -1 and endIdx == -1 and              │
│          strings.TrimSpace(line) == markerEnd: ← 严格    │
│         endIdx = i                                         │
│     ↑ 解释段含 marker 字面量但 TrimSpace 后 ≠ marker       │
│       常量字符串，被正确跳过                               │
└──────────────────────────────────────────────────────────────┘
```

## Components

### 1. `internal/steering/inject.go` — injectToFile marker 行识别逻辑

修改 line 196-199 两处 `strings.Contains` 调用为 `strings.TrimSpace(...) == ...`：

| 旧（line 196） | 新 |
|---|---|
| `strings.Contains(line, InjectMarkerStart)` | `strings.TrimSpace(line) == InjectMarkerStart` |

| 旧（line 198） | 新 |
|---|---|
| `strings.Contains(line, InjectMarkerEnd)` | `strings.TrimSpace(line) == InjectMarkerEnd` |

函数签名 + 行为对外契约保持不变（仍返回 `(bool, *InjectSkip)`，missing-marker
仍走 `InjectSkip{Reason: "missing inject marker, run free-kiro init first"}`）。

### 2. `internal/steering/inject_test.go` — 新增反向测试

`TestInject_IgnoresMarkerInProse` 测试用例：

1. 在 t.TempDir() 建 .kiro/specs/foo 与 .kiro/steering/{product,structure,tech}.md
   三个 always 文档（用 fixture 或 inline string）
3. 构造一个目标文件 CLAUDE.md-like：
   - line 1-N：手写指令
   - line 30：解释段含 marker 字符串字面量
       ```
       本文件末尾由 `<!-- free-kiro-managed:start -->` /
       `<!-- free-kiro-managed:end -->` marker 包裹的 markdown 块
       ```
   - line N+1：编码规范段（与 init template line 30+ 同结构）
   - line M：独立 marker start `<!-- free-kiro-managed:start -->`
   - line M+1：空 marker block
   - line M+2：独立 marker end `<!-- free-kiro-managed:end -->`
5. 调 inject 注入 product/structure/tech 内容
6. 断言：
   - 解释段 line 30-31 字面量**未变**：`strings.Contains(content,
     "本文件末尾由 \`<!-- free-kiro-managed:start -->\`")` 仍为 true
   - 真 marker block (line M+1) 内含 `## product.md` + `## structure.md`
     + `## tech.md` 三个二级标题
   - start / end marker 行**保留原样**（不在循环里被消耗）

### 3. 文档注释更新

`internal/steering/inject.go:181-184` 注释从

```
// The marker lines are located by line substring match — they may
// appear anywhere on a line.
```

改为

```
// The marker lines are located by EXACT line match (after
// strings.TrimSpace) — they must occupy a whole line. Prose that
// quotes the marker literals (e.g. the init template's
// "本文件末尾由 ... marker 包裹的 markdown 块" explanation paragraph)
// is intentionally NOT matched.
```

## Data Model

无新增数据结构。`InjectMarkerStart` / `InjectMarkerEnd` 常量值保持不变。

## Testing Strategy

### 单元测试

```bash
# 新增 + 旧测试一起跑
go test -race -count=1 -run 'Inject' ./internal/steering/...

# 期望：
# - TestStore_InjectAll_WritesBlockIntoMarkerRegion PASS（已有）
# - TestStore_InjectAll_SkipsMissingMarker PASS（已有）
# - TestStore_InjectAll_OnlyGlob PASS（已有）
# - TestStore_InjectAll_NoAlwaysDocs PASS（已有）
# - TestInject_IgnoresMarkerInProse PASS（新增，AC-4 / AC-6）
```

### 端到端（仓库根）

```bash
go install ./cmd/free-kiro
free-kiro init --ide claude-code --overwrite-instructions
free-kiro steering inject

# 断言 marker 仍在文件末尾（不是解释段）
grep -c "free-kiro-managed:start" CLAUDE.md   # ≥ 2（解释段 1 + 真 marker 1）
grep -n "free-kiro-managed:start" CLAUDE.md
# line 33 / 34 是解释段引用（保持不变）
# line 85 / 88 是真 marker（被注入内容填好）

# 断言 inject 没污染解释段
sed -n '30,40p' CLAUDE.md   # 含 product/structure 实际内容的就是错误
```

### 反向验证（确认 bug 修了）

```bash
# before fix:
# free-kiro steering inject 会把 product/structure/tech 写到 line 33-34
# 之间（解释段的 marker 字符串字面量），破坏文件结构
# → grep 显示 line 33-34 之间有 ## product.md
# → grep -n 显示 marker 在解释段位置（line 33-34）

# after fix:
# → grep -n 显示 marker 仍在 line 85+（init 写入位置）
# → line 33-34 解释段保持 verbatim
```

## Compatibility / Rollout

### 兼容性边界
- marker 字符串本身**不变**（`<!-- free-kiro-managed:start -->` / `<!-- free-kiro-managed:end -->`）
- inject 命令 CLI flag 不变（`--dry-run` / `--only`）
- inject 输出格式不变（仍以 InjectBlockHeader 开头 + 按字母序 always 文档）
- 不动 init 模板措辞

### 风险与回滚

- **风险 A**：未来有人写 ` <!-- free-kiro-managed:start -->` （前面带空格）。
  现在 `TrimSpace(line) == marker` 会匹配；但 `<!-- free-kiro-managed:start
  -->` 后面跟 prose 不会匹配（中间 prose 不等于 marker 字符串）。**对策**：
  `TestInject_IgnoresMarkerInProse` 显式构造 line 30 含 marker 字符串 +
  真 marker 区块两种情况并存，确保宽松匹配不会回归。

- **风险 B**：strict match 会让现有"非 init 写"的 IDE 指令文件（用户自己
  手动写的 `<!-- free-kiro-managed:start -->` block）也匹配不上。**对策**：
  `free-kiro init` + `free-kiro steering inject` 是官方推荐路径；manual
  marker block 用户必须用 exact-line 写法（这也是项目合规要求）。

- **风险 C**：CI / pre-commit 跑 inject 后会修改 CLAUDE.md 等文件，可能
  在 PR 流程引入"应该自动 commit 但没 commit"的修改。**对策**：本 spec
  不动 pre-commit 脚本；后续 free-kiro IDE 流程应该自动 `git add` marker
  修改后的文件（已在 .githooks/pre-commit line 34 实现）。

### 落地步骤

```
1. internal/steering/inject.go 修改两处 strings.Contains 为 TrimSpace ==
2. internal/steering/inject.go 更新注释（line 181-184）
3. internal/steering/inject_test.go 新增 TestInject_IgnoresMarkerInProse
4. go test -race -count=1 ./internal/steering/... 0 FAIL
5. go install ./cmd/free-kiro（重 build binary）
6. free-kiro init --ide claude-code --overwrite-instructions（重写 4 文件）
7. free-kiro steering inject（端到端验证）
8. 验证 4 文件 marker block 在文件末尾、解释段保持不变
9. git commit + push + free-kiro spec complete fix-inject-marker-match
```

## 引用

- 上游 bug 起源：`internal/steering/inject.go:196` / `inject.go:198`
- 触发 prose：`internal/ide/templates/instructions_zh.md:32-33` /
  `instructions_en.md:34-35` 解释段
- pre-commit 静默失败：`.githooks/pre-commit:26` `free-kiro steering inject`
- workaround commit `3fd9678 fix(ide): 同步 CLAUDE.md / AGENTS.md 等 marker 区域`
- spec trigger：commit `3fd9678` commit message 末尾建议开 fix-inject-marker-match spec