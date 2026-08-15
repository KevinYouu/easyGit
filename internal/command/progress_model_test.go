package command

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// captureStdout 捕获函数执行期间的 stdout 输出
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	return string(out)
}

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
		if strings.Count(view, "○") != 4 {
			t.Errorf("未决步骤应显示 3 个 ○ 加状态行 1 个 ○,共 4 个")
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
		// 步骤进行中:已完成 ✓、运行中 ▶(无 spinner)、未决 ○,状态行 ▶
		m := NewProgressModelWithoutSpinner(commands)
		m.width = 80
		m.currentStep = 1
		m.executing = true
		m.stepStatus = []int{2, 1, 0}
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "✓") || !strings.Contains(view, "▶") || !strings.Contains(view, "○") {
			t.Errorf("进行中状态图标错误: %q", view)
		}

		// 全部完成:步骤 3 个 ✓ + 状态行 ✓
		m.isCompleted = true
		m.currentStep = 3
		m.executing = false
		m.stepStatus = []int{2, 2, 2}
		view = ansi.Strip(m.View().Content)
		if strings.Count(view, "✓") != 4 {
			t.Errorf("完成状态图标错误: %q", view)
		}

		// 失败:状态行 ✗
		m.hasError = true
		view = ansi.Strip(m.View().Content)
		if !strings.Contains(view, "✗") {
			t.Errorf("失败状态图标错误: %q", view)
		}
	})
}

// TestProgressHelpBar 执行中底部帮助栏:≥6 行终端渲染 [q] 前缀,<6 行隐藏,完成时切换为退出提示
func TestProgressHelpBar(t *testing.T) {
	commands := []CommandInfo{
		{Command: "git", Args: []string{"add", "."}, Description: "暂存更改", SuccessMsg: "暂存完成"},
	}

	t.Run("执行中渲染帮助栏", func(t *testing.T) {
		m := NewProgressModelWithoutSpinner(commands)
		m.width, m.height = 80, 24
		m.executing = true
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, "[q]") {
			t.Errorf("执行中应渲染 [q] 退出前缀: %q", view)
		}
		if !strings.Contains(view, i18n.T("form.help.quit")) {
			t.Errorf("执行中应含退出动作文本: %q", view)
		}
	})

	t.Run("极小终端隐藏帮助栏", func(t *testing.T) {
		m := NewProgressModelWithoutSpinner(commands)
		m.width, m.height = 80, 4
		m.executing = true
		view := ansi.Strip(m.View().Content)
		if strings.Contains(view, "[q]") {
			t.Errorf("4 行终端不应渲染帮助栏: %q", view)
		}
	})

	t.Run("完成时保留退出提示", func(t *testing.T) {
		m := NewProgressModelWithoutSpinner(commands)
		m.width, m.height = 80, 24
		m.isCompleted = true
		view := ansi.Strip(m.View().Content)
		if !strings.Contains(view, i18n.T("ui.exiting.success")) {
			t.Errorf("完成时应显示退出提示: %q", view)
		}
		if strings.Contains(view, "[q]") {
			t.Errorf("完成时不应渲染帮助栏: %q", view)
		}
	})
}

// TestProgressI18nStatus 进度状态文案走 i18n,与当前语言一致
func TestProgressI18nStatus(t *testing.T) {
	m := NewProgressModel(nil)
	if m.status != i18n.T("progress.preparing") {
		t.Errorf("初始状态应为 i18n 文案 %q, got %q", i18n.T("progress.preparing"), m.status)
	}
}

// pumpProgress 将 Update 结果转回 *ProgressModel 并返回产生的命令
func pumpProgress(m *ProgressModel, msg tea.Msg) (*ProgressModel, tea.Cmd) {
	nm, cmd := m.Update(msg)
	return nm.(*ProgressModel), cmd
}

// runParallelFlow 驱动并行模型完成一次完整流程:
// 串行段逐条执行(步骤 < parallelFrom),并行段按给定顺序乱序完成。
// 返回最终模型。
func runParallelFlow(t *testing.T, m *ProgressModel, parallelFrom int, completions []stepResult) *ProgressModel {
	t.Helper()

	// 启动第一批(步骤 0)
	cmd := m.startNextBatch()
	var msg tea.Msg = cmd()

	// pendingSteps 记录已启动未完成的步骤
	pending := map[int]bool{}

	for _, cr := range completions {
		start, ok := msg.(StepStartMsg)
		if !ok {
			t.Fatalf("期望 StepStartMsg, got %T", msg)
		}
		pending[start.Step] = true
		var next tea.Cmd
		m, next = pumpProgress(m, msg)

		// 完成一个步骤(乱序完成:由调用方指定)
		if !pending[cr.step] {
			t.Fatalf("步骤 %d 尚未启动", cr.step)
		}
		delete(pending, cr.step)
		m, next = pumpProgress(m, StepCompleteMsg{Step: cr.step, Success: cr.success, Output: cr.output})
		if next == nil {
			if len(pending) > 0 || cr.step != len(m.commands)-1 {
				t.Fatalf("步骤 %d 完成后无新命令, 仍有在飞 %v", cr.step, pending)
			}
			msg = nil
			continue
		}
		msg = next()
	}
	return m
}

