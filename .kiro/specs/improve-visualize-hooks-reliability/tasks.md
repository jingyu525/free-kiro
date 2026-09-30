# improve-visualize-hooks-reliability — Tasks

<!--
任务行格式（lint 强约束格式正确性）：
  - [ ] #N Title           待办
  - [x] #N Title           已完成
  - [ ] #N Title [deps: #A,#B]   依赖任务 A、B
-->

- [x] #1 lint: 给 Issue 加 Baseline bool 字段，Spec() 写位 [deps: none]
- [x] #2 lint: Gate() 改用 Issue.Baseline 判 baseline，跳过 strings.HasPrefix [deps: #1]
- [x] #3 lint: requirements/design 的 ReadFile/Stat 失败显式区分 NotExist vs 其他错误，wrap 报出 [deps: none]
- [x] #4 hooks: Registry 加 1s TTL 缓存，Match() 复用缓存，Add() 写盘后置 nil 强制刷新 [deps: none]
- [x] #5 visualize: notifier 挪进 Server struct，subscribe/unsubscribe/broadcast 加 sync.Mutex [deps: none]
- [x] #6 visualize: handleEvents 启动 watcher 用 sync.Once；Shutdown 走 http.Server.Shutdown(ctx) + 5s timeout；<-s.done 加 time.After fallback [deps: #5]
- [x] #7 tests: 补 6 个 fix 的单测（含并发/race/timeout case）+ 跑 `go test -race ./internal/...` 与 `go vet ./...` 全绿 [deps: #2,#3,#4,#6]