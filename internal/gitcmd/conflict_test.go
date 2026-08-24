package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// ─── 测试仓库构造 ────────────────────────────────────────────────────────────

// setupMergeConflictRepo 构造合并必冲突仓库:
// base 后 feat 与 main 各改 f.txt,当前在 main。
func setupMergeConflictRepo(t *testing.T) string {
	t.Helper()
	tempDir := setupTwoBranchRepo(t) // main + other,f.txt=base

	// other 分支提交冲突改动
	testutilCheckout(t, tempDir, "other")
	writeFile(t, tempDir, "f.txt", "other content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "other change")
	testutilCheckout(t, tempDir, "main")

	// main 分支提交冲突改动
	writeFile(t, tempDir, "f.txt", "main content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "main change")
	return tempDir
}

// setupCherryPickConflictRepo 构造摘取必冲突仓库,返回待摘取的冲突提交哈希
func setupCherryPickConflictRepo(t *testing.T) (string, string) {
	t.Helper()
	tempDir := setupTwoBranchRepo(t)

	testutilCheckout(t, tempDir, "other")
	writeFile(t, tempDir, "f.txt", "side content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "side change")
	sideHash, _ := execOutput(t, tempDir, "rev-parse", "other")

	testutilCheckout(t, tempDir, "main")
	writeFile(t, tempDir, "f.txt", "main content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "main change")
	return tempDir, sideHash
}

func testutilCheckout(t *testing.T, dir, branch string) {
	t.Helper()
	testutilRunGit(t, dir, "checkout", branch)
}

func testutilRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if _, err := execOutput(t, dir, args...); err != nil {
		t.Fatalf("git %v failed: %v", args, err)
	}
}

// ─── 状态检测 ────────────────────────────────────────────────────────────────

func TestMergeAndCherryPickInProgress(t *testing.T) {
	tempDir, sideHash := setupCherryPickConflictRepo(t)

	if isMergeInProgress() || isCherryPickInProgress() {
		t.Fatal("干净仓库不应处于合并/摘取状态")
	}

	runGitAllowFail(t, tempDir, "cherry-pick", sideHash)
	if !isCherryPickInProgress() {
		t.Error("摘取冲突后应处于进行中状态")
	}
	runGitAllowFail(t, tempDir, "cherry-pick", "--abort")

	runGitAllowFail(t, tempDir, "merge", "other")
	if !isMergeInProgress() {
		t.Error("合并冲突后应处于进行中状态")
	}
}

// ─── 合并冲突闭环 ────────────────────────────────────────────────────────────

// TestMergeConflictLoopResolved 核心回归:合并冲突进入闭环,
// 解决文件后 continue 完成合并(生成合并提交),不再直接退出报错。
func TestMergeConflictLoopResolved(t *testing.T) {
	tempDir := setupMergeConflictRepo(t)

	out := runGitAllowFail(t, tempDir, "merge", "other", "--no-ff")
	if !strings.Contains(out, "CONFLICT") {
		t.Fatalf("预置场景应产生冲突:\n%s", out)
	}
	if !isMergeInProgress() {
		t.Fatal("合并应处于进行中状态")
	}

	menuCalls := 0
	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		menuCalls++
		// 解决文件后继续
		os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("merged content\n"), 0644)
		return "continue", nil
	})()

	aborted, err := runConflictResolution(mergeConflictOps)
	if err != nil {
		t.Fatalf("runConflictResolution() error = %v", err)
	}
	if aborted {
		t.Error("用户未中止,aborted 应为 false")
	}
	if menuCalls != 1 {
		t.Errorf("menu called %d times, want 1", menuCalls)
	}
	if isMergeInProgress() {
		t.Error("合并应已完成")
	}
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "merged content") {
		t.Errorf("f.txt = %q, want merged content", content)
	}
	// --no-ff 产生合并提交
	out, _ = execOutput(t, tempDir, "log", "--oneline", "-1")
	if !strings.Contains(out, "Merge") && !strings.Contains(out, "merge") {
		t.Errorf("应产生合并提交:\n%s", out)
	}
}

// TestMergeConflictLoopAbort 中止后恢复合并前状态
func TestMergeConflictLoopAbort(t *testing.T) {
	tempDir := setupMergeConflictRepo(t)

	runGitAllowFail(t, tempDir, "merge", "other", "--no-ff")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "abort", nil
	})()

	aborted, err := runConflictResolution(mergeConflictOps)
	if err != nil {
		t.Fatalf("runConflictResolution() error = %v", err)
	}
	if !aborted {
		t.Error("中止后 aborted 应为 true")
	}
	if isMergeInProgress() {
		t.Error("中止后不应处于合并状态")
	}
	if content := readFileContent(t, tempDir, "f.txt"); !strings.Contains(content, "main content") {
		t.Errorf("f.txt = %q, want 恢复为 main content", content)
	}
}

