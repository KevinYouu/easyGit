package gitcmd

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// resetModeOptions 重置模式选择项:列表式单选表单 4 项单行选项
// (名称 + 说明,由 SelectForm 统一组装)。default 的 Value 为空串,
// 执行时不传模式参数(等同 mixed)。
func resetModeOptions() []config.Option {
	return []config.Option{
		{Label: i18n.T("reset.option.default.name"), Description: i18n.T("reset.option.default.desc"), Value: ""},
		{Label: i18n.T("reset.option.soft.name"), Description: i18n.T("reset.option.soft.desc"), Value: "--soft"},
		{Label: i18n.T("reset.option.mixed.name"), Description: i18n.T("reset.option.mixed.desc"), Value: "--mixed"},
		{Label: i18n.T("reset.option.hard.name"), Description: i18n.T("reset.option.hard.desc"), Value: "--hard"},
	}
}

func Reset() error {
	// 显示开始信息 - 简洁的标题
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Padding(0, 1)

	fmt.Printf("%s\n", headerStyle.Render(i18n.T("reset.title")))

	// 统一数据源:GetCommitsOptions 解析 git log(reset/squash/drop 共用)
	options, commits, err := GetCommitsOptions(0)
	if err != nil {
		return err
	}

	// HEAD 提交加 [HEAD] 标记区分(仅 reset 需要,统一数据源保持纯净标签)
	for i := range options {
		if commits[i].IsHead {
			options[i].Label = "[HEAD] " + options[i].Label
		}
	}

	// 使用统一列表选择表单
	chosen, err := form.ListForm(i18n.T("reset.select.commit"), options, form.ListSingle)
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.commit")+" %w", err)
	}
	choose := chosen[0]

	// 获取选择的提交完整信息
	var selectedCommit Commit
	for _, commit := range commits {
		if commit.Hash == choose {
			selectedCommit = commit
			break
		}
	}

	// 选择重置模式:列表式单选表单,4 项单行选项(名称 + 说明);
	// 预选上次使用的模式(未记忆时默认项在首位,直接 Enter)
	lastMode, _ := config.GetLastChoice(config.LastChoiceResetMode)
	resetModes, err := form.ListForm(i18n.T("reset.select.mode"), resetModeOptions(), form.ListSingle, lastMode)
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.mode")+" %w", err)
	}
	resetMode := resetModes[0]

	// 获取可读的重置模式名称(default 不传参数,直接取选项名)
	resetModeReadable := strings.TrimPrefix(resetMode, "--")
	if resetMode == "" {
		resetModeReadable = i18n.T("reset.option.default.name")
	}

	// 重置模式样式：统一使用 PrimaryColor + Bold
	modeStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)

	// 构建更紧凑的确认信息(完整提交消息,huh Confirm 长文本自动换行)
	confirmDesc := fmt.Sprintf("%s %s  %s %s %s%s",
		i18n.T("reset.confirm.to"),
		lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(selectedCommit.Hash),
		selectedCommit.Message,
		i18n.T("reset.confirm.mode"),
		modeStyle.Render(resetModeReadable),
		getModeDescription(resetMode),
	)

	// 针对 hard 模式添加警告，带黄色警告图标
	if resetMode == "--hard" {
		confirmDesc += "\n" + theme.WarningIconStyle.Render("⚠") + " " + lipgloss.NewStyle().
			Foreground(theme.PrimaryColor).
			Bold(true).
			Render(i18n.T("reset.hard.warning"))
	}

	// 使用自定义确认表单
	confirm := form.Confirm(confirmDesc)

	if confirm {
		// 执行重置操作:default 不传模式参数,其余模式传对应 flag
		resetArgs := []string{"reset"}
		if resetMode != "" {
			resetArgs = append(resetArgs, resetMode)
		}
		resetArgs = append(resetArgs, choose)

		_, err := command.RunCmdWithSpinnerOptions("git", resetArgs,
			fmt.Sprintf(i18n.T("reset.executing.mode"), resetModeReadable),
			fmt.Sprintf(i18n.T("reset.completed.to"), choose, resetModeReadable), true)

		if err != nil {
			return fmt.Errorf(i18n.T("reset.error.git.reset")+" %w", err)
		}

		// 记忆本次选择的模式,下次预选
		config.SaveLastChoice(config.LastChoiceResetMode, resetMode)

		// 显示简洁的成功信息
		fmt.Printf("\n%s %s\n",
			theme.SuccessIconStyle.Render("✓"),
			lipgloss.NewStyle().
				Foreground(theme.PrimaryColor).
				Render(fmt.Sprintf(i18n.T("reset.success.prefix"), choose)))

		// 简洁的操作提示
		switch resetMode {
		case "--soft":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.soft")))
		case "--mixed":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.mixed")))
		case "--hard":
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(i18n.T("reset.hint.hard")))
		}
	} else {
		fmt.Printf("\n%s %s\n",
			theme.InfoIconStyle.Render("ℹ"),
			theme.InfoStyle.Render(i18n.T("reset.cancelled.msg")))
	}
	return nil
}

// 获取重置模式的简短描述(default 不传参数,复用选项说明文案)
// getModeDescription 返回模式简短说明:统一复用选项说明文案,
// 避免与选项列表维护两份描述(默认模式即 default 选项)。
func getModeDescription(mode string) string {
	key := strings.TrimPrefix(mode, "--")
	if key == "" {
		key = "default"
	}
	if !i18n.Has("reset.option." + key + ".desc") {
		return ""
	}
	return lipgloss.NewStyle().
		Foreground(theme.MutedForeground).
		Render(i18n.T("reset.option." + key + ".desc"))
}
