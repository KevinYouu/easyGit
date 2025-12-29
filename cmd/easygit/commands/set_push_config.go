package commands

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/gitcmd"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/spf13/cobra"
)

var setPushConfigCmd = &cobra.Command{
	Use:   "set-push-config",
	Short: i18n.T("push.config.short"),
	Run: func(cmd *cobra.Command, args []string) {
		// 如果第一个参数是 "clear",清除配置
		if len(args) > 0 && args[0] == "clear" {
			clearPushConfig()
			return
		}

		// 否则,设置新配置(会先显示当前配置,然后询问是否修改)
		setPushConfig()
	},
}

func clearPushConfig() {
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

	// 多选远程仓库，预选当前配置的远程
	remotes, err := form.MultiSelectForm(i18n.T("git.select.remotes.first"), allRemotes, currentRemotes)
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

// SetPushConfigCommand 返回 set-push-config 命令
func SetPushConfigCommand() *cobra.Command {
	return setPushConfigCmd
}
