package update

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/version"
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

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name       string
		local      string
		remote     string
		wantResult int
		wantOK     bool
	}{
		{name: "本地低于远程", local: "0.2.4", remote: "0.2.5", wantResult: -1, wantOK: true},
		{name: "本地等于远程", local: "0.2.5", remote: "v0.2.5", wantResult: 0, wantOK: true},
		{name: "本地高于远程", local: "0.3.0", remote: "v0.2.5", wantResult: 1, wantOK: true},
		{name: "忽略 v 前缀", local: "v0.2.5", remote: "0.2.5", wantResult: 0, wantOK: true},
		{name: "段数不同补零比较", local: "0.2", remote: "0.2.0", wantResult: 0, wantOK: true},
		{name: "段数不同大小比较", local: "0.2", remote: "0.2.1", wantResult: -1, wantOK: true},
		{name: "忽略 git describe 后缀", local: "0.2.5-1-gabc123", remote: "0.2.5", wantResult: 0, wantOK: true},
		{name: "忽略 dirty 后缀", local: "0.2.5-dirty", remote: "0.2.5", wantResult: 0, wantOK: true},
		{name: "预发布低于同数字段稳定版", local: "0.2.5-rc.1", remote: "v0.2.5", wantResult: -1, wantOK: true},
		{name: "预发布带构建元数据仍低于稳定版", local: "0.2.5-rc.1+meta", remote: "0.2.5", wantResult: -1, wantOK: true},
		{name: "稳定版高于同数字段预发布", local: "0.2.5", remote: "v0.2.5-rc.1", wantResult: 1, wantOK: true},
		{name: "构建元数据与稳定版相等", local: "1.2.3+meta", remote: "1.2.3", wantResult: 0, wantOK: true},
		{name: "构建元数据含连字符视为稳定版", local: "1.2.3+build-1", remote: "1.2.3", wantResult: 0, wantOK: true},
		{name: "构建元数据含预发布字样仍为稳定版", local: "0.2.5+rc.1-1", remote: "0.2.5", wantResult: 0, wantOK: true},
		{name: "git 提交计数后缀带 dirty 视为同版本", local: "0.2.5-2-gabc123-dirty", remote: "0.2.5", wantResult: 0, wantOK: true},
		{name: "预发布后缀按数字段比较", local: "0.2.5-rc.1", remote: "0.2.5-rc.2", wantResult: -1, wantOK: true},
		{name: "预发布后缀相等视为已最新", local: "0.2.5-rc.1", remote: "0.2.5-rc.1", wantResult: 0, wantOK: true},
		{name: "预发布字母段按字典序", local: "0.2.5-beta.2", remote: "0.2.5-rc.1", wantResult: -1, wantOK: true},
		{name: "短预发布后缀低于带序号", local: "0.2.5-rc", remote: "0.2.5-rc.1", wantResult: -1, wantOK: true},
		{name: "数字标识符低于字母标识符", local: "0.2.5-1", remote: "0.2.5-rc.1", wantResult: -1, wantOK: true},
		{name: "开发版本无法比较", local: "untracked", remote: "0.2.5", wantResult: 0, wantOK: false},
		{name: "远程无法解析", local: "0.2.5", remote: "latest", wantResult: 0, wantOK: false},
		{name: "空版本无法比较", local: "", remote: "0.2.5", wantResult: 0, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, gotOK := compareVersions(tt.local, tt.remote)
			if gotResult != tt.wantResult || gotOK != tt.wantOK {
				t.Errorf("compareVersions(%q, %q) = (%d, %v), want (%d, %v)",
					tt.local, tt.remote, gotResult, gotOK, tt.wantResult, tt.wantOK)
			}
		})
	}
}

