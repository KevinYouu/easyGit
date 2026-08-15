package form

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// RenderHelpBar 单行帮助栏:`[键位]` 前缀渲染、ANSI 剥离后单行不折行、超宽截断、空键位。
func TestRenderHelpBar(t *testing.T) {
	keys := []HelpKey{
		{Key: "↑/↓", Action: "导航"},
		{Key: "Enter", Action: "确认"},
		{Key: "Esc", Action: "取消"},
	}

	t.Run("键位方括号前缀与动作文本均出现", func(t *testing.T) {
		bar := RenderHelpBar(keys, 80)
		plain := ansi.Strip(bar)
		for _, want := range []string{"[↑/↓]", "[Enter]", "[Esc]", "导航", "确认", "取消"} {
			if !strings.Contains(plain, want) {
				t.Errorf("帮助栏缺少 %q: %q", want, plain)
			}
		}
	})

	t.Run("键位段样式:主色加粗无背景", func(t *testing.T) {
		bar := RenderHelpBar(keys, 80)
		seqs := sgrSeqList(bar)
		if len(seqs) < 2 {
			t.Fatalf("帮助栏应含键位/动作样式序列: %q", bar)
		}
		// 首个序列属于第一个键位段([↑/↓]):主色加粗(1 + 前景 38)
		if !slices.Contains(seqs[0], "1") || !slices.Contains(seqs[0], "38") {
			t.Errorf("键位段应主色加粗(含 1;38): %q", bar)
		}
		// 键位前缀无背景:整行不得出现背景色序列(48)
		for _, params := range seqs {
			if slices.Contains(params, "48") {
				t.Errorf("键位段不应含背景色(48): %q", bar)
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

// sgrSeqList 提取字符串中所有 SGR 序列的参数列表(按出现顺序),用于断言样式层次
func sgrSeqList(s string) [][]string {
	var out [][]string
	for {
		start := strings.Index(s, "\x1b[")
		if start < 0 {
			return out
		}
		s = s[start+2:]
		end := strings.Index(s, "m")
		if end < 0 {
			return out
		}
		out = append(out, strings.Split(s[:end], ";"))
		s = s[end+1:]
	}
}

// isResetSeq 判断 SGR 参数列表是否为纯 reset(\x1b[m 或 \x1b[0m)
func isResetSeq(params []string) bool {
	return len(params) == 1 && (params[0] == "" || params[0] == "0")
}

// effectiveBoldAt 模拟终端 SGR 状态机,返回 s 中 text 首次出现处的生效加粗状态
// (bold 由 1 开启、22 或 reset 关闭),用于断言内嵌段边界无字重泄漏。
func effectiveBoldAt(s, text string) bool {
	pos := strings.Index(s, text)
	if pos < 0 {
		return false
	}
	bold := false
	for i := 0; i < pos; {
		if s[i] != '\x1b' {
			i++
			continue
		}
		end := strings.IndexByte(s[i:], 'm')
		if end < 0 || i+end > pos {
			break
		}
		seq := s[i+2 : i+end]
		switch {
		case seq == "" || seq == "0":
			bold = false
		default:
			for p := range strings.SplitSeq(seq, ";") {
				switch p {
				case "1":
					bold = true
				case "22":
					bold = false
				}
			}
		}
		i += end + 1
	}
	return bold
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
		// 名称段加粗、说明段常规字重(解析 SGR 参数,lipgloss v2 合并且色值含数字 1)
		seqs := sgrSeqList(label)
		if len(seqs) != 3 {
			t.Fatalf("标签应含名称段/字重复位段/说明段三段样式: %q", label)
		}
		if !slices.Contains(seqs[0], "1") {
			t.Errorf("名称应加粗: %q", label)
		}
		if !slices.Contains(seqs[1], "22") {
			t.Errorf("说明段前应复位字重(22): %q", label)
		}
		if slices.Contains(seqs[2], "1") {
			t.Errorf("说明不应加粗: %q", label)
		}
		// 按终端状态机验证生效字重:名称加粗、说明常规(防止 Bold 状态贯穿到说明段)
		if !effectiveBoldAt(label, "soft") {
			t.Errorf("名称处应生效加粗: %q", label)
		}
		if effectiveBoldAt(label, "保留工作区更改") {
			t.Errorf("说明处不应生效加粗(字重被名称段贯穿): %q", label)
		}
		// 内嵌段不得输出 reset(含 \x1b[m / \x1b[0m):否则会中断选中态背景色(见下方回归用例)
		for _, params := range seqs {
			if isResetSeq(params) {
				t.Errorf("内嵌段不得输出 reset(会中断选中背景): %q", label)
			}
		}
		if lipgloss.Width(plain) != lipgloss.Width("soft")+1+lipgloss.Width("保留工作区更改") {
			t.Errorf("标签宽度异常: %q (%d)", plain, lipgloss.Width(plain))
		}
	})

	t.Run("选中态:背景色连续覆盖名称与说明(回归)", func(t *testing.T) {
		label := OptionLabel("soft", "保留工作区更改")
		// 模拟 huh SelectedOption 整行样式:背景 + 前景 + 加粗 + 左右内边距
		selected := lipgloss.NewStyle().
			Background(theme.SelectionBg).
			Foreground(theme.SelectionFg).
			Bold(true).
			Padding(0, 1).
			Render(label)
		// 行首序列必须带背景色(48)
		seqs := sgrSeqList(selected)
		if len(seqs) == 0 || !slices.Contains(seqs[0], "48") {
			t.Fatalf("选中行首序列应设置背景色: %q", selected)
		}
		// 名称与说明之间的原始内容不得含 reset,否则说明文本背景被中断
		softIdx := strings.Index(selected, "soft")
		descIdx := strings.Index(selected, "保留工作区更改")
		if softIdx < 0 || descIdx < 0 {
			t.Fatalf("选中行应含名称与说明: %q", selected)
		}
		// 名称与说明之间不得出现 reset(背景被中断);允许其余样式序列(如说明前景切换)
		between := selected[softIdx+len("soft") : descIdx]
		for _, params := range sgrSeqList(between) {
			if isResetSeq(params) {
				t.Errorf("名称与说明之间存在 reset,背景被中断: %q", selected)
			}
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
	m := newListModel("标题", []config.Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}, {Label: "c", Value: "c"}}, ListSingle, false, nil)
	m.width, m.height = 80, 4
	m.applyLayout()
	view := ansi.Strip(m.View().Content)
	if strings.Contains(view, "Enter") {
		t.Errorf("4 行终端不应渲染帮助栏: %q", view)
	}
	// 3 个选项完整,不被裁剪
	for _, want := range []string{"a", "b", "c"} {
		if !strings.Contains(view, want) {
			t.Errorf("4 行终端缺少选项 %q: %q", want, view)
		}
	}
}
