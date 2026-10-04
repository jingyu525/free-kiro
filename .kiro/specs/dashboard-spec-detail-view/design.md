# dashboard-spec-detail-view — Design

<!--
决策记录（30 天后回看"为什么这么做"）：
  - hash + query string（?tab=#1）而不是 path：浏览器后退可精确回退到上一个 tab；URL 仍然 # base 不改 server routing
  - Sparkline 数据从 timeline 端点聚合：复用现有端点，不新增后端；timeline 缺 key 维度 → 显示占位（AC-12 优雅降级）
  - 跨 spec 链接沿用 hash 路由（同 dialog 复用，跳转 = 清当前 hash 再 set 新 hash + tab）
  - Roving tabindex：WAI-ARIA 标准 tab pattern，键盘导航与原生 <dialog> 焦点管理不冲突
  - Sparkline 用 viewBox 60×16 像素：8x 缩放下仍清晰，CSS 控制大小
-->

## Architecture

### Hash 路由 schema

```
URL:                  #/spec/<name>?tab=<tab>

例:
  #/spec/dashboard-frontend-components        # 默认 tab=overview
  #/spec/dashboard-frontend-components?tab=drift
  #/spec/foo?tab=tasks

useHashRoute() 返回:
  { name: string; tab: 'overview' | 'drift' | 'tasks' | 'timeline' }

非法处理:
  - 不含 #/spec/<name>       → { name: '', tab: 'overview' } (dialog 关闭)
  - tab 不在 4 个枚举中     → 降级为 'overview'
  - name 含非法字符 (URL escape) → URLDecoder 解码
```

### Dialog 状态机（升级版）

```
States:
  - 'closed': hash 不匹配 spec/<name>
  - 'opening': hash 变化 → spec 存在 → 准备 dialog content
  - 'loading': useSpec*Query.isLoading
  - 'ready':  query 完成 → 渲染 4 tab + selected tab panel
  - 'error':  query 失败 → RetryBanner
  - 'closing': Esc / close / click outside → dialog close + 焦点归还

Transitions:
  - closed → opening:    hash 变 #/spec/<name>?tab=<t>
  - opening → loading:   dialog.showModal() 触发
  - loading → ready:     query 完成
  - any → closing:       Esc / Close / backdrop / hash 清空
  - closing → closed:    dialog.close() + 焦点归还 trigger

URL 同步:
  open  → navigateToHash(`#/spec/${name}?tab=${tab}`)
  close → navigateToHash('')  // (保留 dialog name 触发关闭但不保存)
  tab 切换 → 只更新 ?tab= 参数，不动 name
  dep 跨 → navigateToHash(`#/spec/${depName}?tab=tasks#task-${id}`)
```

### Sparkline 数据流

```
┌──────────────────────────────────────────────────────────┐
│ 后端 /api/spec/<name>/timeline 返回:                       │
│   Array<{ ts: string; kind: string; message: string }>    │
│   (dashboard-backend-api-extensions 当前实现)             │
└──────────────────────────────────────────────────────────┘
                       ↓
        ┌──────────────────────────────────────┐
        │ entities/spec/sparkline.ts         │
        │  useSparklineData(name, key)       │
        │  → 把 message 按 <key>=<value> 正则│
        │    提取 N 个历史点                │
        │  → 返回 { samples: number[], latest│
        │           delta: number }         │
        └──────────────────────────────────────┘
                       ↓
        ┌──────────────────────────────────────┐
        │ widgets/sparkline-cell/index.tsx   │
        │  <svg viewBox="0 0 60 16">          │
        │    <path d="M0,8 L10,5 L20,7..."/>│
        │    <circle r="2" cx="60" cy="...">│
        │  </svg>                            │
        └──────────────────────────────────────┘
```

**关键：timeline message 格式约定** — 后端 timeline 端点返回 `message: string`，不带 key 维度。本 spec 不要求后端改，而是用"宽松解析"：
- 正则匹配 `"<key>=<number>"` 提取
- 提取不到 → AC-12 占位
- 这是 forward-compatible 妥协：若后端未来加 `key` 字段，前端自动适配

### 跨 spec 跳转状态机

```
用户点 dep 链接:
  1. 浏览器解析 href → hashchange 事件
  2. spec-detail-dialog 的 useEffect 触发 → isOpen 变 true (新 spec)
  3. dialog content 重 mount（react-query queryKey 变）
  4. 焦点归还到对应 task (按 name anchor `#task-<id>`)
  5. useEffect 触发 scrollIntoView({ block: 'nearest' })
  
