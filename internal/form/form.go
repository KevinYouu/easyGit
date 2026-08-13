package form

import (
	"errors"
	"fmt"
	"os"

	key "charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// Form 是 huh.Form 的薄包装:视图末尾附加统一帮助栏。
// huh v2 的 ViewHook 在内容设置前执行且 Form.View 不调用它,
// 无法用于追加内容,因此由包装模型在生产(tea.NewProgram)与渲染测试
// 共用同一构造路径,防止配置漂移。
type Form struct {
	*huh.Form
	helpKeys []HelpKey
	width    int // 最近一次窗口宽度
	height   int // 最近一次窗口高度
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
		return &Form{Form: mf, helpKeys: f.helpKeys, width: f.width, height: f.height}, cmd
	}
	return f, cmd
}

// View 实现 tea.Model:表单视图末尾附加帮助栏;极小终端(<6 行)不渲染,
// 退出中的表单不再渲染(与 huh 原生行为一致)。
func (f *Form) View() tea.View {
	content := f.Form.View()
	if content == "" {
		return tea.NewView("")
	}
	v := tea.NewView(content)
	if f.height >= HelpBarMinTermHeight {
		v = AppendHelpBar(v, f.helpKeys, f.width)
	}
	return v
}

// Run 以本包装模型运行表单。Esc 取消返回 huh.ErrUserAborted,语义与 huh.Form.Run 一致。
// TERM=dumb 时 huh 构造期已自动启用 accessible 模式(纯文本行提示,不依赖 TTY),
// 直接委托 huh 原生 Run,避免 bubbletea 在无 TTY 的 stdin 上 OpenTTY 失败。
func (f *Form) Run() error {
	f.SubmitCmd = tea.Quit
	f.CancelCmd = tea.Interrupt
	if os.Getenv("TERM") == "dumb" {
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
