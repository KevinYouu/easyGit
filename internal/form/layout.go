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
	multiTableExtraLines = 3 // 多选模式额外固定行数:标题 + 空行 + 底部帮助行
	singleTableHelpLines = 1 // 单选模式底部帮助栏一行(≥6 行终端)
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

// CalculateTableHeight 计算表格可视行数;多选模式额外预留标题/空行/底部帮助行,
// 单选模式在 ≥6 行终端为底部帮助栏让出 1 行(<6 行帮助栏不渲染,零开销)
func CalculateTableHeight(height int, multi bool) int {
	reserved := 0
	if multi {
		reserved = multiTableExtraLines
	} else if height >= HelpBarMinTermHeight {
		reserved = singleTableHelpLines
	}
	return max(height-reserved, tableHeightMin)
}

// formFieldHeight 计算 huh Select/MultiSelect 字段高度。
// 内容模型:标题一行 + 每选项一行(紧凑主题无边框/帮助,标题按一行计)。
// 内容不足一屏时按内容高度显示,不渲染底部空白;超出一屏时占满终端高度并滚动;
// 极小终端(<3 行)退化为实际终端高度,避免渲染越界。
// 底部帮助栏占一行:终端高度 ≥ 6 行时字段上限让出 1 行(帮助栏不渲染时零开销)。
// 单选项时保底随选项数收缩(标题 + 选项):固定保底 3 会令 huh 视图
// 多出一个空白行(标题+选项+空行+帮助栏),收紧后恰好标题+选项+帮助栏。
func formFieldHeight(optionCount, termHeight int) int {
	// 帮助栏仅在 ≥6 行终端渲染,故仅在该档位让出 1 行
	if termHeight >= HelpBarMinTermHeight {
		termHeight = termHeight - 1
	}
	floor := max(optionCount+1, 1)
	height := max(min(optionCount+1, termHeight), floor)
	return min(height, max(termHeight, 1))
}

// CalculateSelectHeight 计算 huh Select 字段高度,见 formFieldHeight
func CalculateSelectHeight(optionCount, termHeight int) int {
	return formFieldHeight(optionCount, termHeight)
}

// CalculateMultiSelectHeight 计算 huh MultiSelect 字段高度,规则同单选
func CalculateMultiSelectHeight(optionCount, termHeight int) int {
	return formFieldHeight(optionCount, termHeight)
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
