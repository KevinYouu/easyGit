package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
