# update-contrib-skill-bundle — Design

## Architecture

本 spec 是一次**纯文档 bundle 同步**，不涉及 binary 改动。bundle 形态是
`contrib/skills/free-kiro/` 下的静态 markdown 文件 + 一个 `skill.json`
（版本 + sha256 元数据）。安装路径（`install.sh` / `skill install` /
`skill update`）会读取 `skill.json` 中的 sha256 来校验文件完整性。

变更面：

```
contrib/skills/free-kiro/
├── SKILL.md              ← 主入口：补齐命令树 / flag / 概念
├── skill.json            ← sha256 重新计算
└── references/
    ├── ops.md            ← watch / serve / report / doctor 新 flag
    ├── hooks.md          ← event 名两套命名 + free-kiro 内部事件
    ├── spec.md           ← spec 子命令与变体（已有，局部补）
    ├── prd-fetch.md      ← 不动（已对齐）
    └── troubleshooting.md← 不动（已对齐）
```

## Components

| Component | Responsibility |
|---|---|
| `SKILL.md` 顶层"快速参考"块 | 一个屏幕铺出当前 binary 全部入口；用户第一次打开 bundle 第一眼看到 |
| `SKILL.md` 触发词清单 | AI 助手据这段判断何时调用 skill |
| `SKILL.md` 自检段 | 列出 `skill show` / `doctor` 与退出码契约 |
| `references/spec.md` | 状态机表 + EARS 句式 + tasks 写法 + workflow 变体 + 漂移处理 + analyze advisory |
| `references/ops.md` | watch preset 全集 + serve endpoint + report flag + doctor 检查项 |
| `references/hooks.md` | IDE hook / 项目 hook 分层 + 两套 event 命名 |
| `skill.json` | `sha256.<file>` 字段必须用 `scripts/compute-skill-sha.sh` 重算 |

## Data Model

`skill.json` 顶层字段保持不变：

```jsonc
{
  "name": "free-kiro",
  "version": "0.7.0-dev",            // 与 binary 同步：本仓库 dev 构建保持 0.7.0-dev
  "free_kiro_min_version": "0.7.0",
  "license": "MIT",
  "repository": "https://github.com/jingyu525/free-kiro",
  "sha256": {
    "SKILL.md":                     "<sha256>",
    "references/hooks.md":          "<sha256>",
    "references/ops.md":            "<sha256>",
    "references/prd-fetch.md":      "<sha256>",
    "references/spec.md":           "<sha256>",
    "references/troubleshooting.md":"<sha256>"
  }
}
```

`metadata.free-kiro-min-version` 在 SKILL.md frontmatter 仍为 `0.7.0`（不动）。

## Error Handling

- `scripts/compute-skill-sha.sh` 计算 SHA256 必须用 `shasum -a 256`（macOS）
  或 `sha256sum`（Linux）；当前 macOS 实现即可，CI 走 Linux 时由 CI 镜像
  内的脚本处理（本次不动 CI）。
- 改文件后**必须**重跑 `compute-skill-sha.sh` 刷新 `skill.json`；漏算
  → install 时 exit 1（sha256 mismatch）。
- lint 校验：`free-kiro lint update-contrib-skill-bundle` 必须 exit 0
  才能完成 spec。

## Testing Strategy

| 验证 | 命令 |
|---|---|
| Spec 三件套过 lint | `free-kiro lint update-contrib-skill-bundle` |
| sha256 自洽 | `bash contrib/skills/scripts/compute-skill-sha.sh` 打印的 hash 与 `skill.json` 一致 |
| 命令树覆盖 | 人工对照 `free-kiro --help` 与 SKILL.md"快速参考"块 |
| 可装性 | `free-kiro skill install --dry-run --app claude-code` 不报错（如果机器已装则显示"已是最新"或重新 install） |

## Migration / Rollout

- `skill.json` 的 `version` 字段保持 `0.7.0-dev`（与本仓库当前 dev 构建对齐）。
- 发版时由 release pipeline 改成具体版本号（如 `0.7.1`），不在本 spec 范围。
- 已装用户：`free-kiro skill update` 会拉新版 bundle 自动覆盖。
- 没有破坏性变更：新增命令 / flag 是增量，不删任何已有入口。