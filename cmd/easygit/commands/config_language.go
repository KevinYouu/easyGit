package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// configLanguage 界面语言子流程:单选 en/zh,保存后立即应用,
// 返回配置中心主列表时按新语言重建(即时生效)。
func configLanguage() {
	// 获取当前语言设置,预选对应选项
	currentLang := i18n.GetCurrentLanguage()
	var preselected string
	if currentLang == i18n.LangZH {
		preselected = "zh"
	} else {
		preselected = "en"
	}

	// 创建语言选项(SelectForm 统一组装为「名称 + 单行说明」)
	options := []config.Option{
		{
			Label:       i18n.T("language.option.en"),
			Description: i18n.T("language.option.en.desc"),
			Value:       "en",
		},
		{
			Label:       i18n.T("language.option.zh"),
			Description: i18n.T("language.option.zh.desc"),
			Value:       "zh",
		},
	}

	selectedLangs, err := form.ListFormColumns(i18n.T("language.select.title"), configListSpecs(), options, form.ListSingle, preselected)
	if err != nil {
		logs.Error(i18n.T("language.set.error") + ": " + err.Error())
		return
	}
	selectedLang := selectedLangs[0]

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
