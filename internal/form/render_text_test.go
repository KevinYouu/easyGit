package form

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// renderTextInputField 经生产构造器 NewTextInputForm 渲染多行文本表单:
// 与 bubbletea 运行时一致地执行 Init 消息(含 ctrl+home 定位键),再按终端尺寸渲染。
func renderTextInputField(title, defaultValue string, termHeight int) string {
	value := defaultValue
	form := NewTextInputForm(title, &value)
	form = dispatchInitCmd(form)
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*Form).View().Content
}

// dispatchInitCmd 执行表单 Init 返回的命令树并把产生的消息分发回表单,
// 模拟 tea 运行时(Init → 消息 → Update);批量命令逐条执行。
func dispatchInitCmd(form *Form) *Form {
	return dispatchCmd(form, form.Init())
}

func dispatchCmd(form *Form, cmd tea.Cmd) *Form {
	// 深度上限:光标闪烁等周期命令会自我续期,不能无限展开
	return dispatchCmdDepth(form, cmd, 32)
}

// execCmd 执行命令并限时收消息:cursor.Blink 会阻塞到 530ms 的闪烁节拍才返回,
// 在测试里直接调用会让每个用例白等半秒,超时即视为无关命令丢弃。
func execCmd(cmd tea.Cmd) tea.Msg {
	result := make(chan tea.Msg, 1)
	go func() { result <- cmd() }()
	select {
	case msg := <-result:
		return msg
	case <-time.After(20 * time.Millisecond):
		return nil
	}
}

func dispatchCmdDepth(form *Form, cmd tea.Cmd, budget int) *Form {
	if cmd == nil || budget <= 0 {
		return form
	}
	msg := execCmd(cmd)
	if msg == nil {
		return form
	}
	// 光标闪烁等周期命令会延时后自我续期:执行会拖慢测试且与断言无关,跳过
	if _, ok := msg.(cursor.BlinkMsg); ok {
		return form
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, sub := range batch {
			form = dispatchCmdDepth(form, sub, budget-1)
		}
		return form
	}
	// huh 的字段推进命令可能再次返回命令(nextFieldMsg → nextGroupMsg → …)
	if sub, ok := msg.(tea.Cmd); ok {
		return dispatchCmdDepth(form, sub, budget-1)
	}
	m, next := form.Update(msg)
	return dispatchCmdDepth(m.(*Form), next, budget-1)
}

// TestSquashTextInputRender 合并消息表单(squash 实际构造路径):
// 默认值中的多行正文必须逐行渲染、不被折叠;视图合法、
// 高度不越界、行宽不溢出、标题不被挤掉。
func TestSquashTextInputRender(t *testing.T) {
	title := i18n.T("squash.input.message")
	defaultMessage := "fix: first target\n\nfirst body line 1\nfirst body line 2"

	for _, h := range []int{24, 16, 12, 10} {
		t.Run(fmt.Sprintf("%d行", h), func(t *testing.T) {
			view := renderTextInputField(title, defaultMessage, h)
			plain := ansi.Strip(view)

			if !utf8.ValidString(view) {
				t.Fatal("渲染结果含非法 UTF-8")
			}
			if !strings.Contains(plain, title) {
				t.Errorf("标题 %q 未出现在视图中", title)
			}
			// 多行正文逐行渲染:换行未被折叠成一行,也未丢失
			for line := range strings.SplitSeq(defaultMessage, "\n") {
				if line == "" {
					continue
				}
				if !strings.Contains(plain, line) {
					t.Errorf("正文行 %q 未出现在视图中", line)
				}
			}
			if total := lipgloss.Height(strings.TrimRight(view, "\n")); total > h {
				t.Errorf("渲染总高 %d 越界(终端 %d 行)", total, h)
			}
			for line := range strings.SplitSeq(view, "\n") {
				if lw := lipgloss.Width(ansi.Strip(line)); lw > 80 {
					t.Errorf("行宽 %d 溢出终端 80", lw)
				}
			}
		})
	}
}

// TestSquashTextInputStartsAtFirstLine 长消息必须从第一行开始显示
// (huh 预填后光标在末尾会让视图停在最后几行)。
func TestSquashTextInputStartsAtFirstLine(t *testing.T) {
	title := i18n.T("squash.input.message")
	lines := make([]string, 0, 18)
	for i := 1; i <= 18; i++ {
		lines = append(lines, fmt.Sprintf("line %02d", i))
	}
	message := strings.Join(lines, "\n")

	view := ansi.Strip(renderTextInputField(title, message, 24))
	if !strings.Contains(view, "line 01") {
		t.Fatalf("首行未显示(视图停在了末尾):\n%s", view)
	}
	// 首行必须出现在其余可见行之前
	if first, last := strings.Index(view, "line 01"), strings.LastIndex(view, "line "); first > last {
		t.Errorf("首行位置异常:first=%d last=%d", first, last)
	}
	if strings.Index(view, "line 01") > strings.Index(view, "line 05") {
		t.Errorf("视图未从第一行开始渲染:\n%s", view)
	}
	if strings.Contains(view, "line 18") {
		t.Errorf("超出可视高度的末行不应显示(说明视图被顶到末尾):\n%s", view)
	}
}

// TestTextInputHelpBar 多行文本帮助栏:Enter 提交、ctrl+j 换行、Esc 取消
func TestTextInputHelpBar(t *testing.T) {
	bar := RenderHelpBar(textInputHelpKeys(), 100)
	for _, want := range []string{"ctrl+j", i18n.T("form.help.newline"), i18n.T("form.help.submit"), i18n.T("form.help.cancel")} {
		if !strings.Contains(bar, want) {
			t.Errorf("帮助栏缺少 %q: %q", want, bar)
		}
	}
}

// TestTextInputKeyMapDisablesExternalEditor ctrl+e 必须禁用(会阻塞 TUI)
func TestTextInputKeyMapDisablesExternalEditor(t *testing.T) {
	km := huh.NewDefaultKeyMap()
	if !km.Text.Editor.Enabled() {
		t.Fatal("前提不成立:huh 默认应启用 ctrl+e 外部编辑器")
	}

	applyTextInputKeyMap(km)

	if km.Text.Editor.Enabled() {
		t.Error("ctrl+e 外部编辑器仍启用,阻塞 TUI 的风险未消除")
	}
}
