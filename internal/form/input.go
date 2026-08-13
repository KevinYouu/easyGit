package form

import (
	"errors"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// NewInputForm 构造输入 huh 表单(Input 与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移)。输入值写入 value,占位符与非空校验统一在此配置。
func NewInputForm(title string, value *string) *Form {
	// 直接使用紧凑模式
	return newForm(
		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(title).
					Placeholder(i18n.T("form.input.placeholder")).
					Value(value).
					Validate(func(str string) error {
						if str == "" {
							return errors.New(i18n.T("form.input.empty.error"))
						}
						return nil
					}),
			),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		inputHelpKeys(),
	)
}

func Input(title string, defaultValue string) (string, error) {
	inputValue := defaultValue

	form := NewInputForm(title, &inputValue)
	if err := form.Run(); err != nil {
		return "", err
	}

	return inputValue, nil
}
