package theme

import (
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// 设计令牌 - Shadcn / Neutral Dark 色系
// 参考: docs/ui-shadcn-tui-refactor.md
var (
	// 核心文本颜色
	PrimaryColor    = lipgloss.Color("#fafafa") // foreground / primary (Neutral 50)
	MutedForeground = lipgloss.Color("#a3a3a3") // muted text (Neutral 400)

	// 边框 / 输入 / 选中背景
	BorderColor = lipgloss.Color("#404040") // border / input (Neutral 700)
	SelectionBg = lipgloss.Color("#404040") // selection background (Neutral 700)
	SelectionFg = lipgloss.Color("#fafafa") // selection foreground (Neutral 50)

	// 语义颜色
	SuccessColor = lipgloss.Color("#10b981") // Emerald 500
	ErrorColor   = lipgloss.Color("#ef4444") // Shadcn Red
	WarningColor = lipgloss.Color("#f59e0b") // Amber 500
	InfoColor    = lipgloss.Color("#3b82f6") // Blue 500

	// Diff 颜色
	DiffAddedBg   = lipgloss.Color("#1a3a1a") // 暗绿色背景
	DiffRemovedBg = lipgloss.Color("#4a1515") // 暗红色背景
)

// 语义样式 - 纯前景色，无背景（终端友好）
var (
	// 文本样式：统一使用 PrimaryColor
	ErrorStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor)

	// 图标样式：使用语义颜色
	ErrorIconStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	SuccessIconStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true)

	WarningIconStyle = lipgloss.NewStyle().
				Foreground(WarningColor).
				Bold(true)

	InfoIconStyle = lipgloss.NewStyle().
			Foreground(InfoColor)

	MutedStyle = lipgloss.NewStyle().
			Foreground(MutedForeground)

	SpinnerStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	ProgressStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor)
)

// ─── 原子渲染函数 ────────────────────────────────────────────────────────────

// RenderSelection 渲染选中态文本（背景高亮 + 白色前景 + ❯ 指示符）
func RenderSelection(text string) string {
	return lipgloss.NewStyle().
		Background(SelectionBg).
		Foreground(SelectionFg).
		Bold(true).
		Render("❯ " + text)
}

// RenderMuted 渲染弱化文本
func RenderMuted(text string) string {
	return MutedStyle.Render(text)
}

// RenderBadge 渲染标签，variant: "success" | "error" | "warning" | "info"
func RenderBadge(text, variant string) string {
	return lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true).
		Render(text)
}

// ─── huh 表单主题 ────────────────────────────────────────────────────────────

// GetCompactTheme 返回紧凑主题（主要使用场景）
// huh v2 主题为 Theme 接口，isDark 参数用于暗色/亮色适配（本项目固定暗色）
func GetCompactTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		theme := huh.ThemeBase(true)

		// 基础容器 - 最小化
		theme.Focused.Base = lipgloss.NewStyle().
			MarginTop(0)

		theme.Blurred.Base = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			MarginTop(0)

		// 标题样式
		theme.Focused.Title = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Margin(0).
			MarginTop(0).
			PaddingTop(0)

		theme.Blurred.Title = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Margin(0).
			MarginTop(0).
			PaddingTop(0)

		// 描述样式
		theme.Focused.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		theme.Blurred.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		// 选择器样式 - Neutral Dark 选中态
		theme.Focused.SelectedOption = lipgloss.NewStyle().
			Background(SelectionBg).
			Foreground(SelectionFg).
			Bold(true).
			Padding(0, 1)

		theme.Focused.UnselectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		theme.Blurred.SelectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		theme.Blurred.UnselectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		// 输入框样式
		theme.Focused.TextInput.Cursor = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		theme.Focused.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		theme.Focused.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		// 选中指示符 - ❯ 显示宽 2(与 huh 默认 "> " 同宽,零布局位移);
		// ❯ 与选项的间距由 PaddingRight 布局产生,不用字面空格
		theme.Focused.SelectSelector = lipgloss.NewStyle().
			SetString("❯").
			PaddingRight(1).
			Foreground(PrimaryColor).
			Bold(true)

		theme.Focused.MultiSelectSelector = lipgloss.NewStyle().
			SetString("❯").
			PaddingRight(1).
			Foreground(PrimaryColor).
			Bold(true)

		// 错误样式
		theme.Focused.ErrorIndicator = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

		theme.Focused.ErrorMessage = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Italic(true)

		// 隐藏帮助信息
		hidden := lipgloss.NewStyle().Width(0).Height(0)
		theme.Help.Ellipsis = hidden
		theme.Help.ShortKey = hidden
		theme.Help.ShortDesc = hidden
		theme.Help.ShortSeparator = hidden
		theme.Help.FullKey = hidden
		theme.Help.FullDesc = hidden
		theme.Help.FullSeparator = hidden

		return theme
	})
}

