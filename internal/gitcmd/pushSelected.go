package gitcmd

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

func PushSelected() error {
	fileStatus, err := getFileStatuses()
	if err != nil {
		logs.Error(i18n.T("error.get.file.status"))
		return fmt.Errorf("getFileStatuses: %w", err)
	}
	if len(fileStatus) == 0 {
		logs.Info(i18n.T("push.selected.no.files"))
		return nil
	}

	var selectedFiles []string
	for _, fileStatus := range fileStatus {
		if fileStatus.Status != "" {
			selectedFiles = append(selectedFiles, fileStatus.Path)
		}
	}

	data, err := form.MultiSelectForm(i18n.T("push.select.files"), selectedFiles)
	if err != nil {
		logs.Error(i18n.T("error.multiselect.form"))
		return fmt.Errorf("MultiSelectForm: %w", err)
	}

	if len(data) == 0 {
		logs.Error(i18n.T("push.selected.no.selection"))
		return nil
	}

	options, err := config.GetOptions()
	if err != nil {
		logs.Error(i18n.T("error.get.options"))
		return fmt.Errorf("GetOptions: %w", err)
	}

	_, suffix, err := form.SelectForm(i18n.T("push.select.commit.type"), options)
	if err != nil {
		return fmt.Errorf("SelectForm: %w", err)
	}

	commitMessage, err := form.Input(i18n.T("push.input.commit.message"), suffix+": ")
	if err != nil {
		return fmt.Errorf("input: %w", err)
	}

	// 使用新的命令执行器执行Git操作
	commands := []command.CommandInfo{
		{
			Command:     "git",
			Args:        append([]string{"add"}, data...),
			Description: i18n.T("git.add.selected.description"),
			LoadingMsg:  i18n.T("git.add.selected.loading"),
			SuccessMsg:  i18n.T("git.add.selected.success"),
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

	err = command.RunMultipleCommands(commands)
	if err != nil {
		return err
	}

	// 只有在所有Git操作都成功完成后才记录使用历史
	config.IncrementUsage(suffix)
	return nil
}
