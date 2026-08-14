package command

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// 进度显示默认尺寸常量
const (
	progressBarMaxWidth   = 40 // 进度条最大宽度
	progressDefaultWidth  = 80 // 默认终端宽度(未收到尺寸消息前)
	progressDefaultHeight = 24 // 默认终端高度(未收到尺寸消息前)
)

// ProgressModel 多步骤进度显示模型 - 统一的进度条组件
type ProgressModel struct {
	commands     []CommandInfo
	currentStep  int
	total        int
	status       string
	isCompleted  bool
	hasError     bool
	errorMessage string
	results      []string
	executing    bool
	width        int
	height       int

	// Spinner 相关字段
	showSpinner bool
	spinner     spinnerAnimation
	frame       int

	// 步骤状态跟踪
	stepStatus []int // 0=pending, 1=running, 2=success, 3=failed
}

// spinnerAnimation 实现简单的加载动画
type spinnerAnimation struct {
	frames []string
	fps    time.Duration
}

// 默认加载动画
var defaultSpinnerAnimation = spinnerAnimation{
	frames: theme.GetSpinnerFrames(),
	fps:    time.Second / 10,
}

// StepStartMsg 步骤开始消息
type StepStartMsg struct {
	Step        int
	Description string
}

// StepCompleteMsg 步骤完成消息
type StepCompleteMsg struct {
	Step    int
	Success bool
	Output  string
	Error   error
}

// AllCompleteMsg 所有步骤完成消息
type AllCompleteMsg struct {
	Success bool
}

// NewProgressModel 创建新的进度模型
func NewProgressModel(commands []CommandInfo) *ProgressModel {
	return &ProgressModel{
		commands:    commands,
		currentStep: 0,
		total:       len(commands),
		status:      i18n.T("progress.preparing"),
		isCompleted: false,
		results:     make([]string, len(commands)),
		executing:   false,
		showSpinner: true,
		spinner:     defaultSpinnerAnimation,
		stepStatus:  make([]int, len(commands)),
		width:       progressDefaultWidth,
		height:      progressDefaultHeight,
	}
}

// NewProgressModelWithoutSpinner 创建不带spinner的进度模型
func NewProgressModelWithoutSpinner(commands []CommandInfo) *ProgressModel {
	model := NewProgressModel(commands)
	model.showSpinner = false
	return model
}

// tickMsg 动画帧消息
type tickMsg time.Time

// Init 初始化
func (m *ProgressModel) Init() tea.Cmd {
	if m.showSpinner {
		return tea.Batch(
			m.executeNextCommand(),
			m.tickCmd(),
		)
	}
	return m.executeNextCommand()
}

