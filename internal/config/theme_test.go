package config

import (
	"errors"
	"testing"
)

func TestSaveThemeInvalid(t *testing.T) {
	newTestDB(t)

	err := SaveTheme("sepia")
	if err == nil {
		t.Fatal("SaveTheme(sepia) should fail with invalid theme")
	}
	if !errors.Is(err, ErrInvalidTheme) {
		t.Errorf("SaveTheme(sepia) error = %v, want wrapped ErrInvalidTheme", err)
	}

	// 非法值不应写入数据库
	got, err := GetTheme()
	if err != nil {
		t.Fatalf("GetTheme failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetTheme after invalid save = %q, want empty", got)
	}
}

func TestSaveGetTheme(t *testing.T) {
	newTestDB(t)

	// 未设置时返回空串(调用方按自动检测处理)
	got, err := GetTheme()
	if err != nil {
		t.Fatalf("GetTheme on empty db failed: %v", err)
	}
	if got != "" {
		t.Errorf("GetTheme on empty db = %q, want empty", got)
	}

	// 保存后读取
	if err := SaveTheme(ThemeDark); err != nil {
		t.Fatalf("SaveTheme(dark) failed: %v", err)
	}
	got, err = GetTheme()
	if err != nil {
		t.Fatalf("GetTheme failed: %v", err)
	}
	if got != ThemeDark {
		t.Errorf("GetTheme = %q, want %q", got, ThemeDark)
	}

	// UPSERT 覆盖
	if err := SaveTheme(ThemeLight); err != nil {
		t.Fatalf("SaveTheme(light) failed: %v", err)
	}
	got, err = GetTheme()
	if err != nil {
		t.Fatalf("GetTheme failed: %v", err)
	}
	if got != ThemeLight {
		t.Errorf("GetTheme after upsert = %q, want %q", got, ThemeLight)
	}

	// 三个合法值均可保存
	for _, v := range []string{ThemeAuto, ThemeDark, ThemeLight} {
		if err := SaveTheme(v); err != nil {
			t.Errorf("SaveTheme(%s) failed: %v", v, err)
		}
	}
}
