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

// TestGetRecentCommitsKeepsFullMessage 长提交消息必须完整保留在标签中:
// squash 等命令从标签提取默认消息,截断会导致获取到残缺文本
func TestGetRecentCommitsKeepsFullMessage(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录(GetRecentCommits 在当前目录执行 git)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 超过 50 显示宽度的提交消息(中文+emoji+ASCII)
	longMessage := "这是一个非常长的提交消息用于验证完整保留不被截断🚀以及更多的说明文字 xyz1234567890"
	os.WriteFile(testFile, []byte("update"), 0644)
	testutil.RunGitCommand(t, tempDir, "commit", "-am", longMessage)

	options, hashes, err := GetRecentCommits()
	if err != nil {
		t.Fatalf("GetRecentCommits() error = %v", err)
	}
	if len(hashes) < 2 {
		t.Fatalf("commits = %d, want >= 2", len(hashes))
	}

	// 最新提交在最前,其完整消息必须出现在标签中,不得被省略
	if !strings.Contains(options[0].Label, longMessage) {
		t.Errorf("标签中消息被截断:\n got  %q\n want 包含 %q", options[0].Label, longMessage)
	}
}

// ─── 冲突解决闭环测试 ────────────────────────────────────────────────────────

// setupRebaseConflictRepo 构造双冲突仓库:
// feat(f1 改 f.txt / f2 改 g.txt) 与 main(m1 改 f.txt / m2 改 g.txt) 各两个提交,
// rebase 会依次在 f.txt、g.txt 上冲突两次。
func setupRebaseConflictRepo(t *testing.T) (tempDir string, restore func()) {
	t.Helper()

	tempDir, _ = testutil.CreateTempGitRepo(t)
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	base := func(name, content string) {
		os.WriteFile(filepath.Join(tempDir, name), []byte(content), 0644)
	}
	base("f.txt", "base\n")
	base("g.txt", "base\n")
	testutil.RunGitCommand(t, tempDir, "add", ".")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "base")

	current, _ := getCurrentBranch()
	testutil.RunGitCommand(t, tempDir, "checkout", "-b", "feat")
	base("f.txt", "base\nfeat f\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "feat1")
	base("g.txt", "base\nfeat g\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "feat2")

	testutil.RunGitCommand(t, tempDir, "checkout", current)
	base("f.txt", "base\nmain f\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "main1")
	base("g.txt", "base\nmain g\n")
	testutil.RunGitCommand(t, tempDir, "commit", "-am", "main2")

	return tempDir, func() { os.Chdir(originalDir) }
}

// runGitAllowFail 执行 git 命令,失败不中断测试(用于触发冲突)
func runGitAllowFail(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, _ := cmd.CombinedOutput()
	return string(output)
}

// withConflictMenu 注入假菜单并返回恢复函数(通用冲突闭环菜单)
func withConflictMenu(t *testing.T, menu func(options []config.Option) (string, error)) func() {
	t.Helper()
	old := conflictMenu
	conflictMenu = func(title string, options []config.Option) (string, error) {
		return menu(options)
	}
	return func() { conflictMenu = old }
}

func TestIsRebaseInProgress(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	if isRebaseInProgress() {
		t.Error("isRebaseInProgress() = true, want false before rebase")
	}

	runGitAllowFail(t, tempDir, "rebase", "feat")
	if !isRebaseInProgress() {
		t.Error("isRebaseInProgress() = false, want true after conflict")
	}
}

func TestGetUnmergedFiles(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")
	files := getUnmergedFiles()
	if len(files) != 1 || files[0] != "f.txt" {
		t.Errorf("getUnmergedFiles() = %v, want [f.txt]", files)
	}
}

// TestHandleRebaseConflictMultiConflictLoop 核心回归:两次冲突自动循环,
// 第一次 continue 后不退出,第二次解决后变基完成。
func TestHandleRebaseConflictMultiConflictLoop(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")
	if !isRebaseInProgress() {
		t.Fatal("rebase should be in progress after first conflict")
	}

	menuCalls := 0
	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		menuCalls++
		switch menuCalls {
		case 1: // 第一次冲突:解决 f.txt 后继续
			os.WriteFile(filepath.Join(tempDir, "f.txt"), []byte("base\nresolved f\n"), 0644)
			return "continue", nil
		case 2: // 第二次冲突:解决 g.txt 后继续
			os.WriteFile(filepath.Join(tempDir, "g.txt"), []byte("base\nresolved g\n"), 0644)
			return "continue", nil
		}
		return "abort", nil
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if menuCalls != 2 {
		t.Errorf("menu called %d times, want 2 (multi-conflict loop)", menuCalls)
	}
	if isRebaseInProgress() {
		t.Error("rebase should be completed after resolving both conflicts")
	}

	// 变基结果验证:两个提交都在,内容为解决后的值
	out := runGitAllowFail(t, tempDir, "log", "--oneline")
	if !strings.Contains(out, "main2") || !strings.Contains(out, "main1") {
		t.Errorf("rebase 结果缺少提交:\n%s", out)
	}
	content, _ := os.ReadFile(filepath.Join(tempDir, "f.txt"))
	if !strings.Contains(string(content), "resolved f") {
		t.Errorf("f.txt 内容 = %q, want 含 resolved f", string(content))
	}
}

func TestHandleRebaseConflictSkip(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "skip", nil // 跳过 main1
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if isRebaseInProgress() {
		t.Error("rebase should complete after skip")
	}
	out := runGitAllowFail(t, tempDir, "log", "--oneline")
	if strings.Contains(out, "main1") {
		t.Errorf("main1 应被跳过:\n%s", out)
	}
}

func TestHandleRebaseConflictAbort(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "abort", nil
	})()

	err := handleRebaseConflict()
	if err != nil {
		t.Fatalf("handleRebaseConflict() error = %v", err)
	}
	if isRebaseInProgress() {
		t.Error("rebase should be aborted")
	}
}

