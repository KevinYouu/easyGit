package commands

import (
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/version"
	"github.com/spf13/cobra"
)

func VersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   i18n.T("version.short"),
		Run: func(cmd *cobra.Command, args []string) {
			version.GetVersion()
		},
	}
	return cmd
}