func TestVersionParts(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   []int
		wantOK bool
	}{
		{name: "标准三段", input: "0.2.5", want: []int{0, 2, 5}, wantOK: true},
		{name: "带 v 前缀", input: "v1.2.3", want: []int{1, 2, 3}, wantOK: true},
		{name: "两段", input: "1.2", want: []int{1, 2}, wantOK: true},
		{name: "含构建元数据", input: "1.2.3+meta", want: []int{1, 2, 3}, wantOK: true},
		{name: "含预发布段", input: "1.2.3-beta.1", want: []int{1, 2, 3}, wantOK: true},
		{name: "含非数字段", input: "1.2.x", want: nil, wantOK: false},
		{name: "空字符串", input: "", want: nil, wantOK: false},
		{name: "纯后缀", input: "untracked", want: nil, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := versionParts(tt.input)
			if ok != tt.wantOK {
				t.Fatalf("versionParts(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("versionParts(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("versionParts(%q) = %v, want %v", tt.input, got, tt.want)
				}
			}
		})
	}
}

// makeFakeReleaseServer 构造模拟 GitHub release 的测试服务器，返回服务器与 zip 下载计数
func makeFakeReleaseServer(t *testing.T, zipBytes []byte, sum []byte, zipRequests *int) *httptest.Server {
	t.Helper()

	platform, err := getPlatformName()
	if err != nil {
		t.Fatalf("getPlatformName() failed: %v", err)
	}
	assetName := fmt.Sprintf("easyGit_v9.9.9_%s.zip", platform)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			fmt.Fprint(w, `{"tag_name":"v9.9.9"}`)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			fmt.Fprintf(w, "%x  %s\n", sum, assetName)
		case strings.HasSuffix(r.URL.Path, ".zip"):
			*zipRequests++
			w.Write(zipBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	return server
}

func TestUpdateUnixEndToEnd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// 构造假 release 安装包：zip 内含 easyGit 二进制
	var zipBuf bytes.Buffer
	binaryContent := []byte("fake easyGit binary")
	zipWriter := zip.NewWriter(&zipBuf)
	entry, err := zipWriter.Create("easyGit")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := entry.Write(binaryContent); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	zipBytes := zipBuf.Bytes()
	sum := sha256.Sum256(zipBytes)

	var zipRequests int
	server := makeFakeReleaseServer(t, zipBytes, sum[:], &zipRequests)
	defer server.Close()

	// 预置旧版本二进制作为安装目标
	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "bin", "easyGit")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	// 本地版本为不可解析的开发版（默认 untracked），必然触发下载更新
	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	if err := updateUnixTo(client, target); err != nil {
		t.Fatalf("updateUnixTo() unexpected error: %v", err)
	}

	// 目标内容应被替换为假 release 的二进制
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read updated binary: %v", err)
	}
	if string(content) != string(binaryContent) {
		t.Errorf("expected new binary content, got: %s", content)
	}
	if zipRequests != 1 {
		t.Errorf("expected exactly 1 zip download, got %d", zipRequests)
	}
}

func TestUpdateUnixAlreadyLatest(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// 将本地版本伪装成与远程一致，应跳过下载与安装
	oldVersion := version.Version
	version.Version = "v9.9.9"
	t.Cleanup(func() { version.Version = oldVersion })

	var zipRequests int
	server := makeFakeReleaseServer(t, []byte("zip"), make([]byte, 32), &zipRequests)
	defer server.Close()

	target := filepath.Join(t.TempDir(), "easyGit")
	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	if err := updateUnixTo(client, target); err != nil {
		t.Fatalf("updateUnixTo() unexpected error: %v", err)
	}

	if zipRequests != 0 {
		t.Errorf("expected no downloads when already latest, got %d", zipRequests)
	}
	if _, err := os.Stat(target); err == nil {
		t.Error("target should not be created when already latest")
	}
}