// GetCustomTheme 返回标准主题（较大场景使用）
func GetCustomTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		theme := huh.ThemeBase(true)

		theme.Focused.Base = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			MarginBottom(1)

		theme.Blurred.Base = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderColor).
			Padding(1, 2).
			MarginBottom(1)

		theme.Focused.Title = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Padding(0, 0, 1, 0)

		theme.Blurred.Title = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 0, 1, 0)

		theme.Focused.Description = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		theme.Blurred.Description = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		theme.Focused.SelectedOption = lipgloss.NewStyle().
			Background(SelectionBg).
			Foreground(SelectionFg).
			Bold(true).
			Padding(0, 1)

		theme.Focused.UnselectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		theme.Blurred.SelectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		theme.Blurred.UnselectedOption = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Padding(0, 1)

		theme.Focused.TextInput.Cursor = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		theme.Focused.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		theme.Focused.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		// 选中指示符 - ❯ 显示宽 2(与 huh 默认 "> " 同宽,零布局位移);
		// ❯ 与选项的间距由 PaddingRight 布局产生,不用字面空格
		theme.Focused.SelectSelector = lipgloss.NewStyle().
			SetString("❯").
			PaddingRight(1).
			Foreground(PrimaryColor).
			Bold(true)

		theme.Focused.MultiSelectSelector = lipgloss.NewStyle().
			SetString("❯").
			PaddingRight(1).
			Foreground(PrimaryColor).
			Bold(true)

		theme.Focused.ErrorIndicator = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

		theme.Focused.ErrorMessage = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Italic(true)

		return theme
	})
}

// GetMultiInputTheme 返回单页多输入表单主题:所有字段统一卡片容器,
// 焦点/模糊同尺寸(切换焦点零布局跳变),焦点边框高亮、模糊弱化;
// 提示符 ❯ 与列表选中指示符统一。仅 MultiInput 使用,不污染其他表单。
func GetMultiInputTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		st := huh.ThemeBase(true)

		// 统一卡片:边框 + 左右留白。Focused 与 Blurred 完全同尺寸,
		// 焦点切换仅换边框颜色,高度不跳变
		card := lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			Padding(0, 1)

		st.Focused.Base = card.BorderForeground(PrimaryColor)
		st.Blurred.Base = card.BorderForeground(BorderColor)

		st.Focused.Title = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Blurred.Title = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Focused.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Focused.TextInput.Cursor = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Focused.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		st.Focused.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Focused.TextInput.Text = lipgloss.NewStyle().
			Foreground(PrimaryColor)

		st.Blurred.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.TextInput.Text = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		// 错误提示:红色加粗 + ✗ 图标(lipgloss SetString 作 Render 前缀)
		errorStyle := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			SetString("✗")
		st.Focused.ErrorIndicator = errorStyle
		st.Focused.ErrorMessage = errorStyle

		// 隐藏帮助信息(与 GetCompactTheme 一致,帮助栏由 form 包统一渲染)
		hidden := lipgloss.NewStyle().Width(0).Height(0)
		st.Help.Ellipsis = hidden
		st.Help.ShortKey = hidden
		st.Help.ShortDesc = hidden
		st.Help.ShortSeparator = hidden
		st.Help.FullKey = hidden
		st.Help.FullDesc = hidden
		st.Help.FullSeparator = hidden

		return st
	})
}

