package main

import (
	"github.com/KevinYouu/easyGit/cmd/easygit/commands"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "easyGit",
	Short: "", // Will be set dynamically
	Run: func(cmd *cobra.Command, args []string) {
		commands.VersionCommand().Run(cmd, args)
	},
}

// updateRootCommandDescriptions updates all command descriptions based on current language
func updateRootCommandDescriptions() {
	rootCmd.Short = i18n.T("root.short")

	// Update all subcommands
	for _, cmd := range rootCmd.Commands() {
		switch cmd.Use {
		case "version":
			cmd.Short = i18n.T("version.short")
		case "push-all":
			cmd.Short = i18n.T("push.all.short")
		case "push-selected":
			cmd.Short = i18n.T("push.selected.short")
		case "reset":
			cmd.Short = i18n.T("reset.short")
		case "tag-create":
			cmd.Short = i18n.T("tag.short")
		case "tag-delete":
			cmd.Short = i18n.T("tag.delete.short")
		case "branch-delete":
			cmd.Short = i18n.T("branch.delete.short")
		case "merge":
			cmd.Short = i18n.T("merge.short")
		case "cherry-pick":
			cmd.Short = i18n.T("cherry.pick.short")
		case "update":
			cmd.Short = i18n.T("update.short")
		case "init":
			cmd.Short = i18n.T("init.short")
		case "config":
			cmd.Short = i18n.T("config.short")
		case "rebase":
			cmd.Short = i18n.T("rebase.short")
		case "drop":
			cmd.Short = i18n.T("drop.short")
		}
	}
}

func init() {
	// Add language flag support
	i18n.AddLanguageFlag(rootCmd)

	rootCmd.AddCommand(
		commands.PushAllCommand(),
		commands.PushSelectedCommand(),
		commands.ResetCommand(),
		commands.SquashCommand(),
		commands.TagCommand(),
		commands.TagDeleteCommand(),
		commands.BranchDeleteCommand(),
		commands.MergeCommand(),
		commands.CherryPickCommand(),
		commands.VersionCommand(),
		commands.UpdateCommand(),
		commands.InitCommand(),
		commands.ConfigCommand(),
		commands.RebaseCommand(),
		commands.DropCommand(),
		commands.InternalRebaseEditorCommand(),
	)
}
