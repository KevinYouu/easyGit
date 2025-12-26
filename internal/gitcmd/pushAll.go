package gitcmd

import (
	"fmt"

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

	// 使用新的命令执行器执行Git操作
	commands := []command.CommandInfo{
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
		{
			Command:     "git",
			Args:        []string{"push"},
			Description: i18n.T("git.push.description"),
			LoadingMsg:  i18n.T("git.push.loading"),
			SuccessMsg:  i18n.T("git.push.success"),
		},
	}

	return command.RunMultipleCommands(commands)
}
