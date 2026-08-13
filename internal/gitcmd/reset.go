package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

type Commit struct {
	Hash      string
	Message   string
	Date      string
	Author    string
	Email     string
	IsHead    bool
	Timestamp time.Time
}

// resetMessageMaxWidth 重置模式选择时提交消息的最大显示宽度
const resetMessageMaxWidth = 40

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

	// 使用更详细的git log格式获取提交历史，包含ISO时间戳用于排序
	cmd := exec.Command("git", "log", "--pretty=format:%h|%s|%ad|%an|%ae|%ai", "--date=format:%m-%d %H:%M")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf(i18n.T("error.git.log")+" %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var options = []config.Option{}
	var commits = []Commit{}

	// 解析并存储提交信息（不显示历史记录）
	for i, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) >= 5 {
			hash := parts[0]
			message := parts[1]
			date := parts[2]
			author := parts[3]
			email := parts[4]

			commit := Commit{
				Hash:    hash,
				Message: message,
				Date:    date,
				Author:  author,
				Email:   email,
				IsHead:  i == 0,
			}

			// 解析时间戳用于排序（如果有的话）
			if len(parts) >= 6 {
				if timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", parts[5]); err == nil {
					commit.Timestamp = timestamp
				}
			}

			// 存储提交信息
			commits = append(commits, commit)

			// 限制消息长度，避免过长(按显示宽度截断,正确处理中文/emoji)
			shortMsg := form.SafeTruncate(message, resetMessageMaxWidth)

			// 添加到选择列表，使用纯文本格式以允许背景色正确覆盖
			commitLabel := ""
			if i == 0 {
				// HEAD提交使用纯文本格式，但添加标记以区分
				commitLabel = fmt.Sprintf(
					"[HEAD] %s %s\n%s • %s",
					hash,
					shortMsg,
					date,
					author,
				)
			} else {
				// 普通提交使用纯文本格式
				commitLabel = fmt.Sprintf(
					"%s %s\n%s • %s",
					hash,
					shortMsg,
					date,
					author,
				)
			}
			options = append(options, config.Option{Label: commitLabel, Value: hash})
		}
	}

	// 使用表格选择表单
	_, choose, err := form.TableSelectForm(options)
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.commit")+" %w", err)
	}

	// 获取选择的提交完整信息
	var selectedCommit Commit
	for _, commit := range commits {
		if commit.Hash == choose {
			selectedCommit = commit
			break
		}
	}

	// 选择重置模式:列表式单选表单,4 项单行选项(名称 + 说明)
	_, resetMode, err := form.SelectForm(i18n.T("reset.select.mode"), resetModeOptions())
	if err != nil {
		return fmt.Errorf(i18n.T("reset.error.select.mode")+" %w", err)
	}

	// 获取可读的重置模式名称(default 不传参数,直接取选项名)
	resetModeReadable := strings.TrimPrefix(resetMode, "--")
	if resetMode == "" {
		resetModeReadable = i18n.T("reset.option.default.name")
	}

	// 重置模式样式：统一使用 PrimaryColor + Bold
	modeStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)

	// 构建更紧凑的确认信息
	shortMsg := form.SafeTruncate(selectedCommit.Message, resetMessageMaxWidth)

	confirmDesc := fmt.Sprintf("%s %s  %s %s %s%s",
		i18n.T("reset.confirm.to"),
		lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(selectedCommit.Hash),
		shortMsg,
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
