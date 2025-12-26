package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func TagDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "td",
		Aliases: []string{"tag-delete", "deltag"},
		Short:   i18n.T("tag.delete.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.DeleteAndPushTag(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
