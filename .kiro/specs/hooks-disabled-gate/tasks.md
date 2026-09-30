# hooks-disabled-gate — Tasks

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

- [x] #1 Match() 过滤增加 Disabled gate
- [x] #2 runAgentAction() 增加 Disabled 兜底
- [x] #3 cachedAll() 改深拷贝（新增 cloneHooks helper）
- [x] #4 单元测试覆盖上述三处改动 [deps: #1,#2,#3]
- [x] #5 跑 go test -race ./internal/hooks/... 和 free-kiro lint hooks-disabled-gate [deps: #4]