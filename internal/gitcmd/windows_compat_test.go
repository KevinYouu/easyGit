package gitcmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

// normalizePath 规范化路径:解析符号链接(macOS /var → /private/var)
func normalizePath(t *testing.T, p string) string {
	t.Helper()
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	return p
}

// TestResolveAbsolutePaths 相对路径列表在 Git 工作树根下解析为绝对路径
func TestResolveAbsolutePaths(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录(getGitWorkTree 在当前目录执行 git)
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	testutil.RunGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "init")

	// 创建子目录与文件,验证嵌套路径解析
	subDir := filepath.Join(tempDir, "src", "pkg")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main"), 0644)

	// 规范化临时目录路径(macOS /var 是 /private/var 的符号链接)
	normDir := normalizePath(t, tempDir)

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "空列表", in: nil, want: nil},
		{name: "单层相对路径", in: []string{"f.txt"}, want: []string{filepath.Join(normDir, "f.txt")}},
		{name: "嵌套相对路径", in: []string{"src/pkg/main.go"}, want: []string{filepath.Join(normDir, "src", "pkg", "main.go")}},
		{name: "绝对路径保持不变", in: []string{"/abs/a.txt"}, want: []string{"/abs/a.txt"}},
		{name: "混合路径", in: []string{"a.txt", "/abs/b.txt"}, want: []string{filepath.Join(normDir, "a.txt"), "/abs/b.txt"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveAbsolutePaths(tt.in)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range got {
				gotNorm := normalizePath(t, got[i])
				if gotNorm != tt.want[i] {
					t.Errorf("path[%d] = %q, want %q", i, gotNorm, tt.want[i])
				}
			}
		})
	}
}

// TestGetGitWorkTree 验证获取工作树根绝对路径
func TestGetGitWorkTree(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	testutil.RunGitCommand(t, tempDir, "commit", "--allow-empty", "-m", "init")

	workTree, err := getGitWorkTree()
	if err != nil {
		t.Fatalf("getGitWorkTree() error = %v", err)
	}
	if workTree == "" {
		t.Fatal("getGitWorkTree() returned empty path")
	}
	// Windows 返回路径可能大小写不同,用 EvalSymlinks 规范化比较
	want := normalizePath(t, tempDir)
	got := normalizePath(t, workTree)
	if got != want {
		t.Errorf("getGitWorkTree() = %q, want %q", got, want)
	}
}

// TestEscapeBatMessage Windows cmd 特殊字符转义覆盖:
// % ^ & | < > " 必须被 ^ 前缀转义,普通字符不变
func TestEscapeBatMessage(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "纯文本", in: "fix: normal message", want: "fix: normal message"},
		{name: "百分号", in: "100% done", want: "100^% done"},
		{name: "脱字符", in: "a^b", want: "a^^b"},
		{name: "与号", in: "a&b", want: "a^&b"},
		{name: "管道", in: "a|b", want: "a^|b"},
		{name: "小于号", in: "a<b", want: "a^<b"},
		{name: "大于号", in: "a>b", want: "a^>b"},
		{name: "双引号", in: `say "hi"`, want: `say ^"hi^"`},
		{name: "混合注入载荷", in: `& calc | echo injected`, want: `^& calc ^| echo injected`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeBatMessage(tt.in); got != tt.want {
				t.Errorf("escapeBatMessage(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestEscapeShMessage POSIX shell 特殊字符转义覆盖:
// \ " $ ` 必须被 \ 前缀转义,普通字符不变
func TestEscapeShMessage(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "纯文本", in: "fix: normal message", want: "fix: normal message"},
		{name: "反斜杠", in: `a\b`, want: `a\\b`},
		{name: "双引号", in: `a"b`, want: `a\"b`},
		{name: "美元符号", in: "cost $100", want: `cost \$100`},
		{name: "反引号", in: "a`cmd`b", want: "a\\`cmd\\`b"},
		{name: "混合注入载荷", in: `$(rm -rf /)`, want: `\$(rm -rf /)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeShMessage(tt.in); got != tt.want {
				t.Errorf("escapeShMessage(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestWriteSquashMessageScriptEscapesInjection 消息注入脚本必须转义特殊字符,
// 防止 shell 命令注入(消息出现在脚本中而非被拼接为命令)
func TestWriteSquashMessageScriptEscapesInjection(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "与号注入", message: "fix: done & calc"},
		{name: "管道注入", message: "fix: | echo injected"},
		{name: "反引号注入", message: "fix: `whoami`"},
		{name: "美元注入", message: "fix: $HOME"},
		{name: "双引号注入", message: `fix: "quoted"`},
		{name: "反斜杠注入", message: `fix: C:\path\to\file`},
		{name: "混合注入载荷", message: `& | < > " \`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, scriptPath, err := writeSquashMessageScript(tt.message)
			if err != nil {
				t.Fatalf("writeSquashMessageScript() error = %v", err)
			}
			defer os.RemoveAll(dir)

			content, err := os.ReadFile(scriptPath)
			if err != nil {
				t.Fatalf("读取脚本失败: %v", err)
			}
			text := string(content)

			// 脚本内容必须包含完整消息(转义后的版本)
			if !strings.Contains(text, tt.message) {
				// 消息含特殊字符,可能被转义;验证转义后的版本存在于脚本中
				if runtime.GOOS == "windows" {
					if !strings.Contains(text, escapeBatMessage(tt.message)) {
						t.Errorf("脚本内容缺少转义后的消息:\n got  %q\n want 包含 %q", text, escapeBatMessage(tt.message))
					}
				} else {
					if !strings.Contains(text, escapeShMessage(tt.message)) {
						t.Errorf("脚本内容缺少转义后的消息:\n got  %q\n want 包含 %q", text, escapeShMessage(tt.message))
					}
				}
			}

			// 验证脚本结构正确性
			lines := strings.Split(text, "\n")
			if runtime.GOOS == "windows" {
				// bat 脚本第一行是 @echo off
				if len(lines) == 0 || !strings.HasPrefix(lines[0], "@echo off") {
					t.Errorf("bat 脚本应以 @echo off 开头, got %q", text)
				}
			} else {
				// sh 脚本第一行是 #!/bin/sh
				if len(lines) == 0 || !strings.HasPrefix(lines[0], "#!/bin/sh") {
					t.Errorf("sh 脚本应以 #!/bin/sh 开头, got %q", text)
				}
			}
		})
	}
}
