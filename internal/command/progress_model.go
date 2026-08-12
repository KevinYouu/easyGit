package command

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// 进度条宽度常量
const (
	progressBarMaxWidth    = 40 // 进度条最大宽度
	progressBarMinWidth    = 10 // 进度条最小宽度
	progressBarInset       = 8  // 进度条两侧预留
	progressStatusInset    = 12 // 状态文本前侧预留(标签+图标)
	progressStatusMinWidth = 10 // 状态文本最小宽度
	progressDefaultWidth   = 80 // 默认终端宽度(未收到尺寸消息前)
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
		status:      "Preparing...",
		isCompleted: false,
		results:     make([]string, len(commands)),
		executing:   false,
		showSpinner: true,
		spinner:     defaultSpinnerAnimation,
		stepStatus:  make([]int, len(commands)),
		width:       progressDefaultWidth,
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
		// 终端尺寸变化时更新宽度,进度条按宽度自适应
		m.width = msg.Width
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
		m.status = fmt.Sprintf("Executing: %s", msg.Description)
		m.executing = true
		if len(m.stepStatus) > msg.Step {
			m.stepStatus[msg.Step] = 1 // 标记为运行中
		}
		// 开始执行命令
		return m, m.executeCommand(msg.Step)

	case StepCompleteMsg:
		m.results[msg.Step] = msg.Output
		if msg.Success {
			m.status = fmt.Sprintf("Completed: %s", m.commands[msg.Step].Description)
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
			errorMsg := fmt.Sprintf("Step %d failed: %s", msg.Step+1, msg.Error.Error())

			// 如果有命令输出，添加到错误信息中
			if strings.TrimSpace(msg.Output) != "" {
				errorMsg += fmt.Sprintf("\nOutput: %s", strings.TrimSpace(msg.Output))
			}

			// 添加命令信息
			if msg.Step < len(m.commands) {
				cmd := m.commands[msg.Step]
				errorMsg += fmt.Sprintf("\nCommand: %s %s", cmd.Command, strings.Join(cmd.Args, " "))
			}

			m.errorMessage = errorMsg
			m.status = fmt.Sprintf("Failed: %s", m.commands[msg.Step].Description)
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

	// 标题 - 去掉多余的边距和空行
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true)
	s.WriteString(titleStyle.Render(i18n.T("ui.executing.commands")))
	s.WriteString("\n")

	// 进度条
	progress := float64(m.currentStep) / float64(m.total)
	if m.isCompleted {
		progress = 1.0
	}
	// 进度条宽度自适应:min(40, max(终端宽度-8, 10)),窄屏不折行
	barWidth := min(progressBarMaxWidth, max(m.width-progressBarInset, progressBarMinWidth))
	filled := int(progress * float64(barWidth))

	var progressBar strings.Builder
	for i := range barWidth {
		if i < filled {
			progressBar.WriteString("█")
		} else {
			progressBar.WriteString("░")
		}
	}

	s.WriteString(fmt.Sprintf("%s [%s] %s (%d/%d)\n",
		theme.InfoStyle.Render(i18n.T("ui.progress")),
		theme.ProgressStyle.Render(progressBar.String()),
		lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(fmt.Sprintf("%.0f%%", progress*100)),
		m.currentStep,
		m.total))

	// 当前状态
	statusIcon := "⏳"
	if m.isCompleted {
		if m.hasError {
			statusIcon = "❌"
		} else {
			statusIcon = "✅"
		}
	} else if m.executing {
		statusIcon = "⚡"
	}

	// 状态文本超宽时按显示宽度截断,避免窄屏折行
	statusMaxWidth := max(m.width-progressStatusInset, progressStatusMinWidth)
	statusText := ansi.Truncate(m.status, statusMaxWidth, "...")

	s.WriteString(fmt.Sprintf("%s %s %s\n",
		theme.InfoStyle.Render(i18n.T("ui.status")),
		statusIcon,
		lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(statusText)))

	// 显示当前执行的步骤列表
	s.WriteString("\n")
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

		s.WriteString(fmt.Sprintf("  %s %s\n",
			iconStyle.Render(icon),
			textStyle.Render(fmt.Sprintf(i18n.T("ui.step"), i+1, cmd.Description))))
	}

	// 完成时的提示
	if m.isCompleted {
		s.WriteString("\n")
		hintStyle := lipgloss.NewStyle().
			Foreground(theme.MutedForeground).
			Italic(true)

		if m.hasError {
			s.WriteString(hintStyle.Render(i18n.T("ui.exiting.error")))
		} else {
			s.WriteString(hintStyle.Render(i18n.T("ui.exiting.success")))
		}
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
			errorLines := strings.SplitSeq(model.errorMessage, "\n")
			for line := range errorLines {
				if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Step ") && !strings.HasPrefix(line, "Command:") {
					errorStyle := lipgloss.NewStyle().
						Foreground(theme.PrimaryColor).
						Render(strings.TrimSpace(line))
					fmt.Println(errorStyle)
				}
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
