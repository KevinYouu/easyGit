package gitcmd

import (
	"os"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

func TestCherryPickOption_Structure(t *testing.T) {
	// 测试 cherryPickOptions 的基本结构
	if len(cherryPickOptions) == 0 {
		t.Error("cherryPickOptions should not be empty")
	}

	// 检查每个选项是否有必要的字段
	for i, opt := range cherryPickOptions {
		if opt.Name == "" {
			t.Errorf("cherryPickOptions[%d].Name should not be empty", i)
		}
		if opt.NameKey == "" {
			t.Errorf("cherryPickOptions[%d].NameKey should not be empty", i)
		}
		if opt.DescriptionKey == "" {
			t.Errorf("cherryPickOptions[%d].DescriptionKey should not be empty", i)
		}
	}
}

func TestCherryPickOption_DefaultOption(t *testing.T) {
	// 测试默认选项的存在
	var hasDefault bool
	for _, opt := range cherryPickOptions {
		if opt.Name == "default" {
			hasDefault = true
			if len(opt.Args) != 0 {
				t.Error("default option should have empty Args")
			}
			break
		}
	}

	if !hasDefault {
		t.Error("cherryPickOptions should have a 'default' option")
	}
}

func TestCherryPickOption_NoCommitOption(t *testing.T) {
	// 测试 no-commit 选项
	var hasNoCommit bool
	for _, opt := range cherryPickOptions {
		if opt.Name == "no-commit" {
			hasNoCommit = true
			if len(opt.Args) == 0 {
				t.Error("no-commit option should have Args")
			}
			if opt.Args[0] != "--no-commit" {
				t.Errorf("no-commit option Args[0] = %v, want --no-commit", opt.Args[0])
			}
			break
		}
	}

	if !hasNoCommit {
		t.Error("cherryPickOptions should have a 'no-commit' option")
	}
}

func TestGetAllCommitsForCherryPick_NotGitRepo(t *testing.T) {
	// 在非 git 仓库中测试
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	_, err := getAllCommitsForCherryPick()
	if err == nil {
		t.Error("getAllCommitsForCherryPick() should return error in non-git repository")
	}
}

func TestGetAllCommitsForCherryPick_EmptyRepo(t *testing.T) {
	// 在空的 git 仓库中测试（但需要至少一个提交）
	tmpDir := testutil.SetupTestRepo(t)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建一个初始提交，否则空仓库无法执行 git rev-list
	testutil.CreateTestFile(t, tmpDir, "README.md", "test repo")
	testutil.CommitFile(t, tmpDir, "README.md", "Initial commit")

	commits, err := getAllCommitsForCherryPick()
	if err != nil {
		t.Errorf("getAllCommitsForCherryPick() error = %v", err)
	}

	// 只有一个分支，没有其他可 cherry-pick 的提交
	if len(commits) != 0 {
		t.Errorf("getAllCommitsForCherryPick() returned %d commits, want 0 with single branch", len(commits))
	}
}

func TestGetAllCommitsForCherryPick_WithCommits(t *testing.T) {
	// 创建带有提交的测试仓库
	tmpDir := testutil.SetupTestRepo(t)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 获取当前分支名（可能是 master 或 main）
	currentBranch := testutil.GetCurrentBranch(t, tmpDir)

	// 创建初始提交
	testutil.CreateTestFile(t, tmpDir, "test.txt", "initial content")
	testutil.CommitFile(t, tmpDir, "test.txt", "feat: initial commit")

	// 创建新分支并添加提交
	testutil.CreateBranch(t, tmpDir, "feature-branch")
	testutil.CreateTestFile(t, tmpDir, "test.txt", "feature content")
	testutil.CommitFile(t, tmpDir, "test.txt", "feat: feature commit")

	// 切换回主分支
	testutil.CheckoutBranch(t, tmpDir, currentBranch)

	// 获取可 cherry-pick 的提交
	commits, err := getAllCommitsForCherryPick()
	if err != nil {
		t.Fatalf("getAllCommitsForCherryPick() error = %v", err)
	}

	// 应该能找到 feature-branch 上的提交
	if len(commits) == 0 {
		t.Error("getAllCommitsForCherryPick() should find commits from feature-branch")
	}

	// 验证提交包含预期的信息
	foundFeatureCommit := false
	for _, commit := range commits {
		if commit.Message != "" && commit.Hash != "" {
			foundFeatureCommit = true
			break
		}
	}
	if !foundFeatureCommit {
		t.Error("getAllCommitsForCherryPick() should return valid commit information")
	}
}

func TestExecuteCherryPick_ValidOption(t *testing.T) {
	// 测试 executeCherryPick 使用有效的选项
	tmpDir := testutil.SetupTestRepo(t)

	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建初始提交
	testutil.CreateTestFile(t, tmpDir, "test.txt", "initial")
	testutil.CommitFile(t, tmpDir, "test.txt", "feat: initial")

	// 创建另一个提交
	testutil.CreateTestFile(t, tmpDir, "test.txt", "second")
	commitHash := testutil.CommitFile(t, tmpDir, "test.txt", "feat: second commit")

	// 重置到第一个提交
	testutil.RunGitCommand(t, tmpDir, "reset", "--hard", "HEAD~1")

	// 使用 default 选项 cherry-pick 第二个提交
	commit := Commit{
		Hash:    commitHash,
		Message: "feat: second commit",
	}

	defaultOption := cherryPickOptions[0] // default option
	err := executeCherryPick(commit, defaultOption)

	if err != nil {
		t.Errorf("executeCherryPick() error = %v, want nil", err)
	}
}
