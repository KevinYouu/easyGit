package form

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// filterTestOptions 构造带可区分文本的选项列表(过滤词可命中不同项)
func filterTestOptions() []config.Option {
	return []config.Option{
		{Label: "feat: add search", Value: "a"},
		{Label: "fix: filter bug", Value: "b"},
		{Label: "docs: update readme", Value: "c"},
		{Label: "feat: add filter", Value: "d"},
		{Label: "chore: cleanup", Value: "e"},
	}
}

// pumpFilter 向模型发送按键序列,返回最终模型
func pumpFilter(m *listModel, keys ...string) *listModel {
	for _, k := range keys {
		var msg tea.KeyPressMsg
		switch k {
		case "up":
			msg = tea.KeyPressMsg{Code: tea.KeyUp}
		case "down":
			msg = tea.KeyPressMsg{Code: tea.KeyDown}
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "esc":
			msg = tea.KeyPressMsg{Code: tea.KeyEsc}
		case "backspace":
			msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
		case "space":
			msg = tea.KeyPressMsg{Code: tea.KeySpace}
		default:
			// 可打印字符(单字符或多字符粘贴)
			msg = tea.KeyPressMsg{Code: rune(k[0]), Text: k}
		}
		next, _ := m.Update(msg)
		m = next.(*listModel)
	}
	return m
}

// TestFilterEnterAndMatch 按 / 进入过滤,输入字符即过滤(大小写不敏感)
func TestFilterEnterAndMatch(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "F", "i", "x")
	if !m.filtering {
		t.Fatal("按 / 后应处于过滤输入模式")
	}
	if m.filter != "Fix" {
		t.Fatalf("filter = %q, want %q", m.filter, "Fix")
	}
	// 匹配 fix: filter bug(索引 1)
	if len(m.visible) != 1 || m.visible[0] != 1 {
		t.Fatalf("visible = %v, want [1]", m.visible)
	}
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
}

// TestFilterBackspaceEsc 退格删字符;Esc 清除过滤并恢复全量
func TestFilterBackspaceEsc(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "f", "e", "a", "t")
	// feat 命中 0 和 3
	if len(m.visible) != 2 {
		t.Fatalf("visible = %v, want [0 3]", m.visible)
	}
	m = pumpFilter(m, "backspace", "backspace", "backspace")
	// 删到 "f":命中 0,1,3
	if m.filter != "f" {
		t.Fatalf("filter = %q, want f", m.filter)
	}
	if len(m.visible) != 3 {
		t.Fatalf("visible = %v, want [0 1 3]", m.visible)
	}
	m = pumpFilter(m, "esc")
	if m.filtering {
		t.Fatal("Esc 应退出过滤输入模式")
	}
	if m.filter != "" {
		t.Fatalf("Esc 应清除过滤词, got %q", m.filter)
	}
	if len(m.visible) != len(options) {
		t.Fatalf("Esc 后应恢复全量可见, got %d", len(m.visible))
	}
}

// TestFilterNavigation 过滤视图内导航只遍历匹配项,Enter 确认返回匹配值
func TestFilterNavigation(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "feat", "enter")
	if m.filtering {
		t.Fatal("Enter 应完成过滤")
	}
	if m.filter != "feat" {
		t.Fatalf("filter = %q, want feat", m.filter)
	}
	// 光标在第一个匹配(feat: add search)
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
	// 过滤视图内向下移动:0 -> 3(跳过 1/2 非匹配项)
	m = pumpFilter(m, "down")
	if m.cursor != 3 {
		t.Fatalf("down 后 cursor = %d, want 3", m.cursor)
	}
	// 循环导航关闭时顶部边界停留
	m = pumpFilter(m, "up", "up")
	if m.cursor != 0 {
		t.Fatalf("up 到顶后 cursor = %d, want 0", m.cursor)
	}
	// Enter 确认:单选返回光标项
	m = pumpFilter(m, "enter")
	confirmed := m.confirmed
	if !confirmed {
		t.Fatal("Enter 应确认选择")
	}
	if got := options[m.cursor].Value; got != "a" {
		t.Fatalf("确认值 = %q, want a", got)
	}
}

// TestFilterWrapNavigation 过滤视图内循环导航(wrap)
func TestFilterWrapNavigation(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, true, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "feat", "enter", "up")
	// 顶部按 ↑ 循环到最后一个匹配(3)
	if m.cursor != 3 {
		t.Fatalf("wrap up cursor = %d, want 3", m.cursor)
	}
	m = pumpFilter(m, "down")
	if m.cursor != 0 {
		t.Fatalf("wrap down cursor = %d, want 0", m.cursor)
	}
}

