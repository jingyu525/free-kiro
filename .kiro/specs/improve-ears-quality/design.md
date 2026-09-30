# improve-ears-quality — Design

## Architecture

新增 3 个边界清晰的模块,沿用现有 `internal/lint/` 目录约定:

| 模块 | 职责 | 文件 |
|---|---|---|
| 模板正则 | 6 种 EARS 模板的独立 regex,供 AC checker 与 spec analyze 共享 | `internal/lint/ears.go`(扩) |
| 质量规则 | 10 条 AC 的 checker 实现,每条独立函数 + 单测 | 新建 `internal/lint/quality.go` |
| Baseline | `.baseline.json` schema + 加载 + strict 模式 | 新建 `internal/lint/baseline.go` |

`internal/lint/linter.go` 的 `Spec()` 入口只需新增一处 hook:
1. 加载 `.baseline.json`(若存在)→ 得到 `IgnoredCodes map[string]bool`
2. 调用 `Requirements(text)` 时把基线过滤后的 `Issue` 列表返回
3. CLI 层加 `--strict-baseline` flag → 临时禁用基线

`models.MinAcceptanceCriteria` 常量放在 `internal/spec/models/models.go`
(已有 `FirstPlanningDoc` 等常量,与 spec 类型定义同源)。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `ears.go` 扩展 | 提供 5 个独立模板 regex | `WHENRe`, `WHILERe`, `WHERERe`, `UNLESSRe`, `IFTHENRe` |
| `quality.go` | 10 条 AC 的纯函数 checker | `CheckEtcList`, `CheckSingleSHALLLine`, `CheckFewAC`, `CheckTemplateDiversity`, `CheckACMissingID`, `CheckMeasurableResponse`, `CheckTriggerObservable`, `CheckKeywordMisuse`, `CheckPassiveResponse`, `CheckSingleSHALLLine` |
| `baseline.go` | 加载 / 序列化 `.baseline.json` | `LoadBaseline(specDir) (Baseline, error)`, `Baseline.ShouldIgnore(code string) bool` |
| `requirements.go` 扩展 | 调 quality 9 个函数 + User Stories 升 ERROR | 在现有 `Requirements()` 末尾追加调用 |
| `linter.go` 扩展 | 装配 baseline 过滤 | `Spec(specDir)`, 新增 `SpecWithBaseline(specDir, ignore map[string]bool)` |
| `cli/spec_lint.go` | 加 `--strict-baseline` flag | `lintCmd().Flags().BoolVar(&strictBaseline, ...)` |
| `models.MinAcceptanceCriteria` | AC 最小门槛常量 | `const MinAcceptanceCriteria = 3` |

## Data Model

### `.baseline.json` schema

```json
{
  "schema_version": 1,
  "spec_name": "improve-ears-quality",
  "ignored_issues": ["ears-few-ac", "ears-low-template-diversity"]
}
```

字段约束:
- `schema_version`: 整数,当前 1;不匹配时返回明确错误(不静默忽略)
- `spec_name`: 字符串,必须与所在目录名一致;不一致时返回错误(防止误粘贴)
- `ignored_issues`: 字符串数组,每个元素是一个 Issue Code;未知 Code 也接受
  (前向兼容:新增 checker 不会被旧 baseline 误报"未知 key")

### Issue 扩展

现有 `Issue` 结构体(`ears.go:33-39`)无需新增字段。`Severity` 二值
(`error` / `warning`)足够表示"门禁强度";baseline 是否忽略由 `Issue.Code`
经 `Baseline.ShouldIgnore` 查询决定,在打印前过滤。

## Implementation Details

### 单行单 SHALL 检测(AC-2)

```go
func CheckSingleSHALLLine(text string) []Issue {
    var out []Issue
    for i, line := range strings.Split(text, "\n") {
        if !EARSRe.MatchString(line) {
            continue
        }
        // 仅计算本行 SHALL 出现次数;跨行续行不算
        n := len(SHALLRe.FindAllString(line, -1))
        if n >= 2 {
            out = append(out, Issue{
                Severity: SeverityWarning,
                Code:     "ears-multi-shall-line",
                Message:  fmt.Sprintf("line %d contains %d SHALL keywords — split into separate AC lines", i+1, n),
            })
        }
    }
    return out
}
```

### 模板多样性(AC-4)

```go
func CheckTemplateDiversity(text string) []Issue {
    templates := 0
    for _, re := range []*regexp.Regexp{WHENRe, WHILERe, WHERERe, UNLESSRe, IFTHENRe, SHALLRe} {
        if re.MatchString(text) {
            templates++
        }
    }
    if templates < 2 {
        return []Issue{{Severity: SeverityWarning, Code: "ears-low-template-diversity", ...}}
    }
    return nil
}
```

注:`SHALLRe` 命中代表"无条件基线"也算一种模板。

### 触发词模糊词检测(AC-7)

