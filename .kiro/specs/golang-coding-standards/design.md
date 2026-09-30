# golang-coding-standards — Design

把"规范"分成三层落地：①**承载层**（Markdown 文档）②**门禁层**
（`golangci-lint` 配置 + Makefile/CI 钩子）③**入口层**（CONTRIBUTING.md
索引 + `free-kiro` 的 SessionStart 注入）。三层解耦，文档改不动 lint 配置，
lint 配错不影响文档阅读。

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│  入口层                                                    │
│  CONTRIBUTING.md  ──链接──▶ docs/CODING_STYLE.md          │
│  .kiro/AGENTS.md  ──SessionStart hook──▶ AI 自动读取规范  │
└──────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│  承载层（8 章节 Markdown）                                  │
│  docs/CODING_STYLE.md                                     │
│    1. 命名约定   2. 错误处理   3. 并发   4. 接口设计         │
│    5. 测试       6. 注释/文档  7. 依赖管理  8. AI 协作      │
└──────────────────────────────────────────────────────────┘
                            │
                            ▼
┌──────────────────────────────────────────────────────────┐
│  门禁层                                                    │
│  .golangci.yml   ─配置──▶ golangci-lint v1.61+            │
│  Makefile        ─包装──▶ `make lint` 走 free-kiro 链路   │
│  .github/workflows/ci.yml ──>  push/PR 必跑 lint         │
└──────────────────────────────────────────────────────────┘
```

关键决策：

- 选 `golangci-lint` 而不是裸 `go vet`/`staticcheck`：能用一个 YAML
  聚合 N 个 linter，CI 启动成本低，新加规则只需改 YAML 不动代码。
- 选 Markdown 而不是 godoc/internal package：规范要给**人**看
  （包括 AI agent），不是给 godoc 索引；放仓库根 `docs/` 最容易找。
- 不强制项目改 license header：保留范围灵活性，留给后续 spec。
- 不把规则写成"自定义 linter"：除非 `golangci-lint` 现有 linter 覆盖不到，
  否则不引入额外二进制（避免维护负担）。

## Components

| Component | Responsibility | Key API |
|---|---|---|
| `docs/CODING_STYLE.md` | 人读规范主文档；8 章节，每章 ✅/❌ 对照 | N/A |
| `.golangci.yml` | 聚合 linter 配置；可机械检查的规则 | `version: "2"` schema |
| `CONTRIBUTING.md`（修改） | 新增"编码规范"小节，含 `docs/CODING_STYLE.md` 链接 | N/A |
| `Makefile`（修改） | 新增 `lint-go` target，调用 `golangci-lint run ./...` | `make lint-go` |
| `.github/workflows/ci.yml`（修改） | 新增 lint job，`make lint-go` 失败阻断 merge | N/A |
| `.kiro/AGENTS.md`（修改） | SessionStart 注入时附带"先读 docs/CODING_STYLE.md" | N/A |

## Data Model

无数据结构。这是文档 + 配置文件项目。

## Error Handling

- lint 失败 = exit 1（`golangci-lint` 默认），被 CI 与 `free-kiro lint`
  同一契约复用。
- `golangci.yml` 解析失败 = `golangci-lint` 自带非 0 退出码，无需额外包。
- `make lint-go` 通过 `free-kiro lint` 链路串到 free-kiro 的退出码契约
  （0=OK / 1=ERROR / 2=engine / 3=usage），与现有 `lint` target 行为一致。

## Testing Strategy

- **配置 smoke test**：在 CI 跑 `golangci-lint run ./...` 验证 `.golangci.yml`
  至少能加载且在空仓库 / 当前仓库上不报无关 ERROR。
- **golden 文件**（可选）：若后续加自定义 rule，再上 `testdata/`。
- **人工 review**：8 章节每节至少 1 名项目维护者 sign-off。
- **dogfooding**：规范落地的当个 PR 本身必须过 `make lint-go`，即规范
  自己作为合规示范。

## Migration / Rollout

不需要分阶段迁移。规范是**新增文档 + 新增配置**，对现有代码不强制 reformat。

- 第一波：落文档 + `.golangci.yml`，CI 只 WARN 不阻断（`new-from-revive` 的
  `warn` 模式），观察 1 周。
- 第二波：CI 改为 ERROR 阻断；老代码违规按文件用 `//nolint:reason` 逐项豁免
  直到清理。
- 第三波（不在本 spec 范围）：扫 `//nolint` 残留，每季度清理一批。
