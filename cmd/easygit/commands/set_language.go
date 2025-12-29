package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/spf13/cobra"
)

var SetLanguageCmd = &cobra.Command{
	Use:   "set-language",
	Short: i18n.T("language.set.short"),
	Long:  "Set the default language for easyGit (en/zh)",
	Run: func(cmd *cobra.Command, args []string) {
		setLanguageInteractive()
	},
}

func setLanguageInteractive() {
	// 获取当前语言设置
	currentLang := i18n.GetCurrentLanguage()
	var preselected string
	if currentLang == i18n.LangZH {
		preselected = "zh"
	} else {
		preselected = "en"
	}

	// 创建语言选项
	options := []config.Option{
		{
			Label: i18n.T("language.option.en"),
			Value: "en",
		},
		{
			Label: i18n.T("language.option.zh"),
			Value: "zh",
		},
	}

	// 使用 TUI 选择语言
	_, selectedLang, err := form.SelectForm(i18n.T("language.select.title"), options, preselected)
	if err != nil {
		logs.Error(i18n.T("language.set.error") + ": " + err.Error())
		return
	}

	// 保存语言设置
	if err := config.SaveLanguage(selectedLang); err != nil {
		logs.Error(i18n.T("language.set.error") + ": " + err.Error())
		return
	}

	// 立即应用语言设置
	if selectedLang == "zh" {
		i18n.SetLanguage(i18n.LangZH)
	} else {
		i18n.SetLanguage(i18n.LangEN)
	}

	logs.Success(i18n.T("language.set.success"))
}
