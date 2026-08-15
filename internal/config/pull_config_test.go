package config

import (
	"errors"
	"testing"
)

func TestGetPullBeforePushDefault(t *testing.T) {
	newTestDB(t)

	// 未设置时默认 always
	got, err := GetPullBeforePush()
	if err != nil {
		t.Fatalf("GetPullBeforePush failed: %v", err)
	}
	if got != PullBeforePushAlways {
		t.Errorf("GetPullBeforePush default = %q, want always", got)
	}
}

func TestSaveGetPullBeforePush(t *testing.T) {
	newTestDB(t)

	if err := SavePullBeforePush(PullBeforePushNever); err != nil {
		t.Fatalf("SavePullBeforePush(never) failed: %v", err)
	}
	got, err := GetPullBeforePush()
	if err != nil {
		t.Fatalf("GetPullBeforePush failed: %v", err)
	}
	if got != PullBeforePushNever {
		t.Errorf("GetPullBeforePush = %q, want never", got)
	}

	// UPSERT 覆盖
	if err := SavePullBeforePush(PullBeforePushAlways); err != nil {
		t.Fatalf("SavePullBeforePush(always) failed: %v", err)
	}
	got, _ = GetPullBeforePush()
	if got != PullBeforePushAlways {
		t.Errorf("GetPullBeforePush after overwrite = %q, want always", got)
	}
}

func TestSavePullBeforePushInvalid(t *testing.T) {
	newTestDB(t)

	err := SavePullBeforePush("sometimes")
	if !errors.Is(err, ErrInvalidPullSetting) {
		t.Fatalf("SavePullBeforePush(sometimes) err = %v, want ErrInvalidPullSetting", err)
	}
	// 非法值不落库
	got, _ := GetPullBeforePush()
	if got != PullBeforePushAlways {
		t.Errorf("GetPullBeforePush after invalid save = %q, want always", got)
	}
}
