package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func ResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reset",
		Aliases: []string{"rs"},
		Short:   i18n.T("reset.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Reset(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
