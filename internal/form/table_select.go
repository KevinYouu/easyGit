package form

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"golang.org/x/term"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(theme.BorderColor)

type tableModel struct {
	table        table.Model
	styles       table.Styles
	choices      []config.Option
	selected     bool
	quitting     bool
	width        int
	height       int
	messageWidth int
}

func (m tableModel) Init() tea.Cmd { return nil }

func (m tableModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// 终端尺寸变化时重新计算布局
		m.width = msg.Width
		m.height = msg.Height
		m.applyLayout()

		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.selected = true
			return m, tea.Quit
		case "up", "k":
			// 不循环：检查是否已经在第一行
			if m.table.Cursor() > 0 {
				m.table, cmd = m.table.Update(msg)
				m.rebuildRows(m.table.Cursor()) // 重绘指示列,让 ❯ 跟随光标
			}
			return m, cmd
		case "down", "j":
			// 不循环：检查是否已经在最后一行
			if m.table.Cursor() < len(m.choices)-1 {
				m.table, cmd = m.table.Update(msg)
				m.rebuildRows(m.table.Cursor()) // 重绘指示列,让 ❯ 跟随光标
			}
			return m, cmd
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() tea.View {
	if m.selected || m.quitting {
		return tea.NewView("")
	}

	// 表格 View 自带空 header 行,移除避免渲染空行
	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	content := strings.Join(lines, "\n")
	// 宽屏富余时水平居中
	if ShouldCenterTable(m.width, m.table.Columns()) {
		content = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	}
	v := tea.NewView(content)
	// 底部:分隔线 + 帮助栏一行(≥6 行终端渲染,极小终端零开销);计算高度时已让出该行
	if m.height >= HelpBarMinTermHeight {
		v.SetContent(v.Content + "\n" + theme.GetHorizontalRule(m.width))
		v = AppendHelpBar(v, tableSelectHelpKeys(), m.width)
	}
	// bubbletea v2 全屏模式改为声明式:View 中设置 AltScreen
	v.AltScreen = true
	return v
}

// applyLayout 按当前终端尺寸重建表格:列与行在同一代码路径构建,
// 避免 bubbles/table 在行列数量不一致时渲染越界(renderRow 按下标取列)
func (m *tableModel) applyLayout() {
	columns := CalculateColumns(m.width, false)
	m.messageWidth = CalculateMessageWidth(m.width, false)
	cursor := m.table.Cursor()
	// 表格高度不超过内容行数:内容不足一屏时按内容显示,
	// 超出时占满终端由 viewport 滚动
	height := min(CalculateTableHeight(m.height, false), max(len(m.choices), 1))
	// SetHeight 内部会扣除 header 行,这里补偿使可视行数等于计算值
	height += tableHeaderLines

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	// 显式设置表格宽度:viewport 宽度为 0 时渲染为空(v2 行为)
	t.SetWidth(TotalTableWidth(columns))
	t.SetStyles(m.styles)
	m.table = t
	m.rebuildRows(cursor)
	// 重建后恢复光标位置
	m.table.SetCursor(cursor)
}

// rebuildRows 按当前布局模式重新生成行数据;光标行在最左指示列渲染 ❯
func (m *tableModel) rebuildRows(cursorRow int) {
	rows := make([]table.Row, 0, len(m.choices))
	for i, opt := range m.choices {
		rows = append(rows, m.optionRow(i, cursorRow, opt))
	}
	m.table.SetRows(rows)
}

// optionRow 将单个选项按当前布局模式格式化为表格行;指示列宽 2,
// 与 huh 的 "❯ " 指示符同宽,光标行显示 ❯ 其余留空
func (m *tableModel) optionRow(i, cursorRow int, opt config.Option) table.Row {
	indicator := ""
	if i == cursorRow {
		indicator = "❯"
	}
	switch LayoutMode(m.width) {
	case LayoutCompact:
		return table.Row{indicator, formatCompactCommit(opt.Label, m.width-indicatorColWidth-tableInsetWidth-cellPaddingWidth)}
	case LayoutThreeCol:
		hash, message, date, _ := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{indicator, hash, message, date}
	default:
		hash, message, date, author := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{indicator, hash, message, date, author}
	}
}

func NewTableSelectModel(options []config.Option) *tableModel {
	// 检测终端尺寸
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultTermWidth
		height = defaultTermHeight
	}

	m := &tableModel{
		choices: options,
		width:   width,
		height:  height,
		styles:  defaultTableStyles(),
	}
	m.applyLayout()
	return m
}

// defaultTableStyles 表格统一样式:完全隐藏表头,选中行高亮
func defaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Height(0).
		MaxHeight(0).
		Border(lipgloss.HiddenBorder())
	s.Selected = s.Selected.
		Foreground(theme.SelectionFg).
		Background(theme.SelectionBg).
		Bold(true)
	return s
}

func formatCompactCommit(commitInfo string, maxWidth int) string {
	// 将多行提交信息格式化为紧凑的单行
	lines := strings.Split(commitInfo, "\n")
	if len(lines) >= 2 {
		// 提取关键信息
		firstLine := lines[0]  // hash + message
		secondLine := lines[1] // date + author

		// 解析第一行
		parts := strings.SplitN(firstLine, " ", 2)
		if len(parts) >= 2 {
			hash := parts[0]
			message := parts[1]

			// 解析第二行获取日期
			dateParts := strings.Split(secondLine, " • ")
			date := ""
			if len(dateParts) >= 1 {
				date = strings.TrimSpace(dateParts[0])
			}

			// 格式化并按显示宽度截断
			result := fmt.Sprintf("%s %s (%s)", hash, message, date)
			return SafeTruncate(result, maxWidth)
		}
	}

	// 如果解析失败，返回截断的原始文本
	return SafeTruncate(commitInfo, maxWidth)
}

func parseCommitInfo(commitInfo string, messageWidth int) (hash, message, date, author string) {
	lines := strings.Split(commitInfo, "\n")
	// 单行标签(如 reset 模式 "Soft - 保留工作目录")也能解析,首词作 hash 列
	firstLine := lines[0]
	parts := strings.SplitN(firstLine, " ", 2)
	if len(parts) >= 2 {
		hash = parts[0]
		message = SafeTruncate(parts[1], messageWidth)
	}

	// 两行提交信息(GetRecentCommits 格式)再解析日期与作者
	if len(lines) >= 2 {
		secondLine := lines[1]
		dateParts := strings.Split(secondLine, " • ")
		if len(dateParts) >= 2 {
			date = SafeTruncate(strings.TrimSpace(dateParts[0]), dateColWidth)
			author = SafeTruncate(strings.TrimSpace(dateParts[1]), authorColWidth)
		}
	}

	return hash, message, date, author
}

// TableSelectForm 创建一个基于表格的选择表单
func TableSelectForm(options []config.Option) (label, value string, err error) {
	m := NewTableSelectModel(options)

	// 统一使用全屏模式(紧凑仅影响列布局,不决定 AltScreen)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return "", "", err
	}

	if tableModel, ok := finalModel.(tableModel); ok {
		if tableModel.quitting && !tableModel.selected {
			return "", "", fmt.Errorf("%s", i18n.T("table.user.aborted"))
		}

		if tableModel.selected {
			// 获取选中的行索引
			selectedRow := tableModel.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(options) {
				return options[selectedRow].Label, options[selectedRow].Value, nil
			}
		}
	}

	return "", "", fmt.Errorf("%s", i18n.T("table.no.selection"))
}
