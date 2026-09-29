---
mode: always
description: free-kiro 仓库目录结构与代码组织约定
---

# Structure

## 顶层布局

```
free-kiro/
├── cmd/free-kiro/          # main 入口（Cobra root 命令）
├── internal/               # 全部业务逻辑（私有包）
│   ├── cli/                # Cobra 子命令注册与 flag 绑定
│   ├── ide/                # IDE hook 信封生成（Claude Code / CodeBuddy）
│   ├── lint/               # EARS + 结构 + 依赖图 lint 引擎
│   ├── spec/               # spec 文档读写 + 状态机推进
│   ├── steering/           # steering 注入与作用域解析
│   ├── taskgraph/          # tasks.md 依赖图 → wave 拓扑
│   ├── hooks/              # hook JSON 信封序列化
│   ├── watch/              # fsnotify 文件监听
│   ├── workspace/          # .kiro/ 目录发现与对账
│   ├── models/             # 跨包共享的数据结构（spec/hook/steering）
│   ├── visualize/          # ASCII tree / Mermaid / web dashboard
│   ├── skill/              # self-install 的 skill bundle 读写
│   ├── upgrade/            # 自升级下载 + SHA256 校验
│   ├── errors/             # 包级 sentinel 错误 + wrap helper
│   ├── ide/templates/      # IDE hook 配置文件模板
│   ├── spec/templates/     # spec 文档生成模板
│   └── visualize/static/   # web dashboard 静态资源
├── docs/                   # 面向用户的文档
├── contrib/                # 第三方分发物料（Homebrew formula、skill bundle）
├── testdata/               # 测试 fixture
├── install.sh              # curl|bash 安装脚本
├── .goreleaser.yaml        # 跨平台构建配置
├── go.mod / go.sum
└── README.md
```

## 包约定

- **包名**：小写单词，不复数，不缩写（`spec` 而非 `specs`，`lint` 而非 `l`）
- **导出符号**：每个 exported 类型/函数/常量必须有 godoc 注释
- **IO 隔离**：所有外部 IO（文件、网络、子进程）必须放在显式 adapter 后，
  便于测试时 mock
- **依赖方向**：`cli → models + 各领域包`，领域包之间禁止互相 import
  内部细节（只允许通过 `models` 共享类型）
- **错误处理**：跨包错误用 sentinel（在 `internal/errors`）+ `fmt.Errorf %w`
  wrap；不要把 string repr 当 error identity
- **测试**：单元测试与被测文件同包同目录（`*_test.go`），不开 mirror package

## 模板与生成器分离

- `*/templates/*.md.tmpl` 只放模板字符串
- 生成器是纯函数（输入结构 → 输出字符串），不做 IO
- 模板里不嵌入业务逻辑，逻辑全部前置到生成器