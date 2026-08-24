package gitcmd

import (
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// headSubject 当前 HEAD 提交主题
func headSubject(t *testing.T, dir string) string {
	t.Helper()
	out, err := execOutput(t, dir, "log", "-1", "--pretty=%s")
	if err != nil {
		t.Fatalf("git log failed: %v", err)
	}
	return out
}

// headHash 当前 HEAD 哈希
func headHash(t *testing.T, dir string) string {
	t.Helper()
	out, _ := execOutput(t, dir, "rev-parse", "HEAD")
	return out
}

// ─── 纯函数单元测试 ──────────────────────────────────────────────────────────

// TestIsHeadPushed 已推送提交返回 true;本地新提交与无远程仓库返回 false
func TestIsHeadPushed(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// 无远程
	if isHeadPushed() {
		t.Error("无远程仓库时应为 false")
	}

	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "main")
	if !isHeadPushed() {
		t.Error("已推送的 HEAD 应为 true")
	}

	// 推送后产生新的本地提交
	writeFile(t, tempDir, "f.txt", "new local\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "local only")
	if isHeadPushed() {
		t.Error("未推送的新 HEAD 应为 false")
	}
}

func TestGetHeadSubject(t *testing.T) {
	setupTwoBranchRepo(t) // 最后一个提交为 "base"

	subject, err := getHeadSubject()
	if err != nil {
		t.Fatalf("getHeadSubject() error = %v", err)
	}
	if subject != "base" {
		t.Errorf("subject = %q, want base", subject)
	}
}

func TestGetUpstreamRef(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	if got := getUpstreamRef(); got != "" {
		t.Errorf("无上游时 = %q, want 空", got)
	}

	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "-u", "origin", "main")
	if got := getUpstreamRef(); got != "origin/main" {
		t.Errorf("upstream = %q, want origin/main", got)
	}
}

// ─── 核心执行 ────────────────────────────────────────────────────────────────

