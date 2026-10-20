# TODO

> 已完成的变更不再在此累积，见 git 提交历史与 `docs/features/`。本文件只保留待办，
> 缺口分析（证据与方案）见 [`docs/features/命令覆盖缺口分析.md`](docs/features/命令覆盖缺口分析.md)。

## P0 同步与推送

- [ ] `sync`（或 `pull`）：当前分支 `fetch --prune` → `pull --ff-only`；分叉时提供 `rebase`/`merge` 选项并复用冲突闭环
- [ ] `push`（不提交）：列出 `@{u}..HEAD` 待推送提交 → 选 remote（复用 `SelectAndSaveRemotes`）→ 推送；无 upstream 时走 `push -u`
- [ ] 修 pa/ps 硬伤：`git add -A` 后无改动时跳过 commit 步骤继续 pull/push（进度屏步骤动态生成），当前干净工作区必然中断

## P1 回滚与消息编辑

- [ ] `revert`：复用 `GetCommitsOptions` 多选列表 + `runConflictResolution` 冲突闭环，默认消息 `revert: <原主题>` 且可用 `form.TextInput` 编辑
- [ ] pa/ps/am 支持多行提交信息：单行主题（含历史建议）后可选「补充正文」多行表单，复用 `form.TextInput`

## P2 查看与远端

- [ ] `status`：文件列表 + 选中文件预览 `git diff` / `git diff --cached`（只读）
- [ ] `remote` 管理：列表 + add / set-url / remove（写操作走 `form.Confirm`）

## P3 收尾

- [ ] `branch` 改上游：`git branch --set-upstream-to`，复用 `GetRemoteBranches`
- [ ] `tag push`：推送全部本地缺失标签 + 标签列表
