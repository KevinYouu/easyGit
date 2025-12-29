package gitcmd

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

func PushAll() error {
	options, err := config.GetOptions()
	if err != nil {
		logs.Error(i18n.T("error.get.options"))
		return fmt.Errorf("get options: %w", err)
	}

	_, suffix, err := form.SelectForm(i18n.T("push.select.commit.type"), options)
	if err != nil {
		return fmt.Errorf("select form: %w", err)
	}
	config.IncrementUsage(suffix)

	commitMessage, err := form.Input(i18n.T("push.input.commit.message"), suffix+": ")
	if err != nil {
		return fmt.Errorf("input: %w", err)
	}

	// 获取当前分支
	currentBranch, err := GetCurrentBranch()
	if err != nil {
		logs.Error(i18n.T("error.get.current.branch"))
		return fmt.Errorf("get current branch: %w", err)
	}

	// 选择远程仓库(支持配置持久化和多选)
	remotes, needSave, err := SelectRemoteWithConfig()
	if err != nil {
		return fmt.Errorf("select remote: %w", err)
	}

	// 如果需要保存配置(首次选择或配置变更)
	if needSave {
		err = config.SavePushConfig(remotes)
		if err != nil {
			logs.Error(i18n.T("error.save.push.config"))
		} else {
			remotesStr := strings.Join(remotes, ", ")
			logs.Info(fmt.Sprintf(i18n.T("push.config.saved.remotes"), remotesStr))
		}
	} else {
		// 显示当前使用的配置
		remotesStr := strings.Join(remotes, ", ")
		logs.Info(fmt.Sprintf(i18n.T("push.using.config.remotes"), remotesStr))
	}

	// 构建所有命令列表: add -> commit -> pull -> push(为每个远程创建一个步骤)
	allCommands := []command.CommandInfo{
		{
			Command:     "git",
			Args:        []string{"add", "-A"},
			Description: i18n.T("git.add.all.description"),
			LoadingMsg:  i18n.T("git.add.all.loading"),
			SuccessMsg:  i18n.T("git.add.all.success"),
		},
		{
			Command:     "git",
			Args:        []string{"commit", "-m", commitMessage},
			Description: i18n.T("git.commit.description"),
			LoadingMsg:  i18n.T("git.commit.loading"),
			SuccessMsg:  i18n.T("git.commit.success"),
		},
		{
			Command:     "git",
			Args:        []string{"pull"},
			Description: i18n.T("git.pull.description"),
			LoadingMsg:  i18n.T("git.pull.loading"),
			SuccessMsg:  i18n.T("git.pull.success"),
		},
	}

	// 添加每个远程的推送命令
	for _, remote := range remotes {
		allCommands = append(allCommands, command.CommandInfo{
			Command:     "git",
			Args:        []string{"push", remote, currentBranch},
			Description: fmt.Sprintf(i18n.T("git.push.to.remote"), remote),
			LoadingMsg:  fmt.Sprintf(i18n.T("git.push.loading.remote"), remote),
			SuccessMsg:  fmt.Sprintf(i18n.T("git.push.success.remote"), remote),
		})
	}

	// 使用统一的进度条执行所有命令
	return command.RunMultipleCommands(allCommands)
}
