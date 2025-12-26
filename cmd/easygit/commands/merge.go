package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func MergeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "merge",
		Aliases: []string{"m"},
		Short:   i18n.T("merge.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := gitcmd.MergeIntoCurrent(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
