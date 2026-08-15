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
	{command: "set-language", title: i18n.T("language.select.title"), labels: []string{
		OptionLabel(i18n.T("language.option.en"), i18n.T("language.option.en.desc")),
		OptionLabel(i18n.T("language.option.zh"), i18n.T("language.option.zh.desc")),
	}},
	// reset 模式:列表式单选 4 项,单行「名称 + 说明」(default 值为空,不传参数)
	{command: "reset mode", title: i18n.T("reset.select.mode"), labels: []string{
		OptionLabel(i18n.T("reset.option.default.name"), i18n.T("reset.option.default.desc")),
		OptionLabel(i18n.T("reset.option.soft.name"), i18n.T("reset.option.soft.desc")),
		OptionLabel(i18n.T("reset.option.mixed.name"), i18n.T("reset.option.mixed.desc")),
		OptionLabel(i18n.T("reset.option.hard.name"), i18n.T("reset.option.hard.desc")),
	}},
	{command: "cherry-pick option", title: i18n.T("cherry.pick.select.option"), labels: []string{
		OptionLabel(i18n.T("cherry.pick.option.default.name"), i18n.T("cherry.pick.option.default.description")),
		OptionLabel(i18n.T("cherry.pick.option.no.commit.name"), i18n.T("cherry.pick.option.no.commit.description")),
		OptionLabel(i18n.T("cherry.pick.option.edit.name"), i18n.T("cherry.pick.option.edit.description")),
		OptionLabel(i18n.T("cherry.pick.option.signoff.name"), i18n.T("cherry.pick.option.signoff.description")),
	}},
	// 合并策略:统一表格按 hash/消息分列渲染,名称入 hash 列(8 宽)、说明入消息列,
	// 标签取代表性取值(超长名称/说明会被截断,不参与整行匹配)
	{command: "merge strategy", title: i18n.T("merge.select.strategy"), labels: []string{
		OptionLabel("default", "默认合并行为"),
		OptionLabel("ff-only", "仅在可以快进合并时进行合并"),
		OptionLabel("no-ff", "始终创建合并提交"),
		OptionLabel("squash", "将所有提交压缩为单个提交"),
	}},
	{command: "rebase action", title: i18n.T("rebase.status.in_progress"), labels: []string{
		OptionLabel("continue", "继续执行变基"),
		OptionLabel("skip", "跳过当前提交"),
		OptionLabel("abort", "取消本次变基"),
	}},
	// 提交类型来自配置数据库(用户数据),标签为代表性取值;按 commit.type.desc.<value> 查 i18n 附加说明
	{command: "push commit type", title: i18n.T("push.select.commit.type"), labels: []string{
		OptionLabel("fix", i18n.T("commit.type.desc.fix")),
		OptionLabel("feat", i18n.T("commit.type.desc.feat")),
		OptionLabel("refactor", i18n.T("commit.type.desc.refactor")),
		OptionLabel("build", i18n.T("commit.type.desc.build")),
		OptionLabel("chore", i18n.T("commit.type.desc.chore")),
		OptionLabel("style", i18n.T("commit.type.desc.style")),
		OptionLabel("docs", i18n.T("commit.type.desc.docs")),
		OptionLabel("revert", i18n.T("commit.type.desc.revert")),
		OptionLabel("test", i18n.T("commit.type.desc.test")),
	}},
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
	// cherry-pick 提交列表:统一表格按 hash/消息/日期/作者分列渲染,
	// 标签为提交式两行格式(与 GetRecentCommits 一致),hash 7 位,
	// 消息为代表性取值(日期作者落独立列,不参与整行匹配)
	{command: "cherry-pick commits", title: i18n.T("cherry.pick.select.commits"), labels: multiLabels(50, func(i int) string {
		return fmt.Sprintf("a1b2c%02d 修复登录问题", i)
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

// assertCommandField 校验单选/多选列表渲染:总高 = min(内容+附加行, 终端),
// 可见项 = 表格视口行数(计算与渲染同源),行宽不越界。
// 单选/多选结构相同:标题 + 顶部线 + 表格行 + 底部线 + 帮助栏 = n+4;
// 仅单选在 <6 行终端不渲染附加行。
// 断言独立于 CalculateTableHeight 本身(渲染层验证,不构成循环):
// 钉住的是统一列表组件的内容模型。
func assertCommandField(t *testing.T, view string, labels []string, termHeight, termWidth int, multi bool) {
	t.Helper()
	if !utf8.ValidString(view) {
		t.Fatal("渲染结果含非法 UTF-8")
	}
	n := len(labels)
	extra := 4 // 标题 + 顶部线 + 底部线 + 帮助栏
	if !multi && termHeight < HelpBarMinTermHeight {
		extra = 0 // 极小终端单选:仅表格行
	}
	wantTotal := min(n+extra, termHeight)
	wantVisible := min(CalculateTableHeight(termHeight), n)
	if wantVisible+extra > termHeight {
		// 极小终端(6 行)表格高度触底 3 行,附加行溢出 1 行,如实断言
		wantTotal = wantVisible + extra
	}
	if total := lipgloss.Height(view); total != wantTotal {
		t.Errorf("渲染总高 = %d, want %d(选项 %d,终端 %d 行)", total, wantTotal, n, termHeight)
	}
	got := visibleLabels(view, labels)
	if len(got) != wantVisible {
		t.Fatalf("可见选项 %d 个, want %d(选项 %d,终端 %d 行)", len(got), wantVisible, n, termHeight)
	}
	for line := range strings.SplitSeq(view, "\n") {
		if lw := lipgloss.Width(ansi.Strip(line)); lw > termWidth {
			t.Errorf("行宽 %d 溢出终端 %d", lw, termWidth)
		}
	}
}

func TestCommandSelectRender(t *testing.T) {
	// huh 在 TERM=dumb 下强制固定宽度 80 并忽略 WindowSizeMsg(无障碍输出),
	// 测试固定为普通终端语义,窄屏回归才可验证窗口宽度传导。
	t.Setenv("TERM", "xterm-256color")
	for _, tc := range commandSelectCases {
		// 窄屏(60 列)回归:单行说明不得折行破坏高度模型
		for _, w := range []int{80, 60} {
			for _, h := range []int{24, 12, 10, 8, 6} {
				t.Run(fmt.Sprintf("%s@%d行x%d列", tc.command, h, w), func(t *testing.T) {
					view := renderSelectFieldWidth(tc.title, tc.labels, h, w)
					assertCommandField(t, view, tc.labels, h, w, false)
				})
			}
		}
	}
}

func TestCommandMultiSelectRender(t *testing.T) {
	for _, tc := range commandMultiCases {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("%s@%d行", tc.command, h), func(t *testing.T) {
				view := renderMultiSelectField(tc.title, tc.labels, h)
				assertCommandField(t, view, tc.labels, h, 80, true)
			})
		}
	}
}

// TestCommandColumnsRender 自适应多列布局渲染:列数由 ColumnSpec 声明,
// 名称/单元格列自动宽度不截断(仅超上限截断),弹性列占满剩余宽度。
// 配置中心主列表(2 列:名称 + 摘要)与提交类型删除列表(多选)走此路径。
func TestCommandColumnsRender(t *testing.T) {
	// 配置中心列定义与生产一致:名称列 Auto(上限 24),摘要列 Flex
	specs := []ColumnSpec{
		{Kind: ColumnAuto, MaxWidth: 24},
		{Kind: ColumnFlex},
	}

	// 配置中心主列表(单选):名称 + 当前值摘要;整行 = 两列内容拼接
	configOptions := []config.Option{
		{Label: "Interface Language", Description: i18n.T("config.summary.language.en"), Value: "language"},
		{Label: "Push Config", Description: i18n.T("config.summary.push.not.set"), Value: "push"},
		{Label: "Commit Types", Description: fmt.Sprintf(i18n.T("config.summary.commit.types"), "fix, feat, …"), Value: "commit-types"},
		{Label: "Tag Version Cap", Description: fmt.Sprintf(i18n.T("config.summary.tag.patch"), "999.9.9"), Value: "tag-patch"},
	}
	configLabels := make([]string, len(configOptions))
	for i, o := range configOptions {
		configLabels[i] = o.Label + " " + o.Description
	}

	for _, w := range []int{80, 60} {
		for _, h := range []int{24, 12, 10, 8, 6} {
			t.Run(fmt.Sprintf("config center@%d行x%d列", h, w), func(t *testing.T) {
				view := renderColumnsField(i18n.T("config.select.title"), specs, configOptions, h, w)
				assertCommandField(t, view, configLabels, h, w, false)
				// 名称列自动宽度:名称完整显示不截断(6 行终端表格仅 3 行可视,
				// 第 4 项不在视口内,名称断言仅在全量可见时进行)
				if h >= 8 {
					for _, name := range []string{"Interface Language", "Tag Version Cap"} {
						if !strings.Contains(view, name) {
							t.Errorf("名称 %q 被截断或缺失:\n%s", name, view)
						}
					}
				}
			})
		}
	}

	// 三列场景(如未来「名称 | 使用次数 | 摘要」):Auto + Auto + Flex
	threeColSpecs := []ColumnSpec{
		{Kind: ColumnAuto, MaxWidth: 24},
		{Kind: ColumnAuto, MaxWidth: 8},
		{Kind: ColumnFlex},
	}
	threeColOptions := []config.Option{
		{Label: "fix", Cells: []string{"fix", "12", "Bug fixes"}, Value: "fix"},
		{Label: "feat", Cells: []string{"feat", "3", "New features"}, Value: "feat"},
	}
	threeColLabels := []string{"fix 12 Bug fixes", "feat 3 New features"}
	for _, w := range []int{80, 60} {
		t.Run(fmt.Sprintf("three columns@%d列", w), func(t *testing.T) {
			view := renderColumnsField("Commit Type Stats", threeColSpecs, threeColOptions, 12, w)
			assertCommandField(t, view, threeColLabels, 12, w, false)
		})
	}
}

// TestCommandColumnsMultiRender 多选 + 自适应多列(提交类型删除列表)
func TestCommandColumnsMultiRender(t *testing.T) {
	specs := []ColumnSpec{
		{Kind: ColumnAuto, MaxWidth: 24},
		{Kind: ColumnFlex},
	}
	options := []config.Option{
		{Label: "fix", Description: fmt.Sprintf(i18n.T("config.commit.types.usage"), 12), Value: "fix"},
		{Label: "feat", Description: fmt.Sprintf(i18n.T("config.commit.types.usage"), 3), Value: "feat"},
		{Label: "refactor", Description: fmt.Sprintf(i18n.T("config.commit.types.usage"), 0), Value: "refactor"},
	}
	labels := make([]string, len(options))
	for i, o := range options {
		labels[i] = o.Label + " " + o.Description
	}
	for _, h := range []int{24, 12, 10, 8, 6} {
		t.Run(fmt.Sprintf("delete commit types@%d行", h), func(t *testing.T) {
			view := renderColumnsMultiField(i18n.T("config.commit.types.delete.select"), specs, options, h, 80)
			assertCommandField(t, view, labels, h, 80, true)
		})
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
					m := newListModel(title, options, ListMulti, false, nil)
					m.width, m.height = sz.w, sz.h
					m.applyLayout()
					assertTableView(t, m.View().Content, sz.w, sz.h, options, true)
				})
			}
		}
	})

	t.Run("reset 提交列表", func(t *testing.T) {
		options := tableOptions(20)
		for _, sz := range tableSizes {
			m := newListModel(i18n.T("reset.select.commit"), options, ListSingle, false, nil)
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
	// 超过 300 宽终端消息列宽(无上限后 300 宽为 259 列),确保所有布局下都必须截断
	longMsg := strings.Repeat("这个提交消息非常长用于测试省略行为", 10) + "TAILMARKER987654"
	full := longMsg + "完整尾巴绝不应出现在列表中"
	options := []config.Option{
		{Label: fmt.Sprintf("a1b2c3d %s\n07-12 10:00 • 张三丰", full), Value: "v"},
	}

	// 单选/多选结构相同:标题 + 顶部线 + 提交行 + 底部线 + 帮助行
	renderCases := []struct {
		name         string
		view         func(w, h int) string
		wantNonEmpty int
	}{
		{
			name: "单选",
			view: func(w, h int) string {
				m := newListModel(i18n.T("reset.select.commit"), options, ListSingle, false, nil)
				m.width, m.height = w, h
				m.applyLayout()
				return ansi.Strip(m.View().Content)
			},
			wantNonEmpty: 5, // 标题 + 顶部线 + 提交行 + 底部线 + 帮助行
		},
		{
			name: "多选",
			view: func(w, h int) string {
				m := newListModel(i18n.T("rebase.select.drop_commits"), options, ListMulti, false, nil)
				m.width, m.height = w, h
				m.applyLayout()
				return ansi.Strip(m.View().Content)
			},
			wantNonEmpty: 5, // 标题 + 顶部线 + 提交行 + 底部线 + 帮助行
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

				// 省略必须来自列表组件层 SafeTruncate(parseCommitInfo,与 layout.go 的
				// ellipsis 常量耦合是有意的):若出现 bubbles 单元格兜底截断的 "…"
				// (U+2026),说明组件截断失效、测试被上游兜底掩盖
				if !strings.Contains(view, "...") {
					t.Errorf("超长消息未省略")
				}
				if strings.Contains(view, "…") {
					t.Errorf("省略来自 bubbles 兜底(…),组件层 SafeTruncate 未生效")
				}
				if strings.Contains(view, "完整尾巴") {
					t.Errorf("完整长文本泄漏到列表中")
				}
			})
		}
	}
}

