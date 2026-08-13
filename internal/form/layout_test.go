package form

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

// TestLayoutMode 布局模式判定:列数只由宽度决定,高度不影响
func TestLayoutMode(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		want   LayoutKind
	}{
		{name: "极窄短屏 40x10", width: 40, height: 10, want: LayoutCompact},
		{name: "极窄高屏 40x40", width: 40, height: 40, want: LayoutCompact},
		{name: "宽度边界 59", width: 59, height: 30, want: LayoutCompact},
		{name: "三列下界 60x30", width: 60, height: 30, want: LayoutThreeCol},
		{name: "三列上界 79", width: 79, height: 30, want: LayoutThreeCol},
		{name: "四列下界 80x30", width: 80, height: 30, want: LayoutFull},
		{name: "宽短屏 200x10", width: 200, height: 10, want: LayoutFull},
		{name: "常规 120x30", width: 120, height: 30, want: LayoutFull},
		{name: "超宽 300x60", width: 300, height: 60, want: LayoutFull},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LayoutMode(tt.width); got != tt.want {
				t.Errorf("LayoutMode(%d) = %v, want %v", tt.width, got, tt.want)
			}
		})
	}
}

// TestCalculateMessageWidth 消息列宽边界
func TestCalculateMessageWidth(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		withCheckbox bool
		want         int
	}{
		{name: "极窄 20 无消息列", width: 20, want: 0},
		{name: "极窄 35 无消息列", width: 35, want: 0},
		{name: "三列 60", width: 60, want: 34},
		{name: "四列 80", width: 80, want: 42},
		{name: "四列 120", width: 120, want: 82},
		{name: "超宽 200 封顶", width: 200, want: 160},
		{name: "三列含多选框 60", width: 60, withCheckbox: true, want: 28},
		{name: "四列含多选框 80", width: 80, withCheckbox: true, want: 36},
		{name: "超宽含多选框 200", width: 200, withCheckbox: true, want: 156},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateMessageWidth(tt.width, tt.withCheckbox); got != tt.want {
				t.Errorf("CalculateMessageWidth(%d, %v) = %d, want %d", tt.width, tt.withCheckbox, got, tt.want)
			}
		})
	}
}

// TestCalculateTableHeight 表格高度计算
func TestCalculateTableHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		multi  bool
		want   int
	}{
		{name: "常规 24 行单选让出帮助行", height: 24, multi: false, want: 23},
		{name: "常规 24 行多选预留标题", height: 24, multi: true, want: 21},
		{name: "高屏 50 行多选", height: 50, multi: true, want: 47},
		{name: "矮屏 10 行单选让出帮助行", height: 10, multi: false, want: 9},
		{name: "极矮屏 6 行多选触底", height: 6, multi: true, want: 3},
		{name: "极矮屏 3 行触底", height: 3, multi: false, want: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateTableHeight(tt.height, tt.multi); got != tt.want {
				t.Errorf("CalculateTableHeight(%d, %v) = %d, want %d", tt.height, tt.multi, got, tt.want)
			}
		})
	}
}

// TestCalculateSelectHeight 表单高度 = min(选项数+标题, 终端高度-帮助栏1行):
// 内容不足一屏按内容显示,超出一屏占满终端滚动;帮助栏仅在 ≥6 行终端让出 1 行
func TestCalculateSelectHeight(t *testing.T) {
	tests := []struct {
		name       string
		optionNum  int
		termHeight int
		want       int
	}{
		{name: "大屏内容不足按内容显示", optionNum: 9, termHeight: 24, want: 10},
		{name: "大屏少量选项", optionNum: 4, termHeight: 40, want: 5},
		{name: "大屏海量选项占满让出帮助行", optionNum: 50, termHeight: 24, want: 23},
		{name: "恰好一屏让出帮助行", optionNum: 9, termHeight: 10, want: 9},
		{name: "矮屏滚动让出帮助行", optionNum: 9, termHeight: 8, want: 7},
		{name: "极矮屏触底", optionNum: 50, termHeight: 3, want: 3},
		{name: "极端 2 行不越界", optionNum: 50, termHeight: 2, want: 2},
		{name: "极端 1 行不越界", optionNum: 50, termHeight: 1, want: 1},
		{name: "极端 0 行不 panic", optionNum: 50, termHeight: 0, want: 1},
		{name: "单选项保底", optionNum: 1, termHeight: 24, want: 2},
		{name: "矮屏单选项", optionNum: 1, termHeight: 4, want: 2},
		{name: "零选项不 panic", optionNum: 0, termHeight: 24, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateSelectHeight(tt.optionNum, tt.termHeight); got != tt.want {
				t.Errorf("CalculateSelectHeight(%d, %d) = %d, want %d", tt.optionNum, tt.termHeight, got, tt.want)
			}
			if got := CalculateMultiSelectHeight(tt.optionNum, tt.termHeight); got != tt.want {
				t.Errorf("CalculateMultiSelectHeight(%d, %d) = %d, want %d", tt.optionNum, tt.termHeight, got, tt.want)
			}
		})
	}
}

