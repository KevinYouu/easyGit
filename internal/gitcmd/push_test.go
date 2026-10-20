package gitcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// TestBuildPushSteps 完整步骤序列:干净工作区只有 pull + push(有远程),
// 有改动时才是 add → commit → pull → push;串行段不含 push
func TestBuildPushSteps(t *testing.T) {
	tests := []struct {
		name         string
		hasChanges   bool
		remotes      []string
		pullBefore   bool
		wantArgs     []string
		wantParallel int
	}{
		{
			name:       "干净工作区 + 单远程 + pull",
			hasChanges: false,
			remotes:    []string{"origin"},
			pullBefore: true,
			wantArgs:   []string{"pull", "push origin main"},
			// pull 串行,只有 push 并行
			wantParallel: 1,
		},
		{
			name:       "干净工作区 + 单远程 + 无 pull",
			hasChanges: false,
			remotes:    []string{"origin"},
			pullBefore: false,
			wantArgs:   []string{"push origin main"},
			// 首步即可并行
			wantParallel: 0,
		},
		{
			name:       "有改动 + 多远程 + pull",
			hasChanges: true,
			remotes:    []string{"origin", "backup"},
			pullBefore: true,
			wantArgs: []string{
				"add -A", "commit -m fix: msg\n\nbody", "pull",
				"push origin main", "push backup main",
			},
			wantParallel: 3,
		},
		{
			name:       "有改动 + 无远程(仅本地提交)",
			hasChanges: true,
			remotes:    nil,
			wantArgs:   []string{"add -A", "commit -m fix: msg\n\nbody"},
			// 全部串行
			wantParallel: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, parallelFrom := buildPushSteps(tt.hasChanges, "fix: msg\n\nbody", tt.remotes, "main", tt.pullBefore)

			var got []string
			for _, step := range steps {
				got = append(got, strings.Join(step.Args, " "))
			}
			if strings.Join(got, "|") != strings.Join(tt.wantArgs, "|") {
				t.Errorf("步骤 = %v, want %v", got, tt.wantArgs)
			}
			if parallelFrom != tt.wantParallel {
				t.Errorf("parallelFrom = %d, want %d", parallelFrom, tt.wantParallel)
			}
			for i, step := range steps {
				isPush := len(step.Args) > 0 && step.Args[0] == "push"
				if isPush != (i >= parallelFrom) {
					t.Errorf("第 %d 步 push=%v 与 parallelFrom=%d 不符", i, isPush, parallelFrom)
				}
			}
		})
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
