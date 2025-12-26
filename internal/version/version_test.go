package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	// Version 变量应该在编译时注入
	// 测试它至少是一个字符串
	if Version == "" {
		// 开发环境中可能为空,这是正常的
		t.Log("Version is empty (expected in dev environment)")
	} else {
		t.Logf("Version: %s", Version)
	}
}

func TestGetLogo(t *testing.T) {
	// 测试不应该 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetLogo panicked: %v", r)
		}
	}()

	GetLogo()
}

func TestGetPenguin(t *testing.T) {
	// 测试不应该 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetPenguin panicked: %v", r)
		}
	}()

	GetPenguin()
}

func TestGetDivineBeast(t *testing.T) {
	// 测试不应该 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetDivineBeast panicked: %v", r)
		}
	}()

	GetDivineBeast()
}
