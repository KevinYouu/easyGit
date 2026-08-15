package config

import (
	"testing"
)

func TestGetLastChoiceEmpty(t *testing.T) {
	newTestDB(t)

	got, err := GetLastChoice(LastChoiceResetMode)
	if err != nil {
		t.Fatalf("GetLastChoice on empty db failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetLastChoice on empty db = %q, want empty", got)
	}
}

func TestSaveGetLastChoice(t *testing.T) {
	newTestDB(t)

	// 保存后读取
	if err := SaveLastChoice(LastChoiceResetMode, "--soft"); err != nil {
		t.Fatalf("SaveLastChoice(--soft) failed: %v", err)
	}
	got, err := GetLastChoice(LastChoiceResetMode)
	if err != nil {
		t.Fatalf("GetLastChoice failed: %v", err)
	}
	if got != "--soft" {
		t.Errorf("GetLastChoice after save = %q, want --soft", got)
	}

	// UPSERT 覆盖
	if err := SaveLastChoice(LastChoiceResetMode, "--hard"); err != nil {
		t.Fatalf("SaveLastChoice(--hard) failed: %v", err)
	}
	got, err = GetLastChoice(LastChoiceResetMode)
	if err != nil {
		t.Fatalf("GetLastChoice after overwrite failed: %v", err)
	}
	if got != "--hard" {
		t.Errorf("GetLastChoice after overwrite = %q, want --hard", got)
	}
}

func TestSaveLastChoiceEmptyIgnored(t *testing.T) {
	newTestDB(t)

	// 空值不保存(不覆盖已有记忆)
	if err := SaveLastChoice(LastChoiceMergeStrategy, "ff-only"); err != nil {
		t.Fatalf("SaveLastChoice(ff-only) failed: %v", err)
	}
	if err := SaveLastChoice(LastChoiceMergeStrategy, ""); err != nil {
		t.Fatalf("SaveLastChoice(empty) failed: %v", err)
	}
	got, err := GetLastChoice(LastChoiceMergeStrategy)
	if err != nil {
		t.Fatalf("GetLastChoice failed: %v", err)
	}
	if got != "ff-only" {
		t.Errorf("GetLastChoice after empty save = %q, want ff-only", got)
	}
}

func TestLastChoiceKeysIndependent(t *testing.T) {
	newTestDB(t)

	// 各记忆键互不干扰
	SaveLastChoice(LastChoiceResetMode, "--soft")
	SaveLastChoice(LastChoiceCherryPickOption, "signoff")

	resetGot, _ := GetLastChoice(LastChoiceResetMode)
	cpGot, _ := GetLastChoice(LastChoiceCherryPickOption)
	if resetGot != "--soft" || cpGot != "signoff" {
		t.Errorf("keys interfere: reset=%q cp=%q", resetGot, cpGot)
	}
}
