package theme

import (
	"image/color"
	"os"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// 设计令牌 - Shadcn / Neutral 双色板(暗/亮)
// 参考: docs/ui-shadcn-tui-refactor.md

// ─── 主题模式 ────────────────────────────────────────────────────────────────

// Mode 主题模式:自动(跟随终端背景)/强制深色/强制浅色
// 与 internal/config 的 ThemeAuto/ThemeDark/ThemeLight 字符串保持一致
// (theme_test.TestModesAlignWithConfig 防漂移)
type Mode string

const (
	ModeAuto  Mode = "auto"
	ModeDark  Mode = "dark"
	ModeLight Mode = "light"
)

// ValidMode 校验主题模式是否合法
func ValidMode(mode Mode) bool {
	return mode == ModeAuto || mode == ModeDark || mode == ModeLight
}

// palette 单套色板:中性色令牌(暗/亮各一套,语义色两模式共用)。
// lipgloss v2 中 lipgloss.Color 为函数(返回 color.Color 接口),令牌以接口存储。
type palette struct {
	isDark         bool
	primary        color.Color
	muted          color.Color
	border         color.Color
	selectionBg    color.Color
	selectionFg    color.Color
	selectionMuted color.Color // 选中/聚焦行上的弱化前景色
	diffAddedBg    color.Color
	diffRemovedBg  color.Color
}

// darkPalette Neutral Dark 色系(默认,兼容历史固定暗色行为)
var darkPalette = palette{
	isDark:         true,
	primary:        lipgloss.Color("#fafafa"), // foreground / primary (Neutral 50)
	muted:          lipgloss.Color("#a3a3a3"), // muted text (Neutral 400)
	border:         lipgloss.Color("#404040"), // border / input (Neutral 700)
	selectionBg:    lipgloss.Color("#404040"), // selection background (Neutral 700)
	selectionFg:    lipgloss.Color("#fafafa"), // selection foreground (Neutral 50)
	selectionMuted: lipgloss.Color("#d4d4d4"), // 深灰底上的弱化前景 (Neutral 300)
	diffAddedBg:    lipgloss.Color("#1a3a1a"), // 暗绿色背景
	diffRemovedBg:  lipgloss.Color("#4a1515"), // 暗红色背景
}

// lightPalette Neutral Light 色系(浅色终端)
var lightPalette = palette{
	isDark:         false,
	primary:        lipgloss.Color("#18181b"), // foreground / primary (Neutral 900)
	muted:          lipgloss.Color("#737373"), // muted text (Neutral 500)
	border:         lipgloss.Color("#e4e4e7"), // border / input (Neutral 200)
	selectionBg:    lipgloss.Color("#e4e4e7"), // selection background (Neutral 200)
	selectionFg:    lipgloss.Color("#18181b"), // selection foreground (Neutral 900)
	selectionMuted: lipgloss.Color("#525252"), // 浅灰底上的弱化前景 (Neutral 600)
	diffAddedBg:    lipgloss.Color("#dcfce7"), // 浅绿色背景 (Green 100)
	diffRemovedBg:  lipgloss.Color("#fee2e2"), // 浅红色背景 (Red 100)
}

var (
	currentMode Mode    = ModeDark // 当前生效模式(auto 已解析为 dark/light)
	current     palette = darkPalette
)

// 中性色令牌:调用时求值,ApplyMode 切换后全包自动生效
var (
	PrimaryColor    = current.primary
	MutedForeground = current.muted
	BorderColor     = current.border
	SelectionBg     = current.selectionBg
	SelectionFg     = current.selectionFg
	SelectionMuted  = current.selectionMuted
	DiffAddedBg     = current.diffAddedBg
	DiffRemovedBg   = current.diffRemovedBg
)

// 语义颜色:两模式一致,不随 ApplyMode 变化
var (
	SuccessColor = lipgloss.Color("#10b981") // Emerald 500
	ErrorColor   = lipgloss.Color("#ef4444") // Shadcn Red
	WarningColor = lipgloss.Color("#f59e0b") // Amber 500
	InfoColor    = lipgloss.Color("#3b82f6") // Blue 500
)

func init() {
	applyPalette(darkPalette)
}

// ─── 模式切换 ────────────────────────────────────────────────────────────────

// DetectDarkBackground 探测终端背景色:true=深色,false=浅色。
// 查询失败/非 TTY 时回退深色(与历史固定暗色行为一致)。
func DetectDarkBackground() bool {
	return lipgloss.HasDarkBackground(os.Stdin, os.Stdout)
}

// ResolveMode 将模式解析为具体深/浅:auto 按终端背景检测,非法值同 auto
func ResolveMode(mode Mode) Mode {
	if !ValidMode(mode) || mode == ModeAuto {
		if DetectDarkBackground() {
			return ModeDark
		}
		return ModeLight
	}
	return mode
}

// ApplyMode 应用主题模式(auto 自动解析),返回实际生效模式
func ApplyMode(mode Mode) Mode {
	resolved := ResolveMode(mode)
	currentMode = resolved
	if resolved == ModeLight {
		applyPalette(lightPalette)
	} else {
		applyPalette(darkPalette)
	}
	return resolved
}

// CurrentMode 返回当前生效模式(已解析为 dark/light)
func CurrentMode() Mode { return currentMode }

// IsDark 当前是否为深色模式
func IsDark() bool { return current.isDark }

// applyPalette 应用色板到导出令牌并重建样式
func applyPalette(p palette) {
	current = p
	PrimaryColor = p.primary
	MutedForeground = p.muted
	BorderColor = p.border
	SelectionBg = p.selectionBg
	SelectionFg = p.selectionFg
	SelectionMuted = p.selectionMuted
	DiffAddedBg = p.diffAddedBg
	DiffRemovedBg = p.diffRemovedBg
	rebuildStyles()
}

// 语义样式 - 纯前景色，无背景（终端友好）。
// 样式在构建时捕获色值,由 applyPalette → rebuildStyles 在模式切换时重建。
var (
	ErrorStyle       lipgloss.Style // 文本样式:统一使用 PrimaryColor
	SuccessStyle     lipgloss.Style
	WarningStyle     lipgloss.Style
	InfoStyle        lipgloss.Style
	ErrorIconStyle   lipgloss.Style // 图标样式:使用语义颜色
	SuccessIconStyle lipgloss.Style
	WarningIconStyle lipgloss.Style
	InfoIconStyle    lipgloss.Style
	SpinnerStyle     lipgloss.Style
	ProgressStyle    lipgloss.Style
)

// rebuildStyles 重建包级样式(样式构建时捕获色值,模式切换后必须重建)
func rebuildStyles() {
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

	SpinnerStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	ProgressStyle = lipgloss.NewStyle().
		Foreground(PrimaryColor)
}

// ─── huh 表单主题 ────────────────────────────────────────────────────────────

// GetCompactTheme 返回紧凑主题（主要使用场景）
// huh v2 主题为 Theme 接口,ThemeBase(isDark) 跟随当前生效模式(ApplyMode 已解析)
func GetCompactTheme() huh.Theme {
	return huh.ThemeFunc(func(_ bool) *huh.Styles {
		theme := huh.ThemeBase(current.isDark)

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

// ─── Spinner 帧 ──────────────────────────────────────────────────────────────

// GetSpinnerFrames 获取默认加载动画帧
func GetSpinnerFrames() []string {
	return []string{
		"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
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
