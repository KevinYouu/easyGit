package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func PushSelectedCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "push-selected",
		Aliases: []string{"ps"},
		Short:   i18n.T("push.selected.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.PushSelected(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
