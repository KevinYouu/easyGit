package form

import (
	"errors"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/huh"
)

func Input(title string, defaultValue string) (string, error) {
	inputValue := defaultValue

	// 直接使用紧凑模式
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Placeholder(i18n.T("form.input.placeholder")).
				Value(&inputValue).
				Validate(func(str string) error {
					if str == "" {
						return errors.New(i18n.T("form.input.empty.error"))
					}
					return nil
				}),
		),
	).WithTheme(theme.GetCompactTheme())

	err := form.Run()
	if err != nil {
		return "", err
	}

	return inputValue, nil
}
