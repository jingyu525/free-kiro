# update-contrib-skill-bundle — Tasks

- [x] #1 重写 `contrib/skills/free-kiro/SKILL.md` 的"快速参考"块：补齐 `spec quick` / `spec sync` / `spec show` / `spec list` / `spec analyze` / `spec status` / `spec next`、`upgrade --check --force`、`demo --ide --no-color`、`serve --bind --port --open`、`report --stdout --output`、`watch --preset default|lint|status|reactive|full --command --debounce --root --verbose`、`skill path --app`、skill uninstall --app、hook run --file、steering context 全部命令
- [x] #2 在 SKILL.md 加 `## Spec 变体与旗标` 段：解释 `spec new --workflow requirements-first|design-first` / `--type feature|bugfix` / `--quick` 三个变体 flag [deps: #1]
- [x] #3 修 SKILL.md 中 `kiro spec next` 拼写 → 改为 `free-kiro spec next` [deps: #1]
- [x] #4 扩 SKILL.md"排错"表：增加 `dependency dangling` / `self-reference` / `cannot approve: tasks.md must be generated first` / `spec status 报 drift` / `skill show 报 sha mismatch` 行 [deps: #1]
- [x] #5 扩 SKILL.md"触发词"清单：增加 `升级` / `demo` / `quick spec` / `sync` / `分析一下 spec` / `看板` 触发词 [deps: #1]
- [x] #6 改 `contrib/skills/free-kiro/references/spec.md` 的"Workflow 变体"段：补 `spec sync`（消除漂移） / `spec analyze`（advisory） / `spec status`（看漂移） / `spec list` / `spec show --tree` [deps: #1]
- [x] #7 改 `contrib/skills/free-kiro/references/ops.md` 的 watch 段：列全 5 个 preset（default / lint / status / reactive / full）、加 `--command` 自定义示例、改默认 debounce 为 500ms、加 `--root` / `--verbose` 说明 [deps: #1]
- [x] #8 改 `references/ops.md` 的 serve 段：加 `--bind` / `--port` / `--open` flag 与 `/api/summary` / `/api/specs` / `/api/spec/<n>` 端点列表 [deps: #1]
- [x] #9 改 `references/ops.md` 的 doctor 段：加 `--strict` / `--verbose` flag 与全部 5 项检查项说明 [deps: #1]
- [x] #10 改 `references/ops.md` 的 report 段：加 `--stdout` / `--output` flag + 报告包含 mermaid 图的说明 [deps: #1]
- [x] #11 改 `references/hooks.md`：event 名段补全 free-kiro 内部 `file.save` / `file.create` / `file.delete` / `prompt.submit` / `task.run` / `manual` 与 Kiro 官方 v1 信封两套命名兼容说明 [deps: #1]
- [x] #12 重算 `contrib/skills/free-kiro/skill.json` 的 `sha256.*` 字段：跑 `bash contrib/skills/scripts/compute-skill-sha.sh` 后写回 JSON [deps: #1, #2, #3, #4, #5, #6, #7, #8, #9, #10, #11]
- [x] #13 跑 `free-kiro lint update-contrib-skill-bundle` 校验 spec 三件套 + 跑 `free-kiro skill install --dry-run --app claude-code` 验证 sha256 自洽，exit 0 [deps: #12]