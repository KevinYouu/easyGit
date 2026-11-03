package config

import (
	"errors"
	"testing"
)

func TestGetCommitBodyDefault(t *testing.T) {
	newTestDB(t)

	got, err := GetCommitBody()
	if err != nil {
		t.Fatalf("GetCommitBody failed: %v", err)
	}
	if got != CommitBodyOff {
		t.Errorf("GetCommitBody default = %q, want off", got)
	}
}

func TestSaveGetCommitBody(t *testing.T) {
	newTestDB(t)

	if err := SaveCommitBody(CommitBodyOn); err != nil {
		t.Fatalf("SaveCommitBody(on) failed: %v", err)
	}
	got, err := GetCommitBody()
	if err != nil {
		t.Fatalf("GetCommitBody failed: %v", err)
	}
	if got != CommitBodyOn {
		t.Errorf("GetCommitBody = %q, want on", got)
	}

	if err := SaveCommitBody(CommitBodyOff); err != nil {
		t.Fatalf("SaveCommitBody(off) failed: %v", err)
	}
	got, _ = GetCommitBody()
	if got != CommitBodyOff {
		t.Errorf("GetCommitBody after overwrite = %q, want off", got)
	}
}

func TestSaveCommitBodyInvalid(t *testing.T) {
	newTestDB(t)

	err := SaveCommitBody("sometimes")
	if !errors.Is(err, ErrInvalidCommitBodySetting) {
		t.Fatalf("SaveCommitBody(sometimes) err = %v, want ErrInvalidCommitBodySetting", err)
	}
	got, _ := GetCommitBody()
	if got != CommitBodyOff {
		t.Errorf("GetCommitBody after invalid save = %q, want off", got)
	}
}
