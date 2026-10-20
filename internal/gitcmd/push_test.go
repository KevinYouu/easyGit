package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
)

// TestValidateCommitMessage 提交消息校验:须含类型前缀且主题非空,
// 防止误 Enter 提交 "fix: " 空主题消息。
func TestValidateCommitMessage(t *testing.T) {
	cases := []struct {
		name    string
		message string
		wantErr bool
	}{
		{name: "normal message", message: "fix: 修复空主题提交", wantErr: false},
		{name: "english message", message: "feat: add validation", wantErr: false},
		{name: "empty subject", message: "fix: ", wantErr: true},
		{name: "empty subject no space", message: "fix:", wantErr: true},
		{name: "whitespace subject", message: "fix:   ", wantErr: true},
		{name: "no colon", message: "fix", wantErr: true},
		{name: "plain text", message: "just a message", wantErr: true},
		{name: "empty message", message: "", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCommitMessage(tc.message)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateCommitMessage(%q) err = %v, wantErr = %v", tc.message, err, tc.wantErr)
			}
			if tc.wantErr && !strings.Contains(err.Error(), "subject") {
				t.Fatalf("error message should hint at subject, got: %v", err)
			}
		})
	}
}

// TestShouldPullBeforePush 推送前 pull 决策:无 upstream 时跳过
// (真实仓库验证,避免无参数 git pull 失败终止流程)。
func TestShouldPullBeforePush(t *testing.T) {
	dir := setupTestRepo(t)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	// 无远程:无 upstream,应跳过
	if shouldPullBeforePush() {
		t.Error("无上游分支时应跳过 pull")
	}
}

// TestComposeCommitMessage 主题 + 可选正文拼装:正文按行保留,
// 空正文不留悬空空行,CRLF 规范为 LF,首尾空白收敛
func TestComposeCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		body    string
		want    string
	}{
		{name: "仅主题", subject: "fix: bug", body: "", want: "fix: bug"},
		{name: "主题+多行正文", subject: "fix: bug", body: "line 1\nline 2", want: "fix: bug\n\nline 1\nline 2"},
		{name: "正文首尾空行收敛", subject: "fix: bug", body: "\n\nline 1\n\n", want: "fix: bug\n\nline 1"},
		{name: "正文纯空白视为无正文", subject: "fix: bug", body: "  \n\t", want: "fix: bug"},
		{name: "CRLF 规范为 LF", subject: "fix: bug", body: "line 1\r\nline 2", want: "fix: bug\n\nline 1\nline 2"},
		{name: "主题尾部空白收敛", subject: "fix: bug  ", body: "", want: "fix: bug"},
		{name: "正文内部空行保留", subject: "fix: bug", body: "para 1\n\npara 2", want: "fix: bug\n\npara 1\n\npara 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := composeCommitMessage(tt.subject, tt.body); got != tt.want {
				t.Errorf("composeCommitMessage(%q, %q) = %q, want %q", tt.subject, tt.body, got, tt.want)
			}
		})
	}
}

// TestAppendPushSteps 拉取与并行推送步骤:parallelFrom 等于串行步数,
// pull 可选(未设上游时跳过),每个远程一步
func TestAppendPushSteps(t *testing.T) {
	tests := []struct {
		name         string
		existing     []command.CommandInfo
		remotes      []string
		pullBefore   bool
		wantLen      int
		wantParallel int
	}{
		{
			name:         "无 pull 单远程",
			existing:     []command.CommandInfo{{Command: "git"}, {Command: "git"}},
			remotes:      []string{"origin"},
			wantLen:      3,
			wantParallel: 2,
		},
		{
			name:         "pull + 多远程",
			existing:     []command.CommandInfo{{Command: "git"}},
			remotes:      []string{"origin", "backup"},
			pullBefore:   true,
			wantLen:      4,
			wantParallel: 2,
		},
		{
			name:         "干净工作区(无 add/commit)+ pull + 单远程",
			existing:     nil,
			remotes:      []string{"origin"},
			pullBefore:   true,
			wantLen:      2,
			wantParallel: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, parallelFrom := appendPushSteps(tt.existing, tt.remotes, "main", tt.pullBefore)
			if len(got) != tt.wantLen {
				t.Fatalf("步骤数 = %d, want %d", len(got), tt.wantLen)
			}
			if parallelFrom != tt.wantParallel {
				t.Errorf("parallelFrom = %d, want %d", parallelFrom, tt.wantParallel)
			}
			// 串行段内的命令不得是 push;并行段必须全是 push
			for i, step := range got {
				isPush := len(step.Args) > 0 && step.Args[0] == "push"
				if isPush && i < parallelFrom {
					t.Errorf("第 %d 步是 push,不应在串行段(parallelFrom=%d)", i, parallelFrom)
				}
				if !isPush && i >= parallelFrom {
					t.Errorf("第 %d 步不是 push,不应在并行段", i)
				}
			}
		})
	}
}

// TestCommitSteps 工作区干净时不生成 add/commit 步骤(否则 commit 失败中断推送),
// 有改动时按 add → commit 顺序生成
func TestCommitSteps(t *testing.T) {
	add := command.CommandInfo{Command: "git", Args: []string{"add", "-A"}}

	if steps := commitSteps(false, add, "fix: msg"); len(steps) != 0 {
		t.Fatalf("干净工作区不应生成提交步骤, got %d", len(steps))
	}

	steps := commitSteps(true, add, "fix: msg\n\nbody")
	if len(steps) != 2 {
		t.Fatalf("步骤数 = %d, want 2", len(steps))
	}
	if strings.Join(steps[0].Args, " ") != "add -A" {
		t.Errorf("第 1 步 = %v, want add -A", steps[0].Args)
	}
	want := "commit -m fix: msg\n\nbody"
	if strings.Join(steps[1].Args, " ") != want {
		t.Errorf("第 2 步 = %v, want %s", steps[1].Args, want)
	}
}

// TestPushAllNoChangesNoRemote 干净工作区 + 无远程:提示后直接返回,不进入表单
func TestPushAllNoChangesNoRemote(t *testing.T) {
	config.SetTestDBPath(filepath.Join(t.TempDir(), "test.db"))
	defer config.SetTestDBPath("")

	dir := setupTestRepo(t)
	oldWD, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(oldWD)

	if err := PushAll(); err != nil {
		t.Fatalf("PushAll() error = %v", err)
	}
}
