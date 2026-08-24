package gitcmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// setupReflogRepo 构造带多次提交的仓库(产生 reflog 记录)
func setupReflogRepo(t *testing.T) string {
	t.Helper()
	tempDir := setupTwoBranchRepo(t) // base 提交

	writeFile(t, tempDir, "f.txt", "second\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "second commit")
	return tempDir
}

// ─── 列表解析 ────────────────────────────────────────────────────────────────

// TestListReflogParsing 记录数、最新在前、字段解析(含 HEAD@{N} 标记)
func TestListReflogParsing(t *testing.T) {
	tempDir := setupReflogRepo(t)

	entries, err := listReflog(undoReflogLimit)
	if err != nil {
		t.Fatalf("listReflog() error = %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("entries = %d, want >= 2(base + second)", len(entries))
	}

	first := entries[0]
	if first.Sel != "HEAD@{0}" {
		t.Errorf("最新记录 Sel = %q, want HEAD@{0}", first.Sel)
	}
	if !strings.Contains(first.Action, "commit: second commit") {
		t.Errorf("最新动作 = %q, want 含 commit: second commit", first.Action)
	}
	if first.Hash == "" || first.Date == "" {
		t.Errorf("Hash/Date 不应为空: %+v", first)
	}

	// 第二条应指向更早的提交
	secondHash, _ := execOutput(t, tempDir, "rev-parse", "--short", "HEAD~1")
	if entries[1].Hash != strings.TrimSpace(secondHash) {
		t.Errorf("第二条 Hash = %q, want %q", entries[1].Hash, strings.TrimSpace(secondHash))
	}
}

// TestListReflogPipeInSubject 字段错位回归:提交消息含 "|" 时
// subject 位于末段(格式 %h|%cd|%gs),解析仍完整。
func TestListReflogPipeInSubject(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	writeFile(t, tempDir, "f.txt", "pipe\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "feat: a|b|c pipe subject")

	entries, err := listReflog(undoReflogLimit)
	if err != nil {
		t.Fatalf("listReflog() error = %v", err)
	}
	first := entries[0]
	if !strings.Contains(first.Action, "a|b|c pipe subject") {
		t.Errorf("Action 应保留含 | 的完整消息, got %q", first.Action)
	}
	// Date 段应为 "MM-DD HH:MM"(11 字符),而非消息片段
	if len(first.Date) != 11 || first.Date[2] != '-' || first.Date[5] != ' ' || first.Date[8] != ':' {
		t.Errorf("Date 段格式异常: %q", first.Date)
	}
}

// TestListReflogLimit limit 生效
func TestListReflogLimit(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	for i := range 5 {
		writeFile(t, tempDir, "f.txt", fmt.Sprintf("change %d\n", i))
		testutil.RunGitCommand(t, tempDir, "commit", "-am", "filler")
	}

	entries, err := listReflog(3)
	if err != nil {
		t.Fatalf("listReflog() error = %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("entries = %d, want 3(limit 生效)", len(entries))
	}
}

// ─── 核心执行 ────────────────────────────────────────────────────────────────

// TestExecuteUndoHard hard 回滚:提交与文件内容一并回退,工作区干净
func TestExecuteUndoHard(t *testing.T) {
	tempDir := setupReflogRepo(t)

	entries, _ := listReflog(2)
	target := entries[1] // 回到 base

	if err := executeUndo(target, "--hard"); err != nil {
		t.Fatalf("executeUndo() error = %v", err)
	}

	subject := headSubject(t, tempDir)
	if subject != "base" {
		t.Errorf("HEAD subject = %q, want base", subject)
	}
	dirty, _ := isWorkingDirectoryDirty()
	if dirty {
		t.Error("hard 回滚后工作区应干净")
	}
	if content := readFileContent(t, tempDir, "f.txt"); strings.Contains(content, "second") {
		t.Errorf("f.txt = %q, want 已回退", content)
	}
}

// TestExecuteUndoSoft soft 回滚:HEAD 后退但改动保留在暂存区
func TestExecuteUndoSoft(t *testing.T) {
	tempDir := setupReflogRepo(t)

	entries, _ := listReflog(2)
	target := entries[1]

	if err := executeUndo(target, "--soft"); err != nil {
		t.Fatalf("executeUndo() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "base" {
		t.Errorf("HEAD subject = %q, want base", subject)
	}
	out, _ := execOutput(t, tempDir, "diff", "--cached", "--name-only")
	if !strings.Contains(out, "f.txt") {
		t.Errorf("soft 模式改动应保留在暂存区:\n%s", out)
	}
}

// TestExecuteUndoDefaultMode 空模式等同 mixed(reset 不传 flag)
func TestExecuteUndoDefaultMode(t *testing.T) {
	tempDir := setupReflogRepo(t)

	entries, _ := listReflog(2)

	if err := executeUndo(entries[1], ""); err != nil {
		t.Fatalf("executeUndo() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "base" {
		t.Errorf("HEAD subject = %q, want base", subject)
	}
	out, _ := execOutput(t, tempDir, "status", "--porcelain")
	if !strings.Contains(out, "M") && !strings.Contains(out, "f.txt") {
		t.Errorf("mixed 模式改动应保留在工作区:\n%s", out)
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestUndoFlow 完整流程:选检查点 → 选 hard → 确认 → 回滚生效
func TestUndoFlow(t *testing.T) {
	tempDir := setupReflogRepo(t)

	restoreSelect := stubMenu(t, &undoSelectEntry, "HEAD@{1}")
	defer restoreSelect()
	restoreMode := stubUndoMode(t, "--hard")
	defer restoreMode()
	restoreConfirm := stubBool(t, &undoConfirm, true)
	defer restoreConfirm()

	if err := Undo(); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "base" {
		t.Errorf("HEAD subject = %q, want base", subject)
	}
}

// TestUndoFlowCancelled 确认取消时不做任何事
func TestUndoFlowCancelled(t *testing.T) {
	tempDir := setupReflogRepo(t)
	oldSubject := headSubject(t, tempDir)

	restoreSelect := stubMenu(t, &undoSelectEntry, "HEAD@{1}")
	defer restoreSelect()
	restoreMode := stubUndoMode(t, "--hard")
	defer restoreMode()
	restoreConfirm := stubBool(t, &undoConfirm, false)
	defer restoreConfirm()

	if err := Undo(); err != nil {
		t.Fatalf("Undo() error = %v", err)
	}
	if headSubject(t, tempDir) != oldSubject {
		t.Error("取消后 HEAD 不应变化")
	}
}

// TestUndoFlowEsc 选择检查点时 Esc 直接退出
func TestUndoFlowEsc(t *testing.T) {
	setupReflogRepo(t)

	oldSelect := undoSelectEntry
	undoSelectEntry = func(options []config.Option) (string, error) { return "", errUserAbortedStub }
	defer func() { undoSelectEntry = oldSelect }()

	if err := Undo(); err != nil {
		t.Errorf("Esc 取消不应报错, got %v", err)
	}
}

// stubUndoMode 注入重置模式选择结果
func stubUndoMode(t *testing.T, mode string) func() {
	t.Helper()
	old := undoModeSelect
	undoModeSelect = func(options []config.Option, preselected string) (string, error) {
		return mode, nil
	}
	return func() { undoModeSelect = old }
}
