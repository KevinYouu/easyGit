package form

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// ListKind 列表选择类型:单选 Enter 确认,多选 Space 切换 + Enter 确认。
type ListKind int

const (
	ListSingle ListKind = iota
	ListMulti
)

// listModel 统一列表模型:单选(原表格单选)与多选(原表格多选)
// 共用同一套布局/渲染/键位逻辑,差异仅由 kind 决定。
type listModel struct {
	table        table.Model
	styles       table.Styles
	choices      []config.Option
	kind         ListKind
	wrap         bool         // 光标到顶端/底端时是否循环到另一端
	specs        []ColumnSpec // 非 nil 时启用自适应多列布局(任意列数,宽度策略见 ColumnSpec)
	colWidths    []int        // 自适应多列布局各列宽(applyLayout 计算)
	selected     map[int]bool // 仅多选使用
	confirmed    bool
	quitting     bool
	width        int
	height       int
	messageWidth int
	title        string
}

func (m *listModel) Init() tea.Cmd { return nil }

func (m *listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "q":
			// q:单选 = 取消,多选 = 确认当前选择
			if m.kind == ListSingle {
				m.quitting = true
			} else {
				m.confirmed = true
			}
			return m, tea.Quit
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case " ", "space":
			// 仅多选支持空格切换选中
			if m.kind == ListMulti {
				cursor := m.table.Cursor()
				m.selected[cursor] = !m.selected[cursor]
				m.rebuildRows(cursor)
			}
			return m, nil
		case "up", "k":
			if m.table.Cursor() > 0 {
				m.table, cmd = m.table.Update(msg)
				m.rebuildRows(m.table.Cursor()) // 重绘指示列,让 ❯ 跟随光标
			} else if m.wrap && len(m.choices) > 0 {
				// 循环导航:顶部按 ↑ 跳到最后一个选项
				m.table.SetCursor(len(m.choices) - 1)
				m.rebuildRows(len(m.choices) - 1)
			}
			return m, cmd
		case "down", "j":
			if m.table.Cursor() < len(m.choices)-1 {
				m.table, cmd = m.table.Update(msg)
				m.rebuildRows(m.table.Cursor()) // 重绘指示列,让 ❯ 跟随光标
			} else if m.wrap && len(m.choices) > 0 {
				// 循环导航:底部按 ↓ 跳到第一个选项
				m.table.SetCursor(0)
				m.rebuildRows(0)
			}
			return m, cmd
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *listModel) View() tea.View {
	if m.confirmed || m.quitting {
		return tea.NewView("")
	}

	// 表格 View 自带空 header 行,移除避免渲染空行
	tableView := m.table.View()
	lines := strings.Split(tableView, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	content := strings.Join(lines, "\n")
	// 多选始终渲染完整结构;单选在 <6 行终端零开销(无标题/分隔线/帮助栏)
	if m.kind == ListMulti || m.height >= HelpBarMinTermHeight {
		// 标题行:多选拼已选计数,单选仅标题
		titleView := lipgloss.NewStyle().
			Foreground(theme.PrimaryColor).
			Bold(true).
			Render(m.title)
		if m.kind == ListMulti {
			count := 0
			for _, selected := range m.selected {
				if selected {
					count++
				}
			}
			countView := lipgloss.NewStyle().
				Foreground(theme.MutedForeground).
				Render(fmt.Sprintf(i18n.T("table.multi.selected.count"), count))
			titleView += countView
		}
		titleLine := SafeTruncate(titleView, max(m.width-cellPaddingWidth, footerMinWidth))
		// 帮助栏:多选含 Space 切换,单选仅导航键
		helpKeys := selectHelpKeys()
		if m.kind == ListMulti {
			helpKeys = multiSelectHelpKeys()
		}
		footer := SafeTruncate(RenderHelpBar(helpKeys, max(m.width-cellPaddingWidth, footerMinWidth)), max(m.width-cellPaddingWidth, footerMinWidth))
		content = fmt.Sprintf("%s\n%s\n%s\n%s\n%s", titleLine, theme.GetHorizontalRule(m.width), content, theme.GetHorizontalRule(m.width), footer)
	}
	// 宽屏富余时水平居中
	if ShouldCenterTable(m.width, m.table.Columns()) {
		content = lipgloss.PlaceHorizontal(m.width, lipgloss.Center, content)
	}
	v := tea.NewView(content)
	// bubbletea v2 全屏模式改为声明式:View 中设置 AltScreen
	v.AltScreen = true
	return v
}

// applyLayout 按当前终端尺寸重建表格:列与行在同一代码路径构建,
// 避免 bubbles/table 在行列数量不一致时渲染越界(renderRow 按下标取列);
// 布局参数统一由 kind 决定(多选含复选框列并预留标题行);
// 自适应多列模式在非紧凑终端按 ColumnSpec 分列,紧凑终端走单列合并。
func (m *listModel) applyLayout() {
	withCheckbox := m.kind == ListMulti
	var columns []table.Column
	if m.specs != nil && LayoutMode(m.width) != LayoutCompact {
		rows := m.cellRows()
		m.colWidths = CalculateAdaptiveColumns(m.width, m.specs, rows, withCheckbox)
		if withCheckbox {
			columns = append(columns, table.Column{Title: "", Width: indicatorColWidth}, table.Column{Title: "", Width: checkboxColWidth})
		} else {
			columns = append(columns, table.Column{Title: "", Width: indicatorColWidth})
		}
		for i := range m.specs {
			columns = append(columns, table.Column{Title: "", Width: m.colWidths[i]})
		}
	} else {
		columns = CalculateColumns(m.width, withCheckbox)
		m.messageWidth = CalculateMessageWidth(m.width, withCheckbox)
	}
	cursor := m.table.Cursor()
	// 表格高度不超过内容行数:内容不足一屏时按内容显示,
	// 超出时占满终端由 viewport 滚动
	height := min(CalculateTableHeight(m.height), max(len(m.choices), 1))
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

// cellRows 自适应多列模式的行单元格矩阵(按列索引取值)
func (m *listModel) cellRows() [][]string {
	rows := make([][]string, len(m.choices))
	for i, opt := range m.choices {
		rows[i] = optionCells(opt)
	}
	return rows
}

// optionCells 多列模式单元格:Option.Cells 非空时按列索引取值,否则取 [Label, Description]。
func optionCells(opt config.Option) []string {
	if len(opt.Cells) > 0 {
		return opt.Cells
	}
	cells := []string{opt.Label}
	if opt.Description != "" {
		cells = append(cells, opt.Description)
	}
	return cells
}

// rebuildRows 按当前布局模式重新生成行数据;光标行在最左指示列渲染 ❯
func (m *listModel) rebuildRows(cursorRow int) {
	rows := make([]table.Row, 0, len(m.choices))
	for i, opt := range m.choices {
		rows = append(rows, m.optionRow(i, cursorRow, opt))
	}
	m.table.SetRows(rows)
}

// optionRow 将单个选项按当前布局模式格式化为表格行;指示列宽 2,
// 与 huh 的 "❯ " 指示符同宽,光标行显示 ❯ 其余留空;多选额外含复选框列。
// 两列模式(非紧凑):名称列自动宽度不截断(仅超上限截断),描述列占满剩余。
func (m *listModel) optionRow(i, cursorRow int, opt config.Option) table.Row {
	indicator := ""
	if i == cursorRow {
		indicator = "❯"
	}
	if m.specs != nil && LayoutMode(m.width) != LayoutCompact {
		row := make(table.Row, 0, len(m.specs)+2)
		row = append(row, indicator)
		cells := optionCells(opt)
		if m.kind == ListMulti {
			checkbox := "[ ]"
			if m.selected[i] {
				checkbox = "[x]"
			}
			row = append(row, checkbox)
		}
		for ci := range m.specs {
			text := ""
			if ci < len(cells) {
				text = cells[ci]
			}
			if ci < len(m.colWidths) {
				text = SafeTruncate(text, m.colWidths[ci])
			}
			row = append(row, text)
		}
		return row
	}
	label := m.rowLabel(opt)
	if m.kind == ListMulti {
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}
		switch LayoutMode(m.width) {
		case LayoutCompact:
			return table.Row{indicator, checkbox, formatCompactCommit(label, m.width-indicatorColWidth-checkboxColWidth-tableInsetWidth-cellPaddingWidth)}
		case LayoutThreeCol:
			hash, message, date, _ := parseCommitInfo(label, m.messageWidth)
			return table.Row{indicator, checkbox, hash, message, date}
		default:
			hash, message, date, author := parseCommitInfo(label, m.messageWidth)
			return table.Row{indicator, checkbox, hash, message, date, author}
		}
	}
	switch LayoutMode(m.width) {
	case LayoutCompact:
		return table.Row{indicator, formatCompactCommit(label, m.width-indicatorColWidth-tableInsetWidth-cellPaddingWidth)}
	case LayoutThreeCol:
		hash, message, date, _ := parseCommitInfo(label, m.messageWidth)
		return table.Row{indicator, hash, message, date}
	default:
		hash, message, date, author := parseCommitInfo(label, m.messageWidth)
		return table.Row{indicator, hash, message, date, author}
	}
}

// rowLabel 选项显示文本:有说明时按「名称 + 说明」纯文本拼接(不嵌 ANSI,
// 与表格单元格渲染兼容);自适应多列模式返回全部单元格空格连接
// (紧凑单列降级)。注意:accessible 纯文本列表已统一改走 optionCells
// 合并(runAccessibleList),此处仅服务表格渲染路径。
func (m *listModel) rowLabel(opt config.Option) string {
	if m.specs != nil {
		return strings.Join(optionCells(opt), " ")
	}
	return optionDisplayText(opt)
}

// optionDisplayText 选项显示文本:说明为空仅名称,否则「名称 + 说明」。
func optionDisplayText(opt config.Option) string {
	if opt.Description == "" {
		return opt.Label
	}
	return opt.Label + " " + opt.Description
}

// NewListModel 构造统一列表模型(生产与渲染测试共用同一构造路径,
// 防止生产配置与测试复刻漂移);默认不循环导航,仅个别命令
// 经 ListFormWrap/NewListModelWrap 显式开启;单选光标落预选项,多选预填选中集。
func NewListModel(title string, options []config.Option, kind ListKind, preselected ...string) *listModel {
	return newListModel(title, options, kind, false, nil, preselected...)
}

// NewListModelWrap 同 NewListModel,开启循环导航:光标在顶部按 ↑/k 跳至末尾,
// 在底部按 ↓/j 跳回顶部。
func NewListModelWrap(title string, options []config.Option, kind ListKind, preselected ...string) *listModel {
	return newListModel(title, options, kind, true, nil, preselected...)
}

// NewListModelColumns 构造自适应多列布局列表(列数不硬编码,由 specs 声明):
// 每行按 ColumnSpec 分列渲染,宽度策略见 CalculateAdaptiveColumns;
// 单元格来源:Option.Cells 非空时按列索引取值,否则取 [Label, Description];
// 极窄终端(< minColumnWidth)自动降级为单列(单元格空格连接后截断)。
func NewListModelColumns(title string, specs []ColumnSpec, options []config.Option, kind ListKind, preselected ...string) *listModel {
	return newListModel(title, options, kind, false, specs, preselected...)
}

// newListModel 构造逻辑实现;wrap 与 specs 为内部参数,导出入口默认零值。
func newListModel(title string, options []config.Option, kind ListKind, wrap bool, specs []ColumnSpec, preselected ...string) *listModel {
	// 检测终端尺寸
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = defaultTermWidth
		height = defaultTermHeight
	}

	m := &listModel{
		choices: options,
		kind:    kind,
		wrap:    wrap,
		specs:   specs,
		width:   width,
		height:  height,
		title:   title,
		styles:  defaultTableStyles(),
	}
	if kind == ListMulti {
		m.selected = make(map[int]bool)
	}
	m.table = table.New(table.WithFocused(true))
	m.applyLayout()

	if len(preselected) > 0 {
		if kind == ListSingle {
			// 单选:光标落到预选值所在行
			if preselected[0] != "" {
				for i, opt := range options {
					if opt.Value == preselected[0] {
						m.rebuildRows(i)
						m.table.SetCursor(i)
						break
					}
				}
			}
		} else {
			// 多选:预填选中集
			for _, value := range preselected {
				for i, opt := range options {
					if opt.Value == value {
						m.selected[i] = true
						break
					}
				}
			}
			m.rebuildRows(m.table.Cursor())
		}
	}
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

// stdinBuf 共享的 stdin 缓冲读取器。
// ListForm 可能连续多次调用(如 reset 先选提交再选模式),必须复用同一缓冲:
// 若每次新建缓冲读取器,前一个表单会提前读走后续输入,后一个表单会误判为 EOF。
var stdinBuf = bufio.NewReader(os.Stdin)

// lineReader 包装共享 stdinBuf 的行级读取器,供 huh accessible 表单使用:
// huh 内部 PromptString 每次新建 bufio.Scanner,Scanner 会从底层 Reader
// 预读大块数据到私有缓冲,若直接传入共享 stdinBuf,多字段表单(如 MultiInput)
// 的后续字段将读到 EOF(数据被废弃 Scanner 吞掉)。包装后每次 Read 至多返回
// 一行,Scanner 消费一行后内部无残留,后续表单可继续从 stdinBuf 读下一行。
type lineReader struct {
	br      *bufio.Reader
	pending []byte
	eof     bool
}

// Read 每次至多返回一行数据;底层 EOF 且已消费完最后一行时返回 io.EOF。
func (l *lineReader) Read(p []byte) (int, error) {
	if len(l.pending) == 0 && !l.eof {
		line, err := l.br.ReadString('\n')
		l.pending = []byte(line)
		if err != nil {
			l.eof = true
			if len(l.pending) == 0 {
				return 0, err
			}
		}
	}
	n := copy(p, l.pending)
	l.pending = l.pending[n:]
	if len(l.pending) == 0 && l.eof {
		return n, io.EOF
	}
	return n, nil
}

// ListForm 列表选择入口:单选返回 1 个值,多选返回全部选中值(按选项顺序)。
// 默认不循环导航;TERM=dumb 时走无障碍纯文本路径(脚本管道兼容)。
func ListForm(title string, options []config.Option, kind ListKind, preselected ...string) ([]string, error) {
	return runListForm(title, options, kind, false, nil, preselected...)
}

// ListFormWrap 同 ListForm,开启循环导航:光标在顶部按 ↑/k 跳至末尾,
// 在底部按 ↓/j 跳回顶部。
func ListFormWrap(title string, options []config.Option, kind ListKind, preselected ...string) ([]string, error) {
	return runListForm(title, options, kind, true, nil, preselected...)
}

// ListFormColumns 同 ListForm,启用自适应多列布局(列数不硬编码,由 specs
// 声明;宽度策略见 ColumnSpec)。用于配置中心等「名称 + 单行说明」列表。
func ListFormColumns(title string, specs []ColumnSpec, options []config.Option, kind ListKind, preselected ...string) ([]string, error) {
	return runListForm(title, options, kind, false, specs, preselected...)
}

// runListForm 列表入口共享实现;wrap 与 specs 为内部参数,导出入口默认零值。
func runListForm(title string, options []config.Option, kind ListKind, wrap bool, specs []ColumnSpec, preselected ...string) ([]string, error) {
	m := newListModel(title, options, kind, wrap, specs, preselected...)

	if isAccessibleMode() {
		return runAccessibleList(os.Stdout, stdinBuf, title, options, kind, preselected...)
	}

	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	listModel, ok := finalModel.(*listModel)
	if !ok {
		return nil, fmt.Errorf("%s", i18n.T("table.no.selection"))
	}

	if listModel.quitting && !listModel.confirmed {
		return nil, fmt.Errorf("%s", i18n.T("table.user.aborted"))
	}

	if listModel.confirmed {
		if listModel.kind == ListSingle {
			// 单选:取光标所在选项
			selectedRow := listModel.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(options) {
				return []string{options[selectedRow].Value}, nil
			}
		} else {
			// 多选:按选项顺序收集选中值
			selectedValues := make([]string, 0, len(options))
			for i, opt := range options {
				if listModel.selected[i] {
					selectedValues = append(selectedValues, opt.Value)
				}
			}
			if len(selectedValues) > 0 {
				return selectedValues, nil
			}
		}
	}

	return nil, fmt.Errorf("%s", i18n.T("table.no.selection"))
}

// runAccessibleList 无障碍纯文本列表(TERM=dumb 替代 TUI):
// 打印「标题 + 编号. 标签」行,从 r 读序号——单选一个序号,
// 多选逗号分隔序号(空行 = 全不选);EOF 视为取消。
// 行文本 = 全部单元格空格连接(自适应多列模式兼容);单选预选项
// 在标签后附加 (current) 标记(替代 TUI 光标落位)。
func runAccessibleList(w io.Writer, r io.Reader, title string, options []config.Option, kind ListKind, preselected ...string) ([]string, error) {
	fmt.Fprintf(w, "%s:\n", title)
	for i, opt := range options {
		var line strings.Builder
		line.WriteString(ansi.Strip(strings.Join(optionCells(opt), " ")))
		if kind == ListSingle {
			if slices.Contains(preselected, opt.Value) {
				line.WriteString(" " + i18n.T("form.accessible.current"))
			}
		}
		fmt.Fprintf(w, "%d. %s\n", i+1, line.String())
	}

	// 用 ReadString 而非 bufio.Scanner:共享 *bufio.Reader 时其内部缓冲必须保留,
	// Scanner 会自行缓冲并吞掉后续表单的输入;ReadString 只消费到行尾。
	br, shared := r.(*bufio.Reader)
	if !shared {
		br = bufio.NewReader(r)
	}
	readLine := func() (line string, ok bool) {
		line, err := br.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			return line, true // 有效行(末尾无换行时 ReadString 同时返回 io.EOF,仍先消费)
		}
		if err != nil {
			return "", false // 无内容且 EOF/出错:结束
		}
		return "", true // 空行(多选 = 全不选;单选 = 非法重试)
	}
	for {
		line, ok := readLine()
		if !ok {
			// EOF:取消
			return nil, huh.ErrUserAborted
		}
		// 多选空行 = 全不选
		if kind == ListMulti && line == "" {
			return nil, nil
		}
		if kind == ListSingle {
			n, err := strconv.Atoi(line)
			if err != nil || n < 1 || n > len(options) {
				fmt.Fprintf(w, "Invalid: must be a number between 1 and %d\n", len(options))
				continue
			}
			return []string{options[n-1].Value}, nil
		}
		// 多选:逗号分隔序号,任一非法则整体报错重试
		parts := strings.Split(line, ",")
		values := make([]string, 0, len(parts))
		valid := true
		for _, part := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil || n < 1 || n > len(options) {
				valid = false
				break
			}
			values = append(values, options[n-1].Value)
		}
		if !valid {
			fmt.Fprintf(w, "Invalid: must be a number between 1 and %d\n", len(options))
			continue
		}
		return values, nil
	}
}

// StringOptions 将纯字符串列表包装为选项(Label = Value = 字符串本身),
// 供 MultiSelectForm 迁移的调用方构造 []config.Option。
func StringOptions(strs []string) []config.Option {
	options := make([]config.Option, len(strs))
	for i, s := range strs {
		options[i] = config.Option{Label: s, Value: s}
	}
	return options
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
	} else {
		// 无空格标签(如 fix/origin/main.go 等非提交类选项):
		// 整体放入消息列,避免渲染成空行
		message = SafeTruncate(firstLine, messageWidth)
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
