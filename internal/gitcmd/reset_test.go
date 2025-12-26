package gitcmd

import (
	"testing"
	"time"
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
