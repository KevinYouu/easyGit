package form

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// 命令 × 单选/多选/表格表单渲染覆盖。
// 按各命令实际使用的表单类型与选项规模,在终端尺寸矩阵下断言:
// 内容不足一屏时按内容显示,超出一屏时占满终端滚动,且无越界行。

// commandSelectCases 各命令的单选表单。
// 标题使用命令实际的 i18n 键(经 i18n.T 取值,与命令渲染一致且与语言无关);
// 选项标签按命令实际格式构造(装饰/本地化),数据库与时间戳等动态内容为代表性取值。
var commandSelectCases = []struct {
	command string
	title   string
	labels  []string
}{
	{command: "set-language", title: i18n.T("language.select.title"), labels: []string{i18n.T("language.option.en"), i18n.T("language.option.zh")}},
	{command: "cherry-pick option", title: i18n.T("cherry.pick.select.option"), labels: []string{
		i18n.T("cherry.pick.option.default.name"),
		i18n.T("cherry.pick.option.no.commit.name"),
		i18n.T("cherry.pick.option.edit.name"),
		i18n.T("cherry.pick.option.signoff.name"),
	}},
	// 合并策略:实际格式 "%s - %s"(名称 - 描述)
	{command: "merge strategy", title: i18n.T("merge.select.strategy"), labels: []string{
		fmt.Sprintf("%s - %s", i18n.T("merge.strategy.default.name"), i18n.T("merge.strategy.default.description")),
		fmt.Sprintf("%s - %s", i18n.T("merge.strategy.ff.only.name"), i18n.T("merge.strategy.ff.only.description")),
		fmt.Sprintf("%s - %s", i18n.T("merge.strategy.no.ff.name"), i18n.T("merge.strategy.no.ff.description")),
		fmt.Sprintf("%s - %s", i18n.T("merge.strategy.squash.name"), i18n.T("merge.strategy.squash.description")),
	}},
	{command: "rebase action", title: i18n.T("rebase.status.in_progress"), labels: []string{
		i18n.T("rebase.action.continue"),
		i18n.T("rebase.action.skip"),
		i18n.T("rebase.action.abort"),
	}},
	// 提交类型来自配置数据库(用户数据),标签为代表性取值
	{command: "push commit type", title: i18n.T("push.select.commit.type"), labels: []string{"fix", "feat", "refactor", "build", "chore", "style", "docs", "revert", "test"}},
	// 分支选项无装饰前缀
	{command: "branch delete", title: i18n.T("branch.delete.select"), labels: []string{"main", "develop", "feature/login", "fix/typo", "release/v1.2.0"}},
	{command: "remote select", title: i18n.T("git.select.remote"), labels: []string{"origin", "upstream", "github"}},
	{command: "remote branch", title: i18n.T("git.select.branch"), labels: []string{"main", "develop", "feature/login", "release/v2.0"}},
	{command: "merge target", title: i18n.T("merge.select.target"), labels: []string{"main", "develop", "feature/login", "fix/typo", "release/v1.2.0"}},
	{command: "rebase target", title: i18n.T("rebase.select.target"), labels: []string{"main", "develop", "feature/login", "fix/typo", "release/v1.2.0"}},
	// 标签选项无装饰前缀(无创建时间时的回退格式;有时间时为 "%s (%s)")
	{command: "tag delete", title: i18n.T("tag.delete.select"), labels: []string{"v1.0.0", "v1.1.0", "v2.0.0-beta"}},
}

// commandMultiCases 各命令的多选表单
var commandMultiCases = []struct {
	command string
	title   string
	labels  []string
}{
	// cherry-pick 提交列表:单行 "[短hash] 日期 (作者) - 消息" 格式(hash 为代表性取值)
	{command: "cherry-pick commits", title: i18n.T("cherry.pick.select.commits"), labels: multiLabels(50, func(i int) string {
		return fmt.Sprintf("[a1b2c3d%02d] 07-12 10:00 (张三丰) - 修复登录问题", i)
	})},
	// push-selected 文件列表
	{command: "push-selected files", title: i18n.T("push.select.files"), labels: multiLabels(12, func(i int) string {
		return fmt.Sprintf("internal/form/file-%02d.go", i)
	})},
	// remote(SelectRemoteWithConfig)与 set-push-config 共用提示语 git.select.remotes.first,
	// 均为多选远端列表(set-push-config 额外带预选,渲染结构一致)
	{command: "remote remotes", title: i18n.T("git.select.remotes.first"), labels: []string{"origin", "upstream", "github", "gitlab"}},
}

// multiLabels 生成 n 个标签
func multiLabels(n int, f func(int) string) []string {
	labels := make([]string, n)
	for i := range n {
		labels[i] = f(i)
	}
	return labels
}

// assertCommandField 校验单选/多选表单渲染:总高 = min(内容, 终端),
// 可见项 = min(选项数, 终端-1),行宽不越界。
// 断言独立于 CalculateSelectHeight 本身(渲染层验证,不构成循环):
// 钉住的是 formFieldHeight 文档化的内容模型 —— 标题一行 + 每选项一行。
func assertCommandField(t *testing.T, view string, labels []string, termHeight, termWidth int) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	n := len(labels)
	// 大屏按内容显示,小屏占满终端滚动
	if total := lipgloss.Height(view); total != min(n+1, termHeight) {
		t.Errorf("渲染总高 = %d, want %d(选项 %d,终端 %d 行)", total, min(n+1, termHeight), n, termHeight)
	}
	got := visibleLabels(view, labels)
	if want := min(n, termHeight-1); len(got) != want {
		t.Fatalf("可见选项 %d 个, want %d(选项 %d,终端 %d 行)", len(got), want, n, termHeight)
	}
	for line := range strings.SplitSeq(view, "\n") {
		if lw := lipgloss.Width(ansi.Strip(line)); lw > termWidth {
			t.Errorf("行宽 %d 溢出终端 %d", lw, termWidth)
		}
	}
}

