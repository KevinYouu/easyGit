package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func CleanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "clean",
		Aliases: []string{"cl"},
		Short:   i18n.T("clean.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Clean(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
