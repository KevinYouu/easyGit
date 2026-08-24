package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func AmendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "amend",
		Aliases: []string{"am"},
		Short:   i18n.T("amend.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Amend(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
