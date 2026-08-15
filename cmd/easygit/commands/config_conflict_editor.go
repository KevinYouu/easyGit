package commands

import (
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// presetEditors 预置编辑器选项(自定义值不在其中,列表预选「自定义」)
var presetEditors = map[string]bool{
	"vim":  true,
	"vi":   true,
	"nano": true,
}

// configConflictEditor 冲突编辑器子流程:单选 自动/vim/vi/nano/自定义,
// 自定义走输入子流程(留空恢复自动),保存后立即生效。
func configConflictEditor() {
	// 获取当前设置,预选对应选项(未设置按自动)
	current, err := config.GetConflictEditor()
	if err != nil {
		logs.Error(i18n.T("conflict.editor.set.error") + ": " + err.Error())
		return
	}

	options := []config.Option{
		{
			Label:       i18n.T("conflict.editor.option.auto"),
			Description: i18n.T("conflict.editor.option.auto.desc"),
			Value:       "",
		},
		{
			Label:       "vim",
			Description: i18n.T("conflict.editor.option.vim.desc"),
			Value:       "vim",
		},
		{
			Label:       "vi",
			Description: i18n.T("conflict.editor.option.vi.desc"),
			Value:       "vi",
		},
		{
			Label:       "nano",
			Description: i18n.T("conflict.editor.option.nano.desc"),
			Value:       "nano",
		},
		{
			Label:       i18n.T("conflict.editor.option.custom"),
			Description: i18n.T("conflict.editor.option.custom.desc"),
			Value:       config.ConflictEditorCustom,
		},
	}

	// 预选:当前值在预置列表内则预选对应项,否则(自定义命令)预选「自定义」
	preselected := current
	if preselected != "" && !presetEditors[preselected] {
		preselected = config.ConflictEditorCustom
	}

	selected, err := form.ListFormColumns(i18n.T("conflict.editor.select.title"), configListSpecs(), options, form.ListSingle, preselected)
	if err != nil {
		// Esc / 取消:返回配置中心主列表
		return
	}

	value := selected[0]
	if value == config.ConflictEditorCustom {
		// 自定义输入;留空 = 清除设置(自动检测)
		value, err = form.Input(i18n.T("conflict.editor.input.title"), current)
		if err != nil {
			return
		}
		value = strings.TrimSpace(value)
	}

	if err := config.SaveConflictEditor(value); err != nil {
		logs.Error(i18n.T("conflict.editor.set.error") + ": " + err.Error())
		return
	}

	logs.Success(i18n.T("conflict.editor.set.success"))
}