// TestSafeTruncate 宽度感知截断:不破坏 UTF-8,显示宽度不超限
func TestSafeTruncate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
	}{
		{name: "ASCII 截断", input: "hello world, this is a long message", maxWidth: 20},
		{name: "中文截断", input: "修复中文截断问题以及一些其他内容", maxWidth: 12},
		{name: "emoji 截断", input: "🚀🚀🚀🚀 发布新版本", maxWidth: 10},
		{name: "中英混合", input: "feat: 添加🚀动画效果支持", maxWidth: 15},
		{name: "不超宽不截断", input: "short", maxWidth: 20},
		{name: "极窄宽度 20 不 panic", input: "a1b2c3d 修复中文消息🚀\n07-12 10:00", maxWidth: 14},
		{name: "超窄 2 无省略号", input: "abcdefghij", maxWidth: 2},
		{name: "负宽度返回空串", input: "abcdefghij", maxWidth: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SafeTruncate(tt.input, tt.maxWidth)

			if !utf8.ValidString(got) {
				t.Errorf("SafeTruncate(%q, %d) 产生非法 UTF-8: %q", tt.input, tt.maxWidth, got)
			}
			if tt.maxWidth < 0 {
				if got != "" {
					t.Errorf("SafeTruncate(%q, %d) = %q, 期望空串", tt.input, tt.maxWidth, got)
				}
				return
			}
			if gotWidth := lipgloss.Width(got); gotWidth > tt.maxWidth {
				t.Errorf("SafeTruncate(%q, %d) 显示宽度 %d 超过限制", tt.input, tt.maxWidth, gotWidth)
			}
			// 空间充足时截断必须带省略号
			if tt.maxWidth > len(ellipsis) && lipgloss.Width(tt.input) > tt.maxWidth {
				if !strings.HasSuffix(got, ellipsis) {
					t.Errorf("SafeTruncate(%q, %d) = %q, 期望以省略号结尾", tt.input, tt.maxWidth, got)
				}
			}
		})
	}
}

// TestParseCommitInfo 提交信息解析:字段按显示宽度截断且合法
func TestParseCommitInfo(t *testing.T) {
	label := "a1b2c3d 修复中文消息🚀以及更长的一些附加说明文字\n07-12 10:00 • 张三丰"

	hash, message, date, author := parseCommitInfo(label, 16)

	if hash != "a1b2c3d" {
		t.Errorf("hash = %q, want a1b2c3d", hash)
	}
	for name, value := range map[string]string{
		"message": message,
		"date":    date,
		"author":  author,
	} {
		if !utf8.ValidString(value) {
			t.Errorf("%s = %q 产生非法 UTF-8", name, value)
		}
	}
	if got := lipgloss.Width(message); got > 16 {
		t.Errorf("message 宽度 %d 超过 16: %q", got, message)
	}
	if got := lipgloss.Width(date); got > dateColWidth {
		t.Errorf("date 宽度 %d 超过 %d: %q", got, dateColWidth, date)
	}
	if got := lipgloss.Width(author); got > authorColWidth {
		t.Errorf("author 宽度 %d 超过 %d: %q", got, authorColWidth, author)
	}

	// 长作者名按显示宽度截断(7 列内容 + 省略号)
	_, _, _, longAuthor := parseCommitInfo("a1b2c3d fix\n07-12 10:00 • Kevin Zhang Wei", 16)
	if longAuthor != "Kevin Z..." {
		t.Errorf("长作者截断 = %q, want %q", longAuthor, "Kevin Z...")
	}

	// 单行标签(如 reset 模式):首词作 hash 列,消息列保留,日期作者为空
	singleHash, singleMsg, singleDate, singleAuthor := parseCommitInfo("Soft - 保留工作目录和暂存区", 20)
	if singleHash != "Soft" {
		t.Errorf("单行标签 hash = %q, want Soft", singleHash)
	}
	if !strings.Contains(singleMsg, "保留工作目录") {
		t.Errorf("单行标签 message = %q, 应包含描述", singleMsg)
	}
	if singleDate != "" || singleAuthor != "" {
		t.Errorf("单行标签日期作者应为空: date=%q author=%q", singleDate, singleAuthor)
	}
}

