package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func PushAllCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "push-all",
		Aliases: []string{"pa"},
		Short:   i18n.T("push.all.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.PushAll(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
