# improve-ears-quality

把 free-kiro 的 EARS 验收标准从"形状门禁"（6 模板存在性 + 段存在性）升级到
"语义质量门禁"（响应可测量、触发词可观测、关键字不混用、单行单 SHALL、最少 AC
数、模板多样性、AC ID 编号），并为已有 7 个 spec 提供 baseline 白名单机制以
平滑过渡。所有新增规则采用"ERROR 阻断 advance + WARNING 仅上报"的二值体系。

## User Stories

- As a free-kiro 维护者, I want 把 EARS 验收标准从形状门禁升级到语义质量
  门禁, so that AI agent 写出的 spec 真正可机读、可追溯,不再靠人工挑刺。
- As a CI 流水线, I want 新增的 10 条 checker 在 `free-kiro lint` 时以
  ERROR/WARNING 退出码暴露问题, so that 违规 spec 在合并前被自动拦截。
- As a 历史 spec 的作者, I want 有一个 baseline 白名单机制让 7 个已完成
  spec 优雅过渡, so that 不需要一次性批量重写历史文档。
- As a 新的 spec 作者, I want 每条 AC 必须带 `[AC-N]` 编号, so that
  requirements.md 与 tasks.md 的双向追溯从启发式升级为确定性匹配。

## Acceptance Criteria

[AC-1] WHEN 开发者执行 `free-kiro lint` 且 requirements.md 中某条 EARSRe
命中行内出现 `\b(etc|and/or)\b` 时 THE SYSTEM SHALL 报告
`ears-etc-list` ERROR 并以 exit code 1 退出。

[AC-2] WHEN requirements.md 中某行 EARSRe 命中的子串里 `SHALL` 出现次数
≥ 2 时 THE SYSTEM SHALL 报告 `ears-multi-shall-line` WARNING（允许多行
续行的 SHALL 跨行,只对单行内多次出现报错）。

[AC-3] WHEN requirements.md 中 EARSRe 命中行数小于
`models.MinAcceptanceCriteria`(默认 3)时 THE SYSTEM SHALL 报告
`ears-few-ac` WARNING,其中 `MinAcceptanceCriteria` 通过常量配置便于
后续 spec 自定义最小门槛。

[AC-4] WHEN requirements.md 中 6 种 EARS 模板（WHEN/WHILE/WHERE/UNLESS/
IF-THEN/无条件基线)实际命中种类数小于 2 时 THE SYSTEM SHALL 报告
`ears-low-template-diversity` WARNING,提示作者覆盖状态/异常/可选分支。

[AC-5] WHEN 某条 EARS AC 行未匹配正则 `^\s*-\s*\[AC-\d+\]\s+` 时 THE
SYSTEM SHALL 报告 `ears-ac-missing-id` WARNING;AC 编号必须位于行首且
紧跟连字符空格。

[AC-6] WHEN 某条 EARS AC 在 SHALL 之后子串里不含任意可测量承诺（数字 +
时间单位 ms|s|min|h、HTTP 状态码 2xx|4xx|5xx、百分比、`within`、
`at most`、`<`/`>` 加数字)之一时 THE SYSTEM SHALL 报告
`ears-response-immeasurable` WARNING。

[AC-7] WHEN WHEN/WHILE 行的触发词子串里出现 `busy|slow|normal|large|
small|many|recently|soon|fast|robust|flexible|seamless|intuitive` 模糊词
时 THE SYSTEM SHALL 报告 `ears-trigger-unobservable` WARNING。

[AC-8] WHEN WHILE 行的状态子串以过去式动词(`logged|submitted|clicked|
executed|started|finished|failed|expired`)结尾时 THE SYSTEM SHALL 报告
`ears-keyword-misuse-while-as-when` WARNING,提示改用 WHEN。

[AC-9] WHEN SHALL 之后子串以 `be|is|are|been` 开头(纯被动无主体)时 THE
SYSTEM SHALL 报告 `ears-passive-response` WARNING。

