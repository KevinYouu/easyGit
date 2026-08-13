package form

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/charmbracelet/x/ansi"
)

// 渲染级测试:直接驱动生产表单构造器,按终端尺寸矩阵渲染并断言,
// 覆盖 SelectForm/MultiSelectForm 高度利用与 TableSelect/TableMultiSelect 溢出问题。
// 按测试功能块拆分:
//   - render_test.go:            基础组件渲染 + 共享辅助函数
//   - render_command_test.go:    命令 × Select/MultiSelect/Table 表单渲染
//   - render_input_confirm_test.go: 命令 × Input/Confirm 表单渲染

// renderSelectField 经生产构造器 NewSelectForm 渲染单选表单
func renderSelectField(title string, labels []string, termHeight int) string {
	opts := make([]config.Option, len(labels))
	for i, l := range labels {
		opts[i] = config.Option{Label: l, Value: l}
	}
	var selected string
	form := NewSelectForm(title, opts, termHeight, &selected)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*Form).View().Content
}

// renderMultiSelectField 经生产构造器 NewMultiSelectForm 渲染多选表单
func renderMultiSelectField(title string, labels []string, termHeight int) string {
	var selected []string
	form := NewMultiSelectForm(title, labels, termHeight, &selected)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*Form).View().Content
}

// visibleLabels 统计渲染文本中按顺序出现的选项标签。
// 选项行形如 "> fix" 或 "  feat"(被 ANSI 样式包裹),先去样式再按行匹配;
// 标签本身也可能内嵌样式(OptionLabel 名称亮 + 说明灰),匹配前同样去样式。
func visibleLabels(view string, labels []string) []string {
	var got []string
	for line := range strings.SplitSeq(view, "\n") {
		plain := ansi.Strip(line)
		for _, l := range labels {
			if strings.Contains(plain, ansi.Strip(l)) {
				got = append(got, l)
				break
			}
		}
	}
	return got
}

func TestSelectRenderUsesAllSpace(t *testing.T) {
	// 与 config.GetDefaultOptions 一致的 9 个提交类型
	labels := []string{"fix", "feat", "refactor", "build", "chore", "style", "docs", "revert", "test"}
	n := len(labels)
	tests := []struct {
		name        string
		termHeight  int
		wantVisible int // 可见选项数:内容不足一屏时全部可见,超出时占满终端滚动
		wantTotal   int // 渲染总高:min(内容高度+帮助1行, 终端高度)
	}{
		{name: "大屏24行按内容显示", termHeight: 24, wantVisible: n, wantTotal: n + 2},
		{name: "大屏14行按内容显示", termHeight: 14, wantVisible: n, wantTotal: n + 2},
		{name: "恰好一屏12行", termHeight: 12, wantVisible: n, wantTotal: n + 2},
		{name: "余1行仍按内容显示", termHeight: 11, wantVisible: n, wantTotal: n + 2},
		{name: "矮屏10行占满滚动", termHeight: 10, wantVisible: n - 1, wantTotal: 10},
		{name: "矮屏9行占满滚动", termHeight: 9, wantVisible: n - 2, wantTotal: 9},
		{name: "矮屏8行占满滚动", termHeight: 8, wantVisible: n - 3, wantTotal: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := renderSelectField("Choose a commit type", labels, tt.termHeight)
			if !utf8.ValidString(view) {
				t.Fatal("渲染结果含非法 UTF-8")
			}
			got := visibleLabels(view, labels)
			if len(got) != tt.wantVisible {
				t.Fatalf("可见选项 %d 个: %v, want %d 个(终端 %d 行)", len(got), got, tt.wantVisible, tt.termHeight)
			}
			for i, l := range got {
				if l != labels[i] {
					t.Fatalf("可见选项顺序错乱: %v, want 前 %d 个", got, tt.wantVisible)
				}
			}
			if h := lipgloss.Height(view); h != tt.wantTotal {
				t.Errorf("渲染总高 %d, want %d(终端 %d 行,内容 %d 行)", h, tt.wantTotal, tt.termHeight, n+1)
			}
		})
	}
}

func TestSelectRenderManyOptions(t *testing.T) {
	labels := make([]string, 20)
	for i := range labels {
		labels[i] = "opt-" + string(rune('A'+i))
	}
	for _, termHeight := range []int{24, 40, 50} {
		view := renderSelectField("很多选项", labels, termHeight)
		got := visibleLabels(view, labels)
		if len(got) != 20 {
			t.Errorf("终端 %d 行:可见 %d 个, want 20 个", termHeight, len(got))
		}
	}
}

