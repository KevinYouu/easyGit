package gitcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

// TestHasPendingChanges 待提交改动的判定必须与 git add -A 的实际结果一致:
// 未跟踪文件不受 status.showUntrackedFiles=no 影响,
// 子模块仅内容脏时不算改动(否则 add -A 无内容可提交,commit 必然失败)
func TestHasPendingChanges(t *testing.T) {
	dir, restore := chdirTempRepo(t)
	defer restore()

	expect := func(want bool, reason string) {
		t.Helper()
		got, err := hasPendingChanges()
		if err != nil {
			t.Fatalf("hasPendingChanges() error = %v", err)
		}
		if got != want {
			t.Errorf("%s: hasPendingChanges() = %v, want %v", reason, got, want)
		}
	}

	writeFile(t, dir, "f.txt", "base\n")
	testutil.RunGitCommand(t, dir, "add", ".")
	testutil.RunGitCommand(t, dir, "commit", "-q", "-m", "base")
	expect(false, "干净工作区")

	writeFile(t, dir, "f.txt", "changed\n")
	expect(true, "修改已跟踪文件")
	testutil.RunGitCommand(t, dir, "commit", "-q", "-am", "changed")

	testutil.RunGitCommand(t, dir, "config", "status.showUntrackedFiles", "no")
	writeFile(t, dir, "untracked.txt", "new\n")
	expect(true, "未跟踪文件(配置隐藏)")
	os.Remove(filepath.Join(dir, "untracked.txt"))
	testutil.RunGitCommand(t, dir, "config", "--unset", "status.showUntrackedFiles")

	subSource := t.TempDir()
	testutil.RunGitCommand(t, subSource, "init", "-q")
	testutil.RunGitCommand(t, subSource, "config", "user.name", "Test User")
	testutil.RunGitCommand(t, subSource, "config", "user.email", "test@example.com")
	writeFile(t, subSource, "sub.txt", "sub\n")
	testutil.RunGitCommand(t, subSource, "add", ".")
	testutil.RunGitCommand(t, subSource, "commit", "-q", "-m", "sub")
	testutil.RunGitCommand(t, dir, "-c", "protocol.file.allow=always", "submodule", "add", "-q", subSource, "sm")
	testutil.RunGitCommand(t, dir, "commit", "-q", "-m", "add submodule")

	writeFile(t, filepath.Join(dir, "sm"), "sub.txt", "dirty\n")
	expect(false, "子模块仅内容脏")

	testutil.RunGitCommand(t, filepath.Join(dir, "sm"), "commit", "-q", "-am", "next")
	expect(true, "子模块指针移动")
}
