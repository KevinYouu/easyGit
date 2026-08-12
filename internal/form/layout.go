package form

import (
	"github.com/charmbracelet/bubbles/table"
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
	checkboxColWidth  = 4   // 多选框列宽
	hashColWidth      = 8   // hash 列宽
	dateColWidth      = 12  // 日期列宽
	authorColWidth    = 10  // 作者列宽
	messageColMin     = 20  // 消息列宽下限
	messageColMax     = 160 // 消息列宽上限
	cellPaddingWidth  = 2   // 单元格左右内边距(1+1)
	tableInsetWidth   = 4   // 紧凑模式外层留白
	centerTableMargin = 16  // 宽屏居中判定余量
)

// 表格高度计算常量
const (
	tableHeightReserved     = 4 // 表格预留高度(边框等)
	tableHeightReservedMore = 2 // 多选模式额外预留(标题/底部帮助)
	tableHeightMin          = 3 // 表格最小高度
)

// huh 表单高度计算常量
const (
	formHeightMin      = 3 // 选择框最小高度
	formHeightReserved = 8 // 表单预留高度(标题/说明等)
)

// 默认终端尺寸(检测失败时回退)
const (
	defaultTermWidth  = 80
	defaultTermHeight = 24
)

// 截断省略号
const ellipsis = "..."

// 各模式消息列固定占位 = 固定列宽之和 + 全部列单元格内边距
const (
	fixedWidthThree = hashColWidth + dateColWidth + 3*cellPaddingWidth                  // 三列:hash/message/date
	fixedWidthFull  = hashColWidth + dateColWidth + authorColWidth + 4*cellPaddingWidth // 四列:hash/message/date/author
)

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

// CalculateMessageWidth 计算消息列宽;紧凑模式无消息列,返回 0
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
		fixed += checkboxColWidth + cellPaddingWidth
	}
	return min(max(width-fixed, messageColMin), messageColMax)
}

// CalculateColumns 根据终端宽度与是否含多选框生成表格列
func CalculateColumns(width int, withCheckbox bool) []table.Column {
	mode := LayoutMode(width)
	switch mode {
	case LayoutCompact:
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: width - checkboxColWidth - tableInsetWidth},
			}
		}
		return []table.Column{{Title: "", Width: width - tableInsetWidth}}
	case LayoutThreeCol:
		messageWidth := CalculateMessageWidth(width, withCheckbox)
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: hashColWidth},
				{Title: "", Width: messageWidth},
				{Title: "", Width: dateColWidth},
			}
		}
		return []table.Column{
			{Title: "", Width: hashColWidth},
			{Title: "", Width: messageWidth},
			{Title: "", Width: dateColWidth},
		}
	default:
		messageWidth := CalculateMessageWidth(width, withCheckbox)
		if withCheckbox {
			return []table.Column{
				{Title: "", Width: checkboxColWidth},
				{Title: "", Width: hashColWidth},
				{Title: "", Width: messageWidth},
				{Title: "", Width: dateColWidth},
				{Title: "", Width: authorColWidth},
			}
		}
		return []table.Column{
			{Title: "", Width: hashColWidth},
			{Title: "", Width: messageWidth},
			{Title: "", Width: dateColWidth},
			{Title: "", Width: authorColWidth},
		}
	}
}

// CalculateTableHeight 计算表格可视行数;multi 多选模式额外预留标题/底部帮助行
func CalculateTableHeight(height int, multi bool) int {
	reserved := tableHeightReserved
	if multi {
		reserved += tableHeightReservedMore
	}
	return max(height-reserved, tableHeightMin)
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

// CalculateSelectHeight 计算单选表单高度,随终端高度与选项数伸缩
func CalculateSelectHeight(optionCount, terminalHeight int) int {
	available := max(terminalHeight-formHeightReserved, formHeightMin)
	return min(max(optionCount, formHeightMin), available)
}

// CalculateMultiSelectHeight 计算多选表单高度,随终端高度与选项数伸缩
func CalculateMultiSelectHeight(optionCount, terminalHeight int) int {
	available := max(terminalHeight-formHeightReserved, formHeightMin)
	return min(max(optionCount+1, formHeightMin), available)
}
