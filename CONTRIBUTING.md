# Contributing to free-kiro

欢迎贡献！free-kiro 是一个**规约驱动开发**的 CLI 项目：所有改动
**先 spec 后代码**。本文档指引你 5 分钟内上手。

## 1. 快速路径

```bash
# 1. 克隆与初始化
git clone https://github.com/jingyu525/free-kiro.git
cd free-kiro
make build                      # 编译到 ./bin/free-kiro

# 2. 跑全量 CI（spec 门禁 + Go lint + race test）
make ci

# 3. 创建你的第一个 spec（任何 > 50 行的改动都要走这步）
free-kiro spec new my-feature --prompt "一句话需求"

# 4. 写完三件套后
free-kiro lint my-feature       # 必须全绿
free-kiro spec approve my-feature
free-kiro spec start my-feature
# ... 实现代码 ...
free-kiro spec complete my-feature
```

## 2. 编码规范（必读）

**所有 Go 代码（含 AI agent 生成）必须遵守以下三份文档**：

- [`.kiro/steering/coding-style.md`](.kiro/steering/coding-style.md) — Go 社区通用编码规范
  （命名 / 错误处理 / 并发 / 接口 / 测试 / 注释 / 依赖），共 7 章。
- [`.kiro/steering/agent-rules.md`](.kiro/steering/agent-rules.md) — AI agent 协作硬性要求 +
  零豁免 / 零死代码政策（违反任意一条 = PR 拒收）。
- [`.kiro/steering/policy.md`](.kiro/steering/policy.md) — free-kiro 项目特定策略（覆盖率门槛、
  TODO 注释 owner、协议合规、commit 格式、代码规模上限、PR 范围约束）。

`coding-style.md` 关键章节速览：

| # | 章节 | 关键约束 |
|---|---|---|
| 1 | 命名约定 | 包名 = 目录名；不写匈牙利命名；缩写词全大写 |
| 2 | 错误处理 | 不吞错误；wrap 用 `%w`；不 panic 做控制流 |
| 3 | 并发 | 优先 `chan`；`context.Context` 传取消；锁不导出 |
| 4 | 接口设计 | 使用方定义；小接口；返回具体类型 |
| 5 | 测试 | 表驱动；`-race` 必跑；`t.TempDir()` |
| 6 | 注释与文档 | 解释"为什么"；导出符号必 godoc |
| 7 | 依赖管理 | 最小依赖；锁版本 |

可机器检查的规则由 `.golangci.yml` 启用 6 个 linter（`govet` / `staticcheck` /
`errcheck` / `gofmt` / `goimports` / `revive`），CI `lint-go` job 会强制
门禁。其余规则由人工 review。

## 3. 工作流

- **改动 < 50 行（typo / 文档 / 注释）**：直接 PR，commit message 用中文
  `类型(范围): 一句话`（如 `docs: 修正错别字`）。
- **改动 ≥ 50 行（feature / refactor / bugfix）**：先 `free-kiro spec new` →
  生成 requirements / design / tasks → `lint` 全绿 → `approve` → `start` →
  实现 → `complete`。详见 `.kiro/AGENTS.md` 与 `docs/WORKFLOW.md`。
- **Bugfix**：用 `free-kiro spec new <name> --type bugfix`，
  生成 `bugfix.md` 而非三件套。

## 4. 开发环境

| 工具 | 版本要求 | 安装 |
|---|---|---|
| Go | ≥ 1.23（项目 go.mod 锁 1.27） | `brew install go` |
| golangci-lint | ≥ v1.61（v1.65+ 更好） | `brew install golangci-lint` |
| free-kiro | latest | `brew install jingyu525/free-kiro/free-kiro` |
| git / make / jq | 系统自带 | — |

## 5. 提 PR 流程

1. 在 fork 的 feature branch 上开发。
2. 跑 `make ci`，三项必须全绿：
   - `lint`：spec 文档结构合规
   - `lint-go`：Go 源码规范
   - `test`：`-race` 通过 + 覆盖率报告
3. commit message 中文，格式：
   ```
   类型(范围): 一句话

   - 详细说明 1
   - 详细说明 2

   测试: <如何验证>
   ```
4. PR 描述里 link 到对应的 spec 文件：`Closes .kiro/specs/<name>/`。
5. 至少 1 名 maintainer approve 才能 merge。

## 6. 反馈与帮助

- 🐛 **Bug** → [GitHub Issues](https://github.com/jingyu525/free-kiro/issues)
- 💡 **功能请求** → 先开 spec，PR spec 文档后再写代码
- 💬 **讨论** → [GitHub Discussions](https://github.com/jingyu525/free-kiro/discussions)
- 📖 **设计文档** → [`docs/`](.)（CLI.md / EARS.md / HOOKS.md / WORKFLOW.md）+ [`.kiro/steering/`](.)（product.md / structure.md / tech.md / coding-style.md / agent-rules.md / policy.md）

---

再次感谢你的贡献！🎉
