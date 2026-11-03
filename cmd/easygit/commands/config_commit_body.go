package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// configCommitBody 补录提交正文开关子流程:单选 on/off,预选当前值。
// 默认 off(只录单行主题);on 时主题确认后再采集可选正文。
func configCommitBody() {
	current, err := config.GetCommitBody()
	if err != nil || (current != config.CommitBodyOn && current != config.CommitBodyOff) {
		current = config.CommitBodyOff
	}

	options := []config.Option{
		{
			Label:       i18n.T("config.commit.body.option.on"),
			Description: i18n.T("config.commit.body.option.on.desc"),
			Value:       config.CommitBodyOn,
		},
		{
			Label:       i18n.T("config.commit.body.option.off"),
			Description: i18n.T("config.commit.body.option.off.desc"),
			Value:       config.CommitBodyOff,
		},
	}

	selected, err := form.ListFormColumns(i18n.T("config.commit.body.select.title"), configListSpecs(), options, form.ListSingle, current)
	if err != nil {
		logs.Error(i18n.T("config.commit.body.set.error") + ": " + err.Error())
		return
	}

	if err := config.SaveCommitBody(selected[0]); err != nil {
		logs.Error(i18n.T("config.commit.body.set.error") + ": " + err.Error())
		return
	}

	logs.Success(i18n.T("config.commit.body.set.success"))
}
