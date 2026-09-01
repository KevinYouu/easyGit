package gitcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
// 必须被双引号包裹后注入 GIT_SEQUENCE_EDITOR,消息脚本路径按 QuoteForEditorEnv 注入
func TestRunInternalRebaseEnvInjection(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	for i := 0; i < 3; i++ {
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

	if err := RunInternalRebase("HEAD~1", "squash", []string{"abc123", "def456"}, "fix: squash test"); err != nil {
		t.Fatalf("RunInternalRebase() error = %v", err)
	}

	// 变基参数
	if len(gotArgs) < 2 || gotArgs[0] != "rebase" || gotArgs[1] != "-i" {
		t.Fatalf("args = %v, want [rebase -i ...]", gotArgs)
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

	// 消息编辑器:脚本路径注入必须与 QuoteForEditorEnv 规则一致
	editor := env["GIT_EDITOR"]
	scriptPath := strings.Trim(editor, `"`)
	if editor != QuoteForEditorEnv(scriptPath) {
		t.Errorf("GIT_EDITOR 未按 QuoteForEditorEnv 注入: got %q, want %q", editor, QuoteForEditorEnv(scriptPath))
	}
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(scriptPath, ".bat") {
			t.Errorf("Windows 脚本应为 .bat, got %q", scriptPath)
		}
	} else if !strings.HasSuffix(scriptPath, ".sh") {
		t.Errorf("非 Windows 脚本应为 .sh, got %q", scriptPath)
	}
}

// TestRunInternalRebaseScriptCleanup 脚本临时目录在调用结束后被清理
func TestRunInternalRebaseScriptCleanup(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	testutil.RunGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "base")
	testutil.RunGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "second")

	oldExe := executablePath
	executablePath = func() (string, error) { return `/usr/bin/easyGit`, nil }
	defer func() { executablePath = oldExe }()

	oldRun := runGitCommand
	var capturedEnv []string
	runGitCommand = func(args, env []string) error {
		capturedEnv = append([]string{}, env...)
		return nil
	}
	defer func() { runGitCommand = oldRun }()

	if err := RunInternalRebase("HEAD~1", "squash", []string{"abc123"}, "msg"); err != nil {
		t.Fatalf("RunInternalRebase() error = %v", err)
	}

	// 调用结束后脚本目录必须已删除
	for _, kv := range capturedEnv {
		if strings.HasPrefix(kv, "GIT_EDITOR=") {
			editor := strings.Trim(strings.TrimPrefix(kv, "GIT_EDITOR="), `"`)
			dir := filepath.Dir(editor)
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				t.Errorf("脚本临时目录未被清理: %s", dir)
			}
		}
	}
}

// TestWriteSquashMessageScript 消息注入脚本平台自适应:
// Windows 生成 .bat(cmd 语法),其他平台生成 .sh(sh 语法);
// 消息原样写入不丢失(含特殊字符)。
func TestWriteSquashMessageScript(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "普通消息", message: "fix: squash commits"},
		{name: "中文消息", message: "合并提交:修复登录问题"},
		{name: "特殊字符", message: `feat: "quoted" & <tag> 100%`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, scriptPath, err := writeSquashMessageScript(tt.message)
			if err != nil {
				t.Fatalf("writeSquashMessageScript() error = %v", err)
			}
			defer os.RemoveAll(dir)

			if scriptPath == "" {
				t.Fatal("scriptPath 为空")
			}
			if !strings.HasPrefix(scriptPath, dir) {
				t.Errorf("scriptPath %q 不在临时目录 %q 内", scriptPath, dir)
			}

			content, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("读取脚本失败: %v", err)
			}
			text := string(content)

			if runtime.GOOS == "windows" {
				if !strings.HasSuffix(scriptPath, ".bat") {
					t.Errorf("Windows 脚本应为 .bat, got %q", scriptPath)
				}
				if !strings.HasPrefix(text, "@echo off") {
					t.Errorf("bat 脚本应以 @echo off 开头, got %q", text)
				}
			} else {
				if !strings.HasSuffix(scriptPath, ".sh") {
					t.Errorf("非 Windows 脚本应为 .sh, got %q", scriptPath)
				}
				if !strings.HasPrefix(text, "#!/bin/sh") {
					t.Errorf("sh 脚本应以 #!/bin/sh 开头, got %q", text)
				}
			}
			// 消息必须原样出现在脚本中(断言不丢失,不依赖跨平台语法细节)
			if !strings.Contains(text, tt.message) {
				t.Errorf("脚本内容缺少完整消息:\n got  %q\n want 包含 %q", text, tt.message)
			}
		})
	}
}
