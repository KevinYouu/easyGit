package form

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/cursor"
	key "charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// ─── 自绘多输入表单 ──────────────────────────────────────────────────────────
//
// 背景:huh Input 的 Inline 布局固定「标题+简介+输入框」拼接,简介无法
// 放到输入框之后;而配置中心与 tc 的多字段表单要求矮终端内:
//   1. 简介(desc)不混入标题——每行 = 标题列 + 输入框 + 行尾弱化简介列
//   2. 不占额外纵向行(简介不用独立行/说明区)
// 因此多输入表单不再经 huh 渲染,自绘 tea.Model:每行三列(标题定宽、
// 输入框自适应、简介行尾弱化),↑/↓/j/k 上下导航,enter 推进/提交,
// esc 取消,校验错误行内显示(不占额外行)。

// multiInputModel 自绘单页多输入表单模型。
type multiInputModel struct {
	specs   []InputSpec
	values  []*string // 提交时写回
	focused int       // 聚焦字段序号
	inputs  []textinput.Model
	preview func([]string) string // 可选实时预览行(nil 省略)
	errMsg  string                // 当前聚焦字段校验错误(行内显示)

	titleMaxW int // 标题列宽(全部字段最大标题宽,对齐输入框起始)
	descMaxW  int // 简介列宽(全部字段最大简介宽,防简介溢出)
	inputW    int // 输入框统一宽度(全部字段一致,desc 列对齐;随最长内容同步变化)
	width     int
	height    int
	done      bool // 提交完成
	canceled  bool // Esc 取消
}

// newMultiInputModel 构造自绘多输入表单:每字段一个 textinput,
// 校验/占位符与单输入一致;preview 非 nil 时顶部渲染实时预览行。
func newMultiInputModel(specs []InputSpec, values []*string, preview func([]string) string) *multiInputModel {
	m := &multiInputModel{
		specs:   specs,
		values:  values,
		preview: preview,
		width:   defaultTermWidth,
		height:  defaultTermHeight,
	}
	m.inputs = make([]textinput.Model, len(specs))
	for i, spec := range specs {
		ti := textinput.New()
		ti.Prompt = " ❯ "
		ti.Placeholder = i18n.T("form.input.placeholder")
		ti.SetValue(*values[i])
		m.inputs[i] = ti
		if w := lipgloss.Width(stepTitle(i, spec.Title)); w > m.titleMaxW {
			m.titleMaxW = w
		}
		if w := lipgloss.Width(spec.Desc); w > m.descMaxW {
			m.descMaxW = w
		}
	}
	return m
}

// currentValues 取当前各字段值(预览行求值用)
func (m *multiInputModel) currentValues() []string {
	vals := make([]string, len(m.inputs))
	for i, ti := range m.inputs {
		vals[i] = ti.Value()
	}
	return vals
}

// validate 校验第 i 个字段:非空(AllowEmpty 跳过)+ 自定义校验
func (m *multiInputModel) validate(i int) error {
	return m.validateWithValue(i, m.inputs[i].Value())
}

// moveTo 切换焦点字段:更新 textinput 焦点与光标命令
func (m *multiInputModel) moveTo(i int) tea.Cmd {
	m.inputs[m.focused].Blur()
	m.focused = i
	return m.inputs[i].Focus()
}

// ─── tea.Model ───────────────────────────────────────────────────────────────

// Init 聚焦首个字段并启动光标闪烁
func (m *multiInputModel) Init() tea.Cmd {
	return m.inputs[0].Focus()
}

// Update 处理键盘导航与输入
func (m *multiInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layoutInputs()
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, quitBinding):
			m.canceled = true
			return m, tea.Quit
		case key.Matches(msg, nextBinding): // enter / ↓ / j:校验后推进或提交
			if err := m.validate(m.focused); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.errMsg = ""
			if m.focused == len(m.specs)-1 {
				m.done = true
				return m, tea.Quit
			}
			return m, m.moveTo(m.focused + 1)
		case key.Matches(msg, prevBinding): // shift+tab / ↑ / k:回退
			if m.focused > 0 {
				m.errMsg = ""
				return m, m.moveTo(m.focused - 1)
			}
		default:
			m.errMsg = ""
			var cmd tea.Cmd
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
			return m, cmd
		}
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}
	return m, nil
}

// layoutInputs 计算输入框统一宽度(每行一致,desc 列对齐):
// 取全部字段内容宽(空时取占位符)的最大值 + 光标余量,上限为
// 终端宽 - 标题列 - 提示符 - 简介列 - 间隔(长内容在框内滚动)。
// 每次渲染前调用:任一行输入变长超过当前宽度时,全部行同步变宽,
// desc 列整体右移,始终保持对齐且紧贴输入框。
func (m *multiInputModel) layoutInputs() {
	w := 0
	for i := range m.inputs {
		ti := m.inputs[i]
		cw := lipgloss.Width(ti.Value())
		if cw == 0 {
			cw = lipgloss.Width(ti.Placeholder)
		}
		if cw > w {
			w = cw
		}
	}
	w += 2 // 光标与余量
	maxW := max(m.width-m.titleMaxW-3-m.descMaxW-2, 10)
	m.inputW = min(w, maxW)
}

