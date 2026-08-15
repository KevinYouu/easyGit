package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// configTheme 界面主题子流程:单选 auto/dark/light,保存后立即应用,
// 返回配置中心主列表时新主题已生效(auto 会重新探测终端背景)。
func configTheme() {
	// 获取当前主题设置,预选对应选项(未设置按自动)
	current, err := config.GetTheme()
	if err != nil || current == "" {
		current = config.ThemeAuto
	}

	// 创建主题选项(SelectForm 统一组装为「名称 + 单行说明」)
	options := []config.Option{
		{
			Label:       i18n.T("theme.option.auto"),
			Description: i18n.T("theme.option.auto.desc"),
			Value:       config.ThemeAuto,
		},
		{
			Label:       i18n.T("theme.option.dark"),
			Description: i18n.T("theme.option.dark.desc"),
			Value:       config.ThemeDark,
		},
		{
			Label:       i18n.T("theme.option.light"),
			Description: i18n.T("theme.option.light.desc"),
			Value:       config.ThemeLight,
		},
	}

	selected, err := form.ListFormColumns(i18n.T("theme.select.title"), configListSpecs(), options, form.ListSingle, current)
	if err != nil {
		logs.Error(i18n.T("theme.set.error") + ": " + err.Error())
		return
	}

	// 保存主题设置
	if err := config.SaveTheme(selected[0]); err != nil {
		logs.Error(i18n.T("theme.set.error") + ": " + err.Error())
		return
	}

	// 立即应用主题设置(auto 内部重新探测终端背景)
	theme.ApplyMode(theme.Mode(selected[0]))

	logs.Success(i18n.T("theme.set.success"))
}