// TestMergeConflictNoSkipOption merge 无 skip 能力:菜单选项不含 skip 值
func TestMergeConflictNoSkipOption(t *testing.T) {
	tempDir := setupMergeConflictRepo(t)
	runGitAllowFail(t, tempDir, "merge", "other", "--no-ff")

	var seenValues []string
	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		for _, o := range options {
			seenValues = append(seenValues, o.Value)
		}
		return "abort", nil
	})()

	runConflictResolution(mergeConflictOps)
	for _, v := range seenValues {
		if v == "skip" {
			t.Error("merge 冲突菜单不应包含 skip 选项")
		}
	}
}

// ─── 摘取冲突闭环 ────────────────────────────────────────────────────────────

// TestCherryPickConflictLoopSkip 摘取冲突选择跳过:提交不落地且状态清理
func TestCherryPickConflictLoopSkip(t *testing.T) {
	tempDir, sideHash := setupCherryPickConflictRepo(t)

	out := runGitAllowFail(t, tempDir, "cherry-pick", sideHash)
	if !strings.Contains(out, "CONFLICT") {
		t.Fatalf("预置场景应产生冲突:\n%s", out)
	}

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "skip", nil
	})()

	aborted, err := runConflictResolution(cherryPickConflictOps)
	if err != nil {
		t.Fatalf("runConflictResolution() error = %v", err)
	}
	if aborted {
		t.Error("skip 不是 abort")
	}
	if isCherryPickInProgress() {
		t.Error("skip 后不应处于摘取状态")
	}
	// 提交数应保持不变(skip 放弃该提交)
	count, _ := execOutput(t, tempDir, "rev-list", "--count", "HEAD")
	if strings.TrimSpace(count) != "2" { // base + main change
		t.Errorf("提交数 = %s, want 2(side change 被跳过)", count)
	}
}

// TestExecuteCherryPickBatchSequential 批次顺序摘取回归:
// 每个提交独立调用 git cherry-pick <hash>,闭环 --continue 只完成当前提交,
// 后续提交必须继续正常摘取(不得跳过)。
// 批次 [side1(撞冲突→解决), side2(g.txt 干净应用)] 两个提交均应落地。
func TestExecuteCherryPickBatchSequential(t *testing.T) {
	// 构造:base(f.txt/g.txt)→ other 上两次提交(f.txt 与 g.txt 各一),
	// main 独立修改 f.txt 制造首提交冲突
	tempDir := setupTwoBranchRepo(t)
	testutilCheckout(t, tempDir, "other")
	writeFile(t, tempDir, "f.txt", "side f\n")
	testutilRunGit(t, tempDir, "commit", "-am", "side one")
	writeFile(t, tempDir, "g.txt", "side g\n")
	testutil.RunGitCommand(t, tempDir, "add", "g.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "side two")
	side2, _ := execOutput(t, tempDir, "rev-parse", "other")
	side1, _ := execOutput(t, tempDir, "rev-parse", "other~1")

	testutilCheckout(t, tempDir, "main")
	writeFile(t, tempDir, "f.txt", "main conflicting\n")
	testutilRunGit(t, tempDir, "commit", "-am", "main change")

	menuCalls := 0
	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		menuCalls++
		if menuCalls > 1 {
			return "abort", nil // 防御:正常路径不应二次进入菜单
		}
		writeFile(t, tempDir, "f.txt", "resolved\n")
		return "continue", nil
	})()

	// 第一个提交:冲突 → 闭环解决完成
	if err := executeCherryPick(Commit{Hash: strings.TrimSpace(side1)}, cherryPickOptions[0]); err != nil {
		t.Fatalf("first pick error = %v", err)
	}
	if menuCalls != 1 || isCherryPickInProgress() {
		t.Fatalf("menuCalls=%d inProgress=%v, want 单次闭环且状态结束", menuCalls, isCherryPickInProgress())
	}

	// 第二个提交:必须继续正常摘取(证明无 sequencer 跨调用遗留)
	if err := executeCherryPick(Commit{Hash: strings.TrimSpace(side2)}, cherryPickOptions[0]); err != nil {
		t.Fatalf("second pick error = %v", err)
	}

	out, _ := execOutput(t, tempDir, "log", "--oneline")
	if !strings.Contains(out, "side one") || !strings.Contains(out, "side two") {
		t.Errorf("批次两个提交均应落地:\n%s", out)
	}
	if content := readFileContent(t, tempDir, "g.txt"); !strings.Contains(content, "side g") {
		t.Errorf("g.txt = %q, want 含 side g", content)
	}
}

