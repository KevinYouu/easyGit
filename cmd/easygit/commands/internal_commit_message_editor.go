package commands

import (
	"fmt"
	"os"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

// InternalCommitMessageEditorCommand 隐藏子命令,作为 GIT_EDITOR 由 git 调用:
// 将 EASYGIT_COMMIT_MESSAGE 原样写入 git 传入的提交信息文件。
// 经可执行文件转发而非临时 shell 脚本:提交信息可含换行与任意特殊字符,
// 不经过 shell 解释,Windows(cmd)与类 Unix 行为一致,换行不会被折叠或截断。
func InternalCommitMessageEditorCommand() *cobra.Command {
	return &cobra.Command{
		Use:    "_internal_commit_message_editor <file>",
		Short:  "Internal command used as GIT_EDITOR",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Fprintln(os.Stderr, i18n.T("internal.editor.file.missing"))
				os.Exit(1)
			}

			if err := gitcmd.WriteCommitMessageFile(args[0], os.Getenv("EASYGIT_COMMIT_MESSAGE")); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", i18n.T("internal.editor.write.failed"), err)
				os.Exit(1)
			}
		},
	}
}
