package update

import (
	"runtime"
	"testing"
)

func TestGetPlatformName(t *testing.T) {
	// 这个测试只能测试当前平台
	platform, err := getPlatformName()

	// 根据当前运行环境验证结果
	switch runtime.GOOS {
	case "linux":
		if err != nil {
			t.Fatalf("expected no error on linux, got: %v", err)
		}
		if runtime.GOARCH == "amd64" && platform != "linux_amd64" {
			t.Errorf("expected linux_amd64, got: %s", platform)
		}
		if (runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64") && platform != "linux_arm64" {
			t.Errorf("expected linux_arm64, got: %s", platform)
		}
	case "darwin":
		if err != nil {
			t.Fatalf("expected no error on darwin, got: %v", err)
		}
		if runtime.GOARCH == "amd64" && platform != "darwin_amd64" {
			t.Errorf("expected darwin_amd64, got: %s", platform)
		}
		if runtime.GOARCH == "arm64" && platform != "darwin_arm64" {
			t.Errorf("expected darwin_arm64, got: %s", platform)
		}
	case "windows":
		if err != nil {
			t.Fatalf("expected no error on windows, got: %v", err)
		}
		if runtime.GOARCH == "amd64" && platform != "windows_amd64" {
			t.Errorf("expected windows_amd64, got: %s", platform)
		}
		if runtime.GOARCH == "arm64" && platform != "windows_arm64" {
			t.Errorf("expected windows_arm64, got: %s", platform)
		}
	default:
		// 对于不支持的平台，应该返回错误
		if err == nil {
			t.Error("expected error for unsupported platform")
		}
	}
}

func TestGetInstallDir(t *testing.T) {
	installDir := getInstallDir()

	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			if installDir != "/opt/homebrew/bin" {
				t.Errorf("expected /opt/homebrew/bin for darwin arm64, got: %s", installDir)
			}
		} else {
			if installDir != "/usr/local/bin" {
				t.Errorf("expected /usr/local/bin for darwin, got: %s", installDir)
			}
		}
	case "linux":
		if installDir != "/usr/local/bin" {
			t.Errorf("expected /usr/local/bin for linux, got: %s", installDir)
		}
	case "windows":
		if installDir != "" {
			t.Errorf("expected empty string for windows, got: %s", installDir)
		}
	default:
		if installDir != "/usr/local/bin" {
			t.Errorf("expected /usr/local/bin as default, got: %s", installDir)
		}
	}
}

func TestHasWritePermission(t *testing.T) {
	// 测试临时目录（应该有写权限）
	tempDir := t.TempDir()
	if !hasWritePermission(tempDir) {
		t.Error("expected write permission in temp directory")
	}

	// 测试无效目录（应该没有写权限）
	if hasWritePermission("/nonexistent/directory/path") {
		t.Error("expected no write permission for nonexistent directory")
	}
}
