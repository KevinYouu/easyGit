# Stash 管理(st)

## 功能说明

`stash (st)`:交互式 stash 管理。主菜单三选:

1. **保存修改到 stash** - 可选消息(`git stash push -m`);无修改时提示跳过
2. **浏览并管理条目** - 列表展示消息/引用/相对时间,选中后可查看 diff、应用(保留条目)、应用并删除(pop)、删除单条;diff 与删除操作后回到列表继续
3. **清空全部** - 显示条目数量,强确认后 `git stash clear`

## 设计理念

### 问题

- ❌ `stash list/pop/apply/drop/show` 全靠记编号(`stash@{N}`),编号随操作动态变化
- ❌ `pop` 冲突后条目保留的行为反直觉,新手常误以为丢失
- ❌ `clear` 一把梭无确认,误触即全部丢失

### 优化方案

- ✅ **可视化列表** - 消息为主列,`stash@{N} · 相对时间` 为说明列,`/` 过滤可用
- ✅ **先看再动** - diff 预览不改变状态,确认内容后再决定应用方式
- ✅ **语义区分** - apply(保留条目)/ pop(删除条目)选项说明写明差异
- ✅ **冲突提示** - 应用冲突时明确告知 pop 条目被 git 自动保留,不会丢失
- ✅ **破坏性防护** - drop 单条与 clear 全部均需确认,clear 前展示条目数

## 实现位置

| 职责 | 文件 |
| --- | --- |
| 核心与交互流程 | `internal/gitcmd/stash.go`(listStashes / saveStash / applyStash / dropStash / clearStashes / manageStashes / Stash)|
| 可空输入组件 | `internal/form/input.go`(InputOptional,stash 可选消息复用)|
| 命令注册 | `cmd/easygit/commands/stash.go`(StashCommand)+ root.go + menu.go |
| i18n | `internal/i18n/en.go` / `zh.go`(`stash.*`)|

## 测试

真实 git 仓库集成回归:

- 引用解析:`parseStashRefNumber` 表驱动(含非法格式)
- 列表解析:多条目顺序(最新在前)、消息/时间字段
- 保存:干净仓库跳过、带消息入栈后工作区干净且消息可检索
- apply/pop 行为差异:条目保留 vs 删除
- 冲突场景:pop 冲突时条目由 git 自动保留 + 返回错误
- drop/clear/show:状态变化断言
- 流程注入:主菜单保存(注入消息)/ 管理→pop(Esc 路径)/ 清空确认通过与取消/空列表清空

## 注意事项

- stash 编号在每次增删后都会变化,列表每次循环重新获取,避免引用过期