回退到原 spec:
  - history.back() → hash 变回 #/spec/<original>?tab=旧
  - dialog useEffect 同上 2-4 流程
```

### Roving tabindex

```tsx
// WAI-ARIA Tabs Pattern
// https://www.w3.org/WAI/ARIA/apg/patterns/tabs/

<div role="tablist" aria-label="Spec detail tabs">
  {TABS.map((t, i) => (
    <button
      role="tab"
      aria-selected={tab === t}
      aria-controls={`panel-${t}`}
      tabIndex={tab === t ? 0 : -1}              // 当前 tab 可 Tab focus
      onClick={() => setTab(t)}
      onKeyDown={(e) => {
        if (e.key === 'ArrowRight') focusNext();
        if (e.key === 'ArrowLeft')  focusPrev();
        if (e.key === 'Home')       focusFirst();
        if (e.key === 'End')        focusLast();
      }}
    >{t}</button>
  ))}
</div>
```

### Focus flow 完整图

```
打开 dialog:
  hash 变 → useEffect → dialog.showModal()
  → useEffect 找第一个 focusable: tab[0] (overview)
  → 聚焦 tab[0]
  → announce(`Spec ${name} details, tab ${tab}`)

Tab 切换 (ArrowRight):
  当前 tab → 下一个 tab
  setTab(next) → aria-selected 改
  焦点 → 下一个 tab button

点击 dep 链接:
  hash 变 #/spec/<dep>?tab=tasks#task-<id>
  → dialog content 重 mount
  → useEffect: query 完成 → querySelector(`#task-${id}`) → focus
  → scrollIntoView

关闭 dialog:
  Esc / Close / backdrop / 外部 navigateToHash('')
  → dialog.close()
  → trigger.focus()
  → announce('Spec details closed')
```

## Components

### 新建 / 修改

| Module | 改动 |
|---|---|
| `shared/lib/hash-router.ts` | `useHashRoute()` 返回 `{name, tab}`；保留 `HashRoute` type 给 legacy |
| `entities/spec/sparkline.ts` | 新建：`useSparklineData(name, key)` hook |
| `widgets/spec-detail-dialog/index.tsx` | 修改：标题栏加 prev/next/copy 3 button + tab roving tabindex + URL 同步 |
| `widgets/task-list/index.tsx` | 修改：deps 渲染为 `<a>` 链接 |
| `widgets/drift-table/index.tsx` | 修改：每行末加 `<td>` 含 `<SparklineCell>` |
| `widgets/sparkline-cell/index.tsx` | 新建：`<SparklineCell name key />` |
| `widgets/spec-detail-dialog/components/index.tsx` | 新建：`<DialogToolbar>` 包裹 prev/next/copy/close |
| `shared/styles/components.css` | 加 `.sparkline` / `.dialog-toolbar` 样式 + `@media (prefers-reduced-motion)` |

### 数据契约

`useHashRoute()` 返回类型：

```ts
export type TabId = 'overview' | 'drift' | 'tasks' | 'timeline';
export interface HashRouteState {
  name: string;
  tab: TabId;
}
export function useHashRoute(): HashRouteState;
```

`useSparklineData()` 返回类型：

```ts
export interface SparklineData {
  samples: number[];
  latest: number;
  delta: number;     // latest - samples[0] (undefined → 0)
}
export function useSparklineData(name: string, key: string): SparklineData;
```

## Data Model

无 Go 端 schema 变更。复用现有 `/api/spec/<name>/timeline` 返回 `{ts, kind, message}`。

如果后端未来在 timeline 事件加 `key` 字段，前端 `useSparklineData` 可平滑扩展（regex 解析 message 退化为读取结构化字段）。

## Error Handling

| 错误 | 来源 | UI |
|---|---|---|
| 非法 tab 值 | URL hash | `useHashRoute` 降级为 `overview` + `announce('Unknown tab, defaulting to overview')` |
| clipboard 失败 | `navigator.clipboard.writeText` 抛错（HTTP/无权限） | `announce('Copy failed, link: ' + location.href, 'assertive')` + 显示 fallback tooltip |
| timeline 无 key | 后端未返 key 数据 | `<span aria-label="no history">—</span>` 占位 |
| 历史点 < 2 个 | 不够画折线 | `<svg>` 只画当前点一个 circle |
| 跨 spec 跳转时新 spec 详情加载失败 | useSpec*Query 失败 | RetryBanner 显示，新 tab 内可点 retry |

- **零吞错误**（.kiro/steering/agent-rules.md §2）：所有异常显式处理
- **错误分类**：announce('assertive') 用于阻断用户（clipboard fail），announce('polite') 用于信息通知

## Testing Strategy

### 类型检查（CI）

```bash
pnpm exec tsc --noEmit                                    # 0 error
pnpm run lint:fsd                                         # ✓ FSD layers OK
pnpm exec vite build --mode production                    # build OK
```

### E2E（升级 e2e-dashboard.mjs）

加场景：
- 7) GET /api/spec/<name>/timeline 返回非空数组（已有，但 sparkline 解析需 message 格式约定）
- 8) deep-link `#/spec/<name>?tab=drift` 渲染 drift tab 内容
- 9) `<svg.sparkline>` 元素存在（即使占位）

