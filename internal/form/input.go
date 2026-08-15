package form

import (
	"errors"
	"fmt"

	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// newInputField 构造输入字段:占位符、提示符与非空校验统一在此配置,
// NewInputForm 与 NewMultiInputForm 共用同一构造路径,防止两份配置漂移。
// desc 可选:空串时不渲染描述行;allowEmpty 为 true 时跳过非空校验;
// validate 可选:非空校验通过后执行的自定义校验。提示符统一 ❯(与列表选中指示符一致)。
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
	// 直接使用紧凑模式
	return newForm(
		huh.NewForm(
			huh.NewGroup(newInputField(title, value, "", false, nil)),
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

// NewMultiInputForm 构造单页多输入表单:所有输入框同页堆叠,Enter 逐字段推进、
// 末字段 Enter 提交,shift+tab 回退(huh 默认键位经 WithPosition 按字段位置启停);
// 生产与渲染测试共用同一构造路径,与 NewInputForm 同占位符、同非空校验。
// values 与 specs 一一对应,每个字段输入写入对应指针。
func NewMultiInputForm(specs []InputSpec, values []*string) *Form {
	// 所有字段放入同一个组:单页堆叠渲染,单 tea 程序单渲染器,
	// 避免连续内联表单在退出后残留内容与新表单上下堆叠
	fields := make([]huh.Field, 0, len(specs))
	for i, spec := range specs {
		// 标准数字序号(1. 2. …)标识输入步骤,聚焦时与标题同加粗
		title := stepTitle(i, spec.Title)
		fields = append(fields, newInputField(title, values[i], spec.Desc, spec.AllowEmpty, spec.Validate))
	}
	return newForm(
		huh.NewForm(
			huh.NewGroup(fields...),
		).WithTheme(theme.GetMultiInputTheme()).
			WithShowHelp(false),
		multiInputHelpKeys(),
	)
}

// NewCompactMultiInputForm 构造紧凑版单页多输入表单(矮终端友好):
// 字段无边框、标题与输入同行(Input.Inline)、字段间无空行,校验/键位/
// 占位符与 NewMultiInputForm 完全一致;preview 非 nil 时在字段组顶部
// 渲染实时预览行(huh Note + 值绑定,任一字段输入变化即重新求值)。
// 用于配置中心版本号上限编辑等需在 8-12 行终端完整操作的场景。
func NewCompactMultiInputForm(specs []InputSpec, values []*string, preview func([]string) string) *Form {
	fields := make([]huh.Field, 0, len(specs)+1)
	if preview != nil {
		// 绑定全部字段值:输入变化触发 TitleFunc 重算,实时预览组合结果
		fields = append(fields, huh.NewNote().TitleFunc(
			func() string {
				vals := make([]string, len(values))
				for i, v := range values {
					vals[i] = *v
				}
				return preview(vals)
			},
			values, // bindings:字段值指针数组,hashstructure 解引用比较
		))
	}
	for i, spec := range specs {
		title := stepTitle(i, spec.Title)
		// inline 行内 title/desc/输入直接拼接,desc 与提示符前加空格分隔
		desc := spec.Desc
		if desc != "" {
			desc = " " + desc
		}
		fields = append(fields,
			newInputField(title, values[i], desc, spec.AllowEmpty, spec.Validate).
				Inline(true).
				Prompt(" ❯ "))
	}
	return newForm(
		huh.NewForm(
			huh.NewGroup(fields...),
		).WithTheme(theme.GetCompactMultiInputTheme()).
			WithShowHelp(false),
		multiInputHelpKeys(),
	)
}

// stepTitle 为第 idx 个字段标题加标准数字序号前缀(如 "1. "、"2. "),
// 渲染层装饰,无语言差异,不进 i18n;序号随标题行样式渲染,
// 聚焦时与标题同加粗高亮
func stepTitle(idx int, title string) string {
	return fmt.Sprintf("%d. ", idx+1) + title
}

func Input(title string, defaultValue string) (string, error) {
	inputValue := defaultValue

	form := NewInputForm(title, &inputValue)
	if err := form.Run(); err != nil {
		return "", err
	}

	return inputValue, nil
}

// MultiInput 单页收集多个输入,返回结果与 specs 一一对应;
// 错误处理与 Input 一致(取消时上抛 ErrUserAborted)。
// 空 specs 直接返回空结果(构造前置守卫,避免空字段组)。
func MultiInput(specs []InputSpec) ([]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	values := make([]string, len(specs))
	ptrs := make([]*string, len(specs))
	for i, spec := range specs {
		values[i] = spec.Default
		ptrs[i] = &values[i]
	}

	form := NewMultiInputForm(specs, ptrs)
	if err := form.Run(); err != nil {
		return nil, err
	}

	return values, nil
}

// MultiInputCompact 同 MultiInput,使用紧凑布局(无边框/字段同行/无空行),
// 适合矮终端;preview 非 nil 时顶部渲染实时预览行(每次输入变化以当前
// 值重新求值,如版本号组合结果),可返回 nil 省略。
func MultiInputCompact(specs []InputSpec, preview func([]string) string) ([]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}

	values := make([]string, len(specs))
	ptrs := make([]*string, len(specs))
	for i, spec := range specs {
		values[i] = spec.Default
		ptrs[i] = &values[i]
	}

	form := NewCompactMultiInputForm(specs, ptrs, preview)
	if err := form.Run(); err != nil {
		return nil, err
	}

	return values, nil
}
