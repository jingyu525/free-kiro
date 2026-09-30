# separate-codestyle-projectspecific — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

依赖不能：
  - 自引用（#1 不能依赖 #1）
  - 指向不存在的任务
  - 形成循环

完成后把 [ ] 改成 [x]；wave 视图用 `free-kiro task list <spec>`。
-->

- [x] #1 新建 docs/AGENT_RULES.md，承载原 docs/CODING_STYLE.md 第 8 章 AI agent 协作（含 8.1 硬性要求 5 条、8.3 上下文注入、8.4 失败处置、8.5 由工具强制、8.6 零豁免政策 ≥ 4 条、8.7 零死代码政策 ≥ 4 条）
- [x] #2 新建 docs/POLICY.md，承载散落在 docs/CODING_STYLE.md 第 5/6/7/8 章的 free-kiro 项目策略（覆盖率门槛、TODO 注释 owner、协议合规、内部依赖路径、commit message 中文、函数与文件行数上限、每个 PR 解决 1 个 spec）≥ 6 条 [deps: #1]
- [x] #3 修改 docs/CODING_STYLE.md：删除整章第 8 章"AI agent 协作"；从第 5/6/7 章移除迁移走的项目特定条款（覆盖率硬阈值、TODO owner、协议合规、内部依赖路径）；顶部声明改为"仅覆盖 Go 社区通用规范"并附 docs/AGENT_RULES.md 与 docs/POLICY.md 链接 [deps: #1,#2]
- [x] #4 修改 CLAUDE.md：硬约束摘录段改为指向 docs/AGENT_RULES.md（AI 协作硬性要求）与 docs/POLICY.md（项目策略）；原第 4 条"改动 > 50 行先 free-kiro spec new"保留为工具特定硬约束并明确标注 [deps: #1,#2,#3]
- [x] #5 检查 .golangci.yml / Makefile / CONTRIBUTING.md / .kiro/AGENTS.md / docs/STEERING.md 中是否有指向 docs/CODING_STYLE.md 第 8 章或被迁移条款的链接/锚点引用，若有则同步更新 [deps: #3]
- [x] #6 在 PR 描述附 `grep -rn "docs/CODING_STYLE.md#8" .` 与 `grep -rn "覆盖率\|TODO(jingyu)\|GPL\|jingyu525" docs/ CLAUDE.md` 的输出，确认无悬空引用与遗漏条款 [deps: #3,#4,#5]
- [x] #7 跑 `free-kiro lint separate-codestyle-projectspecific` 与 `make lint-go` 双双绿色；更新 .kiro/specs/separate-codestyle-projectspecific/tasks.md 把所有 [ ] 改为 [x] [deps: #6]
