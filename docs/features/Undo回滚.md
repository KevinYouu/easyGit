# Undo 回滚(un / reflog 后悔药)

## 功能说明

`undo (un)`:基于 reflog 的交互式回滚。列出最近 50 个检查点(哈希/动作描述/时间/`HEAD@{N}` 标记),选中后选择重置模式(与 `reset (rs)` 共用模式选项与跨会话记忆),确认后恢复。

## 设计理念

### 问题

- ❌ `git reflog` 找哈希 + 手敲 `git reset --hard <hash>` 是新手最不会的"后悔药"组合
- ❌ reflog 动作描述(commit:/checkout:/rebase 等)对新手晦涩,但恰恰是定位"回到哪之前"的关键信息
- ❌ `--hard` 一把梭无警告

### 优化方案

- ✅ **可视化检查点列表** - 短哈希 + 原始动作描述 + 时间 + `HEAD@{N}`,`/` 过滤可用
- ✅ **模式复用** - soft/mixed/hard 选项与说明、上次选择记忆全部复用 reset 实现,心智一致
- ✅ **hard 强提示** - 确认框内联警告:丢弃检查点之后提交并覆盖本地修改且不可恢复
- ✅ **执行后记忆** - 模式写回同一 LastChoiceResetMode 键,rs/un 互通

## 实现位置

| 职责 | 文件 |
| --- | --- |
| 列表与核心执行 | `internal/gitcmd/undo.go`(listReflog / executeUndo / Undo)|
| 模式选项复用 | `internal/gitcmd/reset.go`(resetModeOptions / getModeDescription)|
| 命令注册 | `cmd/easygit/commands/undo.go`(UndoCommand)+ root.go + menu.go |
| i18n | `internal/i18n/en.go` / `zh.go`(`undo.*`)|

## 测试

真实 git 仓库集成回归:

- 列表解析:记录数、最新在前(HEAD@{0})、动作描述含提交消息、短哈希与 HEAD~1 一致
- limit 生效:5 条提交仅返回 3 条
- 三种模式落地断言:hard(工作区干净+内容回退)/ soft(改动保留在暂存区)/ 默认 mixed(改动保留在工作区)
- 流程注入:选检查点 → hard → 确认通过回滚生效;取消时 HEAD 不变;Esc 直接退出

## 注意事项

- 回滚目标以 reflog 记录的短哈希为准(reset 接受短哈希);checkout 类条目同样可作恢复点
- 与 `rs` 共用重置模式选项函数与记忆键,保证两处行为同步演进
