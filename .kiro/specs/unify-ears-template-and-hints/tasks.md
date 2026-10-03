# unify-ears-template-and-hints — Tasks

每完成一条把 `[ ]` 改成 `[x]`；wave 视图用 `free-kiro task list unify-ears-template-and-hints`。

依赖图：模板 + 文档锚点 + 测试三者互不依赖，全部并行；验证 wave 依赖三个 wave 1 任务都完成。

## Wave 1 — 改动（全部可并行）

- [ ] #1 改 `internal/spec/templates/requirements.md.tmpl` line 26-31：6 条 AC 示范前各加裸格式 `[AC-1]` … `[AC-6]` 前缀（保留 `<TODO:...>` 占位符），去掉 `- ` bullet 前缀（避免 `extractEARSLines` 漏计导致 `ears-few-ac` 误报）[no deps]
- [ ] #2 改 `docs/EARS.md` 三处 heading 加 Pandoc 风格 `{#english-slug}` 显式锚点：`## 五种模板 + 无条件基线 {#five-templates}`、`## 模糊词黑名单 {#vague-words}`、`## 语义质量门禁（新增 10 条） {#semantic-quality-gates}`，与 `internal/lint/{requirements,bugfix,quality}.go` 现有 Hint 字符串里的 slug 对齐 [no deps]
- [ ] #3 新建 `internal/lint/hints_test.go`：实现 `extractHeadingSlugs`（接受 Pandoc `{#slug}` 显式 + GitHub 隐式 slug，CJK 保留）+ `TestLintHintsResolve`（扫 `internal/lint/{requirements,bugfix,quality}.go` 所有 `Hint:` 字符串里的 `docs/EARS.md#<anchor>`，断言每个 anchor 都在 valid 集合里，失败时打印文件:行号 + anchor + valid keys）[no deps]

## Wave 2 — 端到端验证（依赖 Wave 1）

- [ ] #4 跑 `go test -cover ./internal/lint/...` 确认 `TestLintHintsResolve` PASS + 覆盖率 ≥ 90%；跑 `free-kiro spec new <test-spec> --prompt "x" --quick` + `free-kiro lint <test-spec>` 确认不报 `ears-ac-missing-id`；手动注入一个 broken anchor 验证测试 fail 信息含文件:行号 [deps: #1, #2, #3]

## Wave 3 — 审批上线

- [ ] #5 `free-kiro spec approve unify-ears-template-and-hints` → `free-kiro spec start unify-ears-template-and-hints` → git commit + push → `free-kiro spec complete unify-ears-template-and-hints` 进入 done 状态 [deps: #4]