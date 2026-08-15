package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// TestInputWithSuggestionsHelpbar 带历史建议的输入表单帮助栏追加
// ↑/↓ 历史提示;无建议时不追加。
func TestInputWithSuggestionsHelpbar(t *testing.T) {
	var value string
	form := NewInputFormWithSuggestions("消息", &value, nil, []string{"fix: 旧消息", "feat: 新功能"})
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	view := ansi.Strip(m.(*Form).View().Content)

	if !strings.Contains(view, i18n.T("form.help.history")) {
		t.Fatalf("帮助栏应含历史提示:\n%s", view)
	}

	// 无建议时无历史提示
	var value2 string
	form2 := NewInputFormWithValidate("消息", &value2, nil)
	form2.Init()
	m2, _ := form2.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	view2 := ansi.Strip(m2.(*Form).View().Content)
	if strings.Contains(view2, i18n.T("form.help.history")) {
		t.Fatalf("无建议时帮助栏不应含历史提示:\n%s", view2)
	}
}
