# unify-ears-template-and-hints — Design

## Architecture

3 个独立子任务 + 1 个装配测试，按文件分散：

| 模块 | 职责 | 文件 |
|---|---|---|
| 模板修正 | 在 `requirements.md.tmpl` 的 6 条 AC 示范行加 `[AC-N]` 前缀 | `internal/spec/templates/requirements.md.tmpl` |
| 文档锚点 | 在 `docs/EARS.md` 的 3 个 heading 加 Pandoc `{#slug}` 显式锚点 | `docs/EARS.md` |
| 锚点解析测试 | 新测试 `TestLintHintsResolve` 校验所有 Hint 字符串能在 docs/EARS.md 解析 | `internal/lint/hints_test.go`(新建) |

不引入新依赖（free-kiro 的精简哲学：Go stdlib + 3 个直接依赖保持不变）。
heading slug 提取用 stdlib `regexp` + `strings` + `unicode` 实现，对中文 + 英文 + 数字混合场景足够。

## Components

### 1. 模板修正（AC-1）

`internal/spec/templates/requirements.md.tmpl` line 26-31 现状：

```
- WHEN <TODO:trigger> THE SYSTEM SHALL <TODO:response>.
- WHILE <TODO:state> THE SYSTEM SHALL <TODO:behavior>.
- WHERE <TODO:feature flag / option> THE SYSTEM SHALL <TODO:behavior>.
- UNLESS <TODO:exception> THE SYSTEM SHALL <TODO:default behavior>.
- IF <TODO:condition> THEN <TODO:event> THE SYSTEM SHALL <TODO:response>.
- THE SYSTEM SHALL <TODO:ubiquitous baseline requirement>.
```

改成：

```
[AC-1] WHEN <TODO:trigger> THE SYSTEM SHALL <TODO:response>.
[AC-2] WHILE <TODO:state> THE SYSTEM SHALL <TODO:behavior>.
[AC-3] WHERE <TODO:feature flag / option> THE SYSTEM SHALL <TODO:behavior>.
[AC-4] UNLESS <TODO:exception> THE SYSTEM SHALL <TODO:default behavior>.
[AC-5] IF <TODO:condition> THEN <TODO:event> THE SYSTEM SHALL <TODO:response>.
[AC-6] THE SYSTEM SHALL <TODO:ubiquitous baseline requirement>.
```

**注意**：采用**裸格式** `[AC-N]`（无 `-` 前缀）。原因：`extractEARSLines`
(`internal/lint/quality.go:33`) 用的 `acIDPrefixRe = ^\s*\[AC-\d+\]` 只认裸格式
（带 `-` 前缀会被漏掉，导致 `ears-few-ac` 误报 0 AC）。`CheckACMissingID`
(`quality.go:238`) 同时认两种，所以裸格式双赢。

`improve-ears-quality/requirements.md` 已经用裸格式（line 21 `[AC-1] WHEN ...`），
本 spec 与其保持一致。

### 2. docs/EARS.md 加显式锚点（AC-3）

3 处 heading 加 Pandoc `{#english-slug}` 显式锚点：

| 行号 | heading 文本 | 锚点 |
|---|---|---|
| `docs/EARS.md:7` | `## 五种模板 + 无条件基线` | `{#five-templates}` |
| `docs/EARS.md:105` | `## 模糊词黑名单` | `{#vague-words}` |
| `docs/EARS.md:65` | `## 语义质量门禁（新增 10 条）` | `{#semantic-quality-gates}` |

GitHub 2017+ 已支持 Pandoc `{#id}` 显式锚点语法；GitHub 渲染时该语法被忽略
（不显示），但 `href="#five-templates"` 等链接可直接定位。

### 3. `TestLintHintsResolve` 测试（AC-2 + AC-4）

新建 `internal/lint/hints_test.go`（与现有 `ears_test.go` / `quality_test.go`
同目录、同包），实现：

```go
package lint

import (
    "os"
    "regexp"
    "strings"
    "testing"
)

// docsAnchorHintRe matches `docs/EARS.md#<anchor>` inside Hint: strings.
// Captures the anchor in group 1.
var docsAnchorHintRe = regexp.MustCompile(`docs/EARS\.md#([\w-]+)`)

// hintLineRe matches a Go file line that carries a Hint: field, capturing
// the Hint string (group 1) and the line number (group 2 via index loop).
var hintLineRe = regexp.MustCompile(`(?i)Hint:\s*"([^"]+)"`)

// TestLintHintsResolve — every Hint: string of the form `docs/EARS.md#<a>`
// in `internal/lint/{requirements,bugfix,quality}.go` must point at an
// anchor that actually exists in docs/EARS.md.
//
// Fails with file:line + anchor name + the full set of valid anchors so
// the broken link is obvious from the test output.
func TestLintHintsResolve(t *testing.T) {
    docs, err := os.ReadFile("../../docs/EARS.md")
    if err != nil {
        t.Fatalf("read docs/EARS.md: %v", err)
    }
    valid := extractHeadingSlugs(string(docs))

    files := []string{
        "requirements.go",
        "bugfix.go",
        "quality.go",
    }
    for _, name := range files {
        src, err := os.ReadFile(name)
        if err != nil {
            t.Fatalf("read %s: %v", name, err)
        }
        // Scan for Hint: "..." lines, capture 1-based line number.
        lines := strings.Split(string(src), "\n")
        for i, line := range lines {
            for _, m := range hintLineRe.FindAllStringSubmatch(line, -1) {
                hint := m[1]
                for _, am := range docsAnchorHintRe.FindAllStringSubmatch(hint, -1) {
                    anchor := am[1]
                    if !valid[anchor] {
                        t.Errorf("%s:%d Hint %q anchors %q which is not in docs/EARS.md (valid: %v)",
                            name, i+1, hint, anchor, sortedKeys(valid))
                    }
                }
            }
        }
    }
}

