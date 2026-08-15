package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// TestGetRecentCommitsKeepsFullMessage 长提交消息必须完整保留在标签中:
// squash 等命令从标签提取默认消息,截断会导致获取到残缺文本
func TestGetRecentCommitsKeepsFullMessage(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录(GetRecentCommits 在当前目录执行 git)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 超过 50 显示宽度的提交消息(中文+emoji+ASCII)
	longMessage := "这是一个非常长的提交消息用于验证完整保留不被截断🚀以及更多的说明文字 xyz1234567890"
	os.WriteFile(testFile, []byte("update"), 0644)
	testutil.RunGitCommand(t, tempDir, "commit", "-am", longMessage)

	options, hashes, err := GetRecentCommits()
	if err != nil {
		t.Fatalf("GetRecentCommits() error = %v", err)
	}
	if len(hashes) < 2 {
		t.Fatalf("commits = %d, want >= 2", len(hashes))
	}

	// 最新提交在最前,其完整消息必须出现在标签中,不得被省略
	if !strings.Contains(options[0].Label, longMessage) {
		t.Errorf("标签中消息被截断:\n got  %q\n want 包含 %q", options[0].Label, longMessage)
	}
}

// ─── 冲突解决闭环测试 ────────────────────────────────────────────────────────

// setupRebaseConflictRepo 构造双冲突仓库:
// feat(f1 改 f.txt / f2 改 g.txt) 与 main(m1 改 f.txt / m2 改 g.txt) 各两个提交,
// rebase 会依次在 f.txt、g.txt 上冲突两次。
func setupRebaseConflictRepo(t *testing.T) (tempDir string, restore func()) {
	t.Helper()

	tempDir, _ = testutil.CreateTempGitRepo(t)
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	base := func(name, content string) {
		os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
	}
	base("f.txt", "base\n")
	base("g.txt", "base\n")
	testutil.RunGitCommand(t, tempDir, "add", ".")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "base")

	current, _ := getCurrentBranch()
	testutil.RunGitCommand(t, tempDir, "checkout", "-b", "feat")
	base("f.txt", "base\nfeat f\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "feat1")
	base("g.txt", "base\nfeat g\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "feat2")

	testutil.RunGitCommand(t, tempDir, "checkout", current)
	base("f.txt", "base\nmain f\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "main1")
	base("g.txt", "base\nmain g\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "main2")

	return tempDir, func() { os.Chdir(originalDir) }
}

// runGitAllowFail 执行 git 命令,失败不中断测试(用于触发冲突)
func runGitAllowFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, _ := cmd.CombinedOutput()
	return string(output)
}

// withConflictMenu 注入假菜单并返回恢复函数
func withConflictMenu(t *testing.T, menu func(options []config.Option) (string, error)) func() {
	t.Helper()
	old := rebaseConflictMenu
	rebaseConflictMenu = menu
	return func() { rebaseConflictMenu = old }
}

func TestIsRebaseInProgress(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	if isRebaseInProgress() {
		t.Error("isRebaseInProgress() = true, want false before rebase")
	}

	runGitAllowFail(t, tempDir, "rebase", "feat")
	if !isRebaseInProgress() {
		t.Error("isRebaseInProgress() = false, want true after conflict")
	}
}

func TestGetUnmergedFiles(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")
	files := getUnmergedFiles()
	if len(files) != 1 || files[0] != "f.txt" {
		t.Errorf("getUnmergedFiles() = %v, want [f.txt]", files)
	}
}

// TestHandleRebaseConflictMultiConflictLoop 核心回归:两次冲突自动循环,
// 第一次 continue 后不退出,第二次解决后变基完成。
func TestHandleRebaseConflictMultiConflictLoop(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")
	if !isRebaseInProgress() {
		t.Fatal("rebase should be in progress after first conflict")
	}

	menuCalls := 0
	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		menuCalls++
		switch menuCalls {
		case 1: // 第一次冲突:解决 f.txt 后继续
			os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("base\nresolved f\n"), 0644)
			return "continue", nil
		case 2: // 第二次冲突:解决 g.txt 后继续
			os.WriteFile(filepath.Join(tempDir, "g.txt"), []byte("base\nresolved g\n"), 0644)
			return "continue", nil
		}
		return "abort", nil
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if menuCalls != 2 {
		t.Errorf("menu called %d times, want 2 (multi-conflict loop)", menuCalls)
	}
	if isRebaseInProgress() {
		t.Error("rebase should be completed after resolving both conflicts")
	}

	// 变基结果验证:两个提交都在,内容为解决后的值
	out := runGitAllowFail(t, tempDir, "log", "--oneline")
	if !strings.Contains(out, "main2") || !strings.Contains(out, "main1") {
		t.Errorf("rebase 结果缺少提交:\n%s", out)
	}
	content, _ := os.ReadFile(filepath.Join(tempDir, "f.txt"))
	if !strings.Contains(string(content), "resolved f") {
		t.Errorf("f.txt 内容 = %q, want 含 resolved f", string(content))
	}
}

func TestHandleRebaseConflictSkip(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "skip", nil // 跳过 main1
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if isRebaseInProgress() {
		t.Error("rebase should complete after skip")
	}
	out := runGitAllowFail(t, tempDir, "log", "--oneline")
	if strings.Contains(out, "main1") {
		t.Errorf("main1 应被跳过:\n%s", out)
	}
}

func TestHandleRebaseConflictAbort(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "abort", nil
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if isRebaseInProgress() {
		t.Error("rebase should be aborted")
	}
}

func TestHandleRebaseConflictQuit(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "quit", nil
	})()

	err := handleRebaseConflict()
	if err == nil {
		t.Fatal("handleRebaseConflict() should return error on quit")
	}
	if !isRebaseInProgress() {
		t.Error("rebase should remain pending after quit")
	}
}