// TestFilterNoMatch 无匹配时视图为空,导航安全,标题行显示过滤状态
func TestFilterNoMatch(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "zzz")
	if len(m.visible) != 0 {
		t.Fatalf("visible = %v, want empty", m.visible)
	}
	if m.cursor != -1 {
		t.Fatalf("cursor = %d, want -1", m.cursor)
	}
	// 无匹配时导航不 panic 且不移动
	m = pumpFilter(m, "enter", "down", "up")
	if len(m.visible) != 0 {
		t.Fatalf("导航后 visible 应仍为空, got %v", m.visible)
	}
	// 过滤输入模式标题行显示输入状态
	m = pumpFilter(m, "/")
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, "zzz") || !strings.Contains(view, "▍") {
		t.Fatalf("过滤输入视图应含过滤词与光标指示符:\n%s", view)
	}
	// 完成过滤后视图显示过滤结果提示
	m = pumpFilter(m, "enter")
	view = ansi.Strip(m.View().Content)
	if !strings.Contains(view, "zzz") {
		t.Fatalf("已过滤视图应含过滤词:\n%s", view)
	}
}

// TestFilterMultiSelect 多选过滤:空格切换仍按 choices 索引,确认返回选中值
func TestFilterMultiSelect(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListMulti, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "feat", "enter")
	// 选中第一个匹配(feat: add search, choices 索引 0)
	m = pumpFilter(m, "space")
	if !m.selected[0] {
		t.Fatal("space 应选中 choices 索引 0")
	}
	// 切到第二个匹配(choices 索引 3)并选中
	m = pumpFilter(m, "down", "space")
	if !m.selected[3] {
		t.Fatal("space 应选中 choices 索引 3")
	}
	// 确认
	m = pumpFilter(m, "enter")
	if !m.confirmed {
		t.Fatal("enter 应确认")
	}
}

// TestFilterRenderTitle 过滤状态渲染:输入模式帮助栏切换为清除/完成
func TestFilterRenderTitle(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 120
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "fix")
	view := ansi.Strip(m.View().Content)
	if !strings.Contains(view, i18n.T("form.help.filter.clear")) {
		t.Fatalf("过滤输入模式帮助栏应含清除筛选:\n%s", view)
	}
	if !strings.Contains(view, i18n.T("form.help.filter.done")) {
		t.Fatalf("过滤输入模式帮助栏应含完成筛选:\n%s", view)
	}
	// 普通模式帮助栏含 / 筛选提示
	m2 := newListModel("标题", options, ListSingle, false, nil)
	m2.width = 120
	m2.height = 20
	m2.applyLayout()
	view2 := ansi.Strip(m2.View().Content)
	if !strings.Contains(view2, i18n.T("form.help.filter")) {
		t.Fatalf("普通模式帮助栏应含筛选提示:\n%s", view2)
	}
}

// TestFilterCaseInsensitive 大小写不敏感匹配
func TestFilterCaseInsensitive(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	m = pumpFilter(m, "/", "FIX")
	if len(m.visible) != 1 || m.visible[0] != 1 {
		t.Fatalf("大写 FIX 应命中 fix, visible = %v", m.visible)
	}
}

func TestFilterPaste(t *testing.T) {
	options := filterTestOptions()
	m := newListModel("标题", options, ListSingle, false, nil)
	m.width = 100
	m.height = 20
	m.applyLayout()

	// 模拟粘贴多字符(Text 为整段)
	next, _ := m.Update(tea.KeyPressMsg{Code: 'a', Text: "add search"})
	m = next.(*listModel)
	if m.filtering {
		t.Fatal("非过滤模式不应进入过滤")
	}
	// 按 / 进入后再粘贴
	m = pumpFilter(m, "/")
	next, _ = m.Update(tea.KeyPressMsg{Code: 'f', Text: "feat: add filter"})
	m = next.(*listModel)
	if m.filter != "feat: add filter" {
		t.Fatalf("粘贴后 filter = %q", m.filter)
	}
	if len(m.visible) != 1 || m.visible[0] != 3 {
		t.Fatalf("粘贴过滤 visible = %v, want [3]", m.visible)
	}
}

var _ = fmt.Sprintf
