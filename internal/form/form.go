package form

import (
	"errors"
	"fmt"
	"os"
	"strings"

	key "charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// Form 是 huh.Form 的薄包装:视图末尾附加统一帮助栏。
// huh v2 的 ViewHook 在内容设置前执行且 Form.View 不调用它,
// 无法用于追加内容,因此由包装模型在生产(tea.NewProgram)与渲染测试
// 共用同一构造路径,防止配置漂移。
type Form struct {
	*huh.Form
	helpKeys          []HelpKey
	dividerAfterTitle bool // 标题下方渲染全宽分隔线(列表类表单,见 newForm 子类设置)
	width             int  // 最近一次窗口宽度
	height            int  // 最近一次窗口高度
}

// newForm 包装 huh 表单并绑定帮助栏键位。
// Esc 取消:huh 默认 Quit 仅绑 ctrl+c,Esc 只挂在已禁用的过滤键上,
// 并入 Quit 使帮助栏「Esc 取消」提示与实际行为一致。
// 筛选键必须在此禁用且先于 WithKeyMap:字段键位按值拷贝(Select/MultiSelect
// 的 WithKeyMap 各取一份),事后修改不传播;而 Esc 在表单级先于字段分发匹配
// Quit,若 "/" 可进入筛选模式,筛选中的 Esc 会误终止表单丢弃已选内容。
func newForm(f *huh.Form, keys []HelpKey) *Form {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	km.Select.Filter = key.NewBinding(key.WithKeys("/"), key.WithDisabled())
	km.MultiSelect.Filter = key.NewBinding(key.WithKeys("/"), key.WithDisabled())
	f.WithKeyMap(km)
	// accessible 模式统一输入源:列表表单经共享 stdinBuf 读取,
	// Input/Confirm/MultiInput 若不指向同一缓冲,管道输入会被列表的
	// 预读缓冲吞掉导致后续表单读不到数据(TERM=dumb 脚本管线回归)。
	// 包装为 lineReader 防止 huh 内部 Scanner 预读吞掉多字段的后续输入。
	if isAccessibleMode() {
		f.WithInput(&lineReader{br: stdinBuf})
	}
	return &Form{
		Form:     f,
		helpKeys: keys,
		width:    defaultTermWidth,
		height:   defaultTermHeight,
	}
}

// Init 实现 tea.Model
func (f *Form) Init() tea.Cmd { return f.Form.Init() }

// Update 实现 tea.Model:透传消息并跟踪窗口尺寸。
// huh 内部经 compat.ViewModel 调用,更新后仍返回 *huh.Form(单页表单无模型替换),
// 此处以新指针重建包装;其余情况保持包装身份,避免兼容层类型无法满足 tea.Model。
func (f *Form) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		f.width, f.height = wsm.Width, wsm.Height
	}
	m, cmd := f.Form.Update(msg)
	if mf, ok := m.(*huh.Form); ok {
		return &Form{Form: mf, helpKeys: f.helpKeys, dividerAfterTitle: f.dividerAfterTitle, width: f.width, height: f.height}, cmd
	}
	return f, cmd
}

// View 实现 tea.Model:列表类表单在标题下方渲染分隔线,视图末尾先渲染
// 底部分隔线再附加帮助栏;极小终端(<6 行)不渲染,退出中的表单不再渲染
// (与 huh 原生行为一致)。
func (f *Form) View() tea.View {
	content := f.Form.View()
	if content == "" {
		return tea.NewView("")
	}
	// 标题下方分隔线:短标题不折行,首行即标题行;多行标题(动态消息)
	// 只在首行后插线,与列表类表单(固定标题)行为一致
	if f.dividerAfterTitle && f.height >= HelpBarMinTermHeight {
		if head, tail, found := strings.Cut(content, "\n"); found {
			content = head + "\n" + theme.GetHorizontalRule(f.width) + "\n" + tail
		} else {
			content = content + "\n" + theme.GetHorizontalRule(f.width)
		}
	}
	v := tea.NewView(content)
	if f.height >= HelpBarMinTermHeight {
		v.SetContent(v.Content + "\n" + theme.GetHorizontalRule(f.width))
		v = AppendHelpBar(v, f.helpKeys, f.width)
	}
	return v
}

// isAccessibleMode 判断 huh 是否处于 accessible 纯文本模式。
// 与 huh v2.0.3 构造期判定耦合(huh/form.go NewForm:TERM=dumb 时
// WithAccessible(true)):该模式直接打印 option.Key 且不再包外层样式,
// 无 reset 的内嵌序列会泄漏到输出(selectOptionLabel 据此剥离样式);
// 包装模型的 Run 也据此委托 huh 原生 Run,避免无 TTY 时 bubbletea OpenTTY 失败。
// 若 huh 放宽 accessible 判定(如空 TERM / NO_COLOR),需同步本函数。
func isAccessibleMode() bool {
	return os.Getenv("TERM") == "dumb"
}

// Run 以本包装模型运行表单。Esc 取消返回 huh.ErrUserAborted,语义与 huh.Form.Run 一致。
// accessible 模式(纯文本行提示,不依赖 TTY)下直接委托 huh 原生 Run,
// 避免 bubbletea 在无 TTY 的 stdin 上 OpenTTY 失败。
func (f *Form) Run() error {
	f.SubmitCmd = tea.Quit
	f.CancelCmd = tea.Interrupt
	if isAccessibleMode() {
		return f.Form.Run()
	}
	// 与 huh 默认一致输出到 stderr:stdout 重定向管道时 TUI 画面不污染输出
	if _, err := tea.NewProgram(f, tea.WithOutput(os.Stderr)).Run(); err != nil {
		if errors.Is(err, tea.ErrInterrupted) {
			return huh.ErrUserAborted
		}
		return fmt.Errorf("huh: %w", err)
	}
	return nil
}
