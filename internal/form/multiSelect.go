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
		height = 24 // 默认值
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

	// 计算合适的高度，确保多选项可见
	availableHeight := height - 6
	compactHeight := len(options) + 1
	if compactHeight > availableHeight {
		compactHeight = availableHeight
	}
	if compactHeight < 3 {
		compactHeight = 3
	}
	if compactHeight > 8 {
		compactHeight = 8 // 限制最大高度
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Height(compactHeight).
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
