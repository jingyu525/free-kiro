# CI 集成指南 — 5 步接入 free-kiro spec-lint 到你的团队仓库

本文档配套 [`top1-demo-onboarding` spec](../.kiro/specs/top1-demo-onboarding/)。
它描述如何把 free-kiro 的 spec-lint 门禁接入到任意团队的 GitHub 仓库。
如果你的 CI 不是 GitHub Actions,见末尾"其他 CI"段。

## 前置条件

- 团队仓库根目录有 `.kiro/` workspace(已跑 `free-kiro init`)
- `.kiro/specs/<name>/` 下至少有一份完整的 spec(requirements.md + design.md + tasks.md 三件套)
- 团队已接受"PR 改动必须过 lint 门禁"的纪律

## 5 步接入

### 第 1 步:复制 yaml 模板到你的仓库

```bash
curl -sL https://raw.githubusercontent.com/jingyu525/free-kiro/main/.github/workflows/spec-lint.yml.example \
  -o .github/workflows/spec-lint.yml
```

或者从 free-kiro 本地 clone 复制:

```bash
# 在你团队主项目的根:
cp path/to/free-kiro/.github/workflows/spec-lint.yml.example \
   .github/workflows/spec-lint.yml
```

`.example` 后缀是故意保留的 — 这样 free-kiro 仓库本身不会自动跑这个 workflow。

### 第 2 步:配置 paths 触发器

默认模板的 `on.pull_request.paths` 覆盖三类文件:

- `.kiro/specs/**` — spec 文档本体
- `.kiro/steering/**` — steering 上下文文档(影响 spec 生成)
- `.kiro/.current` — 当前活跃 spec 指针

如果你的 spec 目录有自定义路径(例如 `.mono/specs/**`),修改 `paths:` 字段:

```yaml
on:
  pull_request:
    branches: [main]
    paths:
      - ".kiro/specs/**"
      - ".mono/specs/**"   # ← 加你的路径
```

### 第 3 步:提交并打开 PR

```bash
git add .github/workflows/spec-lint.yml
git commit -m "ci: 加入 free-kiro spec-lint 门禁"
git push -u origin feat/ci-spec-lint
gh pr create --title "ci: 加入 free-kiro spec-lint 门禁" --body "..."
```

打开 PR 后,GitHub Actions 会自动跑 `free-kiro lint`;任何 spec 文档错误都会
在 PR 检查页签显示。CI 阻断 merge 直到修复。

### 第 4 步:(可选)加 status badge

在 README 顶部加 badge:

```markdown
[![spec-lint](https://github.com/<owner>/<repo>/actions/workflows/spec-lint.yml/badge.svg)](https://github.com/<owner>/<repo>/actions/workflows/spec-lint.yml)
```

替换 `<owner>` / `<repo>` 为你的 GitHub owner / repo 名。

### 第 5 步:团队 onboarding 文档链接到本文

把本文档 URL 加入 `CONTRIBUTING.md` 的"开发流程"段,让新成员知道
"改了 .kiro/ 下的文件,会自动过 lint 门禁,失败时按 docs/EARS.md
修复"。

## 工作流做了什么

| 步骤 | 行为 | 退出码语义 |
|---|---|---|
| `actions/checkout@v4` | 拉 PR 代码 | n/a |
| `setup-free-kiro@v0` | 装 free-kiro 二进制到 PATH | n/a |
| `free-kiro lint` | EARS + 结构门禁 | 1 = ERROR(阻断);0 = OK |
| `free-kiro doctor` | 环境/工作区诊断 | 仅警告(continue-on-error) |
| `free-kiro report` | 写 `.kiro/REPORT.md` | 仅警告(continue-on-error) |
| 上传 artifact | 14 天保留 `.kiro/REPORT.md` | n/a |

`free-kiro lint` 的退出码契约详见 [`CLI.md`](CLI.md#退出码契约)。

## 退出码契约

| free-kiro lint 退出码 | GitHub Actions | 含义 |
|---|---|---|
| 0 | ✅ step pass | 所有 spec 合规 |
| 1 | ❌ step fail | spec 错误,阻断 merge |
| 2 | ❌ step fail | workspace 缺失 / IO 失败(罕见) |
| 3 | ❌ step fail | usage error(罕见,通常说明 setup action 配置错) |

## 高级:本地先验证再 push

推荐让团队成员本地先跑 `free-kiro lint`,避免 CI round-trip:

```bash
# 添加到团队 Makefile(或 package.json scripts):
spec-lint:
    free-kiro lint

# 个人 pre-commit hook(可选):
# 在 .git/hooks/pre-commit 加:
#   free-kiro lint || { echo "lint failed; see docs/EARS.md"; exit 1; }
```

## 其他 CI 系统

### GitLab CI

`.gitlab-ci.yml`:

```yaml
spec-lint:
  image: golang:1.27
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/jingyu525/free-kiro/main/install.sh | bash
  script:
    - free-kiro lint
  rules:
    - changes:
        - .kiro/specs/**/*
        - .kiro/steering/**/*
```

### Jenkins / Buildkite / Drone

类似 pattern:下载 free-kiro 二进制到 `$PATH`,在 spec 改动时跑 `free-kiro lint`。
具体 step 略。

## 故障排查

| 症状 | 处置 |
|---|---|
| `setup-free-kiro` 失败 | 检查 `version:` 是否有效;v0.7.0 是当前稳定 |
| `free-kiro: command not found` | 在 step 里加 `echo $PATH` 确认 setup action 成功 |
| lint 报 `missing-requirements` | spec 文档不齐;检查 `requirements.md` 是否提交 |
| lint 报 `placeholder-ac` | spec 还有 `<TODO:…>` 占位符;按 `docs/EARS.md` 替换 |
| 改了 spec 但 CI 不跑 | 检查 `paths:` 是否覆盖了你的改动路径 |

---

如有问题,在 [free-kiro issues](https://github.com/jingyu525/free-kiro/issues) 开 issue。