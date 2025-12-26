package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func BranchDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bd",
		Aliases: []string{"branch-delete", "delbranch"},
		Short:   i18n.T("branch.delete.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.DeleteBranch(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
