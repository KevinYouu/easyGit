package form

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// 渲染级测试:直接构造真实 huh 字段/表格模型,按终端尺寸矩阵渲染并断言,
// 覆盖 SelectForm/MultiSelectForm 高度利用与 TableSelect/TableMultiSelect 溢出问题。

// renderSelectField 构造与 SelectForm 完全一致的 huh 单选表单
func renderSelectField(title string, labels []string, termHeight int) string {
	opts := make([]huh.Option[string], len(labels))
	for i, l := range labels {
		opts[i] = huh.NewOption(l, l)
	}
	var selected string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title(title).
			Options(opts...).
			Value(&selected).
			Height(CalculateSelectHeight(len(labels), termHeight)).
			Filtering(false),
	)).WithTheme(theme.GetCompactTheme()).WithShowHelp(false)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

// renderMultiSelectField 构造与 MultiSelectForm 完全一致的 huh 多选表单
func renderMultiSelectField(title string, labels []string, termHeight int) string {
	opts := make([]huh.Option[string], len(labels))
	for i, l := range labels {
		opts[i] = huh.NewOption(l, l)
	}
	var selected []string
	form := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title(title).
			Options(opts...).
			Value(&selected).
			Height(CalculateMultiSelectHeight(len(labels), termHeight)),
	)).WithTheme(theme.GetCompactTheme()).WithShowHelp(false)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

// visibleLabels 统计渲染文本中按顺序出现的选项标签。
// 选项行形如 "> fix" 或 "  feat"(被 ANSI 样式包裹),先去样式再按行匹配。
func visibleLabels(view string, labels []string) []string {
	var got []string
	for line := range strings.SplitSeq(view, "\n") {
		plain := ansi.Strip(line)
		for _, l := range labels {
			if strings.Contains(plain, l) {
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
		wantTotal   int // 渲染总高:min(内容高度, 终端高度)
	}{
		{name: "大屏24行按内容显示", termHeight: 24, wantVisible: n, wantTotal: n + 1},
		{name: "大屏14行按内容显示", termHeight: 14, wantVisible: n, wantTotal: n + 1},
		{name: "恰好一屏12行", termHeight: 12, wantVisible: n, wantTotal: n + 1},
		{name: "余1行仍按内容显示", termHeight: 11, wantVisible: n, wantTotal: n + 1},
		{name: "矮屏10行占满滚动", termHeight: 10, wantVisible: n, wantTotal: 10},
		{name: "矮屏9行占满滚动", termHeight: 9, wantVisible: n - 1, wantTotal: 9},
		{name: "矮屏8行占满滚动", termHeight: 8, wantVisible: n - 2, wantTotal: 8},
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
		wantTotal   int // 渲染总高:min(内容高度, 终端高度)
	}{
		{name: "大屏24行按内容显示", termHeight: 24, wantVisible: n, wantTotal: n + 1},
		{name: "大屏14行按内容显示", termHeight: 14, wantVisible: n, wantTotal: n + 1},
		{name: "恰好一屏13行", termHeight: 13, wantVisible: n, wantTotal: n + 1},
		{name: "矮屏12行占满滚动", termHeight: 12, wantVisible: n - 1, wantTotal: 12},
		{name: "矮屏10行占满滚动", termHeight: 10, wantVisible: n - 3, wantTotal: 10},
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
	if multi {
		extraRows = 2 // 标题/count 行 + 底部帮助行(空行不计)
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

// ===== 内置命令表单渲染覆盖 =====
// 按各命令实际使用的表单类型与选项规模,在终端尺寸矩阵下断言:
// 内容不足一屏时按内容显示,超出一屏时占满终端滚动,且无越界行

// commandSelectCases 各命令的单选表单(选项数/标题与命令一致)
var commandSelectCases = []struct {
	command string
	title   string
	labels  []string
}{
	{command: "set-language", title: "选择语言", labels: []string{"中文", "English"}},
	{command: "cherry-pick option", title: "选择模式", labels: []string{"default", "no-commit", "edit", "signoff"}},
	{command: "merge strategy", title: "合并策略", labels: []string{"default", "ff-only", "no-ff", "squash"}},
	{command: "rebase action", title: "rebase 操作", labels: []string{"--continue", "--skip", "--abort"}},
	{command: "push commit type", title: "提交类型", labels: []string{"fix", "feat", "refactor", "build", "chore", "style", "docs", "revert", "test"}},
	{command: "branch delete", title: "删除分支", labels: []string{"main", "develop", "feature/login", "fix/typo", "release/v1.2.0"}},
	{command: "remote select", title: "选择远端", labels: []string{"origin", "upstream", "github"}},
	{command: "remote branch", title: "选择分支", labels: []string{"main", "develop", "feature/login", "release/v2.0"}},
	{command: "tag delete", title: "删除标签", labels: []string{"v1.0.0", "v1.1.0", "v2.0.0-beta"}},
}

// commandMultiCases 各命令的多选表单
var commandMultiCases = []struct {
	command string
	title   string
	labels  []string
}{
	// cherry-pick 提交列表:单行 "[hash8] 日期 (作者) - 消息" 格式
	{command: "cherry-pick commits", title: "选择提交", labels: multiLabels(50, func(i int) string {
		return fmt.Sprintf("[a1b2c3d%02d] 07-12 10:00 (张三丰) - 修复登录问题", i)
	})},
	// push-selected 文件列表
	{command: "push-selected files", title: "选择文件", labels: multiLabels(12, func(i int) string {
		return fmt.Sprintf("internal/form/file-%02d.go", i)
	})},
	// set-push-config 远端列表
	{command: "set-push-config remotes", title: "选择远端", labels: []string{"origin", "upstream", "github", "gitlab"}},
}

// multiLabels 生成 n 个标签
func multiLabels(n int, f func(int) string) []string {
	labels := make([]string, n)
	for i := range n {
		labels[i] = f(i)
	}
	return labels
}

// assertCommandField 校验单选/多选表单渲染:总高 = min(内容, 终端),
// 可见项 = min(选项数, 终端-1),行宽不越界。
// 断言独立于 CalculateSelectHeight 本身(渲染层验证,不构成循环):
// 钉住的是 formFieldHeight 文档化的内容模型 —— 标题一行 + 每选项一行。
func assertCommandField(t *testing.T, view string, labels []string, termHeight, termWidth int) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	n := len(labels)
	// 大屏按内容显示,小屏占满终端滚动
	if total := lipgloss.Height(view); total != min(n+1, termHeight) {
		t.Errorf("渲染总高 = %d, want %d(选项 %d,终端 %d 行)", total, min(n+1, termHeight), n, termHeight)
	}
	got := visibleLabels(view, labels)
	if want := min(n, termHeight-1); len(got) != want {
		t.Fatalf("可见选项 %d 个, want %d(选项 %d,终端 %d 行)", len(got), want, n, termHeight)
	}
	for line := range strings.SplitSeq(view, "\n") {
		if lw := lipgloss.Width(ansi.Strip(line)); lw > termWidth {
			t.Errorf("行宽 %d 溢出终端 %d", lw, termWidth)
		}
	}
}

func TestCommandSelectRender(t *testing.T) {
	for _, tc := range commandSelectCases {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderSelectField(tc.title, tc.labels, h)
				assertCommandField(t, view, tc.labels, h, 80)
			})
		}
	}
}

func TestCommandMultiSelectRender(t *testing.T) {
	for _, tc := range commandMultiCases {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderMultiSelectField(tc.title, tc.labels, h)
				assertCommandField(t, view, tc.labels, h, 80)
			})
		}
	}
}

