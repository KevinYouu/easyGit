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

	// 文件选项带状态列(状态自动宽度,路径列占满):M/A/D 一目了然,
	// 零新增步骤,仅信息增强
	var fileOptions []config.Option
	for _, fs := range fileStatus {
		if fs.Status == "" {
			continue
		}
		fileOptions = append(fileOptions, config.Option{
			Label: fs.Path,
			Value: fs.Path,
			Cells: []string{fs.Status, fs.Path},
		})
	}

	data, err := form.ListFormColumns(
		i18n.T("push.select.files"),
		[]form.ColumnSpec{
			{Kind: form.ColumnAuto, MaxWidth: form.MaxAutoColumnWidth},
			{Kind: form.ColumnFlex},
		},
		fileOptions,
		form.ListMulti,
	)
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
	options = applyCommitTypeDescriptions(options)

	suffixes, err := form.ListFormWrap(i18n.T("push.select.commit.type"), options, form.ListSingle)
	if err != nil {
		return fmt.Errorf("SelectForm: %w", err)
	}
	suffix := suffixes[0]

	commitMessage, subject, err := commitMessageInputs(suffix + ": ")
	if err != nil {
		return err
	}

	remotes, remotesErr := GetAllRemotes()
	hasRemote := remotesErr == nil && len(remotes) > 0

	if !hasRemote {
		logs.Info(i18n.T("push.no.remote.commit.only"))
	}

	allCommands := []command.CommandInfo{
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
	}

	if hasRemote {
		currentBranch, err := GetCurrentBranch()
		if err != nil {
			logs.Error(i18n.T("error.get.current.branch"))
			return fmt.Errorf("get current branch: %w", err)
		}

		remotes, err = SelectAndSaveRemotes()
		if err != nil {
			return fmt.Errorf("select remote: %w", err)
		}

		pullBefore := shouldPullBeforePush()
		allCommands, parallelFrom := appendPushSteps(allCommands, remotes, currentBranch, pullBefore)

		if err := command.RunMultipleCommandsParallel(allCommands, parallelFrom); err != nil {
			return err
		}
	} else if err := command.RunMultipleCommandsParallel(allCommands, len(allCommands)); err != nil {
		return err
	}

	config.IncrementUsage(suffix)
	config.AddRecentCommitMessage(subject)
	return nil
}