func TestMultiSelectRenderUsesAllSpace(t *testing.T) {
	labels := make([]string, 12)
	for i := range labels {
		labels[i] = "file-" + string(rune('A'+i)) + ".go"
	}
	n := len(labels)
	tests := []struct {
		name        string
		termHeight  int
		wantVisible int // 可见选项数:内容不足一屏时全部可见,超出时占满终端滚动
		wantTotal   int // 渲染总高:min(内容高度+帮助1行, 终端高度)
	}{
		{name: "大屏24行按内容显示", termHeight: 24, wantVisible: n, wantTotal: n + 2},
		{name: "大屏14行按内容显示", termHeight: 14, wantVisible: n, wantTotal: n + 2},
		{name: "恰好一屏13行", termHeight: 13, wantVisible: n - 1, wantTotal: 13},
		{name: "矮屏12行占满滚动", termHeight: 12, wantVisible: n - 2, wantTotal: 12},
		{name: "矮屏10行占满滚动", termHeight: 10, wantVisible: n - 4, wantTotal: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := renderMultiSelectField("Select files", labels, tt.termHeight)
			if !utf8.ValidString(view) {
				t.Fatal("渲染结果含非法 UTF-8")
			}
			got := visibleLabels(view, labels)
			if len(got) != tt.wantVisible {
				t.Fatalf("可见选项 %d 个: %v, want %d 个(终端 %d 行)", len(got), got, tt.wantVisible, tt.termHeight)
			}
			if h := lipgloss.Height(view); h != tt.wantTotal {
				t.Errorf("渲染总高 %d, want %d(终端 %d 行,内容 %d 行)", h, tt.wantTotal, tt.termHeight, n+1)
			}
		})
	}
}

// 表格渲染测试:覆盖窄屏(紧凑/三列)、宽屏(四列+居中)、中文/emoji 完整性
var tableSizes = []struct{ w, h int }{
	{40, 10}, // 极窄:紧凑单列
	{40, 40}, // 窄高:紧凑单列
	{60, 30}, // 三列边界
	{79, 24}, // 三列
	{80, 24}, // 四列边界
	{120, 30},
	{200, 10}, // 宽矮:消息列达上限
	{300, 60}, // 超宽:居中
}

var commitLabels = []string{
	"a1b2c3d 修复中文消息🚀以及更长的一些附加说明文字\n07-12 10:00 • 张三丰",
	"e5f6g7h feat: add long feature message with ascii words\n07-13 09:30 • Kevin Zhang",
	"8a9b0c1 修复emoji图标😀与中文混合截断问题\n07-14 11:20 • 李四",
}

// testOptions 构造 20 个提交选项(含中文/emoji/长文本)
func testOptions() []config.Option {
	options := make([]config.Option, 0, 20)
	for _, l := range commitLabels {
		options = append(options, config.Option{Label: l, Value: "v"})
	}
	for i := range 17 {
		options = append(options, config.Option{Label: commitLabels[i%len(commitLabels)], Value: "v"})
	}
	return options
}

// assertTableView 校验表格视图:UTF-8 合法、行宽不溢出、高度策略正确:
// 内容不足一屏时按内容显示(总高=内容高),超出一屏时占满终端
func assertTableView(t *testing.T, view string, w, h int, options []config.Option, multi bool) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	n := len(options)
	expectRows := min(CalculateTableHeight(h, multi), n)
	extraRows := 0
	// 单选在 ≥6 行终端含底部帮助栏一行;多选固定 标题/count 行 + 底部帮助行(空行不计)
	if multi {
		extraRows = 2
	} else if h >= HelpBarMinTermHeight {
		extraRows = 1
	}

	nonEmpty := 0
	for line := range strings.SplitSeq(strings.TrimRight(view, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		nonEmpty++
		if lw := lipgloss.Width(line); lw > w {
			t.Errorf("行宽 %d 溢出终端 %d: %q", lw, w, line)
		}
	}
	if nonEmpty != expectRows+extraRows {
		t.Errorf("非空行数 = %d, want %d(表格 %d 行 + 附加 %d 行)", nonEmpty, expectRows+extraRows, expectRows, extraRows)
	}

	// 总高策略:内容不足一屏(含附加行)时按内容显示,否则占满终端
	contentHeight := n + extraRows
	if multi {
		contentHeight = n + 3 // 标题 + 空行 + 底部帮助行
	}
	wantTotal := min(contentHeight, h)
	if total := lipgloss.Height(view); total != wantTotal {
		t.Errorf("渲染总高 = %d, want %d(内容 %d 行,终端 %d 行)", total, wantTotal, contentHeight, h)
	}

	// 高度足够时每个提交的 hash 前缀都应可见
	if expectRows == n {
		for _, opt := range options {
			hash := strings.SplitN(opt.Label, " ", 2)[0]
			if !strings.Contains(view, hash) {
				t.Errorf("提交 %s 未出现在视图中", hash)
			}
		}
	}
}

func TestTableSelectRender(t *testing.T) {
	options := testOptions()
	for _, sz := range tableSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := NewTableSelectModel(options)
			m.width, m.height = sz.w, sz.h
			m.applyLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, false)
		})
	}
}

func TestTableMultiSelectRender(t *testing.T) {
	options := testOptions()
	for _, sz := range tableSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := NewTableMultiSelectModel("测试多选", options)
			m.width, m.height = sz.w, sz.h
			m.updateLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, true)
		})
	}
}