// View 渲染:预览行 + 每行(标题列/输入框/行尾简介) + 分隔线 + 帮助栏。
// 简介用弱化色置于行尾,不混入标题、不占额外行;校验错误以红色替换
// 聚焦行的行尾内容(不增加行高,矮终端友好)。
func (m *multiInputModel) View() tea.View {
	var sb strings.Builder

	// 输入框统一宽度:全部字段一致(desc 列对齐),随最长内容同步变化
	m.layoutInputs()

	if m.preview != nil {
		sb.WriteString(mutedStyle.Render(m.preview(m.currentValues())))
		sb.WriteString("\n")
	}

	for i, spec := range m.specs {
		focused := i == m.focused
		titleStyle := blurredTitleStyle
		if focused {
			titleStyle = focusedTitleStyle
		}
		// 输入框宽度统一(见 layoutInputs):简介紧贴输入框且各 desc 列对齐
		m.inputs[i].SetWidth(m.inputW)
		sb.WriteString(titleStyle.Width(m.titleMaxW).Render(stepTitle(i, spec.Title)))
		sb.WriteString(m.inputs[i].View())

		// 行尾:校验错误(红)优先,否则弱化简介
		tail := ""
		if focused && m.errMsg != "" {
			tail = errorStyle.Render("✗ " + m.errMsg)
		} else if spec.Desc != "" {
			tail = mutedStyle.Render(spec.Desc)
		}
		if tail != "" {
			sb.WriteString(" " + tail)
		}
		if i < len(m.specs)-1 {
			sb.WriteString("\n")
		}
	}

	v := tea.NewView(sb.String())
	v.AltScreen = true
	if m.height >= HelpBarMinTermHeight {
		v.SetContent(v.Content + "\n" + theme.GetHorizontalRule(m.width))
		v = AppendHelpBar(v, multiInputHelpKeys(), m.width)
	}
	return v
}

// writeBack 将各输入框值写回外部指针(提交/accessible 完成时)
func (m *multiInputModel) writeBack() {
	for i, ti := range m.inputs {
		*m.values[i] = ti.Value()
	}
}

// run 运行表单:accessible 纯文本(TERM=dumb)或 tea TUI。
// 取消返回 huh.ErrUserAborted;提交成功写回 values 并返回 nil。
func (m *multiInputModel) run() error {
	if isAccessibleMode() {
		return runAccessibleMultiInput(os.Stdout, stdinBuf, m)
	}

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return huh.ErrUserAborted
		}
		return err
	}
	mm, ok := final.(*multiInputModel)
	if !ok {
		return huh.ErrUserAborted
	}
	if mm.canceled {
		return huh.ErrUserAborted
	}
	mm.writeBack()
	return nil
}

// ─── accessible 纯文本模式(TERM=dumb) ──────────────────────────────────────

// runAccessibleMultiInput 逐字段行输入:每字段打印「序号+标题(默认值)」,
// 读一行输入;空行 = 保留默认值,非空行经校验(失败重读并提示);
// EOF 视为取消。与列表/确认的管道输入共享 stdinBuf(lineReader 防预读吞行)。
func runAccessibleMultiInput(w io.Writer, r *bufio.Reader, m *multiInputModel) error {
	for i, spec := range m.specs {
		for {
			fmt.Fprintf(w, "%s (默认: %s)\n> ", stepTitle(i, spec.Title), *m.values[i])
			line, err := r.ReadString('\n')
			if err != nil && line == "" {
				// EOF 无输入:视为取消
				return huh.ErrUserAborted
			}
			v := strings.TrimSpace(line)
			if v == "" {
				// 空行 = 保留默认值
				m.inputs[i].SetValue(*m.values[i])
				break
			}
			if verr := m.validateWithValue(i, v); verr != nil {
				fmt.Fprintf(w, "%s\n", verr.Error())
				continue
			}
			m.inputs[i].SetValue(v)
			break
		}
	}
	m.writeBack()
	return nil
}

// validateWithValue 以给定值校验字段:非空(AllowEmpty 跳过)+ 自定义校验
func (m *multiInputModel) validateWithValue(i int, v string) error {
	spec := m.specs[i]
	if v == "" && !spec.AllowEmpty {
		return errors.New(i18n.T("form.input.empty.error"))
	}
	if spec.Validate != nil {
		return spec.Validate(v)
	}
	return nil
}

// ─── 键位与样式 ─────────────────────────────────────────────────────────────

var (
	quitBinding = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	// 上下导航:keymap 匹配先于 textinput 分发,方向键/j/k 不进入输入
	nextBinding = key.NewBinding(key.WithKeys("enter", "down", "j"))
	prevBinding = key.NewBinding(key.WithKeys("shift+tab", "up", "k"))
)

var (
	focusedTitleStyle = lipgloss.NewStyle().
				Foreground(theme.PrimaryColor).
				Bold(true)
	blurredTitleStyle = lipgloss.NewStyle().
				Foreground(theme.MutedForeground)
	mutedStyle = lipgloss.NewStyle().
			Foreground(theme.MutedForeground)
	errorStyle = lipgloss.NewStyle().
			Foreground(theme.ErrorColor).
			Bold(true)
)

// lipglossWidth 字符显示宽度(CJK 按 2 列计算)