// TestTableListLongMessageFullDisplay 宽屏完整显示:
// 终端宽度足够时消息完整显示(截断只发生在实际宽度不足时),
// 300 宽终端消息列 259 列 > 208 列长消息,完整文本出现、无省略号、不折行
func TestTableListLongMessageFullDisplay(t *testing.T) {
	longMsg := strings.Repeat("这个提交消息非常长用于测试省略行为", 6) + "TAILMARKER987654"
	full := longMsg + "完整尾巴应完整显示"
	options := []config.Option{
		{Label: fmt.Sprintf("a1b2c3d %s\n07-12 10:00 • 张三丰", full), Value: "v"},
	}

	renderCases := []struct {
		name string
		view func(w, h int) string
	}{
		{
			name: "单选",
			view: func(w, h int) string {
				m := newListModel(i18n.T("reset.select.commit"), options, ListSingle, false, nil)
				m.width, m.height = w, h
				m.applyLayout()
				return ansi.Strip(m.View().Content)
			},
		},
		{
			name: "多选",
			view: func(w, h int) string {
				m := newListModel(i18n.T("rebase.select.drop_commits"), options, ListMulti, false, nil)
				m.width, m.height = w, h
				m.applyLayout()
				return ansi.Strip(m.View().Content)
			},
		},
	}

	for _, rc := range renderCases {
		for _, sz := range []struct{ w, h int }{{300, 60}} {
			t.Run(fmt.Sprintf("%s %dx%d", rc.name, sz.w, sz.h), func(t *testing.T) {
				view := rc.view(sz.w, sz.h)

				// 非空行数固定:提交不得折行
				lines := 0
				for line := range strings.SplitSeq(view, "\n") {
					if strings.TrimSpace(line) != "" {
						lines++
					}
				}
				if lines != 5 {
					t.Errorf("非空行数 = %d, want 5(提交折行为多行)", lines)
				}

				// 完整文本出现,无任何形式的省略
				if !strings.Contains(view, full) {
					t.Errorf("宽屏下长消息未完整显示")
				}
				if strings.Contains(view, "...") || strings.Contains(view, "…") {
					t.Errorf("宽屏下长消息仍被省略")
				}
			})
		}
	}
}
