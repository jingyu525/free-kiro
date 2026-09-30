# improve-ears-quality — Tasks

每完成一条把 `[ ]` 改成 `[x]`;wave 视图用 `free-kiro task list improve-ears-quality`。

依赖图设计原则:基础设施先行 → 规则实现并行 → 装配集成 → 验证 → 审批上线。

## Wave 1 — 基础设施(全可并行)

- [x] #1 ears.go 新增 5 个独立模板正则(WHENRe/WHILERe/WHERERe/UNLESSRe/IFTHENRe)并在 ears_test.go 加单测 [no deps]
- [x] #2 internal/spec/models/models.go 新增 `MinAcceptanceCriteria = 3` 常量 + 单测 [no deps]
- [x] #3 internal/lint/quality.go 新建文件骨架(包级 imports + extractTrigger 公共函数 + Check 函数签名占位) [no deps]
- [x] #4 internal/lint/baseline.go 新建:Baseline struct + LoadBaseline + ShouldIgnore + 5 类边界单测 [no deps]
- [x] #5 internal/cli/spec_lint.go 加 `--strict-baseline` bool flag 并透传到 linter.Spec [no deps]
- [x] #6 internal/lint/ears_test.go 补 placeholder-ac / missing-current / 文档级 missing-requirements/missing-bugfix/missing-tasks/missing-design 共 6 个缺失的测试路径 [no deps]

## Wave 2 — 质量规则实现(依赖 #3 骨架,可全部并行)

- [x] #7 quality.go 实现 CheckEtcList (AC-1) + CheckSingleSHALLLine (AC-2) + 各 1 阳 1 阴单测 [deps: #3]
- [x] #8 quality.go 实现 CheckFewAC (AC-3) + CheckTemplateDiversity (AC-4) + 各 1 阳 1 阴单测 [deps: #1, #2, #3]
- [x] #9 quality.go 实现 CheckACMissingID (AC-5) + CheckMeasurableResponse (AC-6) + 各 1 阳 1 阴单测 [deps: #3]
- [x] #10 quality.go 实现 CheckTriggerObservable (AC-7) + CheckKeywordMisuse (AC-8) + 各 1 阳 1 阴单测 [deps: #1, #3]
- [x] #11 quality.go 实现 CheckPassiveResponse (AC-9) + 1 阳 1 阴单测 [deps: #3]

## Wave 3 — 装配集成(依赖 Wave 1+2)

- [x] #12 requirements.go: 在 Requirements() 末尾依次调用 #7-#11 的 Check 函数;`no-user-stories` 改 SeverityError(AC-10);加 Requirements() 整体集成单测覆盖"所有新规则同时命中"的复合场景 [deps: #7, #8, #9, #10, #11]
- [x] #13 linter.go: Spec() / Gate() 装配 baseline 过滤;`baseline-parse-error` 错误路径单测 [deps: #4, #5]

## Wave 4 — 文档与历史 spec 保护

- [x] #14 docs/EARS.md "关键约束" 一节新增 10 条规则的描述表(Code/严重级别/触发条件/❌ 反例);"模糊词黑名单" 一节改为"`etc`/`and/or` 由 lint ERROR 拦截,其余 9 个由 spec analyze info 提示" [no deps]
- [x] #15 给 .kiro/specs/ 下 7 个已完成 spec(`add-status-subcommand`/`cli-version-flag`/`enforce-golang-standards-zero-exemptions`/`golang-coding-standards`/`per-ide-instructions`/`refactor-go-best-practices`/`top1-demo-onboarding`)各生成一份空 `.baseline.json`,schema_version=1 + spec_name 与目录名一致 + 空 ignored_issues 数组 [no deps]

## Wave 5 — 端到端验证

- [x] #16 跑 `go test -cover ./internal/lint/... ./internal/spec/...`,确认覆盖率 ≥ 90%;跑 `free-kiro lint` 对 7 个历史 spec + 本 spec,确认历史 spec 不变红、本 spec 全 AC 通过 [deps: #12, #13, #14, #15]

## Wave 6 — 审批上线

- [x] #17 `free-kiro spec approve improve-ears-quality` → `free-kiro spec start improve-ears-quality`(全部 16 任务完成后 [deps: #16]
- [x] #18 实现全部完成 + commit + push,跑 `free-kiro spec complete improve-ears-quality` 进入 done 状态 [deps: #17]