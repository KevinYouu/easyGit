package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/spf13/cobra"
)

// configListSpecs 配置中心列表的列定义:名称列自动宽度(上限引用
// form.MaxAutoColumnWidth,单一事实来源),说明/摘要列弹性占满剩余宽度。
// 列数不硬编码,后续新增列(如使用次数、时间)追加 spec 即可。
func configListSpecs() []form.ColumnSpec {
	return []form.ColumnSpec{
		{Kind: form.ColumnAuto, MaxWidth: form.MaxAutoColumnWidth},
		{Kind: form.ColumnFlex},
	}
}

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
		selected, err := form.ListFormColumns(i18n.T("config.select.title"), configListSpecs(), options, form.ListSingle)
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