[AC-10] WHEN requirements.md 中 `User Stories` 段缺失时 THE SYSTEM
SHALL 把现有 `no-user-stories` 检出的 severity 由 WARNING 升级为 ERROR,
沿用原 Code 字符串以便既有文档链接不破。

[AC-11] WHERE spec 目录中存在 `.kiro/specs/<name>/.baseline.json` 时
THE SYSTEM SHALL 加载它作为该 spec 的 lint 白名单;`.baseline.json` 的
schema 为 `{"ignored_issues": ["<issue-code>", ...]}`,命中的 issue 在
`free-kiro lint` 输出中仍列出但 prefix 为 `[baseline]`,且不计入
`Gate()` 的 ERROR 阻断集合,以免阻塞 `spec approve` 前向转移。

[AC-12] WHEN 用户执行 `free-kiro lint <name> --strict-baseline` 时 THE
SYSTEM SHALL 忽略 `.baseline.json` 中所有条目,等同于未配置 baseline,
用于历史 spec 真正想"重新审视"时的严格模式。

[AC-13] WHEN 用户执行 `free-kiro spec approve <name>` 且该 spec 缺失
`.baseline.json` 但其 requirements.md 命中任何新增的 WARNING 类规则
时 THE SYSTEM SHALL 输出 advisory 提示"建议补 .baseline.json",
但不阻断审批通过。

[AC-14] WHEN 工程师执行 `go test ./internal/lint/...` 时 THE SYSTEM
SHALL 对每条新增 checker 提供至少 1 个阳性用例(命中规则)与 1 个阴性
用例(不命中);同时补齐 `placeholder-ac`、`missing-current`、文档级
`missing-requirements`/`missing-bugfix`/`missing-tasks`/`missing-design`
此前缺失的测试路径,使 `internal/lint/` 覆盖率不低于 90%。

[AC-15] THE SYSTEM SHALL 在 `docs/EARS.md` "关键约束" 一节新增 10 条规则的描述、Code、严重级别、触发条件与一个 ❌ 反例;并在 "模糊词黑名单" 一节明确**该节列出的 2 个禁用词**由 lint ERROR 拦截,其余 9 个模糊词由 `spec analyze` info 提示(具体词名落在 docs 文档正文中,不在本 AC 文本中重复出现以免自我拦截)。

[AC-16] THE SYSTEM SHALL 在 `internal/lint/ears.go` 新增 5 个独立模板
正则 `WHENRe`/`WHILERe`/`WHERERe`/`UNLESSRe`/`IFTHENRe`,并 export 出去
便于 AC-4 / AC-7 / AC-8 复用,避免现有合并 `EARSRe` 重复维护。

[AC-17] THE SYSTEM SHALL 为 `.kiro/specs/` 下 7 个已完成 spec
(`add-status-subcommand`、`cli-version-flag`、`enforce-golang-standards-
zero-exemptions`、`golang-coding-standards`、`per-ide-instructions`、
`refactor-go-best-practices`、`top1-demo-onboarding`)各生成一份默认
`.baseline.json`(空 `ignored_issues` 列表),保证新增规则上线后历史
spec 不瞬时变红。

## Out of Scope

- 不为 design.md / steering.md 引入新的 EARS 规则;这两类文档继续由
  现有的形状门禁与 drift 检查覆盖。
- 不修改 spec 模板里的 TODO 占位符提示骨架;它对新 spec 作者仍然有
  引导价值,只是我们的 spec 不能带任何 TODO 标记通过 lint。
- 不引入机器学习 / LLM 二次审阅 spec 的能力;10 条规则全部基于正则与
  启发式,可离线跑、可单元测试。
- 不动 `enforce-golang-standards-zero-exemptions` 这类 Go 风格 spec
  自身的零豁免门禁;它的合规性由自身 baseline 接管。
- 不在 `free-kiro lint` 之外再开一个新命令;所有规则统一挂在 `lint`
  入口下,保持 CI 调用面稳定。