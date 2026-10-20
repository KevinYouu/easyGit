package gitcmd

import (
	"os"
	"path/filepath"
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
