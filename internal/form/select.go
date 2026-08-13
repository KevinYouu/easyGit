package form

import (
	"os"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
	"golang.org/x/term"
)

// NewSelectForm 构造单选 huh 表单(SelectForm 与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移)。height 为终端高度,字段高度按
// CalculateSelectHeight 计算;选中值写入 selected。
func NewSelectForm(title string, options []config.Option, height int, selected *string) *Form {
	selectOpts := make([]huh.Option[string], len(options))
	for i, opt := range options {
		// 统一单行组装:名称亮加粗 + 说明灰(Description 为空则纯名称,零变化)
		selectOpts[i] = huh.NewOption(OptionLabel(opt.Label, opt.Description), opt.Value)
	}

	// huh v2 已修复 v1 滚动 bug(光标跟随可见区),高屏展示更多选项,
	// 不再固定默认 10 行
	return newForm(
		huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(title).
					Options(selectOpts...).
					Value(selected).
					Height(CalculateSelectHeight(len(options), height)).
					Filtering(false),
			),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		selectHelpKeys(),
	)
}

func SelectForm(title string, options []config.Option, preselected ...string) (label, value string, err error) {
	// 使用统一的紧凑布局
	var selectedValue string
	// 如果提供了预选值，使用预选值初始化
	if len(preselected) > 0 && preselected[0] != "" {
		selectedValue = preselected[0]
	}

	// 按终端高度动态设置 Height
	_, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		height = defaultTermHeight
	}

	form := NewSelectForm(title, options, height, &selectedValue)
	if err = form.Run(); err != nil {
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