// 表格类命令:drop/squash 多选提交、reset 提交列表、reset 模式
func TestCommandTableRender(t *testing.T) {
	// GetRecentCommits 两行格式的提交选项(%h 短 hash 为 7 字符)
	tableOptions := func(n int) []config.Option {
		options := make([]config.Option, n)
		for i := range n {
			options[i] = config.Option{
				Label: fmt.Sprintf("a1b2c%02d 修复中文消息🚀以及附加说明\n07-12 10:00 • 张三丰", i),
				Value: "v",
			}
		}
		return options
	}

	t.Run("drop/squash 多选提交", func(t *testing.T) {
		options := tableOptions(20)
		for _, sz := range tableSizes {
			m := &tableMultiModel{
				choices:  options,
				selected: make(map[int]bool),
				width:    sz.w,
				height:   sz.h,
				title:    "选择提交",
				styles:   defaultTableStyles(),
				table:    table.New(table.WithFocused(true)),
			}
			m.updateLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, true)
		}
	})

	t.Run("reset 提交列表", func(t *testing.T) {
		options := tableOptions(20)
		for _, sz := range tableSizes {
			m := NewTableSelectModel(options)
			m.width, m.height = sz.w, sz.h
			m.applyLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, false)
		}
	})

	t.Run("reset 模式仅3项", func(t *testing.T) {
		options := []config.Option{
			{Label: "--soft 保留工作区更改", Value: "--soft"},
			{Label: "--mixed 保留但取消暂存", Value: "--mixed"},
			{Label: "--hard 丢弃所有更改", Value: "--hard"},
		}
		for _, sz := range tableSizes {
			m := NewTableSelectModel(options)
			m.width, m.height = sz.w, sz.h
			m.applyLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, false)
		}
	})
}

// renderInputField 构造与 Input 完全一致的 huh 输入表单
func renderInputField(title string, termHeight int) string {
	var value string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Placeholder("输入内容").
			Value(&value),
	)).WithTheme(theme.GetCompactTheme())
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

// renderConfirmField 构造与 Confirm 完全一致的 huh 确认表单
func renderConfirmField(title string, termHeight int) string {
	var confirmed bool
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(title).
			Value(&confirmed),
	)).WithTheme(theme.GetCompactTheme()).WithShowHelp(false)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

func TestInputConfirmRender(t *testing.T) {
	// 输入/确认表单无 Height 设置,内容高度固定,只断言不越界
	for _, h := range []int{24, 10, 4} {
		for name, view := range map[string]string{
			"input":   renderInputField("标签名称", h),
			"confirm": renderConfirmField("确认删除该分支?", h),
		} {
			if !utf8.ValidString(view) {
				t.Fatalf("%s 渲染结果含非法 UTF-8", name)
			}
			if total := lipgloss.Height(view); total > h || total == 0 {
				t.Errorf("%s@%d行: 渲染总高 %d 越界或为空", name, h, total)
			}
			for line := range strings.SplitSeq(view, "\n") {
				if lw := lipgloss.Width(ansi.Strip(line)); lw > 80 {
					t.Errorf("%s@%d行: 行宽 %d 溢出", name, h, lw)
				}
			}
		}
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
			assertTableView(t, m.View().Content, sz.w, sz.h, options, true)
		})
	}
}