type stepResult struct {
	step    int
	success bool
	output  string
}

// TestParallelProgressModel 并行段状态机:
// 串行段逐条推进 → 到达 parallelFrom 一次性启动全部剩余步骤 → 乱序完成 → 全部结束后 AllComplete
func TestParallelProgressModel(t *testing.T) {
	commands := []CommandInfo{
		{Command: "git", Args: []string{"add", "."}, Description: "暂存更改"},
		{Command: "git", Args: []string{"commit", "-m", "x"}, Description: "提交更改"},
		{Command: "git", Args: []string{"pull"}, Description: "拉取更新"},
		{Command: "git", Args: []string{"push", "origin"}, Description: "推送到 origin"},
		{Command: "git", Args: []string{"push", "github"}, Description: "推送到 github"},
	}

	t.Run("构造与越界保护", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)
		if m.parallelFrom != 3 {
			t.Errorf("parallelFrom 应为 3, got %d", m.parallelFrom)
		}
		if NewProgressModelParallel(commands, 99).parallelFrom != -1 {
			t.Errorf("越界 parallelFrom 应回退为全串行")
		}
		if NewProgressModelParallel(commands, -1).parallelFrom != -1 {
			t.Errorf("负 parallelFrom 应回退为全串行")
		}
		if NewProgressModel(commands).parallelFrom != -1 {
			t.Errorf("普通模型应全串行")
		}
	})

	t.Run("串行推进到并行段,乱序完成全成功", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)

		// 第一批只启动步骤 0
		cmd := m.startNextBatch()
		if msg := cmd().(StepStartMsg); msg.Step != 0 {
			t.Fatalf("第一批应启动步骤 0, got %d", msg.Step)
		}
		if m.inFlight != 1 {
			t.Errorf("串行段在飞数应为 1, got %d", m.inFlight)
		}

		// 步骤 0 → 1 → 2 逐条推进
		m, _ = pumpProgress(m, StepStartMsg{Step: 0, Description: "x"})
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 0, Success: true})
		if msg := cmd().(StepStartMsg); msg.Step != 1 {
			t.Fatalf("完成后应启动步骤 1, got %d", msg.Step)
		}
		m, _ = pumpProgress(m, StepStartMsg{Step: 1, Description: "x"})
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 1, Success: true})
		if msg := cmd().(StepStartMsg); msg.Step != 2 {
			t.Fatalf("完成后应启动步骤 2, got %d", msg.Step)
		}
		m, _ = pumpProgress(m, StepStartMsg{Step: 2, Description: "x"})
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 2, Success: true})

		// 步骤 2 完成 → 并行段一次性启动步骤 3,4
		batch, ok := cmd().(tea.BatchMsg)
		if !ok {
			t.Fatalf("并行段应返回 BatchMsg, got %T", cmd())
		}
		if len(batch) != 2 {
			t.Fatalf("并行段应启动 2 个步骤, got %d", len(batch))
		}
		if m.inFlight != 2 {
			t.Errorf("并行段在飞数应为 2, got %d", m.inFlight)
		}
		if !strings.Contains(m.status, "2") {
			t.Errorf("并行状态文案应含远程数, got %q", m.status)
		}

		// 两个步骤同时 running
		m, _ = pumpProgress(m, StepStartMsg{Step: 3, Description: "x"})
		m, _ = pumpProgress(m, StepStartMsg{Step: 4, Description: "x"})
		if m.stepStatus[3] != 1 || m.stepStatus[4] != 1 {
			t.Errorf("并行步骤应同时 running: %v", m.stepStatus)
		}

		// 乱序完成:步骤 4 先完成,仍有在飞,不启动新批次
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 4, Success: true})
		if cmd != nil {
			t.Errorf("还有在飞步骤时不应启动新批次")
		}
		if m.inFlight != 1 || m.stepStatus[4] != 2 {
			t.Errorf("步骤 4 应标记成功且剩余在飞 1: inFlight=%d status=%v", m.inFlight, m.stepStatus)
		}

		// 步骤 3 完成 → 全部结束
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 3, Success: true})
		all, ok := cmd().(AllCompleteMsg)
		if !ok || !all.Success {
			t.Fatalf("全部成功应 AllComplete(true), got %T %+v", cmd(), all)
		}
		if m.completedCount() != 5 {
			t.Errorf("完成数应为 5, got %d", m.completedCount())
		}
		if m.hasError {
			t.Errorf("不应有错误")
		}
	})

	t.Run("并行段部分失败:等待全部结束并收集失败", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)
		cmd := m.startNextBatch()
		m, _ = pumpProgress(m, cmd().(StepStartMsg))
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 0, Success: true})
		m, _ = pumpProgress(m, cmd().(StepStartMsg))
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 1, Success: true})
		m, _ = pumpProgress(m, cmd().(StepStartMsg))
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 2, Success: true})
		batch := cmd().(tea.BatchMsg)
		for _, c := range batch {
			m, _ = pumpProgress(m, c().(StepStartMsg))
		}

		// 步骤 3 失败(带输出),步骤 4 成功
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 3, Success: false, Output: "remote rejected"})
		if cmd != nil {
			t.Errorf("步骤 4 仍在飞,不应结束")
		}
		if !m.hasError || len(m.failedSteps) != 1 || m.failedSteps[0] != 3 {
			t.Errorf("应记录失败步骤 3: hasError=%v failedSteps=%v", m.hasError, m.failedSteps)
		}
		if m.errorMessage != "remote rejected" {
			t.Errorf("应收集失败输出, got %q", m.errorMessage)
		}
		if m.stepStatus[3] != 3 || m.stepStatus[4] != 1 {
			t.Errorf("失败/运行状态错: %v", m.stepStatus)
		}

		// 步骤 4 完成后才结束,结果为失败
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 4, Success: true})
		all, ok := cmd().(AllCompleteMsg)
		if !ok || all.Success {
			t.Fatalf("有失败应 AllComplete(false)")
		}
		if m.completedCount() != 5 {
			t.Errorf("失败步骤也应计入完成数, got %d", m.completedCount())
		}
	})

	t.Run("串行段失败立即终止,不启动后续", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)
		cmd := m.startNextBatch()
		m, _ = pumpProgress(m, cmd().(StepStartMsg))
		m, cmd = pumpProgress(m, StepCompleteMsg{Step: 0, Success: false, Output: "commit failed"})
		all, ok := cmd().(AllCompleteMsg)
		if !ok || all.Success {
			t.Fatalf("串行段失败应直接 AllComplete(false)")
		}
		if m.pendingStart != 1 || m.stepStatus[1] != 0 {
			t.Errorf("后续步骤不应启动: pendingStart=%d status=%v", m.pendingStart, m.stepStatus)
		}
	})
}