func TestHandleRebaseConflictQuit(t *testing.T) {
	tempDir, restore := setupRebaseConflictRepo(t)
	defer restore()

	runGitAllowFail(t, tempDir, "rebase", "feat")

	defer withConflictMenu(t, func(options []config.Option) (string, error) {
		return "quit", nil
	})()

	err := handleRebaseConflict()
	if err == nil {
		t.Fatal("handleRebaseConflict() should return error on quit")
	}
	if !isRebaseInProgress() {
		t.Error("rebase should remain pending after quit")
	}
}

// ─── 冲突编辑器解析测试 ──────────────────────────────────────────────────────

func TestResolveConflictEditor(t *testing.T) {
	files := []string{"f.txt", "g.txt"}

	tests := []struct {
		name     string
		editor   string
		wantProg string
		wantArgs []string
		perFile  bool
	}{
		{
			name:     "vim 直通",
			editor:   "vim",
			wantProg: "vim",
			wantArgs: []string{"f.txt", "g.txt"},
		},
		{
			name:     "code 自动补 -w",
			editor:   "code",
			wantProg: "code",
			wantArgs: []string{"-w", "f.txt", "g.txt"},
		},
		{
			name:     "code 已带 -w 不重复",
			editor:   "code -w",
			wantProg: "code",
			wantArgs: []string{"-w", "f.txt", "g.txt"},
		},
		{
			name:     "code --wait 不重复",
			editor:   "code --wait",
			wantProg: "code",
			wantArgs: []string{"--wait", "f.txt", "g.txt"},
		},
		{
			name:     "带参数自定义命令",
			editor:   "code -r --new-window",
			wantProg: "code",
			wantArgs: []string{"-r", "--new-window", "-w", "f.txt", "g.txt"},
		},
		{
			name:     "subl 自动补等待",
			editor:   "subl",
			wantProg: "subl",
			wantArgs: []string{"-w", "f.txt", "g.txt"},
		},
		{
			name:     "Windows 引号路径识别",
			editor:   `"C:\Program Files\Microsoft VS Code\code.exe"`,
			wantProg: `C:\Program Files\Microsoft VS Code\code.exe`,
			wantArgs: []string{"-w", "f.txt", "g.txt"},
		},
		{
			name:     "Windows 引号路径带参数",
			editor:   `"C:\Program Files\Sublime Text\subl.exe" --wait`,
			wantProg: `C:\Program Files\Sublime Text\subl.exe`,
			wantArgs: []string{"--wait", "f.txt", "g.txt"},
		},
		{
			name:     "notepad 逐文件标志",
			editor:   "notepad",
			wantProg: "notepad",
			wantArgs: []string{"f.txt", "g.txt"},
			perFile:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, args, perFile := resolveConflictEditor(tt.editor, files)
			if prog != tt.wantProg {
				t.Errorf("program = %q, want %q", prog, tt.wantProg)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("args = %v, want %v", args, tt.wantArgs)
			}
			for i := range args {
				if args[i] != tt.wantArgs[i] {
					t.Fatalf("args = %v, want %v", args, tt.wantArgs)
				}
			}
			if perFile != tt.perFile {
				t.Errorf("perFile = %v, want %v", perFile, tt.perFile)
			}
		})
	}
}

