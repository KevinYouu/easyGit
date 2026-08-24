package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func StashCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stash",
		Aliases: []string{"st"},
		Short:   i18n.T("stash.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Stash(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