// GetCompactMultiInputTheme 返回紧凑多输入表单主题(矮终端友好):
// 无边框卡片、字段间无空行(FieldSeparator 单换行)、标题与输入同行
// (配合 Input.Inline),5 字段 + 预览 + 帮助 ≈ 9 行;焦点切换仅标题色变化,
// 高度零跳变。错误提示与帮助隐藏同 GetMultiInputTheme。
// 仅紧凑 MultiInput(配置中心版本号上限)使用,不影响其他表单。
func GetCompactMultiInputTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		st := huh.ThemeBase(true)

		// 无边框、无内边距:每字段一行(Inline 模式)
		st.FieldSeparator = lipgloss.NewStyle().SetString("\n")
		st.Focused.Base = lipgloss.NewStyle()
		st.Blurred.Base = lipgloss.NewStyle()
		st.Focused.Card = st.Focused.Base
		st.Blurred.Card = st.Blurred.Base

		// 预览行(Note):弱化色,与字段区分
		st.Focused.NoteTitle = lipgloss.NewStyle().
			Foreground(MutedForeground)
		st.Blurred.NoteTitle = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Focused.Title = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Blurred.Title = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Focused.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.Description = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Focused.TextInput.Cursor = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Focused.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		st.Focused.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

		st.Focused.TextInput.Text = lipgloss.NewStyle().
			Foreground(PrimaryColor)

		st.Blurred.TextInput.Prompt = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.TextInput.Text = lipgloss.NewStyle().
			Foreground(MutedForeground)

		st.Blurred.TextInput.Placeholder = lipgloss.NewStyle().
			Foreground(MutedForeground).
			Italic(true)

		// 错误提示:红色加粗 + ✗ 图标
		errorStyle := lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true).
			SetString("✗")
		st.Focused.ErrorIndicator = errorStyle
		st.Focused.ErrorMessage = errorStyle

		// 隐藏帮助信息(帮助栏由 form 包统一渲染)
		hidden := lipgloss.NewStyle().Width(0).Height(0)
		st.Help.Ellipsis = hidden
		st.Help.ShortKey = hidden
		st.Help.ShortDesc = hidden
		st.Help.ShortSeparator = hidden
		st.Help.FullKey = hidden
		st.Help.FullDesc = hidden
		st.Help.FullSeparator = hidden

		return st
	})
}

// ─── Spinner 帧 ──────────────────────────────────────────────────────────────

// GetSpinnerFrames 获取默认加载动画帧
func GetSpinnerFrames() []string {
	return []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
	}
}

// GetPulseSpinnerFrames 获取脉冲动画帧
func GetPulseSpinnerFrames() []string {
	return []string{
		"●", "◐", "◑", "◒", "◓", "◔", "◕", "◖", "◗", "◘",
	}
}

// GetDotsSpinnerFrames 获取点状动画帧
func GetDotsSpinnerFrames() []string {
	return []string{
		"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷",
	}
}

// GetArrowSpinnerFrames 获取箭头旋转动画帧
func GetArrowSpinnerFrames() []string {
	return []string{
		"→", "↘", "↓", "↙", "←", "↖", "↑", "↗",
	}
}

// GetSpinnerStyle 获取加载动画样式
func GetSpinnerStyle() lipgloss.Style {
	return SpinnerStyle
}

// ─── 工具函数 ─────────────────────────────────────────────────────────────────

// GetHorizontalRule 获取分隔线
func GetHorizontalRule(width int) string {
	return lipgloss.NewStyle().
		Foreground(BorderColor).
		Width(width).
		Render(strings.Repeat("─", width))
}

// GetStatusIcon 根据状态获取图标
func GetStatusIcon(status string) string {
	icons := map[string]string{
		"success":  "✓",
		"error":    "✗",
		"warning":  "⚠",
		"info":     "ℹ",
		"pending":  "○",
		"complete": "✓",
		"running":  "▶",
	}

	if icon, exists := icons[status]; exists {
		return icon
	}
	return "•"
}

// GetProgressBarStyle 获取进度条样式
func GetProgressBarStyle() lipgloss.Style {
	return ProgressStyle
}
