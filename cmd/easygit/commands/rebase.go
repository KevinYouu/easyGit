package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func RebaseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rebase",
		Aliases: []string{"r"},
		Short:   i18n.T("rebase.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.RebaseIntoCurrent(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
