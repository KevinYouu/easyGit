package gitcmd

import (
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// ─── 列表解析 ────────────────────────────────────────────────────────────────

func TestListStashesEmpty(t *testing.T) {
	setupTwoBranchRepo(t)

	entries, err := listStashes()
	if err != nil {
		t.Fatalf("listStashes() error = %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("entries = %v, want empty", entries)
	}
}

// TestListStashesParsing 解析字段:引用/相对时间/消息;最新在最前
func TestListStashesParsing(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	writeFile(t, tempDir, "f.txt", "change one\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push", "-m", "wip one")
	writeFile(t, tempDir, "f.txt", "change two\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push", "-m", "wip two")

	entries, err := listStashes()
	if err != nil {
		t.Fatalf("listStashes() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	// 最新在前
	if entries[0].Index != "stash@{0}" || !strings.Contains(entries[0].Message, "wip two") {
		t.Errorf("entries[0] = %+v, want stash@{0} 含 wip two", entries[0])
	}
	if entries[1].Index != "stash@{1}" || !strings.Contains(entries[1].Message, "wip one") {
		t.Errorf("entries[1] = %+v, want stash@{1} 含 wip one", entries[1])
	}
	if entries[0].Date == "" {
		t.Error("Date 字段不应为空")
	}
}

// ─── 核心执行 ────────────────────────────────────────────────────────────────

func TestSaveStashCleanRepo(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// 干净仓库:提示跳过,不产生 stash
	if err := saveStash("should skip"); err != nil {
		t.Fatalf("saveStash() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0", n)
	}
}

// TestSaveStashWithMessage 带消息保存:工作区恢复干净 + 消息可检索
func TestSaveStashWithMessage(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "wip content\n")

	if err := saveStash("my custom message"); err != nil {
		t.Fatalf("saveStash() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Fatalf("stash count = %d, want 1", n)
	}
	dirty, _ := isWorkingDirectoryDirty()
	if dirty {
		t.Error("stash 后工作区应干净")
	}
	out, _ := execOutput(t, tempDir, "stash", "list")
	if !strings.Contains(out, "my custom message") {
		t.Errorf("stash list 应含自定义消息:\n%s", out)
	}
}

// TestApplyStashKeepAndPop apply 保留条目 / pop 删除条目 行为差异
func TestApplyStashKeepAndPop(t *testing.T) {
	// apply:恢复后保留条目
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "apply me\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	if err := applyStash(StashEntry{Index: "stash@{0}"}, false); err != nil {
		t.Fatalf("applyStash(keep) error = %v", err)
	}
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "apply me") {
		t.Errorf("f.txt = %q, want 已恢复", content)
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("apply 后 stash count = %d, want 1(条目保留)", n)
	}

	// pop:恢复后删除条目
	testutil.RunGitCommand(t, tempDir, "checkout", "--", "f.txt")
	writeFile(t, tempDir, "f.txt", "pop me\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	if err := applyStash(StashEntry{Index: "stash@{0}"}, true); err != nil {
		t.Fatalf("applyStash(pop) error = %v", err)
	}
	// 第一步 apply 保留的条目仍在,仅 pop 的那条被删除
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("pop 后 stash count = %d, want 1(apply 保留的条目)", n)
	}
}

// TestApplyStashConflictPop 冲突时 pop 条目被 git 自动保留,并返回错误
func TestApplyStashConflictPop(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	writeFile(t, tempDir, "f.txt", "stashed change\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")
	// 同文件制造不同已提交内容 → pop 必然冲突
	writeFile(t, tempDir, "f.txt", "conflicting committed content\n")

	err := applyStash(StashEntry{Index: "stash@{0}"}, true)
	if err == nil {
		t.Fatal("pop 冲突应返回错误")
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("冲突后 stash count = %d, want 1(git 自动保留)", n)
	}
}

func TestDropStash(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "to drop\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	if err := dropStash(StashEntry{Index: "stash@{0}"}); err != nil {
		t.Fatalf("dropStash() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0", n)
	}
}

func TestClearStashes(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	for range 3 {
		writeFile(t, tempDir, "f.txt", "change\n")
		testutil.RunGitCommand(t, tempDir, "stash", "push", "-m", "batch")
		testutil.RunGitCommand(t, tempDir, "checkout", "--", "f.txt")
	}

	if err := clearStashes(); err != nil {
		t.Fatalf("clearStashes() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0", n)
	}
}

