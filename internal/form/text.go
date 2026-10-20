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

// newTextField 构造多行文本字段(提交信息正文编辑):
// Enter 提交、ctrl+j/alt+enter 换行、esc 取消;默认值原样带入(含换行)。
// 关闭 huh 的外部编辑器(ctrl+e):TUI 自身即编辑器,再调起外部编辑器
// 会阻塞界面(历史问题),且外部编辑器退出后仍要回到表单。
func newTextField(title string, value *string) *huh.Text {
	lines := min(max(strings.Count(*value, "\n")+2, minTextLines), maxTextLines)

	return huh.NewText().
		Title(title).
		Value(value).
		Lines(lines).
		Placeholder(i18n.T("form.input.placeholder")).
		ExternalEditor(false).
		Validate(func(str string) error {
			if strings.TrimSpace(str) == "" {
				return errors.New(i18n.T("form.input.empty.error"))
			}
			return nil
		})
}

// applyTextInputKeyMap 禁用 huh 多行文本默认绑定的 ctrl+e(打开外部编辑器):
// 该键会挂起 TUI 并阻塞在外部编辑器进程上(历史问题),而表单本身已是编辑器;
// ExternalEditor(false) 只影响帮助栏键位展示,实际分发取 keymap,必须在此禁用。
func applyTextInputKeyMap(km *huh.KeyMap) {
	km.Text.Editor = key.NewBinding(key.WithKeys("ctrl+e"), key.WithDisabled())
}

// NewTextInputForm 构造多行文本表单(TextInput 与渲染测试共用构造路径)。
// 启动时派发一次 ctrl+home(textarea 的 InputBegin):huh 预填值后光标停在末尾,
// 视图会从最后几行开始,合并多条提交信息时看不到主题;该键把光标与视图移回首行,
// 之后可正常向下滚动。
func NewTextInputForm(title string, value *string) *Form {
	return newForm(
		huh.NewForm(
			huh.NewGroup(newTextField(title, value)),
		).WithTheme(theme.GetCompactTheme()).
			WithShowHelp(false),
		textInputHelpKeys(),
		applyTextInputKeyMap,
	).WithInitialMsg(tea.KeyPressMsg{Code: tea.KeyHome, Mod: tea.ModCtrl})
}

// TextInput 多行文本输入(提交信息含正文/多段文本的场景):
// 默认值中的换行与空行原样保留,确认后按原样返回,不做任何折叠。
func TextInput(title, defaultValue string) (string, error) {
	textValue := defaultValue
	if err := NewTextInputForm(title, &textValue).Run(); err != nil {
		return "", err
	}
	return textValue, nil
}
