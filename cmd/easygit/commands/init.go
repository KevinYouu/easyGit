package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

func InitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: i18n.T("init.short"),
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Initialize(); err != nil {
				fmt.Println(err)
			}
		},
	}
	return cmd
}