// TestShowStashDiff 仅展示不影响状态
func TestShowStashDiff(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "diff content marker\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	if err := showStashDiff(StashEntry{Index: "stash@{0}"}); err != nil {
		t.Fatalf("showStashDiff() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("show 不应影响 stash 数量, got %d", n)
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestStashFlowSave 主菜单选保存 → 注入可选消息 → 入栈
func TestStashFlowSave(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "flow save\n")

	restoreMenu := stubMenu(t, &stashMainMenu, "save")
	defer restoreMenu()
	restoreInput := stubStashInput(t, "flow message")
	defer restoreInput()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
	out, _ := execOutput(t, tempDir, "stash", "list")
	if !strings.Contains(out, "flow message") {
		t.Errorf("stash list 应含 flow message:\n%s", out)
	}
}

// TestStashFlowManagePop 管理 → 选条目 → 应用并删除
func TestStashFlowManagePop(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "manage pop\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	restoreMenu := stubMenu(t, &stashMainMenu, "manage")
	defer restoreMenu()
	restoreSelect := stubMenu(t, &stashSelectEntry, "stash@{0}")
	defer restoreSelect()
	restoreAction := stubMenu(t, &stashActionMenu, "pop")
	defer restoreAction()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "manage pop") {
		t.Errorf("f.txt = %q, want 已恢复", content)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0", n)
	}
}

// TestStashFlowManageEsc 管理列表 Esc 返回(nil,不报错)
func TestStashFlowManageEsc(t *testing.T) {
	setupTwoBranchRepo(t)

	restoreMenu := stubMenuErr(t, &stashMainMenu, "manage")
	defer restoreMenu()

	// 条目选择层 Esc(stubMenu 的 Err 变体在下方)
	oldSelect := stashSelectEntry
	stashSelectEntry = func(options []config.Option) (string, error) { return "", errUserAbortedStub }
	defer func() { stashSelectEntry = oldSelect }()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
}

var errUserAbortedStub = &abortError{}

type abortError struct{}

func (*abortError) Error() string { return "aborted" }

// TestStashFlowClearWithConfirm 清空强确认通过 → 全部删除
func TestStashFlowClearWithConfirm(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	for range 2 {
		writeFile(t, tempDir, "f.txt", "x\n")
		testutil.RunGitCommand(t, tempDir, "stash", "push")
		testutil.RunGitCommand(t, tempDir, "checkout", "--", "f.txt")
	}

	restoreMenu := stubMenu(t, &stashMainMenu, "clear")
	defer restoreMenu()
	restoreConfirm := stubStashConfirm(t, true)
	defer restoreConfirm()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 0 {
		t.Errorf("stash count = %d, want 0", n)
	}
}

// TestStashFlowClearCancelled 确认取消时不清空
func TestStashFlowClearCancelled(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "keep\n")
	testutil.RunGitCommand(t, tempDir, "stash", "push")

	restoreMenu := stubMenu(t, &stashMainMenu, "clear")
	defer restoreMenu()
	restoreConfirm := stubStashConfirm(t, false)
	defer restoreConfirm()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
	if n := stashCount(t, tempDir); n != 1 {
		t.Errorf("取消清空后 stash count = %d, want 1", n)
	}
}

// TestStashFlowClearEmpty 清空入口在空列表时直接提示跳过
func TestStashFlowClearEmpty(t *testing.T) {
	setupTwoBranchRepo(t)

	restoreMenu := stubMenu(t, &stashMainMenu, "clear")
	defer restoreMenu()

	if err := Stash(); err != nil {
		t.Fatalf("Stash() error = %v", err)
	}
}

// ─── 注入辅助 ────────────────────────────────────────────────────────────────

// stubMenu 注入菜单函数变量(统一签名)并返回恢复函数
func stubMenu(t *testing.T, target *func([]config.Option) (string, error), result string) func() {
	t.Helper()
	old := *target
	*target = func([]config.Option) (string, error) { return result, nil }
	return func() { *target = old }
}

// stubMenuErr 注入返回错误的菜单(Esc 场景)
func stubMenuErr(t *testing.T, target *func([]config.Option) (string, error), _ string) func() {
	t.Helper()
	old := *target
	*target = func([]config.Option) (string, error) { return "", errUserAbortedStub }
	return func() { *target = old }
}

// stubStashInput 注入可选消息输入结果
func stubStashInput(t *testing.T, value string) func() {
	t.Helper()
	old := stashMessageInput
	stashMessageInput = func() (string, error) { return value, nil }
	return func() { stashMessageInput = old }
}

// stubStashConfirm 注入确认结果
func stubStashConfirm(t *testing.T, result bool) func() {
	t.Helper()
	old := stashConfirm
	stashConfirm = func(string) bool { return result }
	return func() { stashConfirm = old }
}
