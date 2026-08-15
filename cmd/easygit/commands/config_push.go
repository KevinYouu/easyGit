package commands

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// configPush 推送配置子流程:子菜单提供「设置推送配置 / 清除推送配置 / 返回」。
func configPush() {
	options := []config.Option{
		{
			Label:       i18n.T("config.push.setup"),
			Description: i18n.T("config.push.setup.desc"),
			Value:       "setup",
		},
		{
			Label:       i18n.T("config.push.clear"),
			Description: i18n.T("config.push.clear.desc"),
			Value:       "clear",
		},
		{
			Label:       i18n.T("config.back"),
			Description: i18n.T("config.back.desc"),
			Value:       "back",
		},
	}

	selected, err := form.ListForm(i18n.T("config.push.menu.title"), options, form.ListSingle)
	if err != nil {
		// 取消:返回配置中心
		return
	}

	switch selected[0] {
	case "setup":
		setPushConfig()
	case "clear":
		clearPushConfig()
	}
}

func clearPushConfig() {
	// 清除前确认(不可撤销)
	if !form.Confirm(i18n.T("config.push.clear.confirm")) {
		return
	}

	err := config.ClearPushConfig()
	if err != nil {
		logs.Error(i18n.T("error.clear.push.config"))
		return
	}

	logs.Success(i18n.T("push.config.cleared"))
}

func setPushConfig() {
	logs.Info(i18n.T("push.config.setup.title"))

	// 获取当前配置（用于预选）
	pushConfig, _ := config.GetPushConfig()
	var currentRemotes []string
	if pushConfig != nil {
		currentRemotes = pushConfig.Remotes
	}

	// 获取所有远程
	allRemotes, err := gitcmd.GetAllRemotes()
	if err != nil {
		logs.Error(i18n.T("error.get.remotes"))
		return
	}

	// 多选远程仓库,预选当前配置的远程
	remotes, err := form.ListForm(i18n.T("git.select.remotes.first"), form.StringOptions(allRemotes), form.ListMulti, currentRemotes...)
	if err != nil {
		logs.Error(i18n.T("error.select.remote"))
		return
	}

	if len(remotes) == 0 {
		logs.Error(i18n.T("error.no.remote.selected"))
		return
	}

	// 保存配置
	err = config.SavePushConfig(remotes)
	if err != nil {
		logs.Error(i18n.T("error.save.push.config"))
		return
	}

	remotesStr := strings.Join(remotes, ", ")
	logs.Success(fmt.Sprintf(i18n.T("push.config.saved.remotes"), remotesStr))
	logs.Info(i18n.T("push.config.will.use.current.branch"))
}
