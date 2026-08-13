package command

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func TestNewProgressModel(t *testing.T) {
	commands := []CommandInfo{
		{
			Command:     "echo",
			Args:        []string{"test1"},
			Description: "Test command 1",
			LoadingMsg:  "Loading 1",
			SuccessMsg:  "Success 1",
		},
		{
			Command:     "echo",
			Args:        []string{"test2"},
			Description: "Test command 2",
			LoadingMsg:  "Loading 2",
			SuccessMsg:  "Success 2",
		},
	}

	model := NewProgressModel(commands)

	if model == nil {
		t.Fatal("NewProgressModel returned nil")
	}

	if len(model.commands) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(model.commands))
	}

	if model.currentStep != 0 {
		t.Errorf("expected currentStep 0, got %d", model.currentStep)
	}

	if model.total != len(commands) {
		t.Errorf("expected total %d, got %d", len(commands), model.total)
	}
}

func TestNewProgressModelWithoutSpinner(t *testing.T) {
	commands := []CommandInfo{
		{
			Command:     "echo",
			Args:        []string{"test"},
			Description: "Test command",
		},
	}

	model := NewProgressModelWithoutSpinner(commands)

	if model == nil {
		t.Fatal("NewProgressModelWithoutSpinner returned nil")
	}

	if len(model.commands) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(model.commands))
	}
}

func TestCommandInfo(t *testing.T) {
	cmd := CommandInfo{
		Command:     "git",
		Args:        []string{"status"},
		Description: "Check git status",
		LoadingMsg:  "Checking status...",
		SuccessMsg:  "Status checked",
	}

	if cmd.Command != "git" {
		t.Errorf("expected command %q, got %q", "git", cmd.Command)
	}

	if len(cmd.Args) != 1 || cmd.Args[0] != "status" {
		t.Errorf("expected args [status], got %v", cmd.Args)
	}
}

// TestProgressView 进度组件 View 渲染:
// 步骤描述可见、进度条宽度自适应、行宽不越界、状态图标随进度切换
func TestProgressView(t *testing.T) {
	commands := []CommandInfo{
		{Command: "git", Args: []string{"add", "."}, Description: "暂存更改", SuccessMsg: "暂存完成"},
		{Command: "git", Args: []string{"commit", "-m", "x"}, Description: "提交更改", SuccessMsg: "提交完成"},
		{Command: "git", Args: []string{"push"}, Description: "推送远端", SuccessMsg: "推送完成"},
	}

	// extractBar 提取进度条行中 [ 与 ] 之间的条体字符数。
	// ░/█ 为多字节 UTF-8,须按 rune 计数而非字节距离
	extractBar := func(view string) int {
		for line := range strings.SplitSeq(view, "\n") {
			plain := ansi.Strip(line)
			start := strings.Index(plain, "[")
			end := strings.Index(plain, "]")
			if start >= 0 && end > start {
				return utf8.RuneCountInString(plain[start+1 : end])
			}
		}
		return -1
	}

	t.Run("初始状态", func(t *testing.T) {
		m := NewProgressModel(commands)
		view := ansi.Strip(m.View().Content)

		for _, desc := range []string{"暂存更改", "提交更改", "推送远端"} {
			if !strings.Contains(view, desc) {
				t.Errorf("步骤描述 %q 未出现在视图中", desc)
			}
		}
		if !strings.Contains(view, "(0/3)") {
			t.Errorf("进度计数错误: %q", view)
		}
		if strings.Count(view, "○") != 3 {
			t.Errorf("未决步骤应显示 3 个 ○")
		}
	})

	t.Run("进度条宽度自适应且行宽不越界", func(t *testing.T) {
		// 宽屏封顶 40;窄屏按实际行开销收缩条体,空间耗尽时条体为 0,
		// 进度条行仍保留前后缀不折行。
		// 80 列下中英文行开销都足以达到封顶,封顶断言与语言无关。
		// 步骤描述使用超长文本,验证步骤行按剩余宽度截断同样不折行
		commands := []CommandInfo{
			{Command: "git", Args: []string{"add", "."}, Description: "暂存所有更改到索引区等待后续提交", SuccessMsg: "暂存完成"},
			{Command: "git", Args: []string{"commit", "-m", "x"}, Description: "提交代码并生成一条新的提交记录", SuccessMsg: "提交完成"},
			{Command: "git", Args: []string{"push"}, Description: "推送本地提交到远程仓库 origin/main 分支", SuccessMsg: "推送完成"},
		}
		for _, width := range []int{80, 50, 40, 30, 24} {
			m := NewProgressModel(commands)
			m.width = width
			view := m.View().Content

			bar := extractBar(view)
			if bar < 0 || bar > progressBarMaxWidth {
				t.Errorf("宽度 %d:进度条长度 %d 非法", width, bar)
			}
			// 宽屏应封顶 40
			if width == 80 && bar != progressBarMaxWidth {
				t.Errorf("宽度 %d:进度条长度 %d, want 40", width, bar)
			}
			// 所有行(标题/进度条/状态/步骤/提示)都不得越界
			for line := range strings.SplitSeq(view, "\n") {
				if lw := lipgloss.Width(ansi.Strip(line)); lw > width {
					t.Errorf("宽度 %d:行宽 %d 越界: %q", width, lw, line)
				}
			}
		}
	})

	t.Run("状态图标切换", func(t *testing.T) {
		// 步骤进行中:已完成 ✓、运行中 ▶(无 spinner)、未决 ○,状态行 ⚡
		m := NewProgressModelWithoutSpinner(commands)
		m.width = 80
		m.currentStep = 1
		m.executing = true
		m.stepStatus = []int{2, 1, 0}
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "✓") || !strings.Contains(view, "▶") || !strings.Contains(view, "⚡") {
			t.Errorf("进行中状态图标错误: %q", view)
		}

		// 全部完成:步骤全 ✓、状态行 ✅、完成提示
		m.isCompleted = true
		m.currentStep = 3
		m.executing = false
		m.stepStatus = []int{2, 2, 2}
		view = ansi.Strip(m.View().Content)
		if strings.Count(view, "✓") != 3 || !strings.Contains(view, "✅") {
			t.Errorf("完成状态图标错误: %q", view)
		}

		// 失败:状态行 ❌
		m.hasError = true
		view = ansi.Strip(m.View().Content)
		if !strings.Contains(view, "❌") {
			t.Errorf("失败状态图标错误: %q", view)
		}
	})
}
