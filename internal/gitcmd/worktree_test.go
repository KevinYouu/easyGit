package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// setupWorktreeRepo 构造主仓库(main + feat 分支)
func setupWorktreeRepo(t *testing.T) string {
	t.Helper()
	return setupTwoBranchRepo(t) // main(当前)+ other
}

// ─── 列表解析 ────────────────────────────────────────────────────────────────

func TestListWorktreesParsing(t *testing.T) {
	tempDir := setupWorktreeRepo(t)

	wtPath := filepath.Join(t.TempDir(), "wt-other")
	testutilRunGit(t, tempDir, "worktree", "add", wtPath, "other")
	// macOS 下 /var 为 /private/var 符号链接,git 输出真实路径
	if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
		wtPath = resolved
	}

	infos, err := listWorktrees()
	if err != nil {
		t.Fatalf("listWorktrees() error = %v", err)
	}
	if len(infos) != 2 {
		t.Fatalf("infos = %d, want 2", len(infos))
	}

	// 首条 = 主工作树
	main := infos[0]
	if !main.IsMain || !strings.HasSuffix(filepath.Clean(main.Path), filepath.Base(tempDir)) && main.Path != tempDir {
		t.Errorf("首条应为主工作树: %+v", main)
	}
	if main.Branch != "main" || !strings.Contains(main.Head, "") && len(main.Head) != 7 {
		t.Errorf("主工作树字段异常: %+v", main)
	}

	// 次条 = 新增工作树
	added := infos[1]
	if added.IsMain {
		t.Error("第二条不应标记为主工作树")
	}
	if filepath.Clean(added.Path) != filepath.Clean(wtPath) {
		t.Errorf("path = %q, want %q", added.Path, wtPath)
	}
	if added.Branch != "other" {
		t.Errorf("branch = %q, want other", added.Branch)
	}
	if len(added.Head) != 7 {
		t.Errorf("head 应为 7 位短哈希, got %q", added.Head)
	}
}

// TestWorktreesInUse 占用集合:两个工作树各占一个分支
func TestWorktreesInUse(t *testing.T) {
	tempDir := setupWorktreeRepo(t)

	wtPath := filepath.Join(t.TempDir(), "wt")
	testutilRunGit(t, tempDir, "worktree", "add", wtPath, "other")

	inUse := worktreesInUse()
	if !inUse["main"] || !inUse["other"] {
		t.Errorf("inUse = %v, want 含 main 与 other", inUse)
	}
	if inUse["not-exist"] {
		t.Error("不存在的分支不应被误报占用")
	}
}

// ─── 核心执行 ────────────────────────────────────────────────────────────────

// TestWorktreeAddExistingBranch 以已有分支创建工作树:目录生成、分支被检出、主仓不受影响
func TestWorktreeAddExistingBranch(t *testing.T) {
	tempDir := setupWorktreeRepo(t)
	wtPath := filepath.Join(t.TempDir(), "wt-existing")

	if err := worktreeAdd(wtPath, "other", false); err != nil {
		t.Fatalf("worktreeAdd() error = %v", err)
	}
	if _, err := os.Stat(wtPath); err != nil {
		t.Fatalf("工作树目录未创建: %v", err)
	}
	// 工作树检出 other 分支
	out, _ := execOutput(t, wtPath, "branch", "--show-current")
	if strings.TrimSpace(out) != "other" {
		t.Errorf("工作树分支 = %q, want other", out)
	}
	// 主仓仍在 main
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "main" {
		t.Errorf("主仓分支 = %q, want main", branch)
	}
}

// TestWorktreeAddNewBranch -b 新建分支并创建工作树
func TestWorktreeAddNewBranch(t *testing.T) {
	tempDir := setupWorktreeRepo(t)
	wtPath := filepath.Join(t.TempDir(), "wt-new")

	if err := worktreeAdd(wtPath, "brand-new", true); err != nil {
		t.Fatalf("worktreeAdd() error = %v", err)
	}
	out, _ := execOutput(t, wtPath, "branch", "--show-current")
	if strings.TrimSpace(out) != "brand-new" {
		t.Errorf("工作树分支 = %q, want brand-new", out)
	}
	// 新分支存在于分支列表
	out, _ = execOutput(t, tempDir, "branch", "--list", "brand-new")
	if !strings.Contains(out, "brand-new") {
		t.Errorf("新分支应在主仓可见:\n%s", out)
	}
}

// TestWorktreeAddDuplicatePath 已存在目录报错且不影响主仓
func TestWorktreeAddDuplicatePath(t *testing.T) {
	tempDir := setupWorktreeRepo(t)
	existing := filepath.Join(t.TempDir(), "taken")
	os.MkdirAll(existing, 0755)
	testutil.CreateTestFile(t, existing, "occupied.txt", "x")

	if err := worktreeAdd(existing, "other", false); err == nil {
		t.Error("非空已存在路径应报错")
	}
	if branch := testutil.GetCurrentBranch(t, tempDir); branch != "main" {
		t.Errorf("失败后主仓应留在 main, got %q", branch)
	}
}

