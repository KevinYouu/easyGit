package main

import (
	"fmt"
	"os/exec"

	"github.com/KevinYouu/easyGit/cmd/easygit/commands"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "easyGit",
	Short: "", // Will be set dynamically
	Run: func(cmd *cobra.Command, args []string) {
		// 无参数:进入交互式主菜单(新手零记忆成本)
		commands.RunMenu()
	},
}

// nonGitCommands 允许在非 git 仓库目录运行的命令:配置类/工具类命令
// 不依赖仓库上下文,统一排除;其余命令在 root 层统一检查 git 仓库,
// 避免各命令重复检查与分散报错。
var nonGitCommands = map[string]bool{
	"init":                    true,
	"version":                 true,
	"update":                  true,
	"config":                  true,
	"_internal_rebase_editor": true, // 变基期间由 git 调用,必然在仓库内
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
	// 非 git 仓库统一检查(子命令继承;root 自身无参数显示版本不检查)
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if cmd.Parent() == nil || nonGitCommands[cmd.Name()] {
			return nil
		}
		if err := exec.Command("git", "rev-parse", "--git-dir").Run(); err != nil {
			return fmt.Errorf("%s", i18n.T("error.not.git.repo"))
		}
		return nil
	}

	// Add language flag support
	i18n.AddLanguageFlag(rootCmd)
	// Add theme flag support (runtime priority, pre-parsed in main.go)
	rootCmd.PersistentFlags().String("theme", "", "Set theme (auto/dark/light)")

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
