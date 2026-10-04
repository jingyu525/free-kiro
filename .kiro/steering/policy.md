---
mode: always
description: free-kiro 项目策略（覆盖率门槛 / TODO owner / 协议合规 / 内部依赖路径 / commit 格式 / 代码规模 / PR 范围）
---

# free-kiro 项目策略（Project Policy）

> 适用对象：本仓库（`github.com/jingyu525/free-kiro`）的所有贡献者与
> AI agent。本文档**不是** Go 编码规范——通用 Go 风格请见
> [`.kiro/steering/coding-style.md`](./coding-style.md)；AI 协作硬性要求请见
> [`.kiro/steering/agent-rules.md`](./agent-rules.md)。
>
> 这里的策略是 **free-kiro 项目** 的约定，不是 Go 社区通用规则。复制
> 到其他 Go 项目时需要重新评估。

## 目录

1. [测试覆盖率门槛](#1-测试覆盖率门槛)
2. [TODO 注释 owner](#2-todo-注释-owner)
3. [依赖协议合规](#3-依赖协议合规)
4. [内部依赖路径](#4-内部依赖路径)
5. [commit message 格式](#5-commit-message-格式)
6. [代码规模上限](#6-代码规模上限)
7. [PR 范围约束](#7-pr-范围约束)

---

## 1. 测试覆盖率门槛

- 全包 `go test -cover` ≥ **70%**。
- 新增 / 修改行覆盖率 ≥ **80%**。
- PR 必须上传 `coverage.out` 并在 review 中说明未覆盖路径。
- Benchmark 与 fuzz 测试（`*_bench_test.go` / `*_fuzz_test.go`）不计入
  覆盖率统计。
- 例外：E2E 测试（如启动 CLI 子进程跑 smoke）允许低于此门槛，但需在
  PR 描述里说明。

## 2. TODO 注释 owner

- 所有 TODO 注释必须带 owner 前缀：`// TODO(<owner>): <一句话说明>`。
- owner 格式：项目维护者 GitHub 用户名（如 `// TODO(jingyu): ...`）。
- **不带 owner 的 `// TODO` 视为违规**，CI / review 拒收。
- 如功能未完成，**不要写 TODO 注释**——按
  [`.kiro/steering/agent-rules.md`](./agent-rules.md) §1「硬性要求」
  "先 spec 后代码"回 spec 阶段补 requirements/design。

## 3. 依赖协议合规

- **禁止引入 GPL 系传染协议依赖**：AGPL / LGPL 静态链接除外仍允许，
  其它 GPL 变种一律拒绝。
- 默认接受的协议：BSD-3 / MIT / Apache-2.0 / ISC / Unlicense。
- 协议合规不在 lint 强制，由人工 review（建议 `go-licence-checker`）。
- 升级第三方依赖时 major 版本单独 PR，并在 PR 描述里说明 breaking
  change；禁止 `go get -u ./...` 一锅端。

## 4. 内部依赖路径

- 本组织内模块统一以 `github.com/jingyu525/<repo>` 路径引用。
- `require` 块按字母序排列。
- 新增组织内模块依赖前，确认上游已发布稳定 tag（避免依赖 main 分支）。
- 内部依赖与第三方依赖在 `go.mod` 中以空行分隔，方便人工 review。

## 5. commit message 格式

- **默认中文**。格式：`类型(范围): 一句话描述`。
- 类型：`feat` / `fix` / `refactor` / `perf` / `test` / `docs` /
  `chore` / `style`。
- 示例：
  ```
  feat(spec): 支持 EARS 验证
  fix(hooks): Disabled 字段贯穿 Match + runAgentAction
  chore(lint): 启用 unused linter + 零死代码 spec 落地
  ```
- 跨多文件但同主题的改动应 squash 成 1 个 commit。
- 不允许无意义的"wip" / "update" / "fix typo" commit。

## 6. 代码规模上限

> 这是本项目的"软上限"，超出时 review 重点关注职责清晰度。

- 每个函数 ≤ **50 行**。超出说明分支过多，拆函数。
- 每个包文件 ≤ **500 行**。超出说明职责不清，拆包或拆文件。
- 每个 `internal/<pkg>` 包 ≥ 1 个 `_test.go` 文件（不可仅有产品代码）。
- 这些上限**不**由 lint 强制，由 review 与人工度量（`cloc` /
  `scc`）发现趋势性问题。

## 7. PR 范围约束

- 每个 PR **只解决 1 个 spec**。多 spec 并行会拖慢 review 与回滚。
- PR 描述必须引用对应 spec 路径（`.kiro/specs/<name>/`）。
- 紧急 hotfix 允许跳过 spec，但 PR 描述必须说明"未走 spec 流程的原因"
  并在事后补 spec 落地。
- 不允许 PR 在 review 中"顺便加个小功能"——拆 PR 或回到对应 spec。

---

> 文档结束。本文档作为项目级 steering（`mode: always`）注入 IDE 指令
> 文件，变更请同步更新 `.kiro/steering/policy.md` 与 `CONTRIBUTING.md`，
> 重跑 `free-kiro steering inject` 重生成 5 份 IDE 指令文件底部块。
