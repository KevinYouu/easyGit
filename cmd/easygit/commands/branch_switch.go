package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func BranchSwitchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "branch-switch",
		Aliases: []string{"sw"},
		Short:   i18n.T("branch.switch.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.SwitchBranch(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}

func BranchCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "branch-create",
		Aliases: []string{"bc"},
		Short:   i18n.T("branch.create.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.CreateBranch(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
