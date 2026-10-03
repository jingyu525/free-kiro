---
mode: always
description: free-kiro 仓库目录结构与代码组织约定
---

# Structure

## 顶层布局

```
free-kiro/
├── cmd/free-kiro/            # main 入口（Cobra root 命令）
├── internal/                 # 全部业务逻辑（私有包）
│   ├── cli/                  # Cobra 子命令注册与 flag 绑定
│   ├── ide/                  # IDE hook 信封生成（Claude Code / CodeBuddy）
│   │   └── templates/        # IDE hook 配置文件模板
│   ├── lint/                 # EARS + 结构 + 依赖图 lint 引擎
│   ├── spec/                 # spec 文档读写 + 状态机推进 + drift 检测
│   │   ├── templates/        # spec 文档生成模板
│   │   ├── analyze.go        # requirement/design 静态分析
│   │   ├── drift.go          # baseline 锁与差异检测
│   │   └── warn.go           # advisory（vague / 重复 AC / 可追溯性）
│   ├── steering/             # steering 注入与作用域解析
│   ├── taskgraph/            # tasks.md 依赖图 → wave 拓扑
│   ├── hooks/                # hook JSON 信封序列化
│   ├── watch/                # fsnotify 文件监听
│   ├── workspace/            # .kiro/ 目录发现与对账
│   ├── models/               # 跨包共享的数据结构（spec/hook/steering）
│   ├── visualize/            # ASCII tree / Mermaid / web dashboard
│   │   ├── server.go         # http.Server + 路由
│   │   ├── server_handlers.go# HTML 渲染 + REST API
│   │   ├── server_sse.go     # fsnotify → SSE 实时推送
│   │   ├── etag.go           # ETag 生成 + 304 缓存协商
│   │   ├── io.go             # 文件 IO 适配层
│   │   ├── middleware.go     # HTTP 中间件
│   │   └── static/           # web dashboard 前端静态资源
│   ├── skill/                # self-install 的 skill bundle 读写
│   ├── upgrade/              # 自升级下载 + SHA256 校验
│   ├── frontmatter/          # YAML frontmatter 解析（spec/steering 共用）
│   ├── text/                 # 文本处理工具（normalize / dedent 等）
│   └── errors/               # 包级 sentinel 错误 + wrap helper
├── docs/                     # 面向用户的文档
├── contrib/                  # 第三方分发物料（Homebrew formula、skill bundle）
├── testdata/                 # 测试 fixture
├── install.sh                # curl|bash 安装脚本
├── .goreleaser.yaml          # 跨平台构建配置
├── .golangci.yml             # golangci-lint v2 配置（启用 staticcheck / errcheck / revive）
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

## dashboard 后端拆分约定

`internal/visualize/` 把 dashboard 服务拆成多个职责单一的文件：

- `server.go`：路由注册 + 生命周期
- `server_handlers.go`：HTML 渲染 + REST handler
- `server_sse.go`：SSE 长连接 + fsnotify 订阅
- `etag.go`：缓存协商逻辑（与 handler 解耦，便于单测）
- `io.go`：文件系统读取
- `middleware.go`：logging / CORS / recovery

新增 dashboard 端点时，先确认要落在哪个文件；横向功能（如指标采集）开新
文件，不要往 `server_handlers.go` 里堆。