// TestParallelProgressView 并行段渲染:多个步骤同时 running 各自显示 spinner 帧,
// 完成/失败后 ✓/✗ 逐个点亮,进度按完成数推进
func TestParallelProgressView(t *testing.T) {
	commands := []CommandInfo{
		{Command: "git", Args: []string{"add", "."}, Description: "暂存更改"},
		{Command: "git", Args: []string{"commit"}, Description: "提交更改"},
		{Command: "git", Args: []string{"pull"}, Description: "拉取更新"},
		{Command: "git", Args: []string{"push", "origin"}, Description: "推送到 origin"},
		{Command: "git", Args: []string{"push", "github"}, Description: "推送到 github"},
	}

	t.Run("并行段多个 running", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)
		m.width = 80
		m.executing = true
		m.stepStatus = []int{2, 2, 2, 1, 1}
		m.status = fmt.Sprintf(i18n.T("progress.executing.parallel"), 2)
		m.frame = 0
		frame := m.spinner.frames[0]
		view := ansi.Strip(m.View().Content)

		if strings.Count(view, frame) != 2 { // 两个并行步骤各一个 spinner 帧(状态行为 ▶)
			t.Errorf("并行步骤应各自显示 spinner 帧: %q", view)
		}
		if strings.Count(view, "✓") != 3 {
			t.Errorf("3 个已完成步骤应显示 ✓: %q", view)
		}
		if !strings.Contains(view, fmt.Sprintf(i18n.T("progress.executing.parallel"), 2)) {
			t.Errorf("状态行应显示并行文案: %q", view)
		}
	})

	t.Run("并行段完成与失败图标", func(t *testing.T) {
		m := NewProgressModelParallel(commands, 3)
		m.width = 80
		m.isCompleted = true
		m.hasError = true
		m.stepStatus = []int{2, 2, 2, 3, 2}
		view := ansi.Strip(m.View().Content)
		if strings.Count(view, "✗") != 2 { // 步骤 4 一个 + 状态行一个
			t.Errorf("失败步骤应显示 ✗: %q", view)
		}
		if !strings.Contains(view, "(5/5)") {
			t.Errorf("失败也计入进度: %q", view)
		}
	})
}

// TestRunMultipleCommandsParallelSummary 失败摘要应列出全部失败步骤
func TestRunMultipleCommandsParallelSummary(t *testing.T) {
	commands := []CommandInfo{
		{Command: "git", Args: []string{"push", "origin"}, Description: "推送到 origin"},
		{Command: "git", Args: []string{"push", "github"}, Description: "推送到 github"},
	}
	m := NewProgressModelParallel(commands, 0)
	m.hasError = true
	m.failedSteps = []int{0, 1}
	m.errorMessage = "rejected 1\nrejected 2"

	out := captureStdout(func() { printExecutionSummary(m) })
	if !strings.Contains(out, "推送到 origin") || !strings.Contains(out, "推送到 github") {
		t.Errorf("摘要应列出全部失败步骤: %q", out)
	}
	if !strings.Contains(out, "rejected 1") || !strings.Contains(out, "rejected 2") {
		t.Errorf("摘要应包含全部错误输出: %q", out)
	}
}
