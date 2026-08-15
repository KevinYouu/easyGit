package form

import (
	"charm.land/bubbles/v2/table"
	"github.com/charmbracelet/x/ansi"
)

// LayoutKind 表格布局模式
type LayoutKind int

// 布局模式:列数只由终端宽度决定,高度仅影响表格可视行数
const (
	LayoutCompact  LayoutKind = iota // 极窄(< minColumnWidth):单列,信息合并为一行
	LayoutThreeCol                   // 窄(< midColumnWidth):hash/message/date,隐藏 author
	LayoutFull                       // 常规(>= midColumnWidth):hash/message/date/author
)

// 布局判定阈值
const (
	minColumnWidth = 60 // 单列紧凑模式最大终端宽度
	midColumnWidth = 80 // 三列模式最大终端宽度
)

// 表格列宽常量
const (
	checkboxColWidth  = 3  // 多选框列宽([x] / [ ])
	indicatorColWidth = 1  // 选中指示列宽(❯ 或空,间距由列右空间与内边距产生)
	hashColWidth      = 8  // hash 列宽
	dateColWidth      = 12 // 日期列宽
	authorColWidth    = 10 // 作者列宽
	messageColMin     = 20 // 消息列宽下限(无上限,富余宽度全部给消息列)
	cellPaddingWidth  = 2  // 单元格左右内边距(1+1)
	tableInsetWidth   = 4  // 紧凑模式外层留白
	centerTableMargin = 16 // 宽屏居中判定余量
)

// 表格高度计算常量
const (
	multiTableExtraLines = 4 // 列表模式额外固定行数:标题 + 顶部线 + 底部线 + 底部帮助行
	tableHeightMin       = 3 // 表格最小高度
	tableHeaderLines     = 1 // 表格模型自带 header 行,SetHeight 内部会扣除
)

// 默认终端尺寸(检测失败时回退)
const (
	defaultTermWidth  = 80
	defaultTermHeight = 24
)

// 截断省略号
const ellipsis = "..."

// 多选表格底部帮助行最小宽度
const footerMinWidth = 10

// 各模式消息列固定占位 = 固定列宽之和 + 全部列单元格内边距;
// 单选/多选最左均为指示列(1 宽 + 2 内边距),多选另加复选框列
const (
	fixedWidthThree = hashColWidth + dateColWidth + 3*cellPaddingWidth + indicatorColWidth + cellPaddingWidth                  // 三列:指示/hash/message/date
	fixedWidthFull  = hashColWidth + dateColWidth + authorColWidth + 4*cellPaddingWidth + indicatorColWidth + cellPaddingWidth // 四列:指示/hash/message/date/author
)

// 自适应多列布局常量(配置中心等「名称 + 单行说明」列表)
const (
	// MaxAutoColumnWidth Auto 列推荐宽度上限:超过的极长内容截断,防止挤占弹性列。
	// 调用方未在 ColumnSpec.MaxWidth 显式指定时的推荐取值(单一事实来源)。
	MaxAutoColumnWidth = 24
)

// ColumnKind 自适应多列布局的列宽策略
// 列数不硬编码:调用方以 ColumnSpec 声明任意列,每列独立宽度策略。
type ColumnKind int

const (
	ColumnAuto ColumnKind = iota // 自动宽度:按该列最长内容,可设 MaxWidth 上限
	ColumnFlex                   // 弹性宽度:占满剩余,可设 MinWidth 下限
)

// ColumnSpec 自适应多列布局的列定义;Title 仅文档用途(表头隐藏)
type ColumnSpec struct {
	Title    string
	Kind     ColumnKind
	MaxWidth int // ColumnAuto 上限(0 = 不限)
	MinWidth int // ColumnFlex 下限(0 = messageColMin)
}

// NameDescColumns 标准「名称 + 说明」双列 spec:名称列自动宽度(上限
// MaxAutoColumnWidth 截断),说明列占满剩余。非提交类文本列表(命令菜单/
// 分支/远程/标签)统一使用,避免走提交解析路径被 hash 列宽截断首词。
func NameDescColumns() []ColumnSpec {
	return []ColumnSpec{
		{Kind: ColumnAuto, MaxWidth: MaxAutoColumnWidth},
		{Kind: ColumnFlex},
	}
}