### Playwright 浏览器验证（手工）

```js
goto http://127.0.0.1:7380/#/spec/dashboard-frontend-components?tab=drift
expect drift panel 内容
click prev button → #/spec/add-status-subcommand?tab=overview
press ArrowRight × 3 → 焦点在 timeline tab
click copy link → clipboard 含 URL
```

### Bundle 预算（CI）

```bash
gzip -c dist/assets/index-<hash>.js | wc -c   # 入口 ≤ 30 KB (新增 sparkline + toolbar + roving tabindex)
gzip -c dist/assets/index-<hash>.css | wc -c  # ≤ 16 KB (sparkline + toolbar 样式)
```

## Migration / Rollout

### 落地步骤

```
1. shared/lib/hash-router.ts 扩展支持 ?tab= query（向后兼容旧 HashRoute）
2. entities/spec/sparkline.ts 新建 useSparklineData
3. widgets/sparkline-cell/index.tsx 新建
4. widgets/spec-detail-dialog/index.tsx 修改：
   - 加 prev/next/copy 3 button + DialogToolbar 包裹
   - tab 切换同步 URL hash
   - roving tabindex
5. widgets/task-list/index.tsx：<code> ↔ <a href="#/spec/<dep>?tab=tasks#task-<id>">
6. widgets/drift-table/index.tsx：每行末加 <td><SparklineCell/></td>
7. shared/styles/components.css：加 .sparkline / .dialog-toolbar 样式
8. 跑 tsc + lint:fsd + build
9. 启 serve + e2e + 浏览器 Playwright 验证
10. commit + spec complete
```

### 兼容性边界

| 项 | 兼容策略 |
|---|---|
| 旧 hash 格式 `#/spec/<name>`（无 `?tab=`） | 默认 `overview` tab，行为不变 |
| 后端 timeline message 格式变化 | useSparklineData 正则走兼容，未匹配 / 占位 |
| clipboard API 不可用（HTTPS only） | fallback `prompt()` + announce('assertive') |
| `<svg>` 渲染 浏览器 | 锁 ≥ Chrome 100 / Safari 15 / Firefox 100，与现有 baseline 一致 |

### 风险与回滚

- **风险 A**：roving tabindex 与现有 React 18 focus management 冲突。**对策**：roving tabindex 只在 tab `<button>` 子树内；dialog 整体焦点管理（开/关）保留 useEffect 实现。
- **风险 B**：cross-spec link 用 `<a>` 但 hashchange 触发 dialog 重 mount 丢滚动位置。**对策**：scrollIntoView({ block: 'nearest' }) 在 dialog content mount 后执行。
- **风险 C**：Sparkline N=60 个点性能问题。**对策**：最多保留最近 60 个 sample（`slice(-60)`）；`requestIdleCallback` 延迟渲染（fallback `setTimeout`）。
- **风险 D**：timeline 端点无 key 数据，Sparkline 全空显示 `—`。**对策**：AC-12 明示占位语义，UI 不破。
- **风险 E**：clipboard `navigator.clipboard` 在 HTTP（非 localhost）下被禁用。**对策**：try/catch 后 `announce('Copy failed, link: ' + location.href, 'assertive')` 让用户手动复制。

### 落地前置

- 三件套全绿 `free-kiro lint dashboard-spec-detail-view`
- 用户显式 `free-kiro spec approve dashboard-spec-detail-view` 后再 `spec start`
- 实施期改 AC 计数 → `free-kiro spec sync dashboard-spec-detail-view` 重 baseline