func TestCommandSelectRender(t *testing.T) {
	for _, tc := range commandSelectCases {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderSelectField(tc.title, tc.labels, h)
				assertCommandField(t, view, tc.labels, h, 80)
			})
		}
	}
}

func TestCommandMultiSelectRender(t *testing.T) {
	for _, tc := range commandMultiCases {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderMultiSelectField(tc.title, tc.labels, h)
				assertCommandField(t, view, tc.labels, h, 80)
			})
		}
	}
}

// 表格类命令:drop/squash 多选提交、reset 提交列表、reset 模式
func TestCommandTableRender(t *testing.T) {
	// GetRecentCommits 两行格式的提交选项(%h 短 hash 为 7 字符)
	tableOptions := func(n int) []config.Option {
		options := make([]config.Option, n)
		for i := range n {
			options[i] = config.Option{
				Label: fmt.Sprintf("a1b2c%02d 修复中文消息🚀以及附加说明\n07-12 10:00 • 张三丰", i),
				Value: "v",
			}
		}
		return options
	}

	t.Run("drop/squash 多选提交", func(t *testing.T) {
		options := tableOptions(20)
		// drop 与 squash 使用不同的 i18n 标题,逐一驱动生产构造器
		for _, title := range []string{i18n.T("rebase.select.drop_commits"), i18n.T("rebase.select.squash_commits")} {
			for _, sz := range tableSizes {
				t.Run(fmt.Sprintf("%s %dx%d", title, sz.w, sz.h), func(t *testing.T) {
					m := NewTableMultiSelectModel(title, options)
					m.width, m.height = sz.w, sz.h
					m.updateLayout()
					assertTableView(t, m.View().Content, sz.w, sz.h, options, true)
				})
			}
		}
	})

	t.Run("reset 提交列表", func(t *testing.T) {
		options := tableOptions(20)
		for _, sz := range tableSizes {
			m := NewTableSelectModel(options)
			m.width, m.height = sz.w, sz.h
			m.applyLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, false)
		}
	})

	t.Run("reset 模式仅3项", func(t *testing.T) {
		options := []config.Option{
			{Label: "--soft 保留工作区更改", Value: "--soft"},
			{Label: "--mixed 保留但取消暂存", Value: "--mixed"},
			{Label: "--hard 丢弃所有更改", Value: "--hard"},
		}
		for _, sz := range tableSizes {
			m := NewTableSelectModel(options)
			m.width, m.height = sz.w, sz.h
			m.applyLayout()
			assertTableView(t, m.View().Content, sz.w, sz.h, options, false)
		}
	})
}

// TestTableListLongMessageSingleLine 列表超长省略:
// 每条提交只占一行,消息超列宽时以省略号截断,完整长文本不泄漏到视图。
// 覆盖两种表格模型:单选(reset 提交列表)与多选(drop/squash)。
func TestTableListLongMessageSingleLine(t *testing.T) {
	// 超过 160 显示宽度(消息列上限),确保所有布局下都必须截断
	longMsg := strings.Repeat("这个提交消息非常长用于测试省略行为", 6) + "TAILMARKER987654"
	full := longMsg + "完整尾巴绝不应出现在列表中"
	options := []config.Option{
		{Label: fmt.Sprintf("a1b2c3d %s\n07-12 10:00 • 张三丰", full), Value: "v"},
	}

	// 单选无标题/帮助行,多选含标题+底部帮助行
	renderCases := []struct {
		name         string
		view         func(w, h int) string
		wantNonEmpty int
	}{
		{
			name: "单选",
			view: func(w, h int) string {
				m := NewTableSelectModel(options)
				m.width, m.height = w, h
				m.applyLayout()
				return ansi.Strip(m.View().Content)
			},
			wantNonEmpty: 1, // 仅一行提交
		},
		{
			name: "多选",
			view: func(w, h int) string {
				m := NewTableMultiSelectModel(i18n.T("rebase.select.drop_commits"), options)
				m.width, m.height = w, h
				m.updateLayout()
				return ansi.Strip(m.View().Content)
			},
			wantNonEmpty: 3, // 标题 + 提交行 + 帮助行
		},
	}

	for _, rc := range renderCases {
		for _, sz := range tableSizes {
			t.Run(fmt.Sprintf("%s %dx%d", rc.name, sz.w, sz.h), func(t *testing.T) {
				view := rc.view(sz.w, sz.h)

				// 非空行数固定:提交不得折行
				lines := 0
				for line := range strings.SplitSeq(view, "\n") {
					if strings.TrimSpace(line) != "" {
						lines++
					}
				}
				if lines != rc.wantNonEmpty {
					t.Errorf("非空行数 = %d, want %d(提交折行为多行)", lines, rc.wantNonEmpty)
				}

				// 省略必须来自仓库层 SafeTruncate 的 "..."(与 layout.go 的
				// ellipsis 常量耦合是有意的):若出现 bubbles 单元格兜底截断
				// 的 "…"(U+2026),说明仓库层截断失效、测试被上游兜底掩盖
				if !strings.Contains(view, "...") {
					t.Errorf("超长消息未省略")
				}
				if strings.Contains(view, "…") {
					t.Errorf("省略来自 bubbles 兜底(…),仓库层 SafeTruncate 未生效")
				}
				if strings.Contains(view, "完整尾巴") {
					t.Errorf("完整长文本泄漏到列表中")
				}
			})
		}
	}
}
