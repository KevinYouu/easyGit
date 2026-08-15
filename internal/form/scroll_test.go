package form

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/charmbracelet/x/ansi"
)

// cursorViewPos 返回光标在当前视图中的行位置(0 起),-1 表示不可见
func cursorViewPos(m *listModel) int {
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "❯") {
			return i
		}
	}
	return -1
}

// TestFindJump 暴力遍历:各种列表长度 × 视口高度 × 初始位置 × 全部按键序列,
// 断言任何单次按键(光标步进 1)都不会让光标在视图中的位置跳变超过 1 行
// (bubbles/table 视口偏移在区域边界会跳 2 行,回归守护)。
func TestFindJump(t *testing.T) {
	for _, rows := range []int{9, 12, 20} {
		for _, vp := range []int{3, 4, 5, 6, 7, 8} {
			if vp >= rows {
				continue
			}
			for _, init := range []int{0, rows - 1, rows / 2} {
				options := make([]config.Option, rows)
				for i := range options {
					options[i] = config.Option{Label: fmt.Sprintf("opt-%02d", i), Value: "v"}
				}
				m := newListModel("标题", options, ListSingle, true, nil)
				m.height = vp + 4 // CalculateTableHeight(h)=h-4,取 min(h-4, rows)=vp
				m.width = 100
				m.applyLayout()
				m.rebuildRows()
				m.cursor = init
				m.adjustScroll()
				m.rebuildRows()

				var walk func(seq []string, depth int)
				jumps := 0
				walk = func(seq []string, depth int) {
					if depth == 0 {
						return
					}
					for _, k := range []string{"up", "down"} {
						before := m.cursor
						prevPos := cursorViewPos(m)
						m.Update(tea.KeyPressMsg{Code: keyCode(k)})
						after := m.cursor
						pos := cursorViewPos(m)
						if pos < 0 {
							t.Fatalf("rows=%d vp=%d init=%d seq=%v: 光标不可见(cursor=%d)", rows, vp, init, append(seq, k), after)
						}
						if after-before == 1 && pos-prevPos > 1 {
							jumps++
							t.Errorf("JUMP rows=%d vp=%d init=%d seq=%v: cursor %d->%d, viewPos %d->%d",
								rows, vp, init, append(seq, k), before, after, prevPos, pos)
						}
						walk(append(seq, k), depth-1)
					}
				}
				walk(nil, 6)
				if jumps > 0 {
					t.Logf("rows=%d vp=%d init=%d: %d jumps", rows, vp, init, jumps)
				}
			}
		}
	}
}

// TestListWrapCursorVisible 回归:列表超出视口高度时循环导航光标行必须可见。
// 旧实现 wrap 分支用 SetCursor 只重排内容不调 viewport 偏移,顶部 ↑ 跳到
// 末尾后光标行正好被顶到可见区下方一行,视图中无任何 ❯(选中状态不可见)。
// 断言循环后视图中出现 ❯ 且落在目标选项行。
func TestListWrapCursorVisible(t *testing.T) {
	options := make([]config.Option, 20)
	for i := range options {
		options[i] = config.Option{Label: "opt-" + string(rune('A'+i)), Value: "v"}
	}

	cursorRowLabel := func(view string) string {
		for line := range strings.SplitSeq(ansi.Strip(view), "\n") {
			if strings.Contains(line, "❯") {
				return strings.TrimSpace(line)
			}
		}
		return ""
	}

	t.Run("顶部↑循环到末尾,光标行可见且为最后一个选项", func(t *testing.T) {
		m := newListModel("标题", options, ListSingle, true, nil)
		m.width, m.height = 100, 12 // 视口高度 8,列表 20 项,必然滚动
		m.applyLayout()
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		if got := m.cursor; got != len(options)-1 {
			t.Fatalf("顶部 ↑ 后光标 = %d, want %d", got, len(options)-1)
		}
		view := m.View().Content
		if row := cursorRowLabel(view); !strings.Contains(row, options[len(options)-1].Label) {
			t.Errorf("循环后光标行不可见: %q, want 含 %s", row, options[len(options)-1].Label)
		}
	})

	t.Run("底部↓循环回顶部,光标行可见且为第一个选项", func(t *testing.T) {
		m := newListModel("标题", options, ListSingle, true, nil)
		m.width, m.height = 100, 12
		m.applyLayout()
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown}) // 逐格走到末尾,期间正常滚动
		for range len(options) - 2 {
			m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		}
		if got := m.cursor; got != len(options)-1 {
			t.Fatalf("步进后光标 = %d, want %d", got, len(options)-1)
		}
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown}) // 底部 ↓ 循环回顶部
		if got := m.cursor; got != 0 {
			t.Fatalf("底部 ↓ 后光标 = %d, want 0", got)
		}
		view := m.View().Content
		if row := cursorRowLabel(view); !strings.Contains(row, options[0].Label) {
			t.Errorf("循环回顶部后光标行不可见: %q, want 含 %s", row, options[0].Label)
		}
	})
}

func keyCode(s string) rune {
	switch s {
	case "up":
		return tea.KeyUp
	case "down":
		return tea.KeyDown
	}
	return 0
}
