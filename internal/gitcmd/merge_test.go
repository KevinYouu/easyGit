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

	// 验证不包含当前分支,且 label 为 "分支名 (local)" 格式
	for _, branch := range branches {
		if branch.Value == currentBranch {
			t.Errorf("getLocalBranches() should not include current branch %s", currentBranch)
		}
		if branch.Label != branch.Value+" (local)" {
			t.Errorf("getLocalBranches() Label = %q, want %q", branch.Label, branch.Value+" (local)")
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

func TestGetRemoteBranches_Labels(t *testing.T) {
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

	// 创建裸仓库作为远程 origin,推送当前分支与一个特性分支
	bareDir := filepath.Join(t.TempDir(), "remote.git")
	testutil.RunGitCommand(t, tempDir, "init", "--bare", bareDir)
	testutil.RunGitCommand(t, tempDir, "remote", "add", "origin", bareDir)

	currentBranch, _ := getCurrentBranch()
	testutil.RunGitCommand(t, tempDir, "push", "-u", "origin", currentBranch)
	testutil.RunGitCommand(t, tempDir, "branch", "feature-1")
	testutil.RunGitCommand(t, tempDir, "push", "origin", "feature-1")

	branches, err := getRemoteBranches(currentBranch)
	if err != nil {
		t.Fatalf("getRemoteBranches() error = %v", err)
	}

	// 应只返回除当前分支外的远端分支,label 保留 origin/ 前缀并带 (remote) 后缀
	if len(branches) != 1 {
		t.Fatalf("getRemoteBranches() returned %d branches, want 1: %v", len(branches), branches)
	}
	if branches[0].Value != "origin/feature-1" {
		t.Errorf("getRemoteBranches() Value = %q, want %q", branches[0].Value, "origin/feature-1")
	}
	if branches[0].Label != "origin/feature-1 (remote)" {
		t.Errorf("getRemoteBranches() Label = %q, want %q", branches[0].Label, "origin/feature-1 (remote)")
	}
}
