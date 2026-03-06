package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func SquashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "squash",
		Aliases: []string{"sq"},
		Short:   i18n.T("squash.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Squash(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
