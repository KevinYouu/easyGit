package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func WorktreeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "worktree",
		Aliases: []string{"wk"},
		Short:   i18n.T("worktree.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Worktree(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
