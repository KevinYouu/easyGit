package form

import (
	"os"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

func SelectForm(title string, options []config.Option, preselected ...string) (label, value string, err error) {
	// 使用统一的紧凑布局
	var selectedValue string
	// 如果提供了预选值，使用预选值初始化
	if len(preselected) > 0 && preselected[0] != "" {
		selectedValue = preselected[0]
	}

	selectOpts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		selectOpts[i] = huh.NewOption(opt.Label, opt.Value)
	}

	// 检测终端高度,动态计算选择框高度
	_, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termHeight = defaultTermHeight
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(selectOpts...).
				Height(CalculateSelectHeight(len(options), termHeight)).
				Value(&selectedValue).
				Filtering(false),
		),
	).WithTheme(theme.GetCompactTheme()).
		WithShowHelp(false)

	err = form.Run()
	if err != nil {
		return "", "", err
	}

	// 找到选中的选项
	for _, opt := range options {
		if opt.Value == selectedValue {
			return opt.Label, opt.Value, nil
		}
	}

	return "", "", nil
}
