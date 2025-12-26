package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func StatusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Aliases: []string{"s"},
		Short:   i18n.T("status.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Status(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
