package form

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/x/ansi"
)

// 命令 × Input/Confirm 表单渲染覆盖。
// 两类表单均无 Height 设置,内容高度固定,任何终端高度下不填充不裁剪。

// renderInputField 构造与 Input 完全一致的 huh 输入表单
// (占位符/非空校验与 internal/form/input.go 保持一致)
func renderInputField(title, defaultValue string, termHeight int) string {
	value := defaultValue
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title(title).
			Placeholder(i18n.T("form.input.placeholder")).
			Value(&value).
			Validate(func(str string) error {
				if str == "" {
					return errors.New(i18n.T("form.input.empty.error"))
				}
				return nil
			}),
	)).WithTheme(theme.GetCompactTheme())
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

// renderConfirmField 构造与 Confirm 完全一致的 huh 确认表单
func renderConfirmField(title string, termHeight int) string {
	var confirmed bool
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(title).
			Value(&confirmed),
	)).WithTheme(theme.GetCompactTheme()).WithShowHelp(false)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*huh.Form).View()
}

// commandInputCases 各命令的输入表单(标题/默认值与命令一致)
var commandInputCases = []struct {
	command      string
	title        string
	defaultValue string
}{
	// push-all / push-selected:提交类型前缀作默认值
	{command: "push-all message", title: "输入提交信息：", defaultValue: "fix: "},
	{command: "push-selected message", title: "输入提交信息：", defaultValue: "feat: "},
	// squash:最早选中提交的消息作默认值
	{command: "squash message", title: "输入新的提交信息：", defaultValue: "修复登录问题"},
	// tag:版本号自动递增作默认值
	{command: "tag version", title: "输入版本号：", defaultValue: "v1.1.0"},
	// tag:提交消息无默认值,显示占位符
	{command: "tag message", title: "输入提交信息：", defaultValue: ""},
}

// commandConfirmCases 各命令的确认表单(动态消息按命令实际 fmt.Sprintf 拼装)
var commandConfirmCases = []struct {
	command string
	title   string
}{
	// branch delete:本地分支确认,动态嵌入分支名
	{command: "branch delete confirm", title: "您确定要删除分支 'feature/login' 吗？"},
	// branch delete:远端分支追加确认
	{command: "branch delete remote confirm", title: "您是否也想删除远程分支？"},
	// drop:动态嵌入提交计数
	{command: "drop confirm", title: "确定要删除这 5 个提交吗？"},
	// merge:存在未提交更改时继续合并
	{command: "merge continue confirm", title: "是否仍要继续合并？"},
	// reset:按 reset.go 实际 fmt.Sprintf 拼装(重置到 + hash + 短消息 + 模式 + 描述),
	// hard 模式追加警告行(警告图标 + 文案)
	{command: "reset confirm", title: fmt.Sprintf("%s %s  %s %s %s%s\n%s %s",
		i18n.T("reset.confirm.to"), "a1b2c3d", "修复登录问题", i18n.T("reset.confirm.mode"),
		"hard", i18n.T("reset.mode.hard.desc"),
		"⚠️", i18n.T("reset.hard.warning"))},
	// tag delete:两行动态消息(标签名 + 影响范围)
	{command: "tag delete confirm", title: "您确定要删除标签 'v1.0.0' 吗？\n这将从本地和远程仓库中删除标签。"},
}

// assertCompactField 校验 Input/Confirm 表单:内容高度固定,
// 任何终端高度下不填充不裁剪,底部无空白,行宽不越界。
func assertCompactField(t *testing.T, view, title string, termHeight, termWidth int) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	// 标题(可能多行动态消息)每一行都应出现在视图中
	for line := range strings.SplitSeq(title, "\n") {
		if !strings.Contains(ansi.Strip(view), line) {
			t.Errorf("标题行 %q 未出现在视图中", line)
		}
	}
	total := lipgloss.Height(strings.TrimRight(view, "\n"))
	if total == 0 || total > termHeight {
		t.Errorf("渲染总高 %d 越界(终端 %d 行)", total, termHeight)
	}
	// 底部无空白:最后一行必须含内容
	trimmed := strings.TrimRight(view, "\n")
	if last := trimmed[strings.LastIndex(trimmed, "\n")+1:]; strings.TrimSpace(ansi.Strip(last)) == "" {
		t.Errorf("底部存在空白行,总高 %d", total)
	}
	for line := range strings.SplitSeq(view, "\n") {
		if lw := lipgloss.Width(ansi.Strip(line)); lw > termWidth {
			t.Errorf("行宽 %d 溢出终端 %d", lw, termWidth)
		}
	}
}

func TestCommandInputRender(t *testing.T) {
	for _, tc := range commandInputCases {
		for _, h := range []int{24, 12, 10, 8, 6, 4} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderInputField(tc.title, tc.defaultValue, h)
				assertCompactField(t, view, tc.title, h, 80)
			})
		}
	}
}

func TestCommandConfirmRender(t *testing.T) {
	for _, tc := range commandConfirmCases {
		for _, h := range []int{24, 12, 10, 8, 6, 4} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderConfirmField(tc.title, h)
				assertCompactField(t, view, tc.title, h, 80)
			})
		}
	}
}