// extractHeadingSlugs returns the set of anchorable IDs in a markdown
// document. Accepts both Pandoc `{#english-slug}` explicit anchors AND
// GitHub-implicit slugs (lowercase, spaces → hyphens, CJK preserved).
func extractHeadingSlugs(md string) map[string]bool {
    var headingRe = regexp.MustCompile(`(?m)^(#{1,6})\s+(.*?)\s*(?:\{#([\w-]+)\})?\s*$`)
    out := map[string]bool{}
    for _, m := range headingRe.FindAllStringSubmatch(md, -1) {
        title := strings.TrimSpace(m[2])
        if m[3] != "" {
            out[m[3]] = true
        }
        out[githubImplicitSlug(title)] = true
    }
    return out
}

// githubImplicitSlug approximates GitHub's auto-slug: lowercase, spaces → '-',
// preserve CJK + ASCII alphanumerics, drop punctuation.
func githubImplicitSlug(title string) string {
    title = strings.ToLower(title)
    var b strings.Builder
    prevHyphen := false
    for _, r := range title {
        switch {
        case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
            b.WriteRune(r)
            prevHyphen = false
        case r == ' ' || r == '-':
            if !prevHyphen {
                b.WriteRune('-')
                prevHyphen = true
            }
        default:
            // CJK + other Unicode letters: preserve.
            if isCJK(r) || unicode.IsLetter(r) || unicode.IsDigit(r) {
                b.WriteRune(r)
                prevHyphen = false
            }
        }
    }
    return strings.Trim(b.String(), "-")
}

func isCJK(r rune) bool {
    return unicode.Is(unicode.Han, r)
}

func sortedKeys(m map[string]bool) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}
```

## Data Model

无新数据结构。改动都在：
- markdown 模板字符串（template）
- markdown 文档（docs/EARS.md）
- 新 Go 测试文件

不动 `.baseline.json` schema、不动 linter Issue 结构。

## Implementation Details

### heading slug 提取规则

GitHub auto-slug（2024+）实际行为：
1. 全部转小写
2. 移除所有标点（包括 `+` `&` `,` `?` `!` 等），但保留 `-` `_`
3. 空格转 `-`
4. 多个 `-` 折叠为一个
5. **CJK 字符保留**（不转 hex，这是新行为；老版本会转 hex）

我们的 `githubImplicitSlug` 实现覆盖 1-5。
唯一不覆盖的 edge case：emoji / 其它特殊字符。在 `docs/EARS.md` 现有 heading 里没有用到。

### 测试覆盖

`TestLintHintsResolve` 必须：

1. **正路径**：现有所有 Hint 字符串都解析（`#five-templates` / `#vague-words` / `#semantic-quality-gates`）
2. **负路径**：手动注入一个 broken anchor（比如 `Hint: "see docs/EARS.md#bogus"`），断言 test fails 且报错信息含 "bogus" + 文件:行号
3. **CJK 路径**：docs/EARS.md 现有中文 heading（如 `## 五种模板 + 无条件基线`）的隐式 slug "五种模板--无条件基线" 也要在 valid 集合里——保证隐式 anchor 也算"valid"

### 端到端验证（AC-4）

```bash
# 1. 起 test-spec 模板
free-kiro spec new test-spec --prompt "test" --quick

# 2. lint 不报 ears-ac-missing-id
free-kiro lint test-spec | grep ears-ac-missing-id
# 期望: 空输出

# 3. lint 引擎所有 Hint 锚点解析通过
go test ./internal/lint/... -run TestLintHintsResolve
# 期望: PASS

# 4. coverage 保持 ≥ 90%
go test -cover ./internal/lint/... ./internal/spec/...
# 期望: PASS，coverage ≥ 90%
```

## Error Handling

- `extractHeadingSlugs` 在 `docs/EARS.md` 缺失时让测试 panic（明显错误，仓库不完整）
- `Hint:` 字符串里的 anchor 不在 valid 集合时，让测试 fail 并打印文件:行号 + anchor + valid keys
- 不修改 linter 本身的错误处理路径（Hint 字段对用户是 advisory）

## Testing Strategy

| 层 | 覆盖 | 文件 |
|---|---|---|
| 单元 | `extractHeadingSlugs` 对 docs/EARS.md 真实数据正确提取 CJK + 显式 + 隐式 slug | `internal/lint/hints_test.go` |
| 单元 | `TestLintHintsResolve` 扫 3 个 lint 文件的 Hint 字段 | 同上 |
| 端到端 | `free-kiro spec new <test-spec> --quick` + `free-kiro lint <test-spec>` 不报 `ears-ac-missing-id` | shell 验证 |

覆盖率门槛：`go test -cover ./internal/lint/...` ≥ 90%（与项目既有门槛一致）。

## Migration / Rollout

**默认上线**：合并到 main 即生效。模板改动让所有 `free-kiro spec new` 调用受益。
docs/EARS.md 锚点改动让 GitHub 链接可点。无 schema / 数据迁移。

**回滚**：3 处改动互不耦合，回滚任何一个文件即可恢复。

**docs/EARS.md 锚点影响范围**：仅 GitHub 渲染（GitHub 已支持 Pandoc `{#id}`）。
本地 markdown 渲染器（VSCode preview / pandoc）也支持 Pandoc 锚点。零兼容性问题。