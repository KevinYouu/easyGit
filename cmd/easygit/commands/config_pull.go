package commands

import (
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// configPull 推送前 pull 策略子流程:单选 always/never,预选当前值。
// 默认 always(推送前先拉取);never 时 pa/ps 跳过 pull 步骤。
func configPull() {
	current, err := config.GetPullBeforePush()
	if err != nil || current == "" {
		current = config.PullBeforePushAlways
	}

	options := []config.Option{
		{
			Label:       i18n.T("config.pull.option.always"),
			Description: i18n.T("config.pull.option.always.desc"),
			Value:       config.PullBeforePushAlways,
		},
		{
			Label:       i18n.T("config.pull.option.never"),
			Description: i18n.T("config.pull.option.never.desc"),
			Value:       config.PullBeforePushNever,
		},
	}

	selected, err := form.ListFormColumns(i18n.T("config.pull.select.title"), configListSpecs(), options, form.ListSingle, current)
	if err != nil {
		logs.Error(i18n.T("config.pull.set.error") + ": " + err.Error())
		return
	}

	if err := config.SavePullBeforePush(selected[0]); err != nil {
		logs.Error(i18n.T("config.pull.set.error") + ": " + err.Error())
		return
	}

	logs.Success(i18n.T("config.pull.set.success"))
}