// CalculateAdaptiveColumns 计算自适应多列布局的列宽(不含单元格内边距):
// Auto 列 = 该列最长内容的显示宽度(上限 MaxWidth);剩余宽度平均分配给
// Flex 列(各自不低于 MinWidth 或 messageColMin)。调用方保证 spec 与
// 单元格数一致,行不足时缺列补空。仅在非紧凑模式(>= minColumnWidth)
// 下使用,紧凑模式走单列合并。
func CalculateAdaptiveColumns(width int, specs []ColumnSpec, rows [][]string, withCheckbox bool) []int {
	widths := make([]int, len(specs))
	fixed := indicatorColWidth + cellPaddingWidth + len(specs)*cellPaddingWidth
	if withCheckbox {
		fixed += checkboxColWidth + cellPaddingWidth
	}
	flexCount := 0
	for i, spec := range specs {
		switch spec.Kind {
		case ColumnAuto:
			maxW := 0
			for _, row := range rows {
				if i < len(row) {
					maxW = max(maxW, ansi.StringWidth(ansi.Strip(row[i])))
				}
			}
			if spec.MaxWidth > 0 {
				maxW = min(maxW, spec.MaxWidth)
			}
			widths[i] = maxW
			fixed += maxW
		case ColumnFlex:
			flexCount++
		}
	}
	if flexCount > 0 {
		remaining := max(width-fixed, 0)
		per := remaining / flexCount
		for i, spec := range specs {
			if spec.Kind == ColumnFlex {
				minW := spec.MinWidth
				if minW == 0 {
					minW = messageColMin
				}
				widths[i] = max(per, minW)
			}
		}
	}
	return widths
}

// LayoutMode 根据终端宽度判定布局模式
func LayoutMode(width int) LayoutKind {
	switch {
	case width < minColumnWidth:
		return LayoutCompact
	case width < midColumnWidth:
		return LayoutThreeCol
	default:
		return LayoutFull
	}
}

// CalculateMessageWidth 计算消息列宽;紧凑模式无消息列,返回 0。
// 消息列无上限:终端越宽消息列越宽,截断只发生在实际宽度不足时
// 单选/多选最左均为 1 宽指示列,多选另加复选框列(3 宽),间距由列右空间与内边距产生
func CalculateMessageWidth(width int, withCheckbox bool) int {
	mode := LayoutMode(width)
	var fixed int
	switch mode {
	case LayoutCompact:
		return 0
	case LayoutThreeCol:
		fixed = fixedWidthThree
	default:
		fixed = fixedWidthFull
	}
	if withCheckbox {
		fixed += indicatorColWidth + checkboxColWidth + 2*cellPaddingWidth
	}
	return max(width-fixed, messageColMin)
}

// CalculateColumns 根据终端宽度与是否含多选框生成表格列;
// 最左为 1 宽选中指示列(光标行显示 ❯),多选模式其后为 3 宽复选框列
func CalculateColumns(width int, withCheckbox bool) []table.Column {
	mode := LayoutMode(width)
	switch mode {
	case LayoutCompact:
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: indicatorColWidth},
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: width - indicatorColWidth - checkboxColWidth - tableInsetWidth - cellPaddingWidth},
			}
		}
		return []table.Column{
			{Title: "", Width: indicatorColWidth},
			{Title: "", Width: width - indicatorColWidth - tableInsetWidth},
		}
	case LayoutThreeCol:
		messageWidth := CalculateMessageWidth(width, withCheckbox)
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: indicatorColWidth},
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: hashColWidth},
				{Title: "", Width: messageWidth},
				{Title: "", Width: dateColWidth},
			}
		}
		return []table.Column{
			{Title: "", Width: indicatorColWidth},
			{Title: "", Width: hashColWidth},
			{Title: "", Width: messageWidth},
			{Title: "", Width: dateColWidth},
		}
	default:
		messageWidth := CalculateMessageWidth(width, withCheckbox)
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: indicatorColWidth},
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: hashColWidth},
				{Title: "", Width: messageWidth},
				{Title: "", Width: dateColWidth},
				{Title: "", Width: authorColWidth},
			}
		}
		return []table.Column{
			{Title: "", Width: indicatorColWidth},
			{Title: "", Width: hashColWidth},
			{Title: "", Width: messageWidth},
			{Title: "", Width: dateColWidth},
			{Title: "", Width: authorColWidth},
		}
	}
}

// CalculateTableHeight 计算表格可视行数;单选与多选均预留 4 行
// (标题 + 顶部线 + 底部线 + 底部帮助行),<6 行终端附加行不渲染但
// 统一按 4 行预留,保证滚动一致
func CalculateTableHeight(height int) int {
	return max(height-multiTableExtraLines, tableHeightMin)
}

// TotalTableWidth 计算表格渲染总宽(列宽之和 + 单元格内边距)
func TotalTableWidth(columns []table.Column) int {
	total := 0
	for _, col := range columns {
		total += col.Width
	}
	return total + len(columns)*cellPaddingWidth
}

// ShouldCenterTable 终端宽度富余时水平居中,避免超宽屏信息挤在左侧
func ShouldCenterTable(width int, columns []table.Column) bool {
	return width-TotalTableWidth(columns) > centerTableMargin
}

// SafeTruncate 按显示宽度截断(宽字符按 2 列计),保证不破坏 UTF-8;
// 空间足够时附加省略号,极窄空间下不加,负宽度返回空串
func SafeTruncate(s string, maxWidth int) string {
	if maxWidth <= len(ellipsis) {
		return ansi.Truncate(s, maxWidth, "")
	}
	return ansi.Truncate(s, maxWidth, ellipsis)
}
