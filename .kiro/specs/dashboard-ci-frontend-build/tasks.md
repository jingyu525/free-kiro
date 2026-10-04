# dashboard-ci-frontend-build — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B

依赖规则：
  - 不能自引用（#1 不能依赖 #1）
  - 不能指向不存在的任务编号
  - 不能形成循环

Wave 划分（按依赖自动计算，free-kiro task list 可视化）：
  - Wave 1：#1 Makefile dashboard-dist + #2 composite action（并行）
  - Wave 2：#3 Makefile ci target / #4 ci.yml 5 job 接入 / #5 docs 同步（并行）
  - Wave 3：#6 本地 + CI 端到端验证

产物文件清单（最终 PR diff）：
  - 新增 .github/actions/build-dashboard/action.yaml（composite action）
  - 修改 .github/workflows/ci.yml（5 job 各加 1 step）
  - 修改 Makefile（dashboard-dist + ci targets + help 同步）
  - 修改 docs/STEERING.md（dashboard frontend dev / CI 段落）
总文件数 ≤ 4；预计 diff ≤ 80 行（少于 dashboard frontend 迁移 ~2000 行）。
-->

## Wave 1 — 本地入口 + CI 复用抽象

- [ ] #1 `Makefile` 新增 `dashboard-dist` target：依赖 `internal/visualize/static/package.json` 与 `pnpm-lock.yaml`；命令序列 `cd internal/visualize/static && pnpm install --frozen-lockfile && pnpm build`；`pnpm` 不在 PATH 时打印 `corepack enable pnpm` 或 `npm i -g pnpm@9` 并 exit 1；目标产物 `internal/visualize/static/dist/index.html` + `dist/assets/index-*.{js,css}` + `dist/assets/{react-vendor,query-vendor}-*.js` 5+ 文件
- [ ] #2 `.github/actions/build-dashboard/action.yaml` 新建 composite action：顺序 5 step `actions/checkout@v4`（可选）→ `actions/setup-node@v4`（node-version: '20'）→ `actions/cache@v4`（key `pnpm-${{ hashFiles('internal/visualize/static/pnpm-lock.yaml') }}`, path `internal/visualize/static/node_modules`, restore-keys fallback `pnpm-`）→ `corepack enable pnpm` → `make dashboard-dist`；input 仅 `working-directory`（默认 `.`，预留 monorepo 升级）；`runs.using: composite` + `runs.steps:`

## Wave 2 — CI 接入 + 本地 CI 入口 + 文档同步

- [ ] #3 `Makefile` 新增 `ci` target：顺序执行 `dashboard-dist` + `test` + `lint-go` + `build`（依赖 `make` 内部 target，等价于 `$(MAKE) dashboard-dist && $(MAKE) test && $(MAKE) lint-go && $(MAKE) build`）；任意 step 非 0 退出立即终止（`set -e` 隐式在 `$(MAKE)`）；更新 `help` target 输出包含 `dashboard-dist` 与 `ci` — [deps: #1]
- [ ] #4 `.github/workflows/ci.yml` 5 个 Go job（`test` / `lint-go` / `drift-check-steering` / `build` / `smoke`）每个 job 在原 setup-go@v5 之后、原 Go 命令之前插入 step `uses: ./.github/actions/build-dashboard`；不动其他 step、不改原 Go 命令；最终每个 job 内部 6 个 step（setup-go → build-dashboard → 原有） — [deps: #1,#2]
- [ ] #5 `docs/STEERING.md` 新增 §"dashboard frontend dev / CI"段落：说明本地 `make dashboard-dist` / `make ci` 等价于 CI 5 job 的 frontend build 步骤；交叉引用 `dashboard-frontend-react-vite-fsd` spec 与本 spec；列出 build 产物的 5 个文件名（index.html + assets/index-*.{js,css} + assets/{react,query}-vendor-*.js）；说明 pnpm cache key 是 `pnpm-lock.yaml` 的 hash — [deps: #1]

## Wave 3 — 端到端验证

- [ ] #6 本地 `make dashboard-dist` 跑通且 `internal/visualize/static/dist/` 5+ 文件存在；本地 `make ci` 全绿；PR push 上后 `gh pr checks <N>` 显示 ci.yml 5 个 Go job 全 pass；可选 `git revert <commit>` 验证回滚干净（composite action 路径独立删除） — [deps: #3,#4,#5]