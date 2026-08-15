package gitcmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// setupTestRepo 在临时目录初始化真实 git 仓库,返回仓库路径。
func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return string(out)
	}

	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test User")

	for i := range 3 {
		if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content "+string(rune('a'+i))), 0644); err != nil {
			t.Fatal(err)
		}
		run("add", ".")
		run("commit", "-q", "-m", "commit message "+string(rune('a'+i)))
	}
	return dir
}

// TestGetCommitsOptions 统一提交数据源:数量限制、HEAD 标记、字段解析。
func TestGetCommitsOptions(t *testing.T) {
	dir := setupTestRepo(t)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	// 全部提交
	options, commits, err := GetCommitsOptions(0)
	if err != nil {
		t.Fatalf("GetCommitsOptions(0) err: %v", err)
	}
	if len(commits) != 3 {
		t.Fatalf("want 3 commits, got %d", len(commits))
	}
	if !commits[0].IsHead {
		t.Error("newest commit should be IsHead")
	}
	if commits[0].Message != "commit message c" {
		t.Errorf("Message = %q, want 'commit message c'", commits[0].Message)
	}
	if commits[0].Author != "Test User" {
		t.Errorf("Author = %q, want 'Test User'", commits[0].Author)
	}
	if commits[0].Email != "test@example.com" {
		t.Errorf("Email = %q, want 'test@example.com'", commits[0].Email)
	}
	if commits[0].Timestamp.IsZero() {
		t.Error("Timestamp should be parsed from ISO format")
	}
	if commits[0].Timestamp.After(time.Now()) {
		t.Error("Timestamp should not be in the future")
	}
	if len(options) != 3 || options[0].Value != commits[0].Hash {
		t.Error("options should align with commits by hash")
	}
	if options[0].Label != commits[0].Hash+" commit message c\n"+commits[0].Date+" • Test User" {
		t.Errorf("Label mismatch: %q", options[0].Label)
	}

	// 限制数量
	limited, lcommits, err := GetCommitsOptions(2)
	if err != nil {
		t.Fatalf("GetCommitsOptions(2) err: %v", err)
	}
	if len(lcommits) != 2 {
		t.Fatalf("want 2 commits, got %d", len(lcommits))
	}
	if len(limited) != 2 {
		t.Fatalf("want 2 options, got %d", len(limited))
	}
}

// TestGetRecentCommits 薄封装:哈希列表与选项对齐。
func TestGetRecentCommits(t *testing.T) {
	dir := setupTestRepo(t)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	options, hashes, err := GetRecentCommits()
	if err != nil {
		t.Fatalf("GetRecentCommits err: %v", err)
	}
	if len(hashes) != 3 {
		t.Fatalf("want 3 hashes, got %d", len(hashes))
	}
	for i, opt := range options {
		if opt.Value != hashes[i] {
			t.Errorf("option %d value %q != hash %q", i, opt.Value, hashes[i])
		}
	}
}
