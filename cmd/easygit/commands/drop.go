package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func DropCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "drop",
		Aliases: []string{"d"},
		Short:   i18n.T("drop.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Drop(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