// TestExecuteAmendMessageOnly 仅改消息:消息更新、文件不变、暂存区不受影响
func TestExecuteAmendMessageOnly(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	oldHash := headHash(t, tempDir)

	if err := executeAmend(nil, "rewritten subject"); err != nil {
		t.Fatalf("executeAmend() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "rewritten subject" {
		t.Errorf("subject = %q, want rewritten subject", subject)
	}
	if newHash := headHash(t, tempDir); newHash == oldHash {
		t.Error("amend 后 HEAD 哈希应变化")
	}
	// 提交数量不变(仍为一个 base 提交)
	out, _ := execOutput(t, tempDir, "rev-list", "--count", "HEAD")
	if strings.TrimSpace(out) != "1" {
		t.Errorf("提交数 = %s, want 1(amend 不新增提交)", out)
	}
}

// TestExecuteAmendAppendFiles 追加文件:改动并入上次提交,消息保持不变
func TestExecuteAmendAppendFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "extra.txt", "appended content\n")

	if err := executeAmend([]string{"extra.txt"}, ""); err != nil {
		t.Fatalf("executeAmend() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "base" {
		t.Errorf("消息应保持不变, got %q", subject)
	}
	// extra.txt 进入上次提交
	out, _ := execOutput(t, tempDir, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(out, "extra.txt") {
		t.Errorf("HEAD 提交应含 extra.txt:\n%s", out)
	}
	// 工作区干净
	dirty, _ := isWorkingDirectoryDirty()
	if dirty {
		t.Error("追加后工作区应干净")
	}
}

// TestExecuteAmendBoth 消息 + 文件同时修改
func TestExecuteAmendBoth(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "both.txt", "both\n")

	if err := executeAmend([]string{"both.txt"}, "new message with file"); err != nil {
		t.Fatalf("executeAmend() error = %v", err)
	}
	if subject := headSubject(t, tempDir); subject != "new message with file" {
		t.Errorf("subject = %q, want new message with file", subject)
	}
	out, _ := execOutput(t, tempDir, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(out, "both.txt") {
		t.Errorf("HEAD 提交应含 both.txt:\n%s", out)
	}
}

// TestExecuteAmendEmptyRepo 无提交仓库报错
func TestExecuteAmendEmptyRepo(t *testing.T) {
	_, restore := chdirTempRepo(t)
	defer restore()

	if _, err := getHeadSubject(); err == nil {
		t.Error("空仓库 getHeadSubject 应报错")
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestAmendFlowMessage 主流程:仅改消息(已推送场景)→ 警告放行 → 强推拒绝
func TestAmendFlowMessage(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "-u", "origin", "main")

	restoreMenu := stubMenu(t, &amendActionMenu, "message")
	defer restoreMenu()

	// 第 1 次确认 = 已推送警告(放行);第 2 次 = 强推询问(拒绝)
	confirms := 0
	warnAccepted := false
	oldConfirm := amendConfirm
	amendConfirm = func(string) bool {
		confirms++
		if confirms == 1 {
			warnAccepted = true
			return true
		}
		return false
	}
	defer func() { amendConfirm = oldConfirm }()

	restoreInput := stubFunc(t, &amendMessageInput, func(current string) (string, error) {
		if current != "base" {
			t.Errorf("输入应预填原消息 base, got %q", current)
		}
		return "flow rewritten", nil
	})
	defer restoreInput()

	if err := Amend(); err != nil {
		t.Fatalf("Amend() error = %v", err)
	}
	if !warnAccepted {
		t.Error("应先弹出已推送警告确认")
	}
	if subject := headSubject(t, tempDir); subject != "flow rewritten" {
		t.Errorf("subject = %q, want flow rewritten", subject)
	}
	if confirms < 2 {
		t.Error("amend 后应询问是否强制推送")
	}
	// 本地已改写,远程仍是旧的(未强推)
	out, _ := execOutput(t, tempDir, "ls-remote", "origin", "refs/heads/main")
	if strings.Contains(out, headHash(t, tempDir)) {
		t.Error("拒绝强推后远程不应包含新哈希")
	}
}

// TestAmendFlowPushedWarningCancel 已推送且用户取消警告 → 不做任何修改
func TestAmendFlowPushedWarningCancel(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "main")
	oldHash := headHash(t, tempDir)

	restoreMenu := stubMenu(t, &amendActionMenu, "message")
	defer restoreMenu()

	oldConfirm := amendConfirm
	amendConfirm = func(string) bool { return false }
	defer func() { amendConfirm = oldConfirm }()

	inputCalled := false
	restoreInput := stubFunc(t, &amendMessageInput, func(current string) (string, error) {
		inputCalled = true
		return "should not happen", nil
	})
	defer restoreInput()

	if err := Amend(); err != nil {
		t.Fatalf("Amend() error = %v", err)
	}
	if headHash(t, tempDir) != oldHash {
		t.Error("取消后 HEAD 不应变化")
	}
	if inputCalled {
		t.Error("取消后不应进入消息输入")
	}
}

// TestAmendFlowFiles 文件追加主流程(未推送场景,不触发强推询问)
func TestAmendFlowFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "flow.txt", "flow content\n")

	restoreMenu := stubMenu(t, &amendActionMenu, "files")
	defer restoreMenu()

	oldSelect := amendFilesSelect
	amendFilesSelect = func(options []config.Option) ([]string, error) {
		return []string{"flow.txt"}, nil
	}
	defer func() { amendFilesSelect = oldSelect }()

	if isHeadPushed() {
		t.Fatal("此用例要求 HEAD 未推送")
	}
	if err := Amend(); err != nil {
		t.Fatalf("Amend() error = %v", err)
	}
	out, _ := execOutput(t, tempDir, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(out, "flow.txt") {
		t.Errorf("HEAD 提交应含 flow.txt:\n%s", out)
	}
}

// TestAmendFlowEsc 操作菜单 Esc 直接退出
func TestAmendFlowEsc(t *testing.T) {
	setupTwoBranchRepo(t)

	oldMenu := amendActionMenu
	amendActionMenu = func(options []config.Option) (string, error) {
		return "", errUserAbortedStub
	}
	defer func() { amendActionMenu = oldMenu }()

	if err := Amend(); err != nil {
		t.Errorf("Esc 取消不应报错, got %v", err)
	}
}
