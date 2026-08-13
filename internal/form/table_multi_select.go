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

type tableMultiModel struct {
	table        table.Model
	styles       table.Styles
	choices      []config.Option
	selected     map[int]bool
	confirmed    bool
	quitting     bool
	width        int
	height       int
	messageWidth int
	title        string
}

func (m *tableMultiModel) Init() tea.Cmd { return nil }

func (m *tableMultiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case " ", "space":
			cursor := m.table.Cursor()
			m.selected[cursor] = !m.selected[cursor]
			m.rebuildRows() // 仅重绘行以显示多选框变化,避免重建表格
			return m, nil
		case "up", "k":
			if m.table.Cursor() > 0 {
				m.table, cmd = m.table.Update(msg)
			}
			return m, cmd
		case "down", "j":
			if m.table.Cursor() < len(m.choices)-1 {
				m.table, cmd = m.table.Update(msg)
			}
			return m, cmd
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *tableMultiModel) updateLayout() {
	// 尺寸变化时重建表格:列与行在同一代码路径构建,
	// 避免 bubbles/table 在行列数量不一致时渲染越界(renderRow 按下标取列)
	columns := CalculateColumns(m.width, true)
	m.messageWidth = CalculateMessageWidth(m.width, true)
	cursor := m.table.Cursor()
	// 表格高度不超过内容行数:内容不足一屏时按内容显示,
	// 超出时占满终端由 viewport 滚动
	height := min(CalculateTableHeight(m.height, true), max(len(m.choices), 1))
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
	m.rebuildRows()
	// 重建后恢复光标位置
	m.table.SetCursor(cursor)
}

// rebuildRows 按当前布局重新生成行数据(含多选框)
func (m *tableMultiModel) rebuildRows() {
	var rows []table.Row
	for i, opt := range m.choices {
		rows = append(rows, m.optionRow(i, opt))
	}
	m.table.SetRows(rows)
}

// optionRow 将单个选项按当前布局模式格式化为表格行(含多选框)
func (m *tableMultiModel) optionRow(i int, opt config.Option) table.Row {
	checkbox := "[ ]"
	if m.selected[i] {
		checkbox = "[x]"
	}

	switch LayoutMode(m.width) {
	case LayoutCompact:
		return table.Row{checkbox, formatCompactCommit(opt.Label, m.width-checkboxColWidth-tableInsetWidth-cellPaddingWidth)}
	case LayoutThreeCol:
		hash, message, date, _ := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{checkbox, hash, message, date}
	default:
		hash, message, date, author := parseCommitInfo(opt.Label, m.messageWidth)
		return table.Row{checkbox, hash, message, date, author}
	}
}

func (m *tableMultiModel) View() tea.View {
	if m.confirmed || m.quitting {
		return tea.NewView("")
	}

	titleView := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(m.title)

	// Count selected
	count := 0
	for _, sel := range m.selected {
		if sel {
			count++
		}
	}
	countView := lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(fmt.Sprintf(i18n.T("table.multi.selected.count"), count))

	// 标题行固定为一行:窄屏下按终端宽度截断,避免折行
	titleLine := SafeTruncate(titleView+countView, max(m.width-cellPaddingWidth, footerMinWidth))

	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	footer := theme.RenderMuted(i18n.T("table.multi.help"))
	// 窄屏下帮助文本可能溢出,按终端宽度截断
	footer = SafeTruncate(footer, max(m.width-cellPaddingWidth, footerMinWidth))

	content := fmt.Sprintf("%s\n\n%s\n%s", titleLine, strings.Join(lines, "\n"), footer)
	// 宽屏富余时水平居中
	if ShouldCenterTable(m.width, m.table.Columns()) {
		content = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	}
	v := tea.NewView(content)
	// bubbletea v2 全屏模式改为声明式:View 中设置 AltScreen
	v.AltScreen = true
	return v
}

// NewTableMultiSelectModel 构建多选表格模型(TableMultiSelectForm 与渲染测试
// 共用同一构造路径,防止模型配置漂移)
func NewTableMultiSelectModel(title string, options []config.Option) *tableMultiModel {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultTermWidth
		height = defaultTermHeight
	}

	m := &tableMultiModel{
		choices:  options,
		selected: make(map[int]bool),
		width:    width,
		height:   height,
		title:    title,
		styles:   defaultTableStyles(),
		table:    table.New(table.WithFocused(true)),
	}

	m.updateLayout()
	return m
}

// TableMultiSelectForm creates a table-based multi-select form suitable for commits
func TableMultiSelectForm(title string, options []config.Option) ([]string, error) {
	m := NewTableMultiSelectModel(title, options)

	// 统一使用全屏模式(紧凑仅影响列布局,不决定 AltScreen)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	if resultModel, ok := finalModel.(*tableMultiModel); ok {
		if resultModel.quitting && !resultModel.confirmed {
			return nil, fmt.Errorf("%s", i18n.T("table.user.aborted"))
		}

		if resultModel.confirmed {
			var selectedValues []string
			for i, opt := range resultModel.choices {
				if resultModel.selected[i] {
					selectedValues = append(selectedValues, opt.Value)
				}
			}
			return selectedValues, nil
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("table.no.selection"))
}
