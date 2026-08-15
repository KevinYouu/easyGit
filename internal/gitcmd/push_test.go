package gitcmd

import (
	"strings"
	"testing"
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
