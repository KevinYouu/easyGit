package form

import (
	"errors"
	"strings"

	key "charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// 多行文本框可视行数:随默认值行数增长并夹在区间内,
// 短消息不占整屏,长消息(合并多条提交信息)可一屏内看到主体。
const (
	minTextLines = 6
	maxTextLines = 14
)

// textValidate 多行文本校验:allowEmpty 放行空内容(可选正文,Enter 跳过)。
func textValidate(allowEmpty bool) func(string) error {
	return func(str string) error {
		if allowEmpty || strings.TrimSpace(str) != "" {
			return nil
		}
		return errors.New(i18n.T("form.input.empty.error"))
	}
}

// newTextField 多行文本字段:Enter 提交、ctrl+j/alt+enter 换行、esc 取消,
// 默认值(含换行)原样带入。
func newTextField(title string, value *string, allowEmpty bool) *huh.Text {
	lines := min(max(strings.Count(*value, "\n")+2, minTextLines), maxTextLines)

	return huh.NewText().
		Title(title).
		Value(value).
		Lines(lines).
		Placeholder(i18n.T("form.input.placeholder")).
		ExternalEditor(false).
		Validate(textValidate(allowEmpty))
}

// applyTextInputKeyMap 禁用 ctrl+e 外部编辑器:该键会挂起 TUI 并阻塞在外编辑器进程上;
// ExternalEditor(false) 只影响帮助栏展示,实际分发取 keymap,必须在此禁用。
func applyTextInputKeyMap(km *huh.KeyMap) {
	km.Text.Editor = key.NewBinding(key.WithKeys("ctrl+e"), key.WithDisabled())
}

// newTextInputForm 多行文本表单统一构造路径(必填/可空共用)。
// 启动时派发 ctrl+home(textarea InputBegin):huh 预填值后光标在末尾,
// 视图会停在最后几行,该键把光标与视图移回首行。
func newTextInputForm(title string, value *string, allowEmpty bool) *Form {
	return newForm(
		huh.NewForm(
			huh.NewGroup(newTextField(title, value, allowEmpty)),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		textInputHelpKeys(),
		applyTextInputKeyMap,
	).WithInitialMsg(tea.KeyPressMsg{Code: tea.KeyHome, Mod: tea.ModCtrl})
}

// NewTextInputForm 构造必填多行文本表单(TextInput 与渲染测试共用构造路径)。
func NewTextInputForm(title string, value *string) *Form {
	return newTextInputForm(title, value, false)
}

// NewOptionalTextInputForm 构造可空多行文本表单(可选正文等,Enter 跳过)。
func NewOptionalTextInputForm(title string, value *string) *Form {
	return newTextInputForm(title, value, true)
}

// TextInput 多行文本输入(提交信息含正文/多段文本的场景):
// 默认值中的换行与空行原样保留,确认后按原样返回,不做任何折叠。
func TextInput(title, defaultValue string) (string, error) {
	value := defaultValue
	if err := NewTextInputForm(title, &value).Run(); err != nil {
		return "", err
	}
	return value, nil
}

// TextInputOptional 可空多行文本输入(可选正文):空内容直接确认返回空串,
// Esc 取消返回错误。
func TextInputOptional(title, defaultValue string) (string, error) {
	value := defaultValue
	if err := NewOptionalTextInputForm(title, &value).Run(); err != nil {
		return "", err
	}
	return value, nil
}
