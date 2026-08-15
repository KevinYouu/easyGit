package form

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// HelpBarMinTermHeight 终端高度低于该值时帮助栏不渲染(避免挤压内容)
const HelpBarMinTermHeight = 6

// HelpKey 帮助栏键位项:键位前缀 + 动作说明(调用方已 i18n)
type HelpKey struct {
	Key    string // "↑/↓"、"Enter"、"Space"、"Esc"、"q"
	Action string // 动作说明(调用方已 i18n)
}

// keyStyle 键位前缀样式:主色加粗文本,无背景色(终端友好,括号本身即视觉边界)。
// 函数化:样式构建时捕获色值,包级 var 会在主题切换后固化过期色。
func keyStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(theme.PrimaryColor).
		Bold(true)
}

// helpActionStyle 动作说明:弱化前景色
func helpActionStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(theme.MutedForeground) }

// 预置键位集合:各交互形态的统一帮助栏内容,全部单行紧凑
func selectHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "↑/↓", Action: i18n.T("form.help.navigate")},
		{Key: "Enter", Action: i18n.T("form.help.confirm")},
		{Key: "Esc", Action: i18n.T("form.help.cancel")},
	}
}

func multiSelectHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "↑/↓", Action: i18n.T("form.help.navigate")},
		{Key: "Space", Action: i18n.T("form.help.toggle")},
		{Key: "Enter", Action: i18n.T("form.help.confirm")},
		{Key: "Esc", Action: i18n.T("form.help.cancel")},
	}
}

func inputHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "Enter", Action: i18n.T("form.help.submit")},
		{Key: "Esc", Action: i18n.T("form.help.cancel")},
	}
}

// multiInputHelpKeys 单页多输入表单帮助栏:↑/↓ 上下导航(含 j/k),
// Enter 在非末字段为「继续」、末字段为「提交」,静态文案统一显示「继续/提交」;
// shift+tab 保留回退(帮助栏不展示,防冗长)
func multiInputHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "↑/↓", Action: i18n.T("form.help.navigate")},
		{Key: "Enter", Action: i18n.T("form.help.next")},
		{Key: "Esc", Action: i18n.T("form.help.cancel")},
	}
}

func confirmHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "←/→", Action: i18n.T("form.help.switch")},
		{Key: "Enter", Action: i18n.T("form.help.confirm")},
		{Key: "Esc", Action: i18n.T("form.help.cancel")},
	}
}

// ProgressHelpKeys 进度屏执行中的帮助栏键位(q 退出)
func ProgressHelpKeys() []HelpKey {
	return []HelpKey{
		{Key: "q", Action: i18n.T("form.help.quit")},
	}
}

// RenderHelpBar 渲染单行帮助栏:`[键位]` 前缀 + 动作说明,键值对间以 · 分隔;
// 超宽时整行 SafeTruncate 防止折行;空键位或负宽度返回空串。
func RenderHelpBar(keys []HelpKey, width int) string {
	if len(keys) == 0 || width < 0 {
		return ""
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, keyStyle().Render("["+k.Key+"]")+" "+helpActionStyle().Render(k.Action))
	}
	return SafeTruncate(strings.Join(parts, " · "), width)
}

// AppendHelpBar 将帮助栏追加到视图末尾(单行);键位为空、视图为空或宽度无效时不附加。
// width 为终端宽度:截断须以真实终端宽度为准,而非视图内容宽度
// (内容可能仅数列宽,按内容截断会把帮助栏裁没)。
func AppendHelpBar(v tea.View, keys []HelpKey, width int) tea.View {
	if len(keys) == 0 || strings.TrimSpace(v.Content) == "" {
		return v
	}
	bar := RenderHelpBar(keys, width)
	if bar == "" {
		return v
	}
	v.SetContent(v.Content + "\n" + bar)
	return v
}

// OptionLabel 构造单行选项标签:名称亮色加粗 + 说明灰色(同一行内嵌 ANSI)。
// desc 为空时仅返回名称(纯文本,与未配置说明时零变化)。
// 说明文案须保持简短(≤ 20 字),避免超宽折行占用额外行高。
func OptionLabel(name, desc string) string {
	if desc == "" {
		return name
	}
	// 内嵌段不输出 reset:lipgloss.Render 默认在段尾输出 reset,会中断 huh
	// 选中态 SelectedOption 为整行设置的背景色,导致说明文本失去高亮;
	// 剥离段尾 reset 后每段仅切换前景/字重,背景从外层序列连续贯穿整行。
	// 说明段前须显式复位字重(\x1b[22m):名称段的 Bold 状态在无 reset 时
	// 持续贯穿,只切前景的说明段会被错误加粗(说明应为常规字重)。
	return optionSegment(lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true), name) +
		" " + ansi.SGR(ansi.AttrNormalIntensity) + optionSegment(lipgloss.NewStyle().Foreground(theme.MutedForeground), desc)
}

// optionSegment 渲染单段内嵌样式并剥离段尾纯 reset 序列(见 OptionLabel 注释)
func optionSegment(style lipgloss.Style, text string) string {
	return stripTrailingReset(style.Render(text))
}

// stripTrailingReset 剥离段尾纯 reset 序列(\x1b[m 与 \x1b[0m 等价);
// 段尾非 reset 时原样返回,不依赖 lipgloss 固定输出某一种编码。
func stripTrailingReset(s string) string {
	idx := strings.LastIndex(s, "\x1b[")
	if idx < 0 {
		return s
	}
	end := strings.Index(s[idx:], "m")
	if end < 0 {
		return s
	}
	params := s[idx+2 : idx+end]
	if params != "" && params != "0" {
		return s
	}
	return s[:idx] + s[idx+end+1:]
}
