# Worktree 管理(wk)

## 功能说明

`worktree (wk)`:交互式工作树管理,主菜单循环三选:

1. **查看工作树** - 列出全部关联工作树(路径 + 分支,主树/裸仓标注)
2. **添加工作树** - 输入路径 → 选择检出方式:
   - 新建分支:`git worktree add -b <branch> <path>`(基于当前 HEAD,名称走 refname 校验)
   - 已有分支:仅列出未被其他工作树占用的分支(git 不允许同一分支双检出)
3. **删除工作树** - 多选(主工作树与裸仓自动排除)→ 确认 → 顺序移除

## 设计理念

### 问题

- ❌ `worktree add/list/remove` 记不住路径规则与参数顺序
- ❌ 同一分支被占用时 `worktree add` 直接报错,新手不知原因
- ❌ 删除时容易误删主工作树

### 优化方案

- ✅ **可视化列表** - porcelain 解析,主树标注 `(main)`/`(主)`,路径+分支一目了然
- ✅ **占用感知** - 添加时过滤已被占用的分支,从源头避免 git 报错
- ✅ **新分支模式** - `-b` 参数表单化,名称校验复用 branch-create 的 validateBranchName
- ✅ **删除防护** - 主工作树/裸仓不可选;确认框说明目录将从磁盘删除
- ✅ **顺序执行** - worktree 共享仓库元数据,不做并行删除

## 实现位置

| 职责 | 文件 |
| --- | --- |
| 列表解析与核心执行 | `internal/gitcmd/worktree.go`(listWorktrees / worktreesInUse / worktreeAdd / worktreeRemovePaths / Worktree)|
| 命令注册 | `cmd/easygit/commands/worktree.go`(WorktreeCommand)+ root.go + menu.go |
| i18n | `internal/i18n/en.go` / `zh.go`(`worktree.*`)|

## 测试

真实 git 仓库集成回归:

- porcelain 解析:主树标记/路径/短哈希(7 位)/分支名;新增工作树后两条记录
- 占用集合:in-use 分支判定准确
- 添加:已有分支检出(工作树分支正确、主仓不受影响)/ `-b` 新建分支并可见 / 非空已存在路径报错且主仓不动
- 移除:多删后目录消失、分支占用解除、主树保留
- 流程注入:add 新分支全流程(含校验)/ remove 多选确认 / list 后回菜单再 Esc 退出

## 注意事项

- macOS 下 `/var` 为 `/private/var` 符号链接,porcelain 输出真实路径;测试对比前需 EvalSymlinks 归一
- `worktree remove` 要求工作树干净(无未提交修改);脏树由 git 报错原样上抛,不静默强删
