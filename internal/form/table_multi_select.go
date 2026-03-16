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

type tableMultiModel struct {
	table        table.Model
	choices      []config.Option
	selected     map[int]bool
	confirmed    bool
	quitting     bool
	compact      bool
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
		m.compact = msg.Height < 15
		m.updateLayout()
		return m, nil

	case tea.KeyMsg:
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
			m.updateLayout() // re-render rows to show checkbox changes
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
	var columns []table.Column

	checkboxWidth := 4

	if m.compact {
		columns = []table.Column{
			{Title: "", Width: checkboxWidth},
			{Title: "", Width: m.width - checkboxWidth - 4},
		}
	} else {
		m.messageWidth = max(calculateMessageWidth(m.width)-checkboxWidth, 10)
		columns = []table.Column{
			{Title: "", Width: checkboxWidth},
			{Title: "", Width: 8},
			{Title: "", Width: m.messageWidth},
			{Title: "", Width: 12},
			{Title: "", Width: 10},
		}
	}
	m.table.SetColumns(columns)
	m.table.SetHeight(calculateTableHeight(m.height, m.compact) - 2) // reserve space for title and footer

	var rows []table.Row
	for i, opt := range m.choices {
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		if m.compact {
			compactInfo := formatCompactCommit(opt.Label, m.width-checkboxWidth-6)
			rows = append(rows, table.Row{checkbox, compactInfo})
		} else {
			hash, message, date, author := parseCommitInfo(opt.Label, m.messageWidth)
			rows = append(rows, table.Row{checkbox, hash, message, date, author})
		}
	}
	m.table.SetRows(rows)
}

func (m *tableMultiModel) View() string {
	if m.confirmed || m.quitting {
		return ""
	}

	titleView := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true).Render(m.title)

	// Count selected
	count := 0
	for _, sel := range m.selected {
		if sel {
			count++
		}
	}
	countView := lipgloss.NewStyle().Foreground(theme.SuccessColor).Render(fmt.Sprintf(" (%d selected)", count))

	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	footer := theme.RenderMuted(i18n.T("table.multi.help"))

	return fmt.Sprintf("%s%s\n\n%s\n%s\n", titleView, countView, strings.Join(lines, "\n"), footer)
}

// TableMultiSelectForm creates a table-based multi-select form suitable for commits
func TableMultiSelectForm(title string, options []config.Option) ([]string, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 80
		height = 24
	}

	compact := height < 15

	m := tableMultiModel{
		choices:  options,
		selected: make(map[int]bool),
		compact:  compact,
		width:    width,
		height:   height,
		title:    title,
		table:    table.New(table.WithFocused(true)),
	}

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).MaxHeight(0).Border(lipgloss.HiddenBorder())
	s.Selected = s.Selected.Foreground(theme.SelectionFg).Background(theme.SelectionBg).Bold(true)
	m.table.SetStyles(s)

	m.updateLayout()

	var p *tea.Program
	if m.compact {
		p = tea.NewProgram(&m)
	} else {
		p = tea.NewProgram(&m, tea.WithAltScreen())
	}

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