// tickCmd 帧更新命令
func (m *ProgressModel) tickCmd() tea.Cmd {
	return tea.Tick(m.spinner.fps, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update 更新状态
func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 终端尺寸变化时更新尺寸,进度条按宽度自适应,帮助栏按高度判定
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		// 如果已经完成（成功或失败），任何按键都退出
		if m.isCompleted {
			return m, tea.Quit
		}
		// 如果正在执行中，只允许 q 或 Ctrl+C 退出
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		if m.showSpinner {
			// 更新加载动画帧
			m.frame = (m.frame + 1) % len(m.spinner.frames)
			if m.isCompleted {
				return m, tea.Tick(time.Second, func(time.Time) tea.Msg {
					return tea.Quit()
				})
			}
			return m, m.tickCmd()
		}
		return m, nil

	case StepStartMsg:
		m.currentStep = msg.Step
		m.status = fmt.Sprintf(i18n.T("progress.executing"), msg.Description)
		m.executing = true
		if len(m.stepStatus) > msg.Step {
			m.stepStatus[msg.Step] = 1 // 标记为运行中
		}
		// 开始执行命令
		return m, m.executeCommand(msg.Step)

	case StepCompleteMsg:
		m.results[msg.Step] = msg.Output
		if msg.Success {
			m.status = fmt.Sprintf(i18n.T("progress.completed"), m.commands[msg.Step].Description)
			if len(m.stepStatus) > msg.Step {
				m.stepStatus[msg.Step] = 2 // 标记为成功
			}
			m.currentStep = msg.Step + 1 // 更新到下一步
			// 继续下一个命令
			if msg.Step+1 < m.total {
				return m, m.executeNextCommand()
			} else {
				// 所有命令完成
				return m, func() tea.Msg { return AllCompleteMsg{Success: true} }
			}
		} else {
			// 命令失败 - 收集详细的错误信息
			m.hasError = true

			// 构建详细的错误消息
			errorMsg := fmt.Sprintf(i18n.T("progress.step.failed"), msg.Step+1, msg.Error.Error())

			// 如果有命令输出，添加到错误信息中
			if strings.TrimSpace(msg.Output) != "" {
				errorMsg += "\n" + fmt.Sprintf(i18n.T("progress.output"), strings.TrimSpace(msg.Output))
			}

			// 添加命令信息
			if msg.Step < len(m.commands) {
				cmd := m.commands[msg.Step]
				errorMsg += "\n" + fmt.Sprintf(i18n.T("progress.command"), cmd.Command, strings.Join(cmd.Args, " "))
			}

			m.errorMessage = errorMsg
			m.status = fmt.Sprintf(i18n.T("progress.failed"), m.commands[msg.Step].Description)
			if len(m.stepStatus) > msg.Step {
				m.stepStatus[msg.Step] = 3 // 标记为失败
			}
			return m, func() tea.Msg { return AllCompleteMsg{Success: false} }
		}

	case AllCompleteMsg:
		m.isCompleted = true
		m.executing = false
		if msg.Success {
			m.status = i18n.T("success.operation.complete")
		}
		// 减少等待时间，让用户更快看到摘要
		return m, tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
			return tea.Quit()
		})
	}

	return m, nil
}

