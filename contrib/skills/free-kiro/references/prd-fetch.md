# prd-fetch — 从 URL / issue / 浏览器抓 PRD 起 spec

`free-kiro spec new` 支持三种外部输入源，互斥：

```bash
free-kiro spec new [<name>] --from-issue <url>
free-kiro spec new [<name>] --from-prd <url>
free-kiro spec new [<name>] --from-browser <url>
```

如果不给 `<name>`，会从抓到的标题自动生成 kebab-case slug。

## `--from-issue <url>`

- URL 形式：`https://github.com/<owner>/<repo>/issues/<n>`
- 或简写：`<owner>/<repo>#<n>`
- 依赖：`gh` CLI（已认证）
- 拉 issue 的 title + body，作为 prompt；配合 `--type bugfix` 走 Bugfix Spec 变体

```bash
free-kiro spec new --from-issue https://github.com/foo/bar/issues/42
free-kiro spec new fix-login --from-issue foo/bar#42 --type bugfix
```

## `--from-prd <url>`

- 任意 HTTP(S) 的 HTML / Markdown / 纯文本
- 走 `net/http` + `golang.org/x/net/html` 抓取，**不执行 JS**
- 适用：Notion 公开页、Confluence、README、公司 wiki、Google Docs 导出版
- 1 MB body 上限、30 s 超时
- 解析时丢弃 `<script>` / `<style>` / `<nav>` / `<footer>` / `<header>` / `<aside>` / `<form>` / `<iframe>` / `<svg>` / `<noscript>`
- 标题取 `<title>`，缺失时用 URL 末段
- 输出 prompt 形如：
  ```
  # Source: <原始 URL>

  # <title>

  <正文（≤32 KB，多余截断）>
  ```

```bash
free-kiro spec new --from-prd https://example.com/our-product-prd
```

## `--from-browser <url>`（需要 bsk）

- 适用：SPA / 需要登录态 / JS 渲染后才显示的内容
- 依赖：browser-skill 的 `bsk` CLI 必须在 PATH 上（参考 [https://github.com/jingyu525/browser-skill](https://github.com/jingyu525/browser-skill)）
- 流程：
  1. 启动 `bsk session`（用户登录态浏览器）
  2. `bsk navigate <url>`
  3. `bsk get-html --out <tmpfile>`（渲染后的 HTML）
  4. 复用 `--from-prd` 的 HTML 解析流水线

### `bsk` 缺失时的错误

```
error: --from-browser requires the `bsk` CLI on PATH (from browser-skill).
       Install: see https://github.com/jingyu525/browser-skill
       Tip:    --from-prd <url> works for non-JS pages without bsk.
```

### 何时用 `--from-prd` 而非 `--from-browser`

- 静态 HTML 页面（多数 Notion 公开页、文档站、博客）
- 不需要登录
- 抓得快、不打扰用户浏览器

### 何时必须用 `--from-browser`

- 单页应用（SPA），初始 HTML 是空壳
- 需要登录态才看得到的内容（Notion 私有页、企业内网）
- URL 必须 JS 执行后才返回正确 HTML

## 三选项互斥校验

如果同时给两个或三个 `--from-*` flag，`free-kiro` 报：

```
error: --from-issue, --from-prd, and --from-browser are mutually exclusive
```

## 调试

```bash
# 单独测抓取结果（不起 spec）：
free-kiro spec new --from-prd <url> --dry-run  # 不存在该 flag 时用 -h 查

# 看错误细节：
free-kiro --verbose spec new --from-prd <url>
```