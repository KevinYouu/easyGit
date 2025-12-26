package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/update"
	"github.com/spf13/cobra"
)

func UpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: i18n.T("update.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := update.UpdateSelf(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