func TestUpdateUnixChecksumMismatch(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// 校验表写入错误的哈希（全零），安装包校验必须失败且不安装
	zipBytes := []byte("zip content")
	var zipRequests int
	server := makeFakeReleaseServer(t, zipBytes, make([]byte, 32), &zipRequests)
	defer server.Close()

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "easyGit")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	err := updateUnixTo(client, target)
	if err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	// 错误信息应包含实际与期望 SHA256，便于诊断
	actualSum := sha256.Sum256(zipBytes)
	expectedSum := make([]byte, 32)
	if !strings.Contains(err.Error(), hex.EncodeToString(actualSum[:])) ||
		!strings.Contains(err.Error(), hex.EncodeToString(expectedSum)) {
		t.Errorf("expected actual/expected SHA256 in error, got: %v", err)
	}
	if zipRequests != 1 {
		t.Errorf("expected 1 zip download before abort, got %d", zipRequests)
	}

	// 校验失败后旧版本必须保持原样
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "old binary" {
		t.Errorf("expected old content preserved, got: %s", content)
	}
}

func TestUpdateUnixDownloadFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// zip 下载返回 500，应报下载失败且不安装
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			fmt.Fprint(w, `{"tag_name":"v9.9.9"}`)
		case strings.HasSuffix(r.URL.Path, ".zip"):
			http.Error(w, "boom", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	target := filepath.Join(t.TempDir(), "easyGit")
	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	err := updateUnixTo(client, target)
	if err == nil {
		t.Fatal("expected download failure error")
	}
	if !strings.Contains(err.Error(), i18n.T("update.failed_download_both")) {
		t.Errorf("expected failed_download_both in error, got: %v", err)
	}
	if _, statErr := os.Stat(target); statErr == nil {
		t.Error("target should not be created when download fails")
	}
}

func TestUpdateUnixExtractFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// 安装包是非法 zip：校验通过后解压必须失败且不安装
	zipBytes := []byte("this is not a zip archive")
	sum := sha256.Sum256(zipBytes)
	var zipRequests int
	server := makeFakeReleaseServer(t, zipBytes, sum[:], &zipRequests)
	defer server.Close()

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "easyGit")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	err := updateUnixTo(client, target)
	if err == nil {
		t.Fatal("expected extract failure error")
	}
	if !strings.Contains(err.Error(), i18n.T("update.failed_extract_zip")) {
		t.Errorf("expected failed_extract_zip in error, got: %v", err)
	}

	// 解压失败后旧版本必须保持原样
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "old binary" {
		t.Errorf("expected old content preserved, got: %s", content)
	}
}

func TestUpdateUnixInstallFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("updateUnix 只用于 Unix 更新流程")
	}

	// 构造假 release 安装包（zip 内含 easyGit 二进制）
	var zipBuf bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuf)
	entry, err := zipWriter.Create("easyGit")
	if err != nil {
		t.Fatalf("failed to create zip entry: %v", err)
	}
	if _, err := entry.Write([]byte("new binary")); err != nil {
		t.Fatalf("failed to write zip entry: %v", err)
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	zipBytes := zipBuf.Bytes()
	sum := sha256.Sum256(zipBytes)

	var zipRequests int
	server := makeFakeReleaseServer(t, zipBytes, sum[:], &zipRequests)
	defer server.Close()

	// 安装步骤失败（非权限类 rename 错误，不触发 sudo 回退）
	origRenameFile := renameFile
	renameFile = func(_, _ string) error { return fmt.Errorf("disk full") }
	t.Cleanup(func() { renameFile = origRenameFile })

	tempDir := t.TempDir()
	target := filepath.Join(tempDir, "bin", "easyGit")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("failed to create target dir: %v", err)
	}
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	client := NewReleaseClientWithBaseURL(server.URL, server.URL)
	err = updateUnixTo(client, target)
	if err == nil {
		t.Fatal("expected install failure error")
	}
	if !strings.Contains(err.Error(), i18n.T("update.failed_install_binary")) {
		t.Errorf("expected failed_install_binary in error, got: %v", err)
	}

	// 安装失败后旧版本保持原样，且暂存文件不残留
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "old binary" {
		t.Errorf("expected old content preserved, got: %s", content)
	}
	matches, err := filepath.Glob(filepath.Join(tempDir, "bin", ".easygit-update-*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected no leftover temp files, got: %v", matches)
	}
}
