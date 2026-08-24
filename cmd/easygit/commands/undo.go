package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func UndoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "undo",
		Aliases: []string{"un"},
		Short:   i18n.T("undo.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.Undo(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