// TestExecuteCherryPickEmptyCommitSkips 内容重复时警告并跳过而非中断批次:
// 冲突解决方案已包含后续提交改动时,该提交变 empty,按 already-applied 同等处理。
func TestExecuteCherryPickEmptyCommitSkips(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// other 分支提交 f.txt 改动
	testutilCheckout(t, tempDir, "other")
	writeFile(t, tempDir, "f.txt", "same content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "dup change")
	dupHash, _ := execOutput(t, tempDir, "rev-parse", "other")

	testutilCheckout(t, tempDir, "main")
	// main 上做出完全相同的内容改动 → 对方摘取必然 empty
	writeFile(t, tempDir, "f.txt", "same content\n")
	testutilRunGit(t, tempDir, "commit", "-am", "identical change")

	menuCalled := false
	oldMenu := conflictMenu
	conflictMenu = func(title string, options []config.Option) (string, error) {
		menuCalled = true
		return "", nil
	}
	defer func() { conflictMenu = oldMenu }()

	err := executeCherryPick(Commit{Hash: strings.TrimSpace(dupHash)}, cherryPickOptions[0])
	if err != nil {
		t.Fatalf("empty commit 应警告跳过而非报错, got %v", err)
	}
	if menuCalled {
		t.Error("empty commit 不应进入冲突闭环")
	}
}

// TestExecuteCherryPickConflictResolved 验证 executeCherryPick 冲突路径接入闭环:
// 解决后返回 viaLoop=true 且 err 为 nil。
func TestExecuteCherryPickConflictResolved(t *testing.T) {
	tempDir, sideHash := setupCherryPickConflictRepo(t)

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("resolved pick\n"), 0644)
		return "continue", nil
	})()

	err := executeCherryPick(Commit{Hash: strings.TrimSpace(sideHash)}, cherryPickOptions[0])
	if err != nil {
		t.Fatalf("executeCherryPick() error = %v", err)
	}
	if isCherryPickInProgress() {
		t.Error("解决后不应处于摘取状态")
	}
	out, _ := execOutput(t, tempDir, "log", "--oneline", "-1")
	if !strings.Contains(out, "side change") {
		t.Errorf("HEAD 应为摘取的提交:\n%s", out)
	}
}

// TestExecuteCherryPickNormalPathNoLoop 常规无冲突路径 viaLoop=false
func TestExecuteCherryPickNormalPathNoLoop(t *testing.T) {
	tempDir, sideHash := setupCherryPickConflictRepo(t)

	// 还原 main 对 f.txt 的独立修改,使摘取无冲突
	testutilRunGit(t, tempDir, "reset", "--hard", "HEAD~1")

	menuCalled := false
	oldMenu := conflictMenu
	conflictMenu = func(title string, options []config.Option) (string, error) {
		menuCalled = true
		return "", nil
	}
	defer func() { conflictMenu = oldMenu }()

	err := executeCherryPick(Commit{Hash: strings.TrimSpace(sideHash)}, cherryPickOptions[0])
	if err != nil {
		t.Fatalf("executeCherryPick() error = %v", err)
	}
	if menuCalled {
		t.Error("无冲突不应进入冲突闭环")
	}
}

// TestExecuteCherryPickConflictAborted 中止时返回 errConflictAborted
func TestExecuteCherryPickConflictAborted(t *testing.T) {
	tempDir, sideHash := setupCherryPickConflictRepo(t)
	_ = tempDir // 仅需仓库环境;中止断言不依赖路径

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "abort", nil
	})()

	err := executeCherryPick(Commit{Hash: strings.TrimSpace(sideHash)}, cherryPickOptions[0])
	if !isErrConflictAborted(err) {
		t.Errorf("err = %v, want errConflictAborted", err)
	}
	if isCherryPickInProgress() {
		t.Error("中止后不应处于摘取状态")
	}
}

func isErrConflictAborted(err error) bool {
	type causer interface{ Unwrap() error }
	for err != nil {
		if err == errConflictAborted {
			return true
		}
		c, ok := err.(causer)
		if !ok {
			return false
		}
		err = c.Unwrap()
	}
	return false
}
