package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// ─── 列表 ────────────────────────────────────────────────────────────────────

// TestListModifiedTrackedFiles 已跟踪修改:暂存与未暂存均列出,未跟踪不混入
func TestListModifiedTrackedFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	writeFile(t, tempDir, "f.txt", "unstaged\n") // 未暂存
	writeFile(t, tempDir, "g.txt", "staged\n")   // 未跟踪 → 先 add 变成已暂存
	testutil.RunGitCommand(t, tempDir, "add", "g.txt")
	writeFile(t, tempDir, "new.txt", "untracked\n") // 未跟踪

	files, err := listModifiedTrackedFiles()
	if err != nil {
		t.Fatalf("listModifiedTrackedFiles() error = %v", err)
	}
	joined := strings.Join(files, "\n")
	if !strings.Contains(joined, "f.txt") {
		t.Errorf("未暂存修改 f.txt 应在列表中: %v", files)
	}
	if !strings.Contains(joined, "g.txt") {
		t.Errorf("已暂存修改 g.txt 应在列表中: %v", files)
	}
	if strings.Contains(joined, "new.txt") {
		t.Errorf("未跟踪文件不应出现在已跟踪修改列表: %v", files)
	}
}

// TestListUntrackedFiles 未跟踪列表:gitignore 忽略项排除
func TestListUntrackedFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	os.WriteFile(filepath.Join(tempDir, "untracked.txt"), []byte("x"), 0644)
	os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte("debug.log\n"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", ".gitignore")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "gitignore")
	os.WriteFile(filepath.Join(tempDir, "debug.log"), []byte("log"), 0644)

	files, err := listUntrackedFiles()
	if err != nil {
		t.Fatalf("listUntrackedFiles() error = %v", err)
	}
	joined := strings.Join(files, "\n")
	if !strings.Contains(joined, "untracked.txt") {
		t.Errorf("untracked.txt 应在列表中: %v", files)
	}
	if strings.Contains(joined, "debug.log") {
		t.Errorf("gitignore 忽略项不应出现(避免误删): %v", files)
	}
	if strings.Contains(joined, "f.txt") {
		t.Errorf("已跟踪文件不应出现在未跟踪列表: %v", files)
	}
}

// TestListCleanRepoEmpty 干净仓库两个列表均为空
func TestListCleanRepoEmpty(t *testing.T) {
	setupTwoBranchRepo(t)

	files, err := listModifiedTrackedFiles()
	if err != nil || len(files) != 0 {
		t.Errorf("modified = %v, err = %v, want empty", files, err)
	}
	files, err = listUntrackedFiles()
	if err != nil || len(files) != 0 {
		t.Errorf("untracked = %v, err = %v, want empty", files, err)
	}
}

// ─── 核心执行 ────────────────────────────────────────────────────────────────

// TestDiscardTrackedChanges 还原后内容回 HEAD 且暂存清空;新文件场景恢复删除
func TestDiscardTrackedChanges(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// 场景1:修改 + 暂存 → 还原
	writeFile(t, tempDir, "f.txt", "dirty staged\n")
	testutil.RunGitCommand(t, tempDir, "add", "f.txt")

	if err := discardTrackedChanges([]string{"f.txt"}); err != nil {
		t.Fatalf("discardTrackedChanges() error = %v", err)
	}
	if content := readFileContent(t, tempDir, "f.txt"); content != "base\n" {
		t.Errorf("f.txt = %q, want base", content)
	}
	dirty, _ := isWorkingDirectoryDirty()
	if dirty {
		t.Error("还原后工作区应干净(含暂存区)")
	}

	// 场景2:已提交文件的删除 → 还原恢复文件
	os.Remove(filepath.Join(tempDir, "f.txt"))
	if err := discardTrackedChanges([]string{"f.txt"}); err != nil {
		t.Fatalf("discardTrackedChanges(delete) error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "f.txt")); err != nil {
		t.Error("被删除的文件应被还原")
	}
}

func TestDeleteUntrackedFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	os.WriteFile(filepath.Join(tempDir, "junk.txt"), []byte("junk"), 0644)

	if err := deleteUntrackedFiles([]string{"junk.txt"}); err != nil {
		t.Fatalf("deleteUntrackedFiles() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "junk.txt")); !os.IsNotExist(err) {
		t.Error("未跟踪文件应已被删除")
	}
}

// ─── 交互流程(注入菜单) ─────────────────────────────────────────────────────

// TestCleanFlowMixed 混合选择:同时还原修改并删除未跟踪文件
func TestCleanFlowMixed(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "dirty\n")
	os.WriteFile(filepath.Join(tempDir, "junk.txt"), []byte("junk"), 0644)

	restoreSelect := stubMultiSelect(t, &cleanSelect, func(options []config.Option) ([]string, error) {
		// 按 Label 找到对应项的 Value
		var values []string
		for _, o := range options {
			if strings.Contains(o.Label, "f.txt") || strings.Contains(o.Label, "junk.txt") {
				values = append(values, o.Value)
			}
		}
		return values, nil
	})
	defer restoreSelect()

	restoreConfirm := stubBool(t, &cleanConfirm, true)
	defer restoreConfirm()

	if err := Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if content := readFileContent(t, tempDir, "f.txt"); content != "base\n" {
		t.Errorf("f.txt = %q, want 已还原", content)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "junk.txt")); !os.IsNotExist(err) {
		t.Error("junk.txt 应已被删除")
	}
}

// TestCleanFlowCancelled 确认取消时什么都不动
func TestCleanFlowCancelled(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	writeFile(t, tempDir, "f.txt", "keep me\n")

	restoreSelect := stubMultiSelect(t, &cleanSelect, func(options []config.Option) ([]string, error) {
		return []string{options[0].Value}, nil
	})
	defer restoreSelect()
	restoreConfirm := stubBool(t, &cleanConfirm, false)
	defer restoreConfirm()

	if err := Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if content := readFileContent(t, tempDir, "f.txt"); content != "keep me\n" {
		t.Errorf("取消后文件不应变化, got %q", content)
	}
}

// TestCleanFlowEmptyRepoClean 干净仓库直接提示退出
func TestCleanFlowEmptyRepoClean(t *testing.T) {
	setupTwoBranchRepo(t)

	called := false
	oldSelect := cleanSelect
	cleanSelect = func(options []config.Option) ([]string, error) {
		called = true
		return nil, nil
	}
	defer func() { cleanSelect = oldSelect }()

	if err := Clean(); err != nil {
		t.Fatalf("Clean() error = %v", err)
	}
	if called {
		t.Error("干净仓库不应弹出选择列表")
	}
}

// stubMultiSelect 注入多选菜单函数变量(统一签名)
func stubMultiSelect(t *testing.T, target *func([]config.Option) ([]string, error), impl func([]config.Option) ([]string, error)) func() {
	t.Helper()
	old := *target
	*target = impl
	return func() { *target = old }
}
