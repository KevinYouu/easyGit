package gitcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

// TestQuoteForEditorEnv 编辑器环境变量路径转义表驱动:
// 仅在路径含空格或反斜杠时加双引号,其余原样返回。
func TestQuoteForEditorEnv(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "Windows 带空格路径",
			in:   `C:\Program Files\easyGit\easyGit.exe`,
			want: `"C:\Program Files\easyGit\easyGit.exe"`,
		},
		{
			name: "Windows 用户目录路径",
			in:   `C:\Users\Kevin\AppData\Local\easyGit\easyGit.exe`,
			want: `"C:\Users\Kevin\AppData\Local\easyGit\easyGit.exe"`,
		},
		{
			name: "Unix 路径无空白",
			in:   "/usr/local/bin/easyGit",
			want: "/usr/local/bin/easyGit",
		},
		{
			name: "纯命令名",
			in:   "easyGit",
			want: "easyGit",
		},
		{
			name: "已含引号不二次包裹",
			in:   `"C:\my path\x.exe"`,
			want: `"C:\my path\x.exe"`,
		},
		{
			name: "仅含反斜杠不加空格也加引号",
			in:   `C:\Users\Kevin\easyGit.exe`,
			want: `"C:\Users\Kevin\easyGit.exe"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := QuoteForEditorEnv(tt.in); got != tt.want {
				t.Errorf("QuoteForEditorEnv(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestRunInternalRebaseEnvInjection 端到端回归:含空格的 Windows 可执行文件路径
// 必须被双引号包裹后注入 GIT_SEQUENCE_EDITOR;合并消息经 GIT_EDITOR(隐藏子命令)
// 与 EASYGIT_COMMIT_MESSAGE 传递,多行正文原样注入不被折叠
func TestRunInternalRebaseEnvInjection(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	for i := range 3 {
		os.WriteFile(filepath.Join(tempDir, "file.txt"), []byte(fmt.Sprintf("content %d", i)), 0644)
		testutil.RunGitCommand(t, tempDir, "add", "file.txt")
		testutil.RunGitCommand(t, tempDir, "commit", "-m", fmt.Sprintf("commit %d", i))
	}
	fakeExe := `C:\Program Files\easyGit\easyGit.exe`
	oldExe := executablePath
	executablePath = func() (string, error) { return fakeExe, nil }
	defer func() { executablePath = oldExe }()

	var gotArgs []string
	var gotEnv []string
	oldRun := runGitCommand
	runGitCommand = func(args, env []string) error {
		gotArgs = append([]string{}, args...)
		gotEnv = append([]string{}, env...)
		return nil
	}
	defer func() { runGitCommand = oldRun }()

	// 多行合并消息:主题 + 空行 + 两段正文
	message := "feat: squash test\n\nfirst body line\nsecond body line"
	if err := RunInternalRebase("HEAD~1", "squash", []string{"abc123", "def456"}, message); err != nil {
		t.Fatalf("RunInternalRebase() error = %v", err)
	}

	// 变基参数:注入消息时必须禁用默认 strip 清理(否则 # 开头的行被删)
	wantArgs := []string{"-c", "commit.cleanup=verbatim", "rebase", "-i", "HEAD~1"}
	if strings.Join(gotArgs, " ") != strings.Join(wantArgs, " ") {
		t.Fatalf("args = %v, want %v", gotArgs, wantArgs)
	}

	env := make(map[string]string)
	for _, kv := range gotEnv {
		if i := strings.Index(kv, "="); i > 0 {
			env[kv[:i]] = kv[i+1:]
		}
	}

	// 序列编辑器:含空格路径必须整体加引号
	wantSeq := `"C:\Program Files\easyGit\easyGit.exe" _internal_rebase_editor`
	if env["GIT_SEQUENCE_EDITOR"] != wantSeq {
		t.Errorf("GIT_SEQUENCE_EDITOR = %q, want %q", env["GIT_SEQUENCE_EDITOR"], wantSeq)
	}

	// 消息编辑器:同一可执行文件内的隐藏子命令,不经 shell 脚本
	wantEditor := `"C:\Program Files\easyGit\easyGit.exe" _internal_commit_message_editor`
	if env["GIT_EDITOR"] != wantEditor {
		t.Errorf("GIT_EDITOR = %q, want %q", env["GIT_EDITOR"], wantEditor)
	}

	// 多行消息必须完整到达消息写入方
	if env["EASYGIT_COMMIT_MESSAGE"] != message {
		t.Errorf("EASYGIT_COMMIT_MESSAGE = %q, want %q", env["EASYGIT_COMMIT_MESSAGE"], message)
	}
}

// TestRebaseArgs 变基参数:注入消息时前置 commit.cleanup=verbatim,
// drop(不注入)与 --root 场景不追加该配置
func TestRebaseArgs(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		injected bool
		want     []string
	}{
		{name: "注入消息", base: "abc123", injected: true, want: []string{"-c", "commit.cleanup=verbatim", "rebase", "-i", "abc123"}},
		{name: "drop 不注入", base: "abc123", injected: false, want: []string{"rebase", "-i", "abc123"}},
		{name: "root 注入", base: "--root", injected: true, want: []string{"-c", "commit.cleanup=verbatim", "rebase", "-i", "--root"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rebaseArgs(tt.base, tt.injected); strings.Join(got, " ") != strings.Join(tt.want, " ") {
				t.Errorf("rebaseArgs(%q, %v) = %v, want %v", tt.base, tt.injected, got, tt.want)
			}
		})
	}
}

// TestRunInternalRebaseDropMode drop 模式不注入消息编辑器与 verbatim 清理
func TestRunInternalRebaseDropMode(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)
	testutil.RunGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "base")

	oldExe := executablePath
	executablePath = func() (string, error) { return "/usr/bin/easyGit", nil }
	defer func() { executablePath = oldExe }()

	var gotArgs []string
	var gotEnv []string
	oldRun := runGitCommand
	runGitCommand = func(args, env []string) error {
		gotArgs = append([]string{}, args...)
		gotEnv = append([]string{}, env...)
		return nil
	}
	defer func() { runGitCommand = oldRun }()

	if err := RunInternalRebase("HEAD~1", "drop", []string{"abc123"}, ""); err != nil {
		t.Fatalf("RunInternalRebase() error = %v", err)
	}

	wantArgs := []string{"rebase", "-i", "HEAD~1"}
	if strings.Join(gotArgs, " ") != strings.Join(wantArgs, " ") {
		t.Errorf("args = %v, want %v", gotArgs, wantArgs)
	}

	env := make(map[string]string)
	for _, kv := range gotEnv {
		if i := strings.Index(kv, "="); i > 0 {
			env[kv[:i]] = kv[i+1:]
		}
	}
	if env["GIT_EDITOR"] != "true" {
		t.Errorf("GIT_EDITOR = %q, want true", env["GIT_EDITOR"])
	}
	if _, ok := env["EASYGIT_COMMIT_MESSAGE"]; ok {
		t.Error("drop 模式不应注入 EASYGIT_COMMIT_MESSAGE")
	}
}

// TestWriteCommitMessageFile 消息文件写入:多行正文/空行/特殊字符原样落盘,
// 尾部规范为单个换行;空消息不改动文件(git 生成的内容保留)
func TestWriteCommitMessageFile(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    string
	}{
		{name: "单行消息", message: "fix: squash commits", want: "fix: squash commits\n"},
		{
			name:    "多行正文插空行",
			message: "feat: title\n\nbody line 1\nbody line 2",
			want:    "feat: title\n\nbody line 1\nbody line 2\n",
		},
		{
			name:    "CRLF 规范为 LF",
			message: "feat: title\r\n\r\nbody",
			want:    "feat: title\n\nbody\n",
		},
		{
			name:    "尾部多余空行收敛",
			message: "feat: title\n\nbody\n\n\n",
			want:    "feat: title\n\nbody\n",
		},
		{
			name:    "特殊字符不经转义",
			message: `feat: "quoted" & <tag> 100% $HOME $(id)`,
			want:    "feat: \"quoted\" & <tag> 100% $HOME $(id)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
			if err := os.WriteFile(path, []byte("placeholder\n"), 0o644); err != nil {
				t.Fatalf("准备消息文件失败: %v", err)
			}

			if err := WriteCommitMessageFile(path, tt.message); err != nil {
				t.Fatalf("WriteCommitMessageFile() error = %v", err)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("读取消息文件失败: %v", err)
			}
			if got := string(content); got != tt.want {
				t.Errorf("消息文件内容 = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("空消息不改动文件", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "COMMIT_EDITMSG")
		if err := os.WriteFile(path, []byte("git generated\n"), 0o644); err != nil {
			t.Fatalf("准备消息文件失败: %v", err)
		}
		if err := WriteCommitMessageFile(path, "\n\n"); err != nil {
			t.Fatalf("WriteCommitMessageFile() error = %v", err)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("读取消息文件失败: %v", err)
		}
		if got := string(content); got != "git generated\n" {
			t.Errorf("空消息不应覆盖文件, got %q", got)
		}
	})
}
