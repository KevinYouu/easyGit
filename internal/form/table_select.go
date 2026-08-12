package form

import (
	"fmt"
	"os"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	case tea.KeyMsg:
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
			}
			return m, cmd
		case "down", "j":
			// 不循环：检查是否已经在最后一行
			if m.table.Cursor() < len(m.choices)-1 {
				m.table, cmd = m.table.Update(msg)
			}
			return m, cmd
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m tableModel) View() string {
	if m.selected || m.quitting {
		return ""
	}

	// 获取表格视图并移除开头的空白行
	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")

	// 移除开头的空白行
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	content := strings.Join(lines, "\n")
	// 宽屏富余时水平居中
	if ShouldCenterTable(m.width, m.table.Columns()) {
		content = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	}
	return content
}

// applyLayout 按当前终端尺寸重建表格:列与行在同一代码路径构建,
// 避免 bubbles/table 在行列数量不一致时渲染越界(renderRow 按下标取列)
func (m *tableModel) applyLayout() {
	columns := CalculateColumns(m.width, false)
	m.messageWidth = CalculateMessageWidth(m.width, false)
	cursor := m.table.Cursor()

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(CalculateTableHeight(m.height, false)),
	)
	t.SetStyles(m.styles)
	m.table = t
	m.rebuildRows()
	// 重建后恢复光标位置
	m.table.SetCursor(cursor)
}

// rebuildRows 按当前布局模式重新生成行数据
func (m *tableModel) rebuildRows() {
	rows := make([]table.Row, 0, len(m.choices))
	for _, opt := range m.choices {
		rows = append(rows, m.optionRow(opt))
	}
	m.table.SetRows(rows)
}

// optionRow 将单个选项按当前布局模式格式化为表格行
func (m *tableModel) optionRow(opt config.Option) table.Row {
	switch LayoutMode(m.width) {
	case LayoutCompact:
		return table.Row{formatCompactCommit(opt.Label, m.width-tableInsetWidth-cellPaddingWidth)}
	case LayoutThreeCol:
		hash, message, date, _ := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{hash, message, date}
	default:
		hash, message, date, author := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{hash, message, date, author}
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
	if len(lines) >= 2 {
		// 解析第一行：hash + message
		firstLine := lines[0]
		parts := strings.SplitN(firstLine, " ", 2)
		if len(parts) >= 2 {
			hash = parts[0]
			message = SafeTruncate(parts[1], messageWidth)
		}

		// 解析第二行：date + author
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
	p := tea.NewProgram(m, tea.WithAltScreen())

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
