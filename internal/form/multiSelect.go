package form

import (
	"os"

	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

func MultiSelectForm(title string, options []string, preselected ...[]string) (Values []string, err error) {
	// 检测终端高度用于高度计算
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = defaultTermHeight
	}

	// 使用统一的紧凑布局
	var selectedValues []string
	// 如果提供了预选值，使用预选值初始化
	if len(preselected) > 0 && len(preselected[0]) > 0 {
		selectedValues = preselected[0]
	}

	selectOpts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		selectOpts[i] = huh.NewOption(opt, opt)
	}

	// 按终端高度与选项数动态计算选择框高度,确保多选项可见
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Height(CalculateMultiSelectHeight(len(options), height)).
				Options(selectOpts...).
				Value(&selectedValues),
		),
	).WithTheme(theme.GetCompactTheme()).
		WithShowHelp(false)

	err = form.Run()
	if err != nil {
		return nil, err
	}

	return selectedValues, nil
}
