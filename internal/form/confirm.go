package form

import (
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// NewConfirmForm 构造确认 huh 表单(Confirm 与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移)。确认结果写入 value。
func NewConfirmForm(title string, value *bool) *Form {
	return newForm(
		huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(title).
					Value(value),
			),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		confirmHelpKeys(),
	)
}

func Confirm(title string) bool {
	var confirmed bool

	if err := NewConfirmForm(title, &confirmed).Run(); err != nil {
		return false
	}

	return confirmed
}
