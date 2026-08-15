package form

import (
	"errors"
	"fmt"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// newInputField 构造输入字段:占位符、提示符与非空校验统一在此配置,
// 单输入与列表输入共用同一构造路径,防止两份配置漂移。
// desc 可选:空串时不渲染描述行;allowEmpty 为 true 时跳过非空校验;
// validate 可选:非空校验通过后执行的自定义校验。提示符统一 ❯(与列表选中指示符一致)。
// 注:多输入表单已自绘(multi_input.go),不经本构造器。
func newInputField(title string, value *string, desc string, allowEmpty bool, validate func(string) error) *huh.Input {
	f := huh.NewInput().
		Title(title).
		Prompt("❯ ").
		Placeholder(i18n.T("form.input.placeholder")).
		Value(value).
		Validate(func(str string) error {
			if str == "" && !allowEmpty {
				return errors.New(i18n.T("form.input.empty.error"))
			}
			if validate != nil {
				return validate(str)
			}
			return nil
		})
	if desc != "" {
		f = f.Description(desc)
	}
	return f
}

// NewInputForm 构造输入 huh 表单(Input 与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移)。输入值写入 value,占位符与非空校验统一在此配置。
func NewInputForm(title string, value *string) *Form {
	return NewInputFormWithValidate(title, value, nil)
}

// NewInputFormWithValidate 构造带自定义校验的输入表单:非空校验通过后
// 执行 validate(如提交消息主题非空),为 nil 时跳过。
func NewInputFormWithValidate(title string, value *string, validate func(string) error) *Form {
	// 直接使用紧凑模式
	return newForm(
		huh.NewForm(
			huh.NewGroup(newInputField(title, value, "", false, validate)),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		inputHelpKeys(),
	)
}

// InputSpec 单页多输入表单的一项:标题、默认值与可选描述
// (Desc 渲染为字段下方的弱化提示行,空串不显示);
// AllowEmpty 为 true 时跳过非空校验(prefix/suffix 等可空字段);
// Validate 可选:非空校验通过后执行的自定义校验(如数字格式)。
type InputSpec struct {
	Title      string
	Default    string
	Desc       string
	AllowEmpty bool
	Validate   func(string) error
}

// MultiInput 单页收集多个输入,返回结果与 specs 一一对应;
// 布局为自绘三列(标题列/输入框/行尾弱化简介列),简介不混入标题、
// 不占额外行;↑/↓/k/j 上下导航,enter 推进/末字段提交,esc 取消;
// preview 非 nil 时顶部渲染实时预览行(每次输入变化以当前值重新求值,
// 如版本号组合结果),传 nil 省略。错误处理与 Input 一致(取消时上抛
// ErrUserAborted)。空 specs 直接返回空结果(构造前置守卫)。
func MultiInput(specs []InputSpec, preview func([]string) string) ([]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	values := make([]string, len(specs))
	ptrs := make([]*string, len(specs))
	for i, spec := range specs {
		values[i] = spec.Default
		ptrs[i] = &values[i]
	}

	m := newMultiInputModel(specs, ptrs, preview)
	if err := m.run(); err != nil {
		return nil, err
	}

	return values, nil
}

// stepTitle 为第 idx 个字段标题加标准数字序号前缀(如 "1. "、"2. "),
// 渲染层装饰,无语言差异,不进 i18n;序号随标题行样式渲染,
// 聚焦时与标题同加粗高亮
func stepTitle(idx int, title string) string {
	return fmt.Sprintf("%d. ", idx+1) + title
}

func Input(title string, defaultValue string) (string, error) {
	return InputWithValidate(title, defaultValue, nil)
}

// InputWithValidate 输入表单,支持自定义校验(非空校验通过后执行),
// 用于提交消息主题非空等业务级校验。
func InputWithValidate(title string, defaultValue string, validate func(string) error) (string, error) {
	inputValue := defaultValue

	form := NewInputFormWithValidate(title, &inputValue, validate)
	if err := form.Run(); err != nil {
		return "", err
	}

	return inputValue, nil
}