// View 渲染视图
func (m *ProgressModel) View() tea.View {
	var s strings.Builder

	// 标题 - 去掉多余的边距和空行,超宽截断避免窄屏折行
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true)
	s.WriteString(titleStyle.Render(ansi.Truncate(i18n.T("ui.executing.commands"), max(m.width, 0), "")))
	s.WriteString("\n")

	// 进度条
	progress := float64(m.currentStep) / float64(m.total)
	if m.isCompleted {
		progress = 1.0
	}
	// 进度条行固定开销(标签 + 括号 + 百分比 + 计数)随语言/进度变化,
	// 先渲染前后缀实测宽度再反推条长:宽屏封顶 40,窄屏收缩条体,
	// 空间耗尽时条体为 0(仅保留前后缀),行不折行
	percentText := fmt.Sprintf("%.0f%%", progress*100)
	rowPrefix := theme.InfoStyle.Render(i18n.T("ui.progress"))
	rowSuffix := fmt.Sprintf("] %s (%d/%d)", percentText, m.currentStep, m.total)
	barWidth := min(progressBarMaxWidth, max(m.width-lipgloss.Width(rowPrefix)-lipgloss.Width(rowSuffix)-2, 0))
	filled := int(progress * float64(barWidth))

	var progressBar strings.Builder
	for i := range barWidth {
		if i < filled {
			progressBar.WriteString("█")
		} else {
			progressBar.WriteString("░")
		}
	}

	s.WriteString(fmt.Sprintf("%s [%s%s\n",
		rowPrefix,
		theme.ProgressStyle.Render(progressBar.String()),
		rowSuffix))

	// 当前状态（用特殊字符，避免 emoji 在不同终端渲染不一致）
	statusIcon := "○"
	if m.isCompleted {
		if m.hasError {
			statusIcon = "✗"
		} else {
			statusIcon = "✓"
		}
	} else if m.executing {
		statusIcon = "▶"
	}

	// 状态行固定开销(标签 + 图标)实测,剩余空间给状态文本,避免窄屏折行
	statusPrefix := fmt.Sprintf("%s %s ", theme.InfoStyle.Render(i18n.T("ui.status")), statusIcon)
	statusMaxWidth := max(m.width-lipgloss.Width(statusPrefix), 0)
	statusText := ansi.Truncate(m.status, statusMaxWidth, "...")

	s.WriteString(fmt.Sprintf("%s%s\n",
		statusPrefix,
		lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(statusText)))

	// 步骤列表顶部:状态行与步骤列表之间以全宽分隔线代替空行
	// (≥6 行终端渲染,极矮终端保留空行零开销)
	if m.height >= form.HelpBarMinTermHeight {
		s.WriteString(theme.GetHorizontalRule(m.width))
		s.WriteString("\n")
	} else {
		s.WriteString("\n")
	}
	for i, cmd := range m.commands {
		var icon string
		var iconStyle lipgloss.Style
		var textStyle lipgloss.Style

		if len(m.stepStatus) > 0 {
			// 使用详细的步骤状态
			switch m.stepStatus[i] {
			case 0: // 等待
				icon = "○"
				iconStyle = lipgloss.NewStyle().Foreground(theme.MutedForeground)
				textStyle = lipgloss.NewStyle().Foreground(theme.MutedForeground)
			case 1: // 运行中
				if m.showSpinner {
					icon = m.spinner.frames[m.frame]
				} else {
					icon = "▶"
				}
				iconStyle = theme.InfoIconStyle
				textStyle = lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)
			case 2: // 成功
				icon = "✓"
				iconStyle = theme.SuccessIconStyle
				textStyle = lipgloss.NewStyle().Foreground(theme.PrimaryColor)
			case 3: // 失败
				icon = "✗"
				iconStyle = theme.ErrorIconStyle
				textStyle = lipgloss.NewStyle().Foreground(theme.PrimaryColor)
			}
		} else {
			// 使用简单的步骤状态（向后兼容）
			if i < m.currentStep {
				icon = "✓"
				iconStyle = theme.SuccessIconStyle
				textStyle = theme.SuccessStyle
			} else if i == m.currentStep && m.executing {
				if m.showSpinner {
					icon = m.spinner.frames[m.frame]
				} else {
					icon = "▶"
				}
				iconStyle = theme.InfoIconStyle
				textStyle = theme.InfoStyle
			} else if i == m.currentStep && m.hasError {
				icon = "✗"
				iconStyle = theme.ErrorIconStyle
				textStyle = theme.ErrorStyle
			} else {
				icon = "○"
				iconStyle = lipgloss.NewStyle().Foreground(theme.MutedForeground)
				textStyle = lipgloss.NewStyle().Foreground(theme.MutedForeground)
			}
		}

		// 步骤行固定开销(缩进 + 图标 + 空格)实测,剩余空间给步骤文本,避免窄屏折行
		stepMaxWidth := max(m.width-2-lipgloss.Width(icon)-1, 0)
		stepText := ansi.Truncate(fmt.Sprintf(i18n.T("ui.step"), i+1, cmd.Description), stepMaxWidth, "...")
		s.WriteString(fmt.Sprintf("  %s %s\n",
			iconStyle.Render(icon),
			textStyle.Render(stepText)))
	}

	// 执行中底部:帮助栏前插入全宽分隔线(条件同帮助栏);
	// 完成时保留上方 ui.exiting.* 提示
	if !m.isCompleted && m.height >= form.HelpBarMinTermHeight {
		s.WriteString(theme.GetHorizontalRule(m.width))
		s.WriteString("\n")
		s.WriteString(form.RenderHelpBar(form.ProgressHelpKeys(), m.width))
	}

	// 完成时的提示,超宽截断避免窄屏折行
	if m.isCompleted {
		s.WriteString("\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(theme.MutedForeground).
			Italic(true)

		hint := i18n.T("ui.exiting.error")
		if !m.hasError {
			hint = i18n.T("ui.exiting.success")
		}
		s.WriteString(hintStyle.Render(ansi.Truncate(hint, max(m.width, 0), "")))
	}

	return tea.NewView(s.String())
}

