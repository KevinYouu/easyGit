package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

// setupSquashRepo 构造含多行提交信息的仓库:
// c1 主题 + 两行正文,c2 纯主题,c3 主题 + 单行正文,返回从旧到新的哈希
func setupSquashRepo(t *testing.T) []string {
	t.Helper()

	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	t.Cleanup(cleanup)

	originalDir, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(originalDir) })
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	commits := []string{
		"feat: first\n\nfirst body line 1\nfirst body line 2",
		"fix: second",
		"chore: third\n\nthird body",
	}
	hashes := make([]string, 0, len(commits))
	for i, message := range commits {
		path := filepath.Join(tempDir, "file.txt")
		if err := os.WriteFile(path, []byte(strings.Repeat("x", i+1)), 0o644); err != nil {
			t.Fatalf("write file: %v", err)
		}
		testutil.RunGitCommand(t, tempDir, "add", "file.txt")
		testutil.RunGitCommand(t, tempDir, "commit", "-m", message)

		hashes = append(hashes, revParseHead(t, tempDir))
	}
	return hashes
}

// revParseHead 取当前 HEAD 的短哈希(与 GetCommitsOptions 的 %h 同长度口径)
func revParseHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse HEAD: %v", err)
	}
	return strings.TrimSpace(string(output))
}

// TestGetCommitMessageBodiesPreservesMultiline 完整消息必须带正文与换行:
// %s 只取主题,合并时会丢掉全部正文(本缺陷的根因)
func TestGetCommitMessageBodiesPreservesMultiline(t *testing.T) {
	hashes := setupSquashRepo(t)

	bodies, err := GetCommitMessageBodies(hashes)
	if err != nil {
		t.Fatalf("GetCommitMessageBodies() error = %v", err)
	}
	if len(bodies) != len(hashes) {
		t.Fatalf("bodies = %d, want %d", len(bodies), len(hashes))
	}

	if !strings.Contains(bodies[0], "first body line 1\nfirst body line 2") {
		t.Errorf("第 1 条正文丢失: %q", bodies[0])
	}
	if got := strings.TrimRight(bodies[1], "\n"); got != "fix: second" {
		t.Errorf("第 2 条消息 = %q, want %q", got, "fix: second")
	}
	if !strings.Contains(bodies[2], "\nthird body") {
		t.Errorf("第 3 条正文丢失: %q", bodies[2])
	}
}

// TestOrderOldestFirst 所选提交按 hashes(新→旧)下标重排为旧→新
func TestOrderOldestFirst(t *testing.T) {
	hashes := []string{"c3", "c2", "c1"}

	tests := []struct {
		name     string
		selected []string
		want     []string
	}{
		{name: "乱序选择", selected: []string{"c3", "c1", "c2"}, want: []string{"c1", "c2", "c3"}},
		{name: "已按旧到新", selected: []string{"c1", "c2"}, want: []string{"c1", "c2"}},
		{name: "单个", selected: []string{"c2"}, want: []string{"c2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := orderOldestFirst(tt.selected, hashes)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Errorf("orderOldestFirst(%v) = %v, want %v", tt.selected, got, tt.want)
			}
		})
	}
}

// TestBuildSquashDefaultMessage 默认消息保留全部文本:每条提交的主题与正文
// 按旧→新排列、空行分隔,换行结构不被折叠
func TestBuildSquashDefaultMessage(t *testing.T) {
	hashes := setupSquashRepo(t)

	message, err := buildSquashDefaultMessage(hashes)
	if err != nil {
		t.Fatalf("buildSquashDefaultMessage() error = %v", err)
	}

	want := "feat: first\n\nfirst body line 1\nfirst body line 2\n\nfix: second\n\nchore: third\n\nthird body"
	if message != want {
		t.Errorf("默认消息 = %q\nwant %q", message, want)
	}
}

// TestBuildSquashDefaultMessageSkipsEmptyBody 空正文提交不产生多余空行,
// 且不会因消息为空而丢段
func TestBuildSquashDefaultMessageSkipsEmptyBody(t *testing.T) {
	hashes := setupSquashRepo(t)

	message, err := buildSquashDefaultMessage([]string{hashes[1], hashes[1]})
	if err != nil {
		t.Fatalf("buildSquashDefaultMessage() error = %v", err)
	}
	if message != "fix: second\n\nfix: second" {
		t.Errorf("默认消息 = %q", message)
	}
}
