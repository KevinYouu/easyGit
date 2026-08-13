package form

import (
	"os"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/theme"
	"golang.org/x/term"
)

// NewMultiSelectForm 构造多选 huh 表单(MultiSelectForm 与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移)。height 为终端高度,字段高度按
// CalculateMultiSelectHeight 计算;选中值写入 selected。
func NewMultiSelectForm(title string, options []string, height int, selected *[]string) *huh.Form {
	selectOpts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		selectOpts[i] = huh.NewOption(opt, opt)
	}

	// huh v2 已修复 v1 滚动 bug(光标跟随可见区),高屏展示更多选项,
	// 不再固定默认 8 行
	return huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(selectOpts...).
				Value(selected).
				Height(CalculateMultiSelectHeight(len(options), height)),
		),
	).WithTheme(theme.GetCompactTheme()).
		WithShowHelp(false)
}

func MultiSelectForm(title string, options []string, preselected ...[]string) (Values []string, err error) {
	// 使用统一的紧凑布局
	var selectedValues []string
	// 如果提供了预选值，使用预选值初始化
	if len(preselected) > 0 && len(preselected[0]) > 0 {
		selectedValues = preselected[0]
	}

	// 按终端高度动态设置 Height
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = defaultTermHeight
	}

	form := NewMultiSelectForm(title, options, height, &selectedValues)
	if err = form.Run(); err != nil {
		return nil, err
	}

	return selectedValues, nil
}
