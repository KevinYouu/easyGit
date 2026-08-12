package form

import (
	"os"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
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

	// 按终端高度动态设置 Height:huh v2 已修复 v1 滚动 bug(光标跟随可见区),
	// 高屏展示更多选项,不再固定默认 10 行
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = defaultTermHeight
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(selectOpts...).
				Value(&selectedValue).
				Height(CalculateSelectHeight(len(options), height)).
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
