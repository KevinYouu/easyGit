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

	commitMessage, err := form.InputWithSuggestions(i18n.T("push.input.commit.message"), suffix+": ", validateCommitMessage, config.GetRecentCommitMessages())
	if err != nil {
		return fmt.Errorf("input: %w", err)
	}

	// 检测是否有远程仓库
	remotes, remotesErr := GetAllRemotes()
	hasRemote := remotesErr == nil && len(remotes) > 0

	// 是否需要 pull(在 hasRemote 块内赋值;未设 upstream 或配置 never 时跳过)
	pullBefore := false

	if !hasRemote {
		logs.Info(i18n.T("push.no.remote.commit.only"))
	}

	// 构建基础命令列表: add -> commit
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
		// 获取当前分支
		currentBranch, err := GetCurrentBranch()
		if err != nil {
			logs.Error(i18n.T("error.get.current.branch"))
			return fmt.Errorf("get current branch: %w", err)
		}

		// 选择远程仓库(支持配置持久化和多选,输出保存/使用日志)
		remotes, err = SelectAndSaveRemotes()
		if err != nil {
			return fmt.Errorf("select remote: %w", err)
		}

		// 是否需要 pull:未设 upstream 或配置 never 时跳过
		pullBefore = shouldPullBeforePush()

		// 添加 pull 步骤
		if pullBefore {
			allCommands = append(allCommands, command.CommandInfo{
				Command:     "git",
				Args:        []string{"pull"},
				Description: i18n.T("git.pull.description"),
				LoadingMsg:  i18n.T("git.pull.loading"),
				SuccessMsg:  i18n.T("git.pull.success"),
			})
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
	}

	// 使用统一的进度条执行所有命令:add → commit → (pull) 串行,
	// 随后所有远程推送一次性并行启动;parallelFrom = 串行步数
	parallelFrom := 2
	if pullBefore {
		parallelFrom = 3
	}
	err = command.RunMultipleCommandsParallel(allCommands, parallelFrom)
	if err != nil {
		return err
	}

	// 只有在所有Git操作都成功完成后才记录使用历史
	config.IncrementUsage(suffix)
	// 记忆本次提交消息(下次 ↑ 直接复用)
	config.AddRecentCommitMessage(commitMessage)
	return nil
}
