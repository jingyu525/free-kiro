# docs-template-marker-explanation — Tasks

```
依赖图（无依赖→依赖他人）：
  wave 1: #1
  wave 2: #2, #3
  wave 3: #4
  wave 4: #5, #6
  wave 5: #7
  wave 6: #8
```

- [x] #1 改 `internal/ide/templates/instructions_zh.md` 的"项目上下文（自动注入的 steering）"段，把顶部 `# free-kiro-managed:` + 底部 `<!-- free-kiro-managed:start/end -->` 两套 marker 的注入源 / 用途 / 修改路径都讲清楚
- [x] #2 改 `internal/ide/templates/instructions_en.md` 同段（英文翻译，marker 字符串原文保留）[deps: #1]
- [x] #3 改 `internal/ide/templates/agents_zh.md` 同段（中文，与 #1 平行）[deps: #1]
- [x] #4 改 `internal/ide/templates/agents_en.md` 同段（英文，与 #2 平行）[deps: #2,#3]
- [x] #5 在 `internal/ide/ide_test.go` 新增 `TestInstructionTemplate_ExplainsBothMarkers` + helper，断言 4 个模板正文同时含两套 marker 字符串 + prependMarker / steering inject 注入源关键词 [deps: #1,#2,#3,#4]
- [x] #6 改 `docs/STEERING.md` §"自动注入到 IDE 指令文件"段，同步说明两套 marker 的注入源 / 用途 / 修改路径 [deps: #1,#2,#3,#4]
- [x] #7 跑验证三件套：`free-kiro lint docs-template-marker-explanation` + `go test -race ./internal/ide/...` + 重编译 free-kiro 后 `free-kiro init --ide auto --overwrite-instructions` 同步 4 个实际 IDE 指令文件 + `free-kiro steering inject` 注入 always-mode 内容，确认顶部 / 底部 marker + 新说明同步落地 [deps: #5,#6]
- [x] #8 跑 `free-kiro spec complete docs-template-marker-explanation` 收尾 [deps: #7]