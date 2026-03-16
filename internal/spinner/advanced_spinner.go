package spinner

import (
	"fmt"
	"strings"
	"time"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AdvancedSpinnerType 高级加载动画类型
type AdvancedSpinnerType int

const (
	SpinnerDefault AdvancedSpinnerType = iota
	SpinnerPulse
	SpinnerDots
	SpinnerArrow
	SpinnerProgress
)

// AdvancedSpinnerModel 高级加载动画模型
type AdvancedSpinnerModel struct {
	spinner      spinner.Model
	spinnerType  AdvancedSpinnerType
	message      string
	submessage   string
	progress     float64
	steps        []string
	currentStep  int
	done         bool
	success      bool
	err          error
	resultMsg    string
	showProgress bool
	showSteps    bool
	elapsedTime  time.Duration
	startTime    time.Time
}

// NewAdvancedSpinner 创建新的高级加载动画
func NewAdvancedSpinner(spinnerType AdvancedSpinnerType, message string) AdvancedSpinnerModel {
	s := spinner.New()

	// 根据类型设置不同的动画帧
	switch spinnerType {
	case SpinnerPulse:
		s.Spinner = spinner.Spinner{
			Frames: theme.GetPulseSpinnerFrames(),
			FPS:    time.Millisecond * 150,
		}
	case SpinnerDots:
		s.Spinner = spinner.Spinner{
			Frames: theme.GetDotsSpinnerFrames(),
			FPS:    time.Millisecond * 100,
		}
	case SpinnerArrow:
		s.Spinner = spinner.Spinner{
			Frames: theme.GetArrowSpinnerFrames(),
			FPS:    time.Millisecond * 200,
		}
	default:
		s.Spinner = spinner.Spinner{
			Frames: theme.GetSpinnerFrames(),
			FPS:    time.Millisecond * 100,
		}
	}

	s.Style = theme.GetSpinnerStyle()

	return AdvancedSpinnerModel{
		spinner:     s,
		spinnerType: spinnerType,
		message:     message,
		startTime:   time.Now(),
	}
}

// SetMessage 设置主消息
func (m *AdvancedSpinnerModel) SetMessage(message string) {
	m.message = message
}

// SetSubmessage 设置子消息
func (m *AdvancedSpinnerModel) SetSubmessage(submessage string) {
	m.submessage = submessage
}

// SetProgress 设置进度
func (m *AdvancedSpinnerModel) SetProgress(progress float64) {
	m.progress = progress
	m.showProgress = true
}

// SetSteps 设置步骤列表
func (m *AdvancedSpinnerModel) SetSteps(steps []string) {
	m.steps = steps
	m.showSteps = true
}

// NextStep 前进到下一步
func (m *AdvancedSpinnerModel) NextStep() {
	if m.currentStep < len(m.steps)-1 {
		m.currentStep++
	}
}

// SetDone 设置完成状态
func (m *AdvancedSpinnerModel) SetDone(success bool, resultMsg string, err error) {
	m.done = true
	m.success = success
	m.resultMsg = resultMsg
	m.err = err
	m.elapsedTime = time.Since(m.startTime)
}

// Init 初始化
func (m AdvancedSpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update 更新
func (m AdvancedSpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, nil
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View 渲染
func (m AdvancedSpinnerModel) View() string {
	if m.done {
		return m.renderComplete()
	}

	var content strings.Builder

	// 标题：简洁的主色标题行
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// 主消息
	content.WriteString(m.renderMainMessage())
	content.WriteString("\n")

	// 进度条（如果启用）
	if m.showProgress {
		content.WriteString(m.renderProgress())
		content.WriteString("\n")
	}

	// 步骤列表（如果启用）
	if m.showSteps {
		content.WriteString(m.renderSteps())
		content.WriteString("\n")
	}

	// 子消息
	if m.submessage != "" {
		content.WriteString(m.renderSubmessage())
		content.WriteString("\n")
	}

	return content.String()
}

// renderHeader 渲染标题 - 简洁单行，无背景块
func (m AdvancedSpinnerModel) renderHeader() string {
	return lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true).
		Render(i18n.T("spinner.easygit.operation"))
}

// renderMainMessage 渲染主消息
func (m AdvancedSpinnerModel) renderMainMessage() string {
	return fmt.Sprintf("%s  %s",
		theme.SpinnerStyle.Render(m.spinner.View()),
		lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(m.message))
}

// renderProgress 渲染进度条
func (m AdvancedSpinnerModel) renderProgress() string {
	width := 40
	filled := int(m.progress * float64(width))

	var progressBar strings.Builder
	for i := range width {
		if i < filled {
			progressBar.WriteString("█")
		} else {
			progressBar.WriteString("░")
		}
	}

	return fmt.Sprintf("  %s %s",
		theme.ProgressStyle.Render(progressBar.String()),
		lipgloss.NewStyle().Foreground(theme.InfoColor).Bold(true).Render(fmt.Sprintf("%.1f%%", m.progress*100)))
}

// renderSteps 渲染步骤列表
func (m AdvancedSpinnerModel) renderSteps() string {
	var steps strings.Builder
	steps.WriteString(fmt.Sprintf("  %s\n", theme.RenderMuted(i18n.T("spinner.step.progress"))))

	for i, step := range m.steps {
		var icon string
		var rendered string
		if i < m.currentStep {
			icon = theme.GetStatusIcon("success")
			rendered = theme.SuccessStyle.Render(step)
		} else if i == m.currentStep {
			icon = theme.GetStatusIcon("running")
			rendered = theme.InfoStyle.Render(step)
		} else {
			icon = theme.GetStatusIcon("pending")
			rendered = theme.RenderMuted(step)
		}

		steps.WriteString(fmt.Sprintf("    %s %s\n", icon, rendered))
	}

	return steps.String()
}

// renderSubmessage 渲染子消息
func (m AdvancedSpinnerModel) renderSubmessage() string {
	return theme.RenderMuted(fmt.Sprintf("  %s", m.submessage))
}

// renderComplete 渲染完成状态 - 纯前景色，无大块背景
func (m AdvancedSpinnerModel) renderComplete() string {
	var content strings.Builder

	if m.success {
		content.WriteString(fmt.Sprintf("%s  %s",
			theme.SuccessStyle.Render(theme.GetStatusIcon("success")),
			theme.SuccessStyle.Render(i18n.T("spinner.operation.complete"))))
	} else {
		content.WriteString(fmt.Sprintf("%s  %s",
			theme.ErrorStyle.Render(theme.GetStatusIcon("error")),
			theme.ErrorStyle.Render(i18n.T("spinner.operation.failed"))))
	}
	content.WriteString("\n")

	// 结果消息
	if m.resultMsg != "" {
		content.WriteString(fmt.Sprintf("   %s\n",
			lipgloss.NewStyle().Foreground(theme.PrimaryColor).Render(m.resultMsg)))
	}

	// 错误详情
	if m.err != nil {
		content.WriteString(theme.ErrorStyle.Render(
			fmt.Sprintf("   "+i18n.T("spinner.error.details"), m.err)))
		content.WriteString("\n")
	}

	// 耗时
	content.WriteString(theme.RenderMuted(
		fmt.Sprintf("   "+i18n.T("spinner.elapsed.time"), m.elapsedTime.Round(time.Millisecond))))

	return content.String()
}