```go
var triggerVagueWords = regexp.MustCompile(`(?i)\b(busy|slow|normal|large|small|many|recently|soon|fast|robust|flexible|seamless|intuitive)\b`)

func CheckTriggerObservable(text string) []Issue {
    var out []Issue
    for i, line := range strings.Split(text, "\n") {
        // 仅检查 WHEN/WHILE 行的"触发词"部分
        if !WHENRe.MatchString(line) && !WHILERe.MatchString(line) {
            continue
        }
        // 提取 WHEN/WHILE 之后、THE SYSTEM SHALL 之前的子串
        trigger := extractTrigger(line)
        if triggerVagueWords.MatchString(trigger) {
            out = append(out, Issue{Severity: SeverityWarning, Code: "ears-trigger-unobservable", ...})
        }
    }
    return out
}
```

### Baseline 加载(AC-11)

```go
type Baseline struct {
    SchemaVersion int      `json:"schema_version"`
    SpecName      string   `json:"spec_name"`
    IgnoredCodes  []string `json:"ignored_issues"`
}

func LoadBaseline(specDir string) (Baseline, error) {
    path := filepath.Join(specDir, ".baseline.json")
    data, err := os.ReadFile(path)
    if errors.Is(err, fs.ErrNotExist) {
        return Baseline{}, nil // 不存在 = 默认空基线
    }
    if err != nil {
        return Baseline{}, fmt.Errorf("read baseline: %w", err)
    }
    var b Baseline
    if err := json.Unmarshal(data, &b); err != nil {
        return Baseline{}, fmt.Errorf("parse baseline: %w", err)
    }
    if b.SchemaVersion != 1 {
        return Baseline{}, fmt.Errorf("baseline schema_version %d not supported (want 1)", b.SchemaVersion)
    }
    expected := filepath.Base(specDir)
    if b.SpecName != "" && b.SpecName != expected {
        return Baseline{}, fmt.Errorf("baseline spec_name %q != directory %q", b.SpecName, expected)
    }
    return b, nil
}

func (b Baseline) ShouldIgnore(code string) bool {
    for _, c := range b.IgnoredCodes {
        if c == code {
            return true
        }
    }
    return false
}
```

`linter.go` 的入口增加一层过滤:

```go
func Spec(specDir string) []Issue {
    base, err := LoadBaseline(specDir)
    if err != nil {
        return []Issue{{Severity: SeverityError, Code: "baseline-parse-error", Message: err.Error()}}
    }
    issues := specIssuesUnfiltered(specDir) // 现有逻辑
    var out []Issue
    for _, iss := range issues {
        if base.ShouldIgnore(iss.Code) {
            iss.Message = "[baseline] " + iss.Message
        }
        out = append(out, iss)
    }
    return out
}
```

`Gate()`(`linter.go:75`)的 ERROR 过滤逻辑要相应改成"忽略 baseline
命中的 issue"。

### CLI `--strict-baseline`(AC-12)

```go
var strictBaseline bool
lintCmd.Flags().BoolVar(&strictBaseline, "strict-baseline", false,
    "ignore .baseline.json (treat as not configured)")
```

传给 `linter.Spec` 时若 `strictBaseline == true` 则传空 Baseline。

## Error Handling

- `.baseline.json` 解析失败 → 返回 `baseline-parse-error` ERROR,告诉用户
  哪一行 JSON 错了;不 panic。
- `schema_version` 不匹配 → 返回明确错误并提示升级 free-kiro 版本。
- `.baseline.json` 的 `spec_name` 与目录不一致 → 返回明确错误,防止误粘贴。
- baseline 中包含未知 Issue Code → 不报错(前向兼容),只在 lint 输出末尾
  输出 INFO 提示"X 个 baseline 条目当前未生效"。

## Testing Strategy

| 层 | 覆盖 | 文件 |
|---|---|---|
| 单元 | 每个 `Check*` 函数至少 1 阳 1 阴 | `internal/lint/quality_test.go`(新建) |
| 单元 | `LoadBaseline` 5 路径(不存在/空/合法/版本错/spec_name 错) | `internal/lint/baseline_test.go`(新建) |
| 单元 | 补齐 placeholder-ac、missing-current、文档级 missing | `internal/lint/ears_test.go`(扩) |
| 集成 | 跑 `free-kiro lint` 对 7 个历史 spec 不报错 | `internal/cli/spec_lint_test.go`(扩)或 e2e |
| 文档 | `docs/EARS.md` 中 10 条规则描述与代码 Code 一致 | 手工 grep 对照 |

覆盖率门槛:`go test -cover ./internal/lint/...` ≥ 90%。

## Migration / Rollout

**默认上线**(合并到 main 即生效):
- AC-1 / AC-10 直接以 ERROR 生效 → 历史 spec 配 `.baseline.json` 保护(AC-17)
- 其余 8 条 AC 默认以 WARNING 上线 → 不阻塞 CI,但 IDE / `make lint` 输出可见

**灰度上线**(可选,若发现 WARNING 噪音过大):
- 通过环境变量 `FREE_KIRO_QUALITY_STRICT=1` 把全部 WARNING 临时升 ERROR,
  仅用于本地调试,不进入 CI 默认。

**回滚**:删除对应 `Check*` 调用行即可,无 schema 变更;`.baseline.json`
即使被误配也不会触发数据迁移,删除文件即恢复严格模式。

**`.baseline.json` 生成**:
- 由本 spec 的任务 #17 自动给 7 个历史 spec 各生成一份(空 list)
- 工程师后续若想让某历史 spec 真正走严格 lint,可编辑该文件的
  `ignored_issues` 数组逐步移除条目。