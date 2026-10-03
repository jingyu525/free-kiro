# fix-inject-marker-match — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

Wave 划分（按依赖自动计算）：
  - Wave 1：#1-#3 修 inject 源码 + 注释 + 测试
  - Wave 2：#4-#5 端到端验证（rebuild + inject + 4 文件检查）
  - Wave 3：#6    commit + complete
-->

## Wave 1 — 修源码 + 加测试

- [ ] #1 `internal/steering/inject.go:196` 把 `strings.Contains(line, InjectMarkerStart)` 改为 `strings.TrimSpace(line) == InjectMarkerStart`；line 198 同样把 `strings.Contains(line, InjectMarkerEnd)` 改为严格行匹配 [no deps]
- [ ] #2 `internal/steering/inject.go:181-184` 注释从「line substring match」改为「EXACT line match (after strings.TrimSpace)」，并补充说明 prose 引用 marker 字符串字面量不会被匹配 [deps: #1]
- [ ] #3 `internal/steering/inject_test.go` 新增 `TestInject_IgnoresMarkerInProse`：构造 prose 段含 marker 字面量 + 文件末尾独立 marker block 两个区域；inject 后断言 prose 保持 verbatim + 真 marker block 内填入 product/structure/tech [deps: #1,#2]

## Wave 2 — 端到端验证

- [ ] #4 `go test -race -count=1 -run 'Inject' ./internal/steering/...` 0 FAIL；`TestStore_InjectAll_WritesBlockIntoMarkerRegion`（已有）+ `TestInject_IgnoresMarkerInProse`（新增）+ `TestStore_InjectAll_SkipsMissingMarker`（已有）+ `TestStore_InjectAll_OnlyGlob`（已有）+ `TestStore_InjectAll_NoAlwaysDocs`（已有）全部 PASS [deps: #3]
- [ ] #5 仓库根跑端到端：`go install ./cmd/free-kiro && free-kiro init --ide claude-code --overwrite-instructions && free-kiro steering inject`；`grep -n "free-kiro-managed:start" CLAUDE.md` 确认真 marker 仍在 line 85+；`sed -n '30,40p' CLAUDE.md` 确认解释段 line 33-34 字面量保持不变；CLAUDE.md / AGENTS.md / .cursorrules / .cursor/rules/free-kiro.md 四个文件 marker block 内含 `## product.md` / `## structure.md` / `## tech.md` 三个 H2 [deps: #4]

## Wave 3 — commit + 后续

- [ ] #6 `git commit -m "fix(steering): inject marker 严格行匹配，避免解释段误识别"` + push + `free-kiro spec complete fix-inject-marker-match` [deps: #5]