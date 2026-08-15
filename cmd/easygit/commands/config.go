package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: i18n.T("config.short"),
	Run: func(cmd *cobra.Command, args []string) {
		runConfigCenter()
	},
}

// ConfigCommand 返回 config 命令(配置中心:所有可配置项的统一 TUI 入口)
func ConfigCommand() *cobra.Command {
	return configCmd
}

// runConfigCenter 配置中心主循环:列出全部可配置项(含当前值摘要),
// 选中某项进入对应子流程,完成后返回主列表;Esc 退出。
// 语言切换后下一轮循环按新语言重建列表,即时生效。
func runConfigCenter() {
	for {
		options := config.BuildConfigOptions()
		selected, err := form.ListForm(i18n.T("config.select.title"), options, form.ListSingle)
		if err != nil {
			// Esc / 取消:退出配置中心
			return
		}

		switch selected[0] {
		case config.ConfigKeyLanguage:
			configLanguage()
		case config.ConfigKeyPush:
			configPush()
		case config.ConfigKeyCommitTypes:
			configCommitTypes()
		case config.ConfigKeyTagPatch:
			configTagPatch()
		}
	}
}