// TestFormatCompactCommit 紧凑格式:窄宽度不 panic 且不破坏 UTF-8
func TestFormatCompactCommit(t *testing.T) {
	label := "a1b2c3d 修复中文消息🚀以及更长的一些附加说明文字\n07-12 10:00 • 张三丰"

	tests := []struct {
		name     string
		maxWidth int
	}{
		{name: "常规宽度", maxWidth: 50},
		{name: "极窄 10", maxWidth: 10},
		{name: "超窄 6", maxWidth: 6},
		{name: "负宽度", maxWidth: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCompactCommit(label, tt.maxWidth)
			if !utf8.ValidString(got) {
				t.Errorf("formatCompactCommit(%d) 产生非法 UTF-8: %q", tt.maxWidth, got)
			}
			if tt.maxWidth > 0 && lipgloss.Width(got) > tt.maxWidth {
				t.Errorf("formatCompactCommit(%d) 宽度 %d 超限", tt.maxWidth, lipgloss.Width(got))
			}
		})
	}
}

// TestCalculateColumns 列数与消息列宽一致
func TestCalculateColumns(t *testing.T) {
	tests := []struct {
		name         string
		width        int
		withCheckbox bool
		wantCols     int
	}{
		{name: "紧凑单列 40", width: 40, wantCols: 1},
		{name: "紧凑含多选框 40", width: 40, withCheckbox: true, wantCols: 2},
		{name: "三列 60", width: 60, wantCols: 3},
		{name: "三列含多选框 60", width: 60, withCheckbox: true, wantCols: 4},
		{name: "四列 80", width: 80, wantCols: 4},
		{name: "四列含多选框 80", width: 80, withCheckbox: true, wantCols: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols := CalculateColumns(tt.width, tt.withCheckbox)
			if len(cols) != tt.wantCols {
				t.Fatalf("CalculateColumns(%d, %v) 列数 = %d, want %d", tt.width, tt.withCheckbox, len(cols), tt.wantCols)
			}
			// 消息列宽与 CalculateMessageWidth 一致
			if msgWidth := CalculateMessageWidth(tt.width, tt.withCheckbox); msgWidth > 0 {
				found := false
				for _, col := range cols {
					if col.Width == msgWidth {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("CalculateColumns(%d, %v) 未包含消息列宽 %d: %+v", tt.width, tt.withCheckbox, msgWidth, cols)
				}
			}
		})
	}
}

// TestTotalTableWidth 表总宽 = 列宽之和 + 单元格内边距,不超出终端宽度
func TestTotalTableWidth(t *testing.T) {
	for _, width := range []int{40, 60, 80, 120, 200, 300} {
		for _, withCheckbox := range []bool{false, true} {
			cols := CalculateColumns(width, withCheckbox)
			total := TotalTableWidth(cols)
			if total > width {
				t.Errorf("宽度 %d (checkbox=%v) 表总宽 %d 溢出终端", width, withCheckbox, total)
			}
		}
	}

	if got := TotalTableWidth(CalculateColumns(80, false)); got != 80 {
		t.Errorf("TotalTableWidth(80 四列) = %d, want 80", got)
	}
}

// TestShouldCenterTable 超宽屏居中判定
func TestShouldCenterTable(t *testing.T) {
	if !ShouldCenterTable(300, CalculateColumns(300, false)) {
		t.Error("300 列终端应居中")
	}
	if ShouldCenterTable(80, CalculateColumns(80, false)) {
		t.Error("80 列终端不应居中")
	}
	if ShouldCenterTable(200, CalculateColumns(200, false)) {
		t.Error("200 列终端富余不足不应居中")
	}
}