// TestWorktreeRemovePaths 删除后目录消失、主工作树保留;占用解除
func TestWorktreeRemovePaths(t *testing.T) {
	tempDir := setupWorktreeRepo(t)
	wt1 := filepath.Join(t.TempDir(), "wt1")
	wt2 := filepath.Join(t.TempDir(), "wt2")
	testutilRunGit(t, tempDir, "worktree", "add", wt1, "other")

	// 第二个工作树用新建分支
	testutilRunGit(t, tempDir, "worktree", "add", "-b", "extra", wt2)

	if err := worktreeRemovePaths([]string{wt1, wt2}); err != nil {
		t.Fatalf("worktreeRemovePaths() error = %v", err)
	}
	for _, p := range []string{wt1, wt2} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s 应已被移除", p)
		}
	}
	// 占用解除后 other 可再次使用
	inUse := worktreesInUse()
	if inUse["other"] || inUse["extra"] {
		t.Errorf("删除后分支占用应解除: %v", inUse)
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestWorktreeFlowAddNew 主菜单 add → 路径 → 新分支模式 → 输入名称 → 创建成功
func TestWorktreeFlowAddNew(t *testing.T) {
	setupWorktreeRepo(t)
	wtPath := filepath.Join(t.TempDir(), "flow-wt")

	restoreMenu := stubMenu(t, &worktreeMenu, "add")
	defer restoreMenu()
	restorePath := stubFunc(t, &worktreePathInput, func() (string, error) { return wtPath, nil })
	defer restorePath()
	restoreMode := stubMenu(t, &worktreeModeSelect, "new")
	defer restoreMode()
	restoreBranch := stubFunc(t, &worktreeBranchInput, func(validate func(string) error) (string, error) {
		if err := validate("flow-branch"); err != nil {
			return "", err
		}
		return "flow-branch", nil
	})
	defer restoreBranch()

	// 注入循环菜单:第二次返回 Esc 退出
	calls := 0
	oldMenu := worktreeMenu
	worktreeMenu = func(options []config.Option) (string, error) {
		calls++
		if calls == 1 {
			return "add", nil
		}
		return "", errUserAbortedStub
	}
	defer func() { worktreeMenu = oldMenu }()

	if err := Worktree(); err != nil {
		t.Fatalf("Worktree() error = %v", err)
	}
	out, _ := execOutput(t, wtPath, "branch", "--show-current")
	if strings.TrimSpace(out) != "flow-branch" {
		t.Errorf("工作树分支 = %q, want flow-branch", out)
	}
}

// TestWorktreeFlowRemove 主菜单 remove → 多选 → 确认(含路径清单)→ 移除成功
func TestWorktreeFlowRemove(t *testing.T) {
	tempDir := setupWorktreeRepo(t)
	wtPath := filepath.Join(t.TempDir(), "flow-rm")
	testutilRunGit(t, tempDir, "worktree", "add", wtPath, "other")

	// macOS 符号链接归一(git porcelain 输出真实路径)
	if resolved, err := filepath.EvalSymlinks(wtPath); err == nil {
		wtPath = resolved
	}

	restoreMenu := stubMenu(t, &worktreeMenu, "remove")
	defer restoreMenu()

	restoreSelect := stubMultiSelect(t, &worktreeSelectRemove, func(options []config.Option) ([]string, error) {
		var values []string
		for _, o := range options {
			if strings.Contains(o.Label, filepath.Base(wtPath)) {
				values = append(values, o.Value)
			}
		}
		return values, nil
	})
	defer restoreSelect()

	var confirmMsg string
	oldConfirm := worktreeConfirm
	worktreeConfirm = func(title string) bool {
		confirmMsg = title
		return true
	}
	defer func() { worktreeConfirm = oldConfirm }()

	if err := removeWorktreeFlow(); err != nil {
		t.Fatalf("removeWorktreeFlow() error = %v", err)
	}
	if !strings.Contains(confirmMsg, wtPath) {
		t.Errorf("删除确认应包含完整路径清单 %q, got:\n%s", wtPath, confirmMsg)
	}
	if _, err := os.Stat(wtPath); !os.IsNotExist(err) {
		t.Error("工作树应已被移除")
	}
}

// TestWorktreeFlowListAndEsc list 展示后回菜单,Esc 正常退出
func TestWorktreeFlowListAndEsc(t *testing.T) {
	setupWorktreeRepo(t)

	calls := 0
	oldMenu := worktreeMenu
	worktreeMenu = func(options []config.Option) (string, error) {
		calls++
		if calls == 1 {
			return "list", nil
		}
		return "", errUserAbortedStub
	}
	defer func() { worktreeMenu = oldMenu }()

	if err := Worktree(); err != nil {
		t.Fatalf("Worktree() error = %v", err)
	}
	if calls != 2 {
		t.Errorf("menu called %d times, want 2(list 后回菜单再退出)", calls)
	}
}
