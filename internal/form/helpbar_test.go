package form

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/charmbracelet/x/ansi"
)

// RenderHelpBar 单行帮助栏:徽章渲染、ANSI 剥离后单行不折行、超宽截断、空键位。
func TestRenderHelpBar(t *testing.T) {
	keys := []HelpKey{
		{Key: "↑/↓", Action: "导航"},
		{Key: "Enter", Action: "确认"},
		{Key: "Esc", Action: "取消"},
	}

	t.Run("键位徽章与动作文本均出现", func(t *testing.T) {
		bar := RenderHelpBar(keys, 80)
		plain := ansi.Strip(bar)
		for _, want := range []string{"↑/↓", "Enter", "Esc", "导航", "确认", "取消"} {
			if !strings.Contains(plain, want) {
				t.Errorf("帮助栏缺少 %q: %q", want, plain)
			}
		}
	})

	t.Run("ANSI 剥离后单行不折行且不越界", func(t *testing.T) {
		for _, width := range []int{80, 60, 50, 40} {
			bar := RenderHelpBar(keys, width)
			plain := ansi.Strip(bar)
			if strings.Contains(plain, "\n") {
				t.Errorf("宽度 %d:帮助栏折行为多行: %q", width, plain)
			}
			if lw := lipgloss.Width(plain); lw > width {
				t.Errorf("宽度 %d:帮助栏宽 %d 越界: %q", width, lw, plain)
			}
		}
	})

	t.Run("超宽整行截断", func(t *testing.T) {
		longKeys := []HelpKey{
			{Key: "↑/↓", Action: strings.Repeat("非常长的动作说明文本", 8)},
			{Key: "Enter", Action: strings.Repeat("另一段非常长的动作说明", 8)},
		}
		bar := RenderHelpBar(longKeys, 30)
		plain := ansi.Strip(bar)
		if lw := lipgloss.Width(plain); lw > 30 {
			t.Errorf("超宽帮助栏宽 %d 越界 30: %q", lw, plain)
		}
		if !strings.HasSuffix(plain, ellipsis) {
			t.Errorf("超宽帮助栏应以省略号结尾: %q", plain)
		}
		if !utf8.ValidString(plain) {
			t.Error("超宽截断产生非法 UTF-8")
		}
	})

	t.Run("空键位返回空串", func(t *testing.T) {
		if bar := RenderHelpBar(nil, 80); bar != "" {
			t.Errorf("空键位应返回空串, got %q", bar)
		}
	})

	t.Run("负宽度返回空串", func(t *testing.T) {
		if bar := RenderHelpBar(keys, -1); bar != "" {
			t.Errorf("负宽度应返回空串, got %q", bar)
		}
	})
}

// AppendHelpBar 空视图不附加、正常视图末尾追加单行。
func TestAppendHelpBar(t *testing.T) {
	v := tea.NewView("标题行")
	appended := AppendHelpBar(v, []HelpKey{{Key: "q", Action: "退出"}}, 80)
	plain := ansi.Strip(appended.Content)
	lines := strings.Split(plain, "\n")
	if len(lines) != 2 || strings.TrimSpace(lines[1]) == "" {
		t.Errorf("附加后应为 2 行且末行非空: %q", plain)
	}
	if !strings.Contains(lines[1], "退出") {
		t.Errorf("末行应含动作说明: %q", lines[1])
	}

	// 空内容不附加
	empty := AppendHelpBar(tea.NewView(""), []HelpKey{{Key: "q", Action: "退出"}}, 80)
	if empty.Content != "" {
		t.Errorf("空视图不应附加帮助栏: %q", empty.Content)
	}
	// 空键位不附加
	noKeys := AppendHelpBar(v, nil, 80)
	if noKeys.Content != v.Content {
		t.Errorf("空键位不应附加帮助栏")
	}
}

// sgrParams 提取字符串中首个 SGR 序列的参数列表,用于断言样式层次
func sgrParams(s string) []string {
	start := strings.Index(s, "\x1b[")
	if start < 0 {
		return nil
	}
	end := strings.Index(s[start:], "m")
	if end < 0 {
		return nil
	}
	return strings.Split(s[start+2:start+end], ";")
}

// OptionLabel 单行选项标签:名称亮色加粗 + 说明灰色;说明为空仅名称。
func TestOptionLabel(t *testing.T) {
	t.Run("有说明:名称与说明同一行且层次分明", func(t *testing.T) {
		label := OptionLabel("soft", "保留工作区更改")
		plain := ansi.Strip(label)
		if !strings.Contains(plain, "soft") || !strings.Contains(plain, "保留工作区更改") {
			t.Errorf("标签应同时含名称与说明: %q", plain)
		}
		if strings.Contains(plain, "\n") {
			t.Errorf("标签折行为多行: %q", plain)
		}
		// 名称段加粗、说明段不加粗(解析 SGR 参数,lipgloss v2 合并且色值含数字 1)
		nameParams := sgrParams(label)
		if !slices.Contains(nameParams, "1") {
			t.Errorf("名称应加粗: %q", label)
		}
		descStart := strings.Index(label, "\x1b[m")
		descParams := sgrParams(label[descStart+len("\x1b[m"):])
		if slices.Contains(descParams, "1") {
			t.Errorf("说明不应加粗: %q", label)
		}
		if lipgloss.Width(plain) != lipgloss.Width("soft")+1+lipgloss.Width("保留工作区更改") {
			t.Errorf("标签宽度异常: %q (%d)", plain, lipgloss.Width(plain))
		}
	})

	t.Run("无说明:返回纯名称", func(t *testing.T) {
		if got := OptionLabel("main", ""); got != "main" {
			t.Errorf("无说明应返回纯名称, got %q", got)
		}
	})

	t.Run("窄屏 60 列安全线内不折行", func(t *testing.T) {
		// 说明 ≤ 20 字时,名称 + 说明总宽不超过 60 列安全线
		label := OptionLabel("default", "推荐,等同 mixed 不传参数")
		plain := ansi.Strip(label)
		if strings.Contains(plain, "\n") {
			t.Errorf("标签折行: %q", plain)
		}
		if lw := lipgloss.Width(plain); lw > 60 {
			t.Errorf("标签宽 %d 超过 60 列安全线: %q", lw, plain)
		}
	})
}

// 极小终端(<6 行)不渲染帮助栏:经生产构造器驱动,内容不被挤压。
func TestHelpBarHiddenOnTinyTerminal(t *testing.T) {
	var selected string
	form := NewSelectForm("标题", []config.Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}, {Label: "c", Value: "c"}}, 4, &selected)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 4})
	view := ansi.Strip(m.(*Form).View().Content)
	if strings.Contains(view, "Enter") {
		t.Errorf("4 行终端不应渲染帮助栏: %q", view)
	}
	// 3 个选项 + 标题 = 4 行,内容完整不被裁剪
	for _, want := range []string{"a", "b", "c"} {
		if !strings.Contains(view, want) {
			t.Errorf("4 行终端缺少选项 %q: %q", want, view)
		}
	}
}
