package testutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// SetupTestRepo 创建临时 Git 仓库
func SetupTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// 初始化仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// 配置用户
	runGitCmd(t, tmpDir, "config", "user.name", "Test User")
	runGitCmd(t, tmpDir, "config", "user.email", "test@example.com")

	return tmpDir
}

// CreateTestFile 创建测试文件
func CreateTestFile(t *testing.T, repoDir, filename, content string) {
	t.Helper()

	path := filepath.Join(repoDir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
}

// CommitFile 提交文件
func CommitFile(t *testing.T, repoDir, filename, message string) string {
	t.Helper()

	runGitCmd(t, repoDir, "add", filename)
	runGitCmd(t, repoDir, "commit", "-m", message)

	// 返回 commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output))
}

// CreateBranch 创建分支
func CreateBranch(t *testing.T, repoDir, branchName string) {
	t.Helper()
	runGitCmd(t, repoDir, "checkout", "-b", branchName)
}

// CheckoutBranch 切换分支
func CheckoutBranch(t *testing.T, repoDir, branchName string) {
	t.Helper()
	runGitCmd(t, repoDir, "checkout", branchName)
}

// CreateTag 创建标签
func CreateTag(t *testing.T, repoDir, tagName string) {
	t.Helper()
	runGitCmd(t, repoDir, "tag", tagName)
}

// runGitCmd 执行 Git 命令
func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, output)
	}
}

// GetCurrentBranch 获取当前分支名
func GetCurrentBranch(t *testing.T, repoDir string) string {
	t.Helper()

	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("failed to get current branch: %v", err)
	}
	return strings.TrimSpace(string(output))
}

// CreateTempGitRepo 创建临时 Git 仓库并返回清理函数
func CreateTempGitRepo(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir := t.TempDir()

	// 初始化仓库
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	// 配置用户
	runGitCmd(t, tmpDir, "config", "user.name", "Test User")
	runGitCmd(t, tmpDir, "config", "user.email", "test@example.com")

	cleanup := func() {
		// t.TempDir() 会自动清理,这里不需要额外操作
	}

	return tmpDir, cleanup
}

// RunGitCommand 执行 Git 命令(导出版本)
func RunGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	runGitCmd(t, dir, args...)
}
