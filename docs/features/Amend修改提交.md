# Amend 修改上次提交(am)

## 功能说明

`amend (am)`:交互式修改上次提交,三种模式单选:

1. **仅修改消息** - 预填原消息,支持 ↑/↓ 复用历史消息;仅校验非空(不强加类型前缀约定)
2. **追加暂存文件** - 多选工作区文件(M/A/D 状态列),`git add` 后 `commit --amend --no-edit`,消息保持不变
3. **两者都做** - 先选文件再改消息

安全机制:

- HEAD 已推送时先警告"将改写历史,需要 force push 覆盖",确认后才继续
- amend 完成后提供 `push --force-with-lease` 安全覆盖询问(仅改写前已推送时)
- 无上游分支时提示跳过自动强推

## 设计理念

### 问题

- ❌ 改消息要进编辑器(`commit --amend` 默认弹 vim)
- ❌ 补文件要两步:`add` 后再 `--amend --no-edit`,忘加 `--no-edit` 会误开编辑器
- ❌ 已推送的提交被 amend 后,忘记强推导致本地远程分叉,新手无从下手
- ❌ `--force` 裸强推会覆盖他人推送,存在覆盖风险

### 优化方案

- ✅ 表单输入替代编辑器,预填原消息减少重打成本
- ✅ 文件多选 + 状态列复用 ps 交互,一步完成追加
- ✅ 改写历史前置警告 + 完成后主动引导强推,闭环防分叉
- ✅ 强推统一使用 `--force-with-lease`(远程有新提交时拒绝覆盖)

## 实现位置

| 职责 | 文件 |
| --- | --- |
| 核心与交互流程 | `internal/gitcmd/amend.go`(getHeadSubject / isHeadPushed / executeAmend / forcePushAmended / Amend)|
| 命令注册 | `cmd/easygit/commands/amend.go`(AmendCommand)+ root.go + menu.go |
| i18n | `internal/i18n/en.go` / `zh.go`(`amend.*`)|

## 测试

真实 git 仓库集成回归:

- isHeadPushed 三态:无远程 / 已推送 / 推送后新提交
- getUpstreamRef:无上游空串 / `-u` 推送后 origin/main
- executeAmend:仅改消息(hash 变化、提交数不变)/ 追加文件(消息不变、内容并入、工作区干净)/ 双改 / 空仓库报错
- 流程注入:仅改消息(预填断言 + 警告放行 + 拒绝强推后远程未变)/ 警告取消(HEAD 不变、不进入输入)/ 文件追加 / Esc 取消

## 注意事项

- 判定"是否需要强推"以**改写前** HEAD 是否在远程为准;amend 后的新提交必然不在远程,事后检查恒为 false
- `executeAmend(nil, msg)` 不执行 add 步骤(`git add` 无路径参数会静默跳过)
