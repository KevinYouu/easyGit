package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// ─── 名称校验 ────────────────────────────────────────────────────────────────

func TestValidateBranchName(t *testing.T) {
	valid := []string{"feat", "feat/x", "feat-x_1", "v1.0.0", "a/b/c"}
	for _, name := range valid {
		if err := validateBranchName(name); err != nil {
			t.Errorf("validateBranchName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",          // 空
		"   ",       // 纯空白
		"feat x",    // 空格
		"-feat",     // 以 - 开头(被解析为选项)
		".hidden",   // 以 . 开头
		"feat.",     // 以 . 结尾
		"a..b",      // 路径穿越
		"feat.lock", // .lock 结尾
		"feat~1",    // ~
		"feat^",     // ^
		"a:b",       // :
		"a?b",       // ?
		"a*b",       // *
		"a[b]",      // [
		"a\\b",      // 反斜杠
		"/abs",      // 以 / 开头
		"ends/",     // 以 / 结尾
		"a//b",      // 连续 /
	}
	for _, name := range invalid {
		if err := validateBranchName(name); err == nil {
			t.Errorf("validateBranchName(%q) = nil, want error", name)
		}
	}
}

// ─── 核心创建执行 ────────────────────────────────────────────────────────────

// TestCreateBranchFromHead 从当前 HEAD 创建并切换到新分支
func TestCreateBranchFromHead(t *testing.T) {
	tempDir := setupTwoBranchRepo(t) // 当前在 main

	if err := createBranch("feat-new", "", false, nil); err != nil {
		t.Fatalf("createBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "feat-new" {
		t.Errorf("current branch = %q, want feat-new", branch)
	}
	// 与 main 同一提交
	headMain, _ := execOutput(t, tempDir, "rev-parse", "main")
	headNew, _ := execOutput(t, tempDir, "rev-parse", "feat-new")
	if headMain != headNew {
		t.Errorf("feat-new (%s) 应与 main (%s) 指向同一提交", headNew, headMain)
	}
}

// TestCreateBranchFromBase 从指定基点(本地分支)创建
func TestCreateBranchFromBase(t *testing.T) {
	tempDir := setupTwoBranchRepo(t) // main + other

	// other 上多一个提交,基点应为 other 的头
	testutil.RunGitCommand(t, tempDir, "checkout", "other")
	writeFile(t, tempDir, "g.txt", "g\n")
	testutil.RunGitCommand(t, tempDir, "add", ".")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "other extra")
	testutil.RunGitCommand(t, tempDir, "checkout", "main")

	if err := createBranch("from-other", "other", false, nil); err != nil {
		t.Fatalf("createBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "from-other" {
		t.Errorf("current branch = %q, want from-other", branch)
	}
	if content := readFileContent(t, tempDir, "g.txt"); !strings.Contains(content, "g\n") {
		t.Errorf("基点应为 other 分支头,g.txt = %q", content)
	}
}

// TestCreateBranchFromRemoteBase 远程引用作为基点
func TestCreateBranchFromRemoteBase(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)
	testutil.RunGitCommand(t, tempDir, "push", "-q", "origin", "other")
	testutil.RunGitCommand(t, tempDir, "branch", "-D", "other") // 仅保留远程引用

	if err := createBranch("from-remote", "origin/other", false, nil); err != nil {
		t.Fatalf("createBranch() error = %v", err)
	}
	if up := upstreamOf(t, tempDir); up == "" {
		t.Log("无上游为预期(仅作基点,未设跟踪)")
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "from-remote" {
		t.Errorf("current branch = %q, want from-remote", branch)
	}
}

// TestCreateBranchPushUpstream 推送新分支并设置 upstream(同步注入执行器)
func TestCreateBranchPushUpstream(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)

	oldRunner := branchCreatePushRunner
	branchCreatePushRunner = func(cmds []command.CommandInfo) error {
		for _, c := range cmds {
			if out, err := exec.Command(c.Command, c.Args...).CombinedOutput(); err != nil {
				return fmt.Errorf("git %v: %v\n%s", c.Args, err, out)
			}
		}
		return nil
	}
	defer func() { branchCreatePushRunner = oldRunner }()

	if err := createBranch("shared", "", true, []string{"origin"}); err != nil {
		t.Fatalf("createBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "shared" {
		t.Errorf("current branch = %q, want shared", branch)
	}
	// 远程已存在该分支
	out, _ := execOutput(t, tempDir, "ls-remote", "--heads", "origin", "shared")
	if !strings.Contains(out, "refs/heads/shared") {
		t.Errorf("远程缺少 shared 分支:\n%s", out)
	}
	if up := upstreamOf(t, tempDir); up != "origin/shared" {
		t.Errorf("upstream = %q, want origin/shared", up)
	}
}

// TestCreateBranchDuplicateName 重名分支报错且不切换
func TestCreateBranchDuplicateName(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	if err := createBranch("other", "", false, nil); err == nil {
		t.Error("重名分支应返回错误")
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "main" {
		t.Errorf("失败后应留在 main, got %q", branch)
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestCreateBranchFlowWithPush 完整流程:输入名称 → 基点 HEAD → 确认推送 → upstream 生效
func TestCreateBranchFlowWithPush(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	createBareRemote(t, tempDir)

	restoreInput := stubString(t, &branchCreateNameInput, func(validate func(string) error) (string, error) {
		if err := validate("flow-branch"); err != nil {
			return "", err
		}
		return "flow-branch", nil
	})
	defer restoreInput()

	restoreBase := stubBaseSelect(t, "")
	defer restoreBase()

	restoreConfirm := stubBool(t, &branchCreatePushConfirm, true)
	defer restoreConfirm()

	restoreRemotes := stubRemotes(t, []string{"origin"})
	defer restoreRemotes()

	// 测试环境无 TTY:注入同步执行器真实跑 git push(生产走并行进度模型)
	oldRunner := branchCreatePushRunner
	branchCreatePushRunner = func(cmds []command.CommandInfo) error {
		for _, c := range cmds {
			if out, err := exec.Command(c.Command, c.Args...).CombinedOutput(); err != nil {
				return fmt.Errorf("git %v: %v\n%s", c.Args, err, out)
			}
		}
		return nil
	}
	defer func() { branchCreatePushRunner = oldRunner }()

	if err := CreateBranch(); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "flow-branch" {
		t.Errorf("current branch = %q, want flow-branch", branch)
	}
	if up := upstreamOf(t, tempDir); up != "origin/flow-branch" {
		t.Errorf("upstream = %q, want origin/flow-branch", up)
	}
}

// TestCreateBranchFlowNoPush 不推送时仅本地创建
func TestCreateBranchFlowNoPush(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	restoreInput := stubString(t, &branchCreateNameInput, func(validate func(string) error) (string, error) {
		return "local-only", nil
	})
	defer restoreInput()

	restoreBase := stubBaseSelect(t, "")
	defer restoreBase()

	restoreConfirm := stubBool(t, &branchCreatePushConfirm, false)
	defer restoreConfirm()

	if err := CreateBranch(); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "local-only" {
		t.Errorf("current branch = %q, want local-only", branch)
	}
	out, _ := execOutput(t, tempDir, "branch", "--list", "local-only")
	if !strings.Contains(out, "local-only") {
		t.Errorf("本地分支应已创建:\n%s", out)
	}
}

// TestCreateBranchFlowInvalidName 非法名称被表单校验拦截(注入层模拟真实路径)
func TestCreateBranchFlowInvalidName(t *testing.T) {
	setupTwoBranchRepo(t)

	var capturedErr error
	restoreInput := stubString(t, &branchCreateNameInput, func(validate func(string) error) (string, error) {
		capturedErr = validate("bad name")
		return "", capturedErr
	})
	defer restoreInput()

	// 注入层直接返回错误 → CreateBranch 取消退出
	if err := CreateBranch(); err != nil {
		t.Fatalf("取消不应报错, got %v", err)
	}
	if capturedErr == nil {
		t.Error("非法名称应触发校验错误")
	}
}

// ─── 注入辅助 ────────────────────────────────────────────────────────────────

// stubString 注入函数变量(签名 func(func(string) error) (string, error))并返回恢复函数
func stubString(t *testing.T, target *func(validate func(string) error) (string, error), impl func(validate func(string) error) (string, error)) func() {
	t.Helper()
	old := *target
	*target = impl
	return func() { *target = old }
}

// stubBaseSelect 注入基点选择结果
func stubBaseSelect(t *testing.T, value string) func() {
	t.Helper()
	old := branchCreateBaseSelect
	branchCreateBaseSelect = func(options []config.Option) (string, error) {
		return value, nil
	}
	return func() { branchCreateBaseSelect = old }
}

// stubBool 注入布尔确认
func stubBool(t *testing.T, target *func(string) bool, result bool) func() {
	t.Helper()
	old := *target
	*target = func(string) bool { return result }
	return func() { *target = old }
}

// stubRemotes 注入远程选择结果
func stubRemotes(t *testing.T, remotes []string) func() {
	t.Helper()
	old := branchCreateRemotesSelect
	branchCreateRemotesSelect = func() ([]string, error) { return remotes, nil }
	return func() { branchCreateRemotesSelect = old }
}
