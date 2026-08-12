package form

import (
	"os"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/theme"
	"golang.org/x/term"
)

func MultiSelectForm(title string, options []string, preselected ...[]string) (Values []string, err error) {
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

	// 按终端高度动态设置 Height:huh v2 已修复 v1 滚动 bug(光标跟随可见区),
	// 高屏展示更多选项,不再固定默认 8 行
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = defaultTermHeight
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(selectOpts...).
				Value(&selectedValues).
				Height(CalculateMultiSelectHeight(len(options), height)),
		),
	).WithTheme(theme.GetCompactTheme()).
		WithShowHelp(false)

	err = form.Run()
	if err != nil {
		return nil, err
	}

	return selectedValues, nil
}
