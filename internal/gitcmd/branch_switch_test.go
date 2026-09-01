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

// ─── 测试辅助 ────────────────────────────────────────────────────────────────

// chdirTempRepo 切换到临时仓库目录,返回恢复函数(gitcmd 函数在当前目录执行 git)。
// 显式指定初始分支 main,不依赖用户环境 init.defaultBranch 配置。
func chdirTempRepo(t *testing.T) (string, func()) {
	t.Helper()
	tempDir := t.TempDir()
	cmd := exec.Command("git", "init", "--initial-branch=main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init --initial-branch=main failed: %v", err)
	}
	testutil.RunGitCommand(t, tempDir, "config", "user.name", "Test User")
	testutil.RunGitCommand(t, tempDir, "config", "user.email", "test@example.com")
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	return tempDir, func() { os.Chdir(originalDir) }
}

// createBareRemote 创建裸仓库并注册为 origin 远程
func createBareRemote(t *testing.T, repoDir string) string {
	t.Helper()
	bareDir := filepath.Join(t.TempDir(), "origin.git")
	cmd := exec.Command("git", "init", "--bare", bareDir)
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init --bare failed: %v", err)
	}
	testutil.RunGitCommand(t, repoDir, "remote", "add", "origin", bareDir)
	return bareDir
}

// stashCount 统计当前 stash 条目数
func stashCount(t *testing.T, dir string) int {
	t.Helper()
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git stash list failed: %v", err)
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return 0
	}
	return len(strings.Split(trimmed, "\n"))
}

// upstreamOf 返回分支上游(无上游返回空串)
func upstreamOf(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	cmd.Dir = dir
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output))
}

// execOutput 执行 git 命令返回输出(失败不中断,由调用方断言)
func execOutput(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

// setupTwoBranchRepo 构造 main(含 f.txt)+ other 双分支仓库,当前在 main
func setupTwoBranchRepo(t *testing.T) string {
	t.Helper()
	tempDir, _ := chdirTempRepo(t)
	writeFile(t, tempDir, "f.txt", "base\n")
	testutil.RunGitCommand(t, tempDir, "add", ".")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "base")
	testutil.RunGitCommand(t, tempDir, "checkout", "-b", "other")
	testutil.RunGitCommand(t, tempDir, "checkout", "main")
	return tempDir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s failed: %v", name, err)
	}
}

func readFileContent(t *testing.T, dir, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatalf("read %s failed: %v", name, err)
	}
	return string(content)
}

// ─── 纯函数单元测试 ──────────────────────────────────────────────────────────

func TestLocalNameForRemoteBranch(t *testing.T) {
	tests := []struct {
		ref  string
		want string
	}{
		{"origin/dev", "dev"},
		{"origin/feat/a-b", "feat/a-b"},
		{"upstream/main", "main"},
		{"main", ""},    // 无 remote 前缀结构
		{"origin/", ""}, // 空分支名
		{"", ""},        // 空引用
	}
	for _, tt := range tests {
		if got := localNameForRemoteBranch(tt.ref); got != tt.want {
			t.Errorf("localNameForRemoteBranch(%q) = %q, want %q", tt.ref, got, tt.want)
		}
	}
}

func TestIsWorkingDirectoryDirty(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	dirty, err := isWorkingDirectoryDirty()
	if err != nil || dirty {
		t.Fatalf("clean repo: dirty = %v, err = %v, want false, nil", dirty, err)
	}

	writeFile(t, tempDir, "f.txt", "base\nmodified\n")
	untracked := filepath.Join(tempDir, "new.txt")
	os.WriteFile(untracked, []byte("x"), 0644)

	dirty, err = isWorkingDirectoryDirty()
	if err != nil || !dirty {
		t.Fatalf("dirty repo: dirty = %v, err = %v, want true, nil", dirty, err)
	}
}

// ─── 分支列表 ────────────────────────────────────────────────────────────────

// TestListSwitchableBranches 列表规则:排除当前分支;本地在前;
// 与本地同名的远程项去重;远程独有项保留且标记 IsRemote。
func TestListSwitchableBranches(t *testing.T) {
	tempDir := setupTwoBranchRepo(t) // main(current) + other
	testutil.RunGitCommand(t, tempDir, "branch", "feat")
	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "other")            // 与本地重名 → 应被去重
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "main:remote-only") // 仅远程存在 → 应保留

	branches, err := listSwitchableBranches()
	if err != nil {
		t.Fatalf("listSwitchableBranches() error = %v", err)
	}

	byRef := make(map[string]SwitchableBranch, len(branches))
	for _, b := range branches {
		if strings.Contains(b.Ref, "HEAD") {
			t.Errorf("列表不应包含 HEAD 引用: %+v", b)
		}
		if b.Ref == "main" {
			t.Errorf("列表不应包含当前分支 main")
		}
		byRef[b.Ref] = b
	}

	// 本地:other / feat
	for _, local := range []string{"other", "feat"} {
		b, ok := byRef[local]
		if !ok {
			t.Errorf("缺少本地分支 %s, got %+v", local, branches)
			continue
		}
		if b.IsRemote {
			t.Errorf("%s 应为本地分支", local)
		}
		if b.Name != local {
			t.Errorf("Name = %q, want %q", b.Name, local)
		}
	}

	// 重名去重:origin(other) 不应出现
	if _, ok := byRef["origin/other"]; ok {
		t.Error("同名本地分支已存在时,origin/other 不应重复出现在列表中")
	}

	// 远程独有:origin/remote-only
	b, ok := byRef["origin/remote-only"]
	if !ok {
		t.Fatalf("缺少远程分支 origin/remote-only, got %+v", branches)
	}
	if !b.IsRemote || b.Name != "remote-only" {
		t.Errorf("origin/remote-only 应为 IsRemote=true Name=remote-only, got %+v", b)
	}
}

