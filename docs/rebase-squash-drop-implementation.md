# easyGit Rebase/Squash/Drop 功能开发文档

本文档总结了 `easyGit` 关于 Git 历史整理相关功能（Rebase, Squash, Drop）的重构与封装成果。

## 1. 核心设计理念：意图驱动 (Intent-Driven)

传统的 Git 交互式变基（`git rebase -i`）依赖于复杂的 todo 文件编辑（通常在 Vim 中）。我们将其重构为“意图驱动”的三个独立工具，用户只需通过 TUI 列表勾选即可完成操作，无需理解底层的 `pick/squash/drop` 指令。

## 2. 功能模块说明

### 2.1 变基分支 (Rebase)
- 命令: `easyGit rebase` (别名: `r`)
- 逻辑: 直接进入分支选择列表，将当前分支变基到选定分支。
- 状态管理: 智能检测变基冲突状态。若检测到进行中的变基，会自动弹出 `Continue`、`Skip`、`Abort` 菜单接管流程。

### 2.2 合并提交 (Squash)
- 命令: `easyGit squash`
- 逻辑: 替换了原有的 `reset --soft` 粗暴实现。现在通过交互式变基实现：
  - 用户从列表中勾选连续的多个提交。
  - 多行文本表单预填合并消息：所选提交的消息按从旧到新拼接（主题 + 正文 + 段间空行），完整保留全部原文。
  - 系统自动执行 `reword` 和 `fixup` 操作，完美融合历史且不破坏后续提交。
- UI: 使用专门定制的 `TableMultiSelectForm` 展示。

### 2.3 删除提交 (Drop)
- 命令: `easyGit drop` (别名: `d`)
- 逻辑: 允许用户从历史记录中直接“剔除”某些不想要的提交。
- 交互: 在列表中勾选一个或多个提交，确认后系统自动在后台重写历史，摘除对应节点。

## 3. 技术实现要点

### 3.1 拦截器机制 (GIT_SEQUENCE_EDITOR)
- 封装了一个隐藏子命令 `_internal_rebase_editor`。
- 执行变基时，注入 `GIT_SEQUENCE_EDITOR="easyGit _internal_rebase_editor"`。
- 该拦截器读取通过环境变量 `EASYGIT_REBASE_CONFIG` 传入的 JSON 配置，自动解析并重写 `git-rebase-todo` 文件，实现非交互式的静默变基。

### 3.2 自动化消息注入 (GIT_EDITOR)
- 在 Squash 模式下，`GIT_EDITOR` 注入同一个可执行文件的隐藏子命令 `_internal_commit_message_editor`，消息经环境变量 `EASYGIT_COMMIT_MESSAGE` 传递，由该子命令整体写入 Git 传入的消息文件（`WriteCommitMessageFile`：CRLF 规范为 LF、尾部收敛为单个换行）。
- 不生成临时 shell 脚本：消息不经 `echo`/重定向解释，多行正文、空行与 `$`、`` ` ``、`&` 等字符原样落盘，Windows 与类 Unix 行为一致（旧实现的 `echo` 在 cmd 下会折叠/截断换行）。
- 注入消息时变基以 `git -c commit.cleanup=verbatim rebase -i` 执行（`rebaseArgs`）：git 默认 `commit.cleanup=strip` 会删除所有以 `#` 开头的行并压缩空行，与「保留全部原文」冲突，主题被删空时还会让提交失败。
- 隐藏子命令的报错文案统一走 i18n（`internal.editor.file.missing` / `read.failed` / `write.failed` / `config.invalid`），输出到 stderr 随 git 输出呈现。
- 可执行文件路径经 `QuoteForEditorEnv` 双引号转义，兼容含空格的 Windows 安装路径（如 `C:\Program Files\...`）。

### 3.3 TUI 组件优化 (`TableMultiSelectForm`)
- 基于 `bubbletea` 和 `bubbles/table` 定制。
- 自动布局: 适配终端高度，大视口显示历史记录。
- 多列对齐: 清晰展示 Hash、Message、Date 和 Author。
- 交互状态: 支持 `Space` 切换选中，实时显示 `(X selected)` 统计。

## 4. 文件变动清单

| 类别 | 文件路径 | 说明 |
| :--- | :--- | :--- |
| 核心逻辑 | `internal/gitcmd/rebase.go` | 变基核心引擎、状态检测、公共提交获取逻辑 |
| | `internal/gitcmd/editor_env.go` | 编辑器环境变量路径转义统一入口 `QuoteForEditorEnv`(Windows 含空格路径) |
| | `internal/gitcmd/squash.go` | 基于变基引擎重写的合并逻辑、默认合并消息拼装(`buildSquashDefaultMessage`) |
| | `internal/gitcmd/commits.go` | 提交列表统一数据源 `GetCommitsOptions` 与完整消息读取 `GetCommitMessageBodies`(单次 `git show -z`,按哈希回填顺序) |
| | `internal/gitcmd/drop.go` | 新增的删除提交逻辑 |
| UI 组件 | `internal/form/table_multi_select.go` | 专门用于 Git 提交多选的表格组件 |
| 命令封装 | `cmd/easygit/commands/rebase.go` | `rebase` 命令定义 |
| | `cmd/easygit/commands/drop.go` | `drop` 命令定义 |
| | `cmd/easygit/commands/internal_rebase_editor.go` | 变基计划自动重写工具 (隐藏) |
| | `cmd/easygit/commands/internal_commit_message_editor.go` | 合并消息写入工具 (隐藏,`GIT_EDITOR`) |
| | `internal/form/text.go` | 多行文本表单 `TextInput`(合并消息编辑) |
| | `cmd/easygit/root.go` | 注册新命令与多语言刷新 |
| 国际化 | `internal/i18n/en.go`, `zh.go` | 添加 Rebase/Squash/Drop 相关中英文文案 |

## 5. 验证方法 (测试用例)
1. 标准变基: 在冲突状态下运行 `easyGit rebase` 验证状态菜单。
2. 多选合并: 运行 `easyGit squash` 勾选中间连续的 2 个提交，验证合并后的消息和历史结构；提交带多行正文时，合并后的消息必须完整保留所有正文行与空行（`TestSquashEndToEndPreservesMultiline` 覆盖）；正文/主题含 `#` 开头的行时必须原样保留（`TestSquashEndToEndPreservesHashLines` 覆盖，依赖 `-c commit.cleanup=verbatim`）。
3. 多选删除: 运行 `easyGit drop` 勾选 3 个不连续提交，验证它们是否在 `git log` 中消失。
