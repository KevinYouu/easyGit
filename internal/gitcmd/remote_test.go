package gitcmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KevinYouu/easyGit/internal/testutil"
)

func TestGetCurrentRemote(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 添加远程仓库
	testutil.RunGitCommand(t, tempDir, "remote", "add", "origin", "https://github.com/test/repo.git")

	// 创建一个初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 测试未设置上游分支的情况
	remote, err := GetCurrentRemote()
	if err != nil {
		t.Errorf("GetCurrentRemote() error = %v", err)
	}
	// 应该返回默认值 "origin"
	if remote != "origin" {
		t.Errorf("GetCurrentRemote() = %v, want %v", remote, "origin")
	}
}

func TestGetAllRemotes(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	tests := []struct {
		name    string
		setup   func()
		want    []string
		wantErr bool
	}{
		{
			name: "no remotes",
			setup: func() {
				// 不添加任何远程
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "single remote",
			setup: func() {
				testutil.RunGitCommand(t, tempDir, "remote", "add", "origin", "https://github.com/test/repo.git")
			},
			want:    []string{"origin"},
			wantErr: false,
		},
		{
			name: "multiple remotes",
			setup: func() {
				testutil.RunGitCommand(t, tempDir, "remote", "add", "upstream", "https://github.com/upstream/repo.git")
			},
			want:    []string{"origin", "upstream"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			got, err := GetAllRemotes()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllRemotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("GetAllRemotes() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestGetRemoteBranches(t *testing.T) {
	tempDir, cleanup := testutil.CreateTempGitRepo(t)
	defer cleanup()

	// 切换到临时目录
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// 添加远程仓库
	testutil.RunGitCommand(t, tempDir, "remote", "add", "origin", "https://github.com/test/repo.git")

	// 创建初始提交
	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)
	testutil.RunGitCommand(t, tempDir, "add", "test.txt")
	testutil.RunGitCommand(t, tempDir, "commit", "-m", "initial commit")

	// 模拟创建远程跟踪分支
	// 注意:在真实的测试环境中,远程分支需要实际的 fetch 操作
	// 这里我们只测试函数不会崩溃
	branches, err := GetRemoteBranches("origin")
	if err != nil {
		t.Errorf("GetRemoteBranches() error = %v", err)
	}

	// 空仓库可能没有远程分支
	if branches == nil {
		branches = []string{}
	}
	t.Logf("Found %d remote branches", len(branches))
}