// TestListSwitchableBranchesEmpty 单分支仓库无可切换分支时报错
func TestListSwitchableBranchesEmpty(t *testing.T) {
	tempDir, restore := chdirTempRepo(t)
	defer restore()

	testutil.CreateTestFile(t, tempDir, "a.txt", "a")
	testutil.RunGitCommand(t, tempDir, "add", ".")
	testutil.CommitFile(t, tempDir, "a.txt", "initial")

	if _, err := listSwitchableBranches(); err == nil {
		t.Error("单分支仓库应返回错误")
	}
}

// ─── 核心切换执行 ────────────────────────────────────────────────────────────

func TestSwitchToBranchLocalBranch(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	err := switchToBranch(SwitchableBranch{Name: "other", Ref: "other"}, false)
	if err != nil {
		t.Fatalf("switchToBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
}

// TestSwitchToBranchRemoteCreatesTracking 远程分支切换自动创建跟踪分支并设 upstream
func TestSwitchToBranchRemoteCreatesTracking(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "other")
	testutil.RunGitCommand(t, tempDir, "branch", "-D", "other") // 删除本地强制走远程路径

	err := switchToBranch(SwitchableBranch{Name: "other", Ref: "origin/other", IsRemote: true}, false)
	if err != nil {
		t.Fatalf("switchToBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
	if up := upstreamOf(t, tempDir); up != "origin/other" {
		t.Errorf("upstream = %q, want origin/other", up)
	}
}

// TestSwitchToBranchAutoStashRoundtrip 自动 stash 流程:
// 未提交修改先入 stash,切换后 pop 回新分支,stash 列表清空。
func TestSwitchToBranchAutoStashRoundtrip(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "base\ndirty\n")

	err := switchToBranch(SwitchableBranch{Name: "other", Ref: "other"}, true)
	if err != nil {
		t.Fatalf("switchToBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
	// 修改跟随到新分支
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "dirty") {
		t.Errorf("f.txt = %q, want 包含 dirty(stash 已恢复)", content)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0(pop 成功清空)", n)
	}
}

// TestSwitchToBranchStashConflictKeepsEntry pop 冲突场景:
// 新分支上同文件内容不同导致冲突,stash 条目必须保留供手动恢复,
// 且切换本身视为成功(err == nil)。
func TestSwitchToBranchStashConflictKeepsEntry(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// other 分支提交与脏修改冲突的内容
	testutil.RunGitCommand(t, tempDir, "checkout", "other")
	writeFile(t, tempDir, "f.txt", "other-content\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "other change")
	testutil.RunGitCommand(t, tempDir, "checkout", "main")

	// main 上未提交修改(与 other 的已提交内容冲突)
	writeFile(t, tempDir, "f.txt", "dirty-main\n")

	err := switchToBranch(SwitchableBranch{Name: "other", Ref: "other"}, true)
	if err != nil {
		t.Fatalf("切换本身应成功, error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("stash count = %d, want 1(冲突时保留条目)", n)
	}
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "<<<<<<<") {
		t.Errorf("f.txt = %q, want 含冲突标记", content)
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestSwitchBranchFlowCarry 脏工作区选择「携带修改」→ 正常切换
func TestSwitchBranchFlowCarry(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "base\nkeep\n")

	restoreDirty := withDirtyAction(t, "carry")
	restoreSelect := withSwitchSelect(t, "other")
	defer restoreDirty()
	defer restoreSelect()

	if err := SwitchBranch(); err != nil {
		t.Fatalf("SwitchBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
	// 携带模式:修改直接带到新分支
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "keep") {
		t.Errorf("f.txt = %q, want 包含 keep", content)
	}
}

// TestSwitchBranchFlowCancel 脏工作区选择「取消」→ 不切换
func TestSwitchBranchFlowCancel(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "base\nkeep\n")

	restoreDirty := withDirtyAction(t, "cancel")
	defer restoreDirty()

	if err := SwitchBranch(); err != nil {
		t.Fatalf("SwitchBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "main" {
		t.Errorf("取消后仍应在 main, got %q", branch)
	}
}

// TestSwitchBranchCleanRepo 干净仓库跳过脏区询问直达分支选择
func TestSwitchBranchCleanRepo(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	dirtyAsked := false
	oldAction := branchSwitchDirtyAction
	branchSwitchDirtyAction = func(options []config.Option) (string, error) {
		dirtyAsked = true
		return "", nil
	}
	defer func() { branchSwitchDirtyAction = oldAction }()

	restoreSelect := withSwitchSelect(t, "other")
	defer restoreSelect()

	if err := SwitchBranch(); err != nil {
		t.Fatalf("SwitchBranch() error = %v", err)
	}
	if dirtyAsked {
		t.Error("干净仓库不应询问脏区处理")
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "other" {
		t.Errorf("current branch = %q, want other", branch)
	}
}

// withDirtyAction 注入脏工作区处理选择
func withDirtyAction(t *testing.T, action string) func() {
	t.Helper()
	old := branchSwitchDirtyAction
	branchSwitchDirtyAction = func(options []config.Option) (string, error) {
		return action, nil
	}
	return func() { branchSwitchDirtyAction = old }
}

// withSwitchSelect 注入分支选择结果(ref 值)
func withSwitchSelect(t *testing.T, ref string) func() {
	t.Helper()
	old := branchSwitchSelect
	branchSwitchSelect = func(options []config.Option) (string, error) {
		return ref, nil
	}
	return func() { branchSwitchSelect = old }
}
