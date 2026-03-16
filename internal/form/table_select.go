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
	choices      []config.Option
	selected     bool
	quitting     bool
	compact      bool
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
		m.compact = msg.Height < 15

		// 重新计算列宽
		var columns []table.Column
		if m.compact {
			columns = []table.Column{
				{Title: "", Width: m.width - 4},
			}
		} else {
			m.messageWidth = calculateMessageWidth(m.width)
			columns = []table.Column{
				{Title: "", Width: 8},
				{Title: "", Width: m.messageWidth},
				{Title: "", Width: 12},
				{Title: "", Width: 10},
			}
		}
		m.table.SetColumns(columns)

		// 重新计算表格高度
		tableHeight := calculateTableHeight(m.height, m.compact)
		m.table.SetHeight(tableHeight)

		// 重新生成行数据（使用新的列宽）
		var rows []table.Row
		for _, opt := range m.choices {
			if m.compact {
				compactInfo := formatCompactCommit(opt.Label, m.width-6)
				rows = append(rows, table.Row{compactInfo})
			} else {
				hash, message, date, author := parseCommitInfo(opt.Label, m.messageWidth)
				rows = append(rows, table.Row{hash, message, date, author})
			}
		}
		m.table.SetRows(rows)

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

	return strings.Join(lines, "\n")
}

func NewTableSelectModel(options []config.Option) *tableModel {
	// 检测终端尺寸
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		height = 24
	}

	// 根据终端高度决定是否使用紧凑模式
	compact := height < 15

	// 动态计算消息列宽度
	messageWidth := calculateMessageWidth(width)

	// 创建表格列
	var columns []table.Column
	if compact {
		// 紧凑模式：单列显示，无标题
		columns = []table.Column{
			{Title: "", Width: width - 4},
		}
	} else {
		// 普通模式：多列显示，使用动态列宽
		columns = []table.Column{
			{Title: "", Width: 8},
			{Title: "", Width: messageWidth},
			{Title: "", Width: 12},
			{Title: "", Width: 10},
		}
	}

	// 创建表格行
	var rows []table.Row
	for _, opt := range options {
		if compact {
			// 紧凑模式：将所有信息合并到一行
			compactInfo := formatCompactCommit(opt.Label, width-6)
			rows = append(rows, table.Row{compactInfo})
		} else {
			// 普通模式：解析提交信息到多列，使用动态消息宽度
			hash, message, date, author := parseCommitInfo(opt.Label, messageWidth)
			rows = append(rows, table.Row{hash, message, date, author})
		}
	}

	// 计算表格高度
	tableHeight := calculateTableHeight(height, compact)

	// 创建表格
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	// 设置表格样式
	s := table.DefaultStyles()
	// 完全隐藏表头，不占用任何高度
	s.Header = lipgloss.NewStyle().
		Height(0).
		MaxHeight(0).
		Border(lipgloss.HiddenBorder())
	s.Selected = s.Selected.
		Foreground(theme.SelectionFg).
		Background(theme.SelectionBg).
		Bold(true)

	t.SetStyles(s)

	return &tableModel{
		table:        t,
		choices:      options,
		compact:      compact,
		width:        width,
		height:       height,
		messageWidth: messageWidth,
	}
}

func calculateMessageWidth(terminalWidth int) int {
	// 根据终端宽度动态计算消息列的宽度
	// 固定列：hash(8) + date(12) + author(10) + 间距(5) = 35
	fixedWidth := 35
	messageWidth := min(max(terminalWidth-fixedWidth, 20), 100)

	return messageWidth
}

func calculateTableHeight(terminalHeight int, compact bool) int {
	if compact {
		// 紧凑模式：减少高度，移除额外空间
		height := max(terminalHeight-2, 3)
		return height
	} else {
		// 普通模式：也减少高度
		height := max(terminalHeight-4, 5)
		return height
	}
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

			// 格式化并截断
			result := fmt.Sprintf("%s %s (%s)", hash, message, date)
			if len(result) > maxWidth {
				return result[:maxWidth-3] + "..."
			}
			return result
		}
	}

	// 如果解析失败，返回截断的原始文本
	if len(commitInfo) > maxWidth {
		return commitInfo[:maxWidth-3] + "..."
	}
	return commitInfo
}

func parseCommitInfo(commitInfo string, messageWidth int) (hash, message, date, author string) {
	lines := strings.Split(commitInfo, "\n")
	if len(lines) >= 2 {
		// 解析第一行：hash + message
		firstLine := lines[0]
		parts := strings.SplitN(firstLine, " ", 2)
		if len(parts) >= 2 {
			hash = parts[0]
			message = parts[1]
			if len(message) > messageWidth {
				message = message[:messageWidth-3] + "..."
			}
		}

		// 解析第二行：date + author
		secondLine := lines[1]
		dateParts := strings.Split(secondLine, " • ")
		if len(dateParts) >= 2 {
			date = strings.TrimSpace(dateParts[0])
			author = strings.TrimSpace(dateParts[1])
			if len(author) > 10 {
				author = author[:7] + "..."
			}
		}
	}

	return hash, message, date, author
}

// TableSelectForm 创建一个基于表格的选择表单
func TableSelectForm(options []config.Option) (label, value string, err error) {
	m := NewTableSelectModel(options)

	// 根据模式选择是否使用全屏
	var p *tea.Program
	if m.compact {
		p = tea.NewProgram(m)
	} else {
		p = tea.NewProgram(m, tea.WithAltScreen())
	}

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
