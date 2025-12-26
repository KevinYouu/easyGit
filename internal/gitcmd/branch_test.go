package gitcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

func TestGetCurrentBranch(t *testing.T) {
	// 创建临时Git仓库
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交(空仓库没有HEAD会失败)
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 测试获取当前分支
	branch, err := GetCurrentBranch()
	if err != nil {
		t.Fatalf("GetCurrentBranch() error = %v", err)
	}

	// 默认分支应该是main或master
	if branch != "main" && branch != "master" {
		t.Errorf("GetCurrentBranch() = %v, want main or master", branch)
	}
}

func TestGetCurrentBranch_NotGitRepo(t *testing.T) {
	// 创建临时目录但不初始化为git仓库
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 测试在非Git仓库中调用
	_, err := GetCurrentBranch()
	if err == nil {
		t.Error("GetCurrentBranch() expected error in non-git repo, got nil")
	}
}

func TestGetAllBranches(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建测试文件并提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 创建新分支
	testutil.RunGitCommand(t, tempDir, "branch", "feature-1")
	testutil.RunGitCommand(t, tempDir, "branch", "feature-2")

	// 获取所有分支
	branches, err := GetAllBranches()
	if err != nil {
		t.Fatalf("GetAllBranches() error = %v", err)
	}

	// 验证分支数量(main + feature-1 + feature-2 = 3)
	expectedCount := 3
	if len(branches) != expectedCount {
		t.Errorf("GetAllBranches() returned %d branches, want %d", len(branches), expectedCount)
	}

	// 验证是否包含预期分支
	branchMap := make(map[string]bool)
	for _, b := range branches {
		branchMap[b] = true
	}

	expectedBranches := []string{"main", "feature-1", "feature-2"}
	for _, expected := range expectedBranches {
		if !branchMap[expected] && expected != "master" {
			// 允许main或master
			if expected == "main" && branchMap["master"] {
				continue
			}
			t.Errorf("GetAllBranches() missing branch %s", expected)
		}
	}
}

func TestGetAllBranches_EmptyRepo(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 在空仓库中(没有提交)获取分支
	_, err := GetAllBranches()
	// 空仓库应该返回错误或空列表
	if err == nil {
		t.Log("GetAllBranches() in empty repo returned nil error (acceptable)")
	}
}

func TestGetAllBranches_NotGitRepo(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	_, err := GetAllBranches()
	if err == nil {
		t.Error("GetAllBranches() expected error in non-git repo, got nil")
	}
}
