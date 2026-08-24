package gitcmd

import (
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/testutil"
)

// TestSplitUpstreamRef 上游引用拆分:最长前缀匹配兼容含斜杠 remote 名
func TestSplitUpstreamRef(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	// 注册含斜杠的 remote 与常规 remote
	createBareRemote(t, tempDir) // origin
	testutilRunGit(t, tempDir, "remote", "add", "my/origin", tempDir)

	tests := []struct {
		upstream   string
		wantRemote string
		wantBranch string
	}{
		{"origin/main", "origin", "main"},
		{"my/origin/main", "my/origin", "main"}, // 含斜杠 remote 优先于按首个斜杠拆分
		{"my/origin/feat/a-b", "my/origin", "feat/a-b"},
	}
	for _, tt := range tests {
		remote, branch := splitUpstreamRef(tt.upstream)
		if remote != tt.wantRemote || branch != tt.wantBranch {
			t.Errorf("splitUpstreamRef(%q) = (%q, %q), want (%q, %q)",
				tt.upstream, remote, branch, tt.wantRemote, tt.wantBranch)
		}
	}

	// 无匹配 remote 时回退按首个斜杠拆分(不报错即可)
	remote, branch := splitUpstreamRef("unknown-remote/main")
	if remote != "unknown-remote" || branch != "main" {
		t.Errorf("fallback split = (%q, %q), want (unknown-remote, main)", remote, branch)
	}
}

// TestListStagedFiles 暂存区列表:暂存后可见,还原后清空
func TestListStagedFiles(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	files, err := listStagedFiles()
	if err != nil || len(files) != 0 {
		t.Fatalf("clean staged = %v, err = %v, want empty", files, err)
	}

	writeFile(t, tempDir, "f.txt", "staged content\n")
	testutil.RunGitCommand(t, tempDir, "add", "f.txt")

	files, err = listStagedFiles()
	if err != nil {
		t.Fatalf("listStagedFiles() error = %v", err)
	}
	if len(files) != 1 || files[0] != "f.txt" {
		t.Errorf("staged = %v, want [f.txt]", files)
	}
}

// ─── 交互流程:暂存额外内容防护 ──────────────────────────────────────────────

// TestAmendFlowStagedExtrasAccepted 追加模式下,未勾选的既有暂存文件触发警告,
// 确认后随 amend 一并并入上次提交。
func TestAmendFlowStagedExtrasAccepted(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)

	writeFile(t, tempDir, "chosen.txt", "chosen\n")
	osWriteFileHelper(t, tempDir, "sneaky.txt", "sneaky\n")
	testutilRunGit(t, tempDir, "add", "chosen.txt", "sneaky.txt")

	restoreMenu := stubMenu(t, &amendActionMenu, "files")
	defer restoreMenu()

	oldSelect := amendFilesSelect
	amendFilesSelect = func(options []config.Option) ([]string, error) {
		return []string{"chosen.txt"}, nil // 只勾选一个
	}
	defer func() { amendFilesSelect = oldSelect }()

	var confirmTitles []string
	oldConfirm := amendConfirm
	amendConfirm = func(title string) bool {
		confirmTitles = append(confirmTitles, title)
		return true
	}
	defer func() { amendConfirm = oldConfirm }()

	if isHeadPushed() {
		t.Fatal("此用例要求 HEAD 未推送")
	}
	if err := Amend(); err != nil {
		t.Fatalf("Amend() error = %v", err)
	}

	// 应弹出一次额外暂存内容警告且列出 sneaky.txt
	found := false
	for _, title := range confirmTitles {
		if strings.Contains(title, "sneaky.txt") {
			found = true
		}
	}
	if !found {
		t.Errorf("确认框应列出未勾选的暂存文件 sneaky.txt, got %v", confirmTitles)
	}

	out, _ := execOutput(t, tempDir, "show", "--name-only", "--pretty=format:", "HEAD")
	if !strings.Contains(out, "chosen.txt") || !strings.Contains(out, "sneaky.txt") {
		t.Errorf("确认后两个文件都应并入上次提交:\n%s", out)
	}
}

// TestAmendFlowStagedExtrasCancelled 警告取消时不做任何修改
func TestAmendFlowStagedExtrasCancelled(t *testing.T) {
	tempDir := setupTwoBranchRepo(t)
	oldHash := headHash(t, tempDir)

	writeFile(t, tempDir, "chosen.txt", "chosen\n")
	osWriteFileHelper(t, tempDir, "sneaky.txt", "sneaky\n")
	testutil.RunGitCommand(t, tempDir, "add", "chosen.txt", "sneaky.txt")

	restoreMenu := stubMenu(t, &amendActionMenu, "files")
	defer restoreMenu()

	oldSelect := amendFilesSelect
	amendFilesSelect = func(options []config.Option) ([]string, error) {
		return []string{"chosen.txt"}, nil
	}
	defer func() { amendFilesSelect = oldSelect }()

	calls := 0
	oldConfirm := amendConfirm
	amendConfirm = func(string) bool {
		calls++
		return false // 第二次确认(额外暂存警告)拒绝
	}
	defer func() { amendConfirm = oldConfirm }()

	if err := Amend(); err != nil {
		t.Fatalf("Amend() error = %v", err)
	}
	if headHash(t, tempDir) != oldHash {
		t.Error("取消后 HEAD 不应变化")
	}
	if calls < 1 {
		t.Error("应弹出额外暂存内容确认")
	}
}

func osWriteFileHelper(t *testing.T, dir, name, content string) {
	t.Helper()
	writeFile(t, dir, name, content)
}