func TestHasWaitFlag(t *testing.T) {
	if !hasWaitFlag([]string{"-w"}) || !hasWaitFlag([]string{"--wait"}) || !hasWaitFlag([]string{"--wait-for-input"}) {
		t.Error("hasWaitFlag 应识别全部等待标志")
	}
	if hasWaitFlag([]string{"-r", "--reuse-window"}) {
		t.Error("hasWaitFlag 不应误判非等待标志")
	}
}

// TestResolveConflictEditorPerFileNotepad 确认 notepad 在非 Windows 平台也标记逐文件
func TestResolveConflictEditorPerFileNotepad(t *testing.T) {
	prog, _, perFile := resolveConflictEditor("notepad", []string{"a.txt"})
	if prog != "notepad" || !perFile {
		t.Errorf("notepad 应标记 perFile, got prog=%q perFile=%v", prog, perFile)
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{"简单命令", "vim", []string{"vim"}},
		{"带参数", "code -w", []string{"code", "-w"}},
		{"Windows 引号路径", `"C:\Program Files\code.exe" -w`, []string{`C:\Program Files\code.exe`, "-w"}},
		{"多空格", "code   -w  --new-window", []string{"code", "-w", "--new-window"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitCommand(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("splitCommand(%q) = %v, want %v", tt.in, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("splitCommand(%q) = %v, want %v", tt.in, got, tt.want)
				}
			}
		})
	}
}

func TestHasUnresolvedMarkers(t *testing.T) {
	tempDir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(tempDir, name)
		os.WriteFile(p, []byte(content), 0644)
		return p
	}

	// 含冲突标记的文件
	unresolved := write("unresolved.txt", "line1\n<<<<<<< HEAD\nours\n=======\ntheirs\n>>>>>>> feat\n")
	// 已解决文件(仅一个标记字面量不作为冲突判定)
	resolved := write("resolved.txt", "line1\nline2\n")
	// 含 "=======" 分隔线但不含 <<<<<<< / >>>>>>> 的文件不应误判
	separator := write("separator.txt", "header\n=======\nbody\n")

	if !hasUnresolvedMarkers([]string{unresolved}) {
		t.Error("hasUnresolvedMarkers 应检测到含 <<<<<<< 的文件")
	}
	if hasUnresolvedMarkers([]string{resolved, separator}) {
		t.Error("hasUnresolvedMarkers 不应误判已解决文件或含 ======= 分隔线的文件")
	}
	if !hasUnresolvedMarkers([]string{resolved, unresolved}) {
		t.Error("多个文件中任一个含冲突标记即应判定未解决")
	}
	if hasUnresolvedMarkers([]string{filepath.Join(tempDir, "not-exist.txt")}) {
		t.Error("读取失败的文件不应判定为未解决")
	}
}