// executeNextCommand 执行下一个命令
func (m *ProgressModel) executeNextCommand() tea.Cmd {
	if m.currentStep >= len(m.commands) {
		return func() tea.Msg { return AllCompleteMsg{Success: true} }
	}

	cmd := m.commands[m.currentStep]
	step := m.currentStep

	return func() tea.Msg {
		// 发送开始消息
		return StepStartMsg{
			Step:        step,
			Description: cmd.Description,
		}
	}
}

// executeCommand 执行具体的命令
func (m *ProgressModel) executeCommand(step int) tea.Cmd {
	cmd := m.commands[step]

	return func() tea.Msg {
		// 创建带上下文的命令执行
		execCmd := exec.Command(cmd.Command, cmd.Args...)

		// 设置工作目录（如果需要）
		// execCmd.Dir = workingDir

		// 执行命令并捕获输出
		output, err := execCmd.CombinedOutput()

		// 如果命令不存在，提供更有用的错误信息
		if err != nil {
			if execCmd.ProcessState == nil {
				// 命令启动失败（通常是命令不存在）
				enhancedErr := fmt.Errorf("failed to start command '%s': %w (make sure the command is installed and in PATH)", cmd.Command, err)
				return StepCompleteMsg{
					Step:    step,
					Success: false,
					Output:  string(output),
					Error:   enhancedErr,
				}
			}
		}

		return StepCompleteMsg{
			Step:    step,
			Success: err == nil,
			Output:  string(output),
			Error:   err,
		}
	}
}

// RunMultipleCommands 使用 Bubble Tea 执行多个命令
func RunMultipleCommands(commands []CommandInfo) error {
	model := NewProgressModel(commands)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run progress UI: %w", err)
	}

	// 检查最终状态并在程序退出后显示摘要
	if progressModel, ok := finalModel.(*ProgressModel); ok {
		if progressModel.hasError {
			// 在程序退出后显示错误摘要，这样不会被清除
			printExecutionSummary(progressModel)
			return fmt.Errorf("command execution failed")
		} else {
			// 成功时也显示摘要
			printExecutionSummary(progressModel)
		}
	}

	return nil
}

// printExecutionSummary 在程序退出后打印执行摘要
func printExecutionSummary(model *ProgressModel) {
	if model.hasError {
		// 显示失败的步骤信息
		if model.currentStep < len(model.commands) {
			failedCmd := model.commands[model.currentStep]
			fmt.Printf("%s", fmt.Sprintf(i18n.T("cmd.failed.step"), model.currentStep+1, failedCmd.Description)+"\n")
			fmt.Printf(i18n.T("cmd.command")+" %s %s\n", failedCmd.Command, strings.Join(failedCmd.Args, " "))
		}

		// 显示详细的错误信息
		if model.errorMessage != "" {
			fmt.Println()
			// 过滤重复行:首行为「步骤失败」行、末行为「命令」行,均已在上方错误摘要展示。
			// 按结构跳过首末非空行,不依赖翻译模板前缀,避免本地化差异误伤命令输出。
			lines := make([]string, 0, 8)
			for line := range strings.SplitSeq(model.errorMessage, "\n") {
				if trimmed := strings.TrimSpace(line); trimmed != "" {
					lines = append(lines, trimmed)
				}
			}
			for i, line := range lines {
				if i == 0 || i == len(lines)-1 {
					continue
				}
				errorStyle := lipgloss.NewStyle().
					Foreground(theme.PrimaryColor).
					Render(line)
				fmt.Println(errorStyle)
			}
		}
	} else {
		// 成功时显示简单的成功信息
		fmt.Println(lipgloss.NewStyle().
			Foreground(theme.PrimaryColor).
			Bold(true).
			Render(i18n.T("success.operation.complete")))
	}

	fmt.Println() // 结尾空行
}
