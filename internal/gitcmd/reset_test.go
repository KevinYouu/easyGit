package gitcmd

import (
	"strings"
	"testing"
	"time"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

func TestCommit_Structure(t *testing.T) {
	commit := Commit{
		Hash:      "abc123",
		Message:   "test commit",
		Date:      "01-15 10:30",
		Author:    "Test User",
		Email:     "test@example.com",
		IsHead:    true,
		Timestamp: time.Now(),
	}

	if commit.Hash != "abc123" {
		t.Errorf("Commit.Hash = %s, want abc123", commit.Hash)
	}
	if commit.Message != "test commit" {
		t.Errorf("Commit.Message = %s, want test commit", commit.Message)
	}
	if !commit.IsHead {
		t.Error("Commit.IsHead = false, want true")
	}
}

func TestGetModeDescription(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want bool // want non-empty string
	}{
		{
			name: "soft mode",
			mode: "--soft",
			want: true,
		},
		{
			name: "mixed mode",
			mode: "--mixed",
			want: true,
		},
		{
			name: "hard mode",
			mode: "--hard",
			want: true,
		},
		{
			name: "unknown mode",
			mode: "--unknown",
			want: false,
		},
		{
			name: "default mode",
			mode: "",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getModeDescription(tt.mode)
			isEmpty := got == ""
			if tt.want && isEmpty {
				t.Errorf("getModeDescription(%s) returned empty string, want non-empty", tt.mode)
			}
			if !tt.want && !isEmpty {
				t.Errorf("getModeDescription(%s) returned non-empty string, want empty", tt.mode)
			}
		})
	}

	// 有效模式须精确等于对应选项说明(单一来源:reset.option.<key>.desc)
	for _, mode := range []string{"", "--soft", "--mixed", "--hard"} {
		key := strings.TrimPrefix(mode, "--")
		if key == "" {
			key = "default"
		}
		want := i18n.T("reset.option." + key + ".desc")
		if got := ansi.Strip(getModeDescription(mode)); got != want {
			t.Errorf("getModeDescription(%q) = %q, want %q", mode, got, want)
		}
	}
}

func TestCommit_TimestampParsing(t *testing.T) {
	// 测试时间戳解析
	timestampStr := "2024-01-15 10:30:45 +0800"
	timestamp, err := time.Parse("2006-01-02 15:04:05 -0700", timestampStr)
	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	commit := Commit{
		Hash:      "abc123",
		Message:   "test",
		Timestamp: timestamp,
	}

	if commit.Timestamp.IsZero() {
		t.Error("Commit.Timestamp should not be zero")
	}

	// 验证年份
	if commit.Timestamp.Year() != 2024 {
		t.Errorf("Commit.Timestamp.Year() = %d, want 2024", commit.Timestamp.Year())
	}
}

func TestCommit_MessageTruncation(t *testing.T) {
	longMessage := "This is a very long commit message that should be truncated when displayed"

	// 模拟消息截断逻辑
	shortMsg := longMessage
	maxLen := 40
	if len(shortMsg) > maxLen {
		shortMsg = shortMsg[:maxLen-3] + "..."
	}

	if len(shortMsg) > maxLen {
		t.Errorf("Truncated message length = %d, want <= %d", len(shortMsg), maxLen)
	}

	if shortMsg[len(shortMsg)-3:] != "..." {
		t.Error("Truncated message should end with '...'")
	}
}

// TestResetModeOptions 重置模式选项:4 项含 default,值顺序与参数语义一致,
// 标签为「名称 + 说明」单行(名称亮 + 说明灰,ANSI 内嵌),无换行。
func TestResetModeOptions(t *testing.T) {
	opts := resetModeOptions()
	wantValues := []string{"", "--soft", "--mixed", "--hard"}
	if len(opts) != len(wantValues) {
		t.Fatalf("重置模式选项 %d 项, want %d", len(opts), len(wantValues))
	}
	for i, opt := range opts {
		if opt.Value != wantValues[i] {
			t.Errorf("第 %d 项 Value = %q, want %q", i, opt.Value, wantValues[i])
		}
		if strings.Contains(opt.Label, "\n") {
			t.Errorf("第 %d 项标签含换行(应单行): %q", i, opt.Label)
		}
		// 名称与说明键:default 的空值映射到 default 键
		key := strings.TrimPrefix(opt.Value, "--")
		if key == "" {
			key = "default"
		}
		plain := ansi.Strip(opt.Label)
		if !strings.Contains(plain, i18n.T("reset.option."+key+".name")) {
			t.Errorf("第 %d 项标签缺名称 %q: %q", i, i18n.T("reset.option."+key+".name"), plain)
		}
		if !strings.Contains(plain, i18n.T("reset.option."+key+".desc")) {
			t.Errorf("第 %d 项标签缺说明 %q: %q", i, i18n.T("reset.option."+key+".desc"), plain)
		}
	}
}
