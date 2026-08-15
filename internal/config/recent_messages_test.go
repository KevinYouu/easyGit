package config

import (
	"fmt"
	"testing"
)

func TestRecentCommitMessagesEmpty(t *testing.T) {
	newTestDB(t)

	if got := GetRecentCommitMessages(); got != nil {
		t.Errorf("GetRecentCommitMessages on empty db = %v, want nil", got)
	}
}

func TestAddRecentCommitMessages(t *testing.T) {
	newTestDB(t)

	// 依次添加,最新在前
	AddRecentCommitMessage("fix: 第一个")
	AddRecentCommitMessage("feat: 第二个")
	got := GetRecentCommitMessages()
	if len(got) != 2 || got[0] != "feat: 第二个" || got[1] != "fix: 第一个" {
		t.Fatalf("messages = %v, want [feat: 第二个 fix: 第一个]", got)
	}

	// 去重:重复消息移到最前(独立表 UNIQUE + created_at 刷新)
	AddRecentCommitMessage("fix: 第一个")
	got = GetRecentCommitMessages()
	if len(got) != 2 || got[0] != "fix: 第一个" {
		t.Fatalf("after dedup messages = %v, want [fix: 第一个 feat: 第二个]", got)
	}

	// 截断:超过上限时只保留最近 RecentMessagesLimit 条
	for i := range RecentMessagesLimit + 3 {
		AddRecentCommitMessage(fmt.Sprintf("msg: 批次 %d", i))
	}
	AddRecentCommitMessage("fix: 最终")
	got = GetRecentCommitMessages()
	if len(got) != RecentMessagesLimit {
		t.Fatalf("messages length = %d, want %d", len(got), RecentMessagesLimit)
	}
	if got[0] != "fix: 最终" {
		t.Errorf("newest message should be first, got %q", got[0])
	}
	// 最旧的 "fix: 第一个" 应被截断
	for _, m := range got {
		if m == "fix: 第一个" {
			t.Errorf("oldest message should be truncated, got %v", got)
		}
	}
}

func TestAddRecentCommitMessagesEmptyIgnored(t *testing.T) {
	newTestDB(t)

	AddRecentCommitMessage("fix: 有效")
	if err := AddRecentCommitMessage(""); err != nil {
		t.Fatalf("AddRecentCommitMessage(empty) failed: %v", err)
	}
	got := GetRecentCommitMessages()
	if len(got) != 1 {
		t.Fatalf("empty message should be ignored, got %v", got)
	}
}
