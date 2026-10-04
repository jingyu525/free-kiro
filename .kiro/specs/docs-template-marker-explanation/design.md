# docs-template-marker-explanation — Design

本次 spec 是**纯文档改动 + 1 个回归测试**，不引入新代码路径、不修改运行时
逻辑。技术决策都在"模板文本怎么写"和"测试断言怎么写"两层。

## Architecture

- **模板层**：`internal/ide/templates/{instructions,agents}_{zh,en}.md` 共
  4 个文件，把"项目上下文（自动注入的 steering）"段改写为同时讲清楚顶部 /
  底部两套 marker 的注入源 + 用途 + 修改路径。
- **测试层**：`internal/ide/ide_test.go` 新增
  `TestInstructionTemplate_ExplainsBothMarkers`，从 embed.FS 读 4 个模板，
  逐个断言两套 marker 字符串 + 注入源关键词（prependMarker / steering
  inject）都出现。
- **同步层**：人工跑 `free-kiro init --ide auto --overwrite-instructions`，
  让 4 个模板的最新内容传播到仓库内 `CLAUDE.md` / `AGENTS.md` /
  `.cursorrules` / `.continuerules` / `.kiro/AGENTS.md` 5 个实际文件。
- **用户文档层**：`docs/STEERING.md` §"自动注入到 IDE 指令文件"段同步说明
  两套 marker 差异，让 docs 路径也能查到。

不改 `internal/ide/ide.go` 的任何函数（prependMarker / IsFreeKiroInstruction
行为冻结）。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `internal/ide/templates/instructions_zh.md` | Claude Code / Cursor / Continue 模板 | 改写 L30-42 "项目上下文"段 |
| `internal/ide/templates/instructions_en.md` | 同上，英文版 | 改写对应段 |
| `internal/ide/templates/agents_zh.md` | OpenCode / CodeBuddy 模板 | 改写对应段 |
| `internal/ide/templates/agents_en.md` | 同上，英文版 | 改写对应段 |
| `internal/ide/ide_test.go` 新增测试 | 防回归 | `TestInstructionTemplate_ExplainsBothMarkers(t *testing.T)` |
| `docs/STEERING.md` | 用户文档 | §"自动注入到 IDE 指令文件"段更新 |

## Data Model

无（纯文档 + 测试，无数据结构变更）。

## Error Handling

无（纯文档；测试断言失败即回归，由 `go test` 标准输出报错）。

## Testing Strategy

### 单元测试

`TestInstructionTemplate_ExplainsBothMarkers` 逻辑：

1. 列出 `embed.FS` 中 `instructions_zh.md` / `instructions_en.md` /
   `agents_zh.md` / `agents_en.md` 共 4 个模板。
2. 对每个模板断言：
   - `strings.Contains(body, "# free-kiro-managed:")` 为 true（顶部 marker 字符串）
   - `strings.Contains(body, "<!-- free-kiro-managed:start -->")` 为 true
   - `strings.Contains(body, "<!-- free-kiro-managed:end -->")` 为 true
   - `strings.Contains(body, "prependMarker")` 为 true（顶部 marker 注入源，英文模板）
     或 中文模板的等效表述（`init` 注入 / `运行时注入`）
   - `strings.Contains(body, "steering inject")` 为 true（底部 marker 注入源）
3. 任意一个模板断言失败 → `t.Fatalf("template %s missing markers: ...")`

### 集成验证

- `go test -race ./internal/ide/...` 全 PASS
- `free-kiro lint` 对当前活跃 spec 全绿（虽然 lint 主要管 spec 文档，
  但要确认本次模板改动没意外触发某条 ears / structural rule）
- `free-kiro init --ide auto --overwrite-instructions` 同步后，肉眼
  diff `CLAUDE.md` / `AGENTS.md` / `.cursorrules` / `.continuerules` /
  `.kiro/AGENTS.md`，确认顶部 `# free-kiro-managed:` 行 + 底部 marker
  块都还在

## Migration / Rollout

无 flag、无分阶段。即时生效：

- 模板改动在 IDE 指令文件下次 `init` / `inject` 时自动传播
- 测试改动随仓库 CI 立即生效
- `docs/STEERING.md` 改动随 docs build 立即生效

## Implementation Notes

### 模板"项目上下文"段改写示意（中文版）

```markdown
## 项目上下文与文件标记

本项目 IDE 指令文件包含**两套** free-kiro 自动维护的标记，区分清楚以免误判：

1. **首行 `# free-kiro-managed:`** —— 由 `free-kiro init` 运行时通过
   `internal/ide/ide.go` 的 `prependMarker` 函数（行 517-545）注入到
   YAML frontmatter 闭合 `---` 之后。用途：让 `IsFreeKiroInstruction`
   识别"这文件是 free-kiro 生成的"，防后续 `init --overwrite-instructions`
   覆盖用户在同路径手写的内容。

2. **文件末尾 `<!-- free-kiro-managed:start -->` ... `<!-- free-kiro-managed:end -->`**
   包裹的 markdown 块 —— 由 `free-kiro steering inject` 在每次 `init` /
   `inject` 运行时自动生成。内容来自 `.kiro/steering/*.md` 中 `mode: always`
   的文档（当前为 `product.md` / `structure.md` / `tech.md`）。

修改路径：
- 想改顶部 marker 语义或位置 → 改 `prependMarker` 函数
- 想改底部 marker 注入内容 → 改 `.kiro/steering/<name>.md` 后跑
  `free-kiro steering inject`
- 想改模板里本段说明本身 → 改 `internal/ide/templates/*.md` 后跑
  `free-kiro init --ide auto --overwrite-instructions`

详见 `docs/STEERING.md` §"自动注入到 IDE 指令文件"。
```

英文版对应翻译，marker 字符串原文保留，注入源用 `prependMarker` / `steering inject` 原词。

### 测试函数骨架

```go
func TestInstructionTemplate_ExplainsBothMarkers(t *testing.T) {
    templates := []string{
        "templates/instructions_zh.md",
        "templates/instructions_en.md",
        "templates/agents_zh.md",
        "templates/agents_en.md",
    }
    for _, p := range templates {
        body, err := templatesFS.ReadFile(p)
        if err != nil { t.Fatalf("read %s: %v", p, err) }
        requireBothMarkers(t, p, string(body))
    }
}

func requireBothMarkers(t *testing.T, path, body string) {
    t.Helper()
    mustContain := []string{
        "# free-kiro-managed:",
        "<!-- free-kiro-managed:start -->",
        "<!-- free-kiro-managed:end -->",
        "steering inject",
    }
    for _, s := range mustContain {
        if !strings.Contains(body, s) {
            t.Errorf("%s missing %q", path, s)
        }
    }
    // prependMarker 或等效中文术语（init 注入 / 运行时注入）
    hasPrependMarker := strings.Contains(body, "prependMarker") ||
        strings.Contains(body, "init 注入") ||
        strings.Contains(body, "运行时注入")
    if !hasPrependMarker {
        t.Errorf("%s missing prependMarker / 等效术语", path)
    }
}
```