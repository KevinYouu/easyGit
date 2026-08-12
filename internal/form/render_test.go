package form

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// 渲染级测试:直接构造真实 huh 字段/表格模型,按终端尺寸矩阵渲染并断言,
// 覆盖 SelectForm/MultiSelectForm 高度利用与 TableSelect/TableMultiSelect 溢出问题。

// renderSelectField 构造与 SelectForm 完全一致的 huh 单选字段
func renderSelectField(title string, labels []string, termHeight int) string {
	opts := make([]huh.Option[string], len(labels))
	for i, l := range labels {
		opts[i] = huh.NewOption(l, l)
	}
	f := huh.NewSelect[string]().
		Title(title).
		Options(opts...).
		Height(CalculateSelectHeight(len(labels), termHeight)).
		Filtering(false)
	f.WithTheme(theme.GetCompactTheme())
	return f.View()
}

// renderMultiSelectField 构造与 MultiSelectForm 完全一致的 huh 多选字段
func renderMultiSelectField(title string, labels []string, termHeight int) string {
	opts := make([]huh.Option[string], len(labels))
	for i, l := range labels {
		opts[i] = huh.NewOption(l, l)
	}
	var selected []string
	f := huh.NewMultiSelect[string]().
		Title(title).
		Options(opts...).
		Height(CalculateMultiSelectHeight(len(labels), termHeight)).
		Value(&selected)
	f.WithTheme(theme.GetCompactTheme())
	return f.View()
}

// visibleLabels 统计渲染文本中按顺序出现的选项标签。
// huh 选项行形如 "│>  fix  │",按"两空格+标签"行内匹配以避开边框字符。
func visibleLabels(view string, labels []string) []string {
	var got []string
	for line := range strings.SplitSeq(view, "\n") {
		for _, l := range labels {
			if strings.Contains(line, "  "+l) {
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
	tests := []struct {
		name       string
		termHeight int
		wantFirst  int // 应可见的选项数(前 wantFirst 个)
	}{
		{name: "常规屏24行全显示", termHeight: 24, wantFirst: 9},
		{name: "矮屏14行全显示", termHeight: 14, wantFirst: 9},
		{name: "矮屏12行显示9个", termHeight: 12, wantFirst: 9}, // 边框2行,内容最多占满
		{name: "矮屏11行显示8个", termHeight: 11, wantFirst: 8},
		{name: "矮屏10行显示7个", termHeight: 10, wantFirst: 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := renderSelectField("Choose a commit type", labels, tt.termHeight)
			if !utf8.ValidString(view) {
				t.Fatal("渲染结果含非法 UTF-8")
			}
			got := visibleLabels(view, labels)
			if len(got) != tt.wantFirst {
				t.Fatalf("可见选项 %d 个: %v, want 前 %d 个(终端 %d 行)", len(got), got, tt.wantFirst, tt.termHeight)
			}
			for i, l := range got {
				if l != labels[i] {
					t.Fatalf("可见选项顺序错乱: %v, want 前 %d 个", got, tt.wantFirst)
				}
			}
			if h := lipgloss.Height(view); h > tt.termHeight {
				t.Errorf("渲染总高 %d 超过终端 %d 行", h, tt.termHeight)
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
	tests := []struct {
		name       string
		termHeight int
		wantFirst  int
	}{
		{name: "常规屏24行全显示", termHeight: 24, wantFirst: 12},
		{name: "矮屏14行显示11个", termHeight: 14, wantFirst: 11}, // 边框2+标题1 占3行
		{name: "矮屏12行显示9个", termHeight: 12, wantFirst: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view := renderMultiSelectField("Select files", labels, tt.termHeight)
			if !utf8.ValidString(view) {
				t.Fatal("渲染结果含非法 UTF-8")
			}
			got := visibleLabels(view, labels)
			if len(got) != tt.wantFirst {
				t.Fatalf("可见选项 %d 个: %v, want 前 %d 个(终端 %d 行)", len(got), got, tt.wantFirst, tt.termHeight)
			}
			if h := lipgloss.Height(view); h > tt.termHeight {
				t.Errorf("渲染总高 %d 超过终端 %d 行", h, tt.termHeight)
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

// assertTableView 校验表格视图:UTF-8 合法、行宽不溢出、行数与高度自适应一致、提交全部可见
func assertTableView(t *testing.T, view string, w, h int, options []config.Option, multi bool) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	n := len(options)
	expectRows := min(CalculateTableHeight(h, multi), n)
	extraRows := 0
	if multi {
		extraRows = 2 // 标题/count 行 + 底部帮助行
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
			assertTableView(t, m.View(), sz.w, sz.h, options, false)
		})
	}
}

func TestTableMultiSelectRender(t *testing.T) {
	options := testOptions()
	for _, sz := range tableSizes {
		t.Run(fmt.Sprintf("%dx%d", sz.w, sz.h), func(t *testing.T) {
			m := &tableMultiModel{
				choices:  options,
				selected: make(map[int]bool),
				width:    sz.w,
				height:   sz.h,
				title:    "测试多选",
				styles:   defaultTableStyles(),
				table:    table.New(table.WithFocused(true)),
			}
			m.updateLayout()
			assertTableView(t, m.View(), sz.w, sz.h, options, true)
		})
	}
}
