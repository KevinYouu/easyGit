package gitcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

func TestGetCurrentBranch_Merge(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	branch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("getCurrentBranch() error = %v", err)
	}

	if branch != "main" && branch != "master" {
		t.Errorf("getCurrentBranch() = %v, want main or master", branch)
	}
}

func TestGetLocalBranches(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 创建其他分支
	testutil.RunGitCommand(t, tempDir, "branch", "feature-1")
	testutil.RunGitCommand(t, tempDir, "branch", "feature-2")

	currentBranch, _ := getCurrentBranch()
	branches, err := getLocalBranches(currentBranch)
	if err != nil {
		t.Fatalf("getLocalBranches() error = %v", err)
	}

	// 应该返回除当前分支外的所有分支
	if len(branches) != 2 {
		t.Errorf("getLocalBranches() returned %d branches, want 2", len(branches))
	}

	// 验证不包含当前分支
	for _, branch := range branches {
		if branch.Value == currentBranch {
			t.Errorf("getLocalBranches() should not include current branch %s", currentBranch)
		}
	}
}

func TestGetLocalBranches_OnlyCurrentBranch(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	currentBranch, _ := getCurrentBranch()
	branches, err := getLocalBranches(currentBranch)
	if err != nil {
		t.Fatalf("getLocalBranches() error = %v", err)
	}

	// 只有当前分支时应该返回空列表
	if len(branches) != 0 {
		t.Errorf("getLocalBranches() returned %d branches, want 0", len(branches))
	}
}

func TestCheckWorkingDirectoryStatus_Clean(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 工作目录干净时不应该返回错误
	err := checkWorkingDirectoryStatus()
	if err != nil {
		t.Errorf("checkWorkingDirectoryStatus() error = %v, want nil", err)
	}
}

func TestMergeStrategy_DefaultValues(t *testing.T) {
	tests := []struct {
		name     string
		strategy MergeStrategy
		wantArgs []string
	}{
		{
			name:     "default strategy",
			strategy: mergeStrategies[0],
			wantArgs: []string{},
		},
		{
			name:     "ff-only strategy",
			strategy: mergeStrategies[1],
			wantArgs: []string{"--ff-only"},
		},
		{
			name:     "no-ff strategy",
			strategy: mergeStrategies[2],
			wantArgs: []string{"--no-ff"},
		},
		{
			name:     "squash strategy",
			strategy: mergeStrategies[3],
			wantArgs: []string{"--squash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.strategy.Args) != len(tt.wantArgs) {
				t.Errorf("strategy.Args length = %d, want %d", len(tt.strategy.Args), len(tt.wantArgs))
				return
			}
			for i, arg := range tt.strategy.Args {
				if arg != tt.wantArgs[i] {
					t.Errorf("strategy.Args[%d] = %s, want %s", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestGetRemoteBranches_NoRemote(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	currentBranch, _ := getCurrentBranch()
	branches, err := getRemoteBranches(currentBranch)

	// 没有远程仓库时应该返回空列表，可能有错误或没有错误
	if err == nil && len(branches) != 0 {
		t.Errorf("getRemoteBranches() with no remote returned %d branches, want 0", len(branches))
	}
}
