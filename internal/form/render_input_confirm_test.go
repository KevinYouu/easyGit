package form

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// 命令 × Input/Confirm 表单渲染覆盖。
// 经生产构造器 NewInputForm/NewConfirmForm 驱动渲染,与命令实际执行的
// 构造路径一致;两类表单均无 Height 设置,内容高度固定,任何终端高度下不填充不裁剪。

// renderInputField 经生产构造器 NewInputForm 渲染输入表单
func renderInputField(title, defaultValue string, termHeight int) string {
	value := defaultValue
	form := NewInputForm(title, &value)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*Form).View().Content
}

// renderConfirmField 经生产构造器 NewConfirmForm 渲染确认表单
func renderConfirmField(title string, termHeight int) string {
	var confirmed bool
	form := NewConfirmForm(title, &confirmed)
	form.Init()
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: termHeight})
	return m.(*Form).View().Content
}

// commandInputCases 各命令的输入表单(标题取命令实际 i18n 键;默认值为代表性数据)
var commandInputCases = []struct {
	command      string
	title        string
	defaultValue string
}{
	// push-all / push-selected:提交类型前缀作默认值
	{command: "push-all message", title: i18n.T("push.input.commit.message"), defaultValue: "fix: "},
	{command: "push-selected message", title: i18n.T("push.input.commit.message"), defaultValue: "feat: "},
	// squash:最早选中提交的消息作默认值(用户数据,代表性取值)
	{command: "squash message", title: i18n.T("squash.input.message"), defaultValue: "修复登录问题"},
	// tag:版本号自动递增作默认值
	{command: "tag version", title: i18n.T("tag.input.version"), defaultValue: "v1.1.0"},
	// tag:提交消息无默认值,显示占位符
	{command: "tag message", title: i18n.T("tag.input.commit.message"), defaultValue: ""},
}

// commandConfirmCases 各命令的确认表单(标题按命令实际 i18n 键与
// fmt.Sprintf 格式拼装,与命令渲染完全一致)
var commandConfirmCases = []struct {
	command string
	title   string
}{
	// branch delete:本地分支确认,动态嵌入分支名(代表性取值)
	{command: "branch delete confirm", title: fmt.Sprintf(i18n.T("branch.delete.confirm"), "feature/login")},
	// branch delete:远端分支追加确认
	{command: "branch delete remote confirm", title: i18n.T("branch.delete.remote.confirm")},
	// drop:动态嵌入提交计数
	{command: "drop confirm", title: fmt.Sprintf(i18n.T("rebase.drop.confirm"), 5)},
	// merge:存在未提交更改时继续合并
	{command: "merge continue confirm", title: i18n.T("merge.confirm.continue.with.changes")},
	// reset:按 reset.go 实际 fmt.Sprintf 拼装(重置到 + hash + 短消息 + 模式 + 描述),
	// hard 模式追加警告行(警告图标 + 文案)
	{command: "reset confirm", title: fmt.Sprintf("%s %s  %s %s %s%s\n%s %s",
		i18n.T("reset.confirm.to"), "a1b2c3d", "修复登录问题", i18n.T("reset.confirm.mode"),
		"hard", i18n.T("reset.option.hard.desc"),
		"⚠", i18n.T("reset.hard.warning"))},
	// tag delete:两行动态消息(标签名 + 影响范围)
	{command: "tag delete confirm", title: fmt.Sprintf(i18n.T("tag.delete.confirm"), "v1.0.0")},
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
