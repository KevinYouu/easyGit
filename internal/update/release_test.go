package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	content := "abc123def  easyGit_v0.2.5_linux_amd64.zip\n" +
		"def456abc  easyGit_v0.2.5_darwin_arm64.zip\n" +
		"invalid line without hash\n" +
		"\n"
	checksums := parseChecksums(content)

	if len(checksums) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(checksums))
	}
	if checksums["easyGit_v0.2.5_linux_amd64.zip"] != "abc123def" {
		t.Errorf("unexpected checksum for linux asset: %s", checksums["easyGit_v0.2.5_linux_amd64.zip"])
	}
	if checksums["easyGit_v0.2.5_darwin_arm64.zip"] != "def456abc" {
		t.Errorf("unexpected checksum for darwin asset: %s", checksums["easyGit_v0.2.5_darwin_arm64.zip"])
	}
}

func TestVerifyChecksum(t *testing.T) {
	tempDir := t.TempDir()
	assetName := "easyGit_v0.2.5_linux_amd64.zip"
	filePath := filepath.Join(tempDir, assetName)

	// 构造测试文件并计算其真实 SHA256
	content := []byte("fake release archive content")
	if err := os.WriteFile(filePath, content, 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	sum := sha256.Sum256(content)
	expected := hex.EncodeToString(sum[:])

	checksums := map[string]string{assetName: expected}
	if err := verifyChecksum(checksums, assetName, filePath); err != nil {
		t.Errorf("expected checksum verification to pass, got: %v", err)
	}

	// 错误的校验和应该失败，且错误信息包含实际与期望哈希
	badHash := strings.Repeat("0", 64)
	badChecksums := map[string]string{assetName: badHash}
	err := verifyChecksum(badChecksums, assetName, filePath)
	if err == nil {
		t.Error("expected checksum verification to fail with wrong hash")
	}
	if !strings.Contains(err.Error(), expected) || !strings.Contains(err.Error(), badHash) {
		t.Errorf("expected both hashes in error message, got: %v", err)
	}

	// 校验表中缺失资产名应该失败
	if err := verifyChecksum(map[string]string{"other.zip": expected}, assetName, filePath); err == nil {
		t.Error("expected checksum verification to fail when asset is missing")
	}
}

func TestReleaseClientLatestVersion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/KevinYouu/easyGit/releases/latest" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header to be set")
		}
		fmt.Fprint(w, `{"tag_name":"v0.2.5"}`)
	}))
	defer server.Close()

	client := newReleaseClient(server.URL, server.URL)
	version, err := client.LatestVersion()
	if err != nil {
		t.Fatalf("LatestVersion() unexpected error: %v", err)
	}
	if version != "v0.2.5" {
		t.Errorf("expected v0.2.5, got: %s", version)
	}
}

func TestReleaseClientLatestVersionAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟 GitHub API 限流（403）
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := newReleaseClient(server.URL, server.URL)
	_, err := client.LatestVersion()
	if err == nil {
		t.Fatal("expected error for HTTP 403")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected status code in error message, got: %v", err)
	}
}

func TestReleaseClientLatestVersionMissingTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	}))
	defer server.Close()

	client := newReleaseClient(server.URL, server.URL)
	if _, err := client.LatestVersion(); err == nil {
		t.Fatal("expected error for missing tag_name")
	}
}

func TestReleaseClientDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "release asset bytes")
	}))
	defer server.Close()

	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "asset.zip")

	client := newReleaseClient(server.URL, server.URL)
	if err := client.DownloadFile(server.URL, destPath); err != nil {
		t.Fatalf("DownloadFile() unexpected error: %v", err)
	}

	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != "release asset bytes" {
		t.Errorf("unexpected downloaded content: %s", content)
	}
}

func TestReleaseClientDownloadFileStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := newReleaseClient(server.URL, server.URL)
	err := client.DownloadFile(server.URL, filepath.Join(t.TempDir(), "asset.zip"))
	if err == nil {
		t.Fatal("expected error for HTTP 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("expected status code in error message, got: %v", err)
	}
}

func TestReleaseClientChecksums(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/checksums.txt") {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		fmt.Fprint(w, "abc123  easyGit_v0.2.5_linux_amd64.zip\n")
	}))
	defer server.Close()

	client := newReleaseClient(server.URL, server.URL)
	checksums, err := client.Checksums("v0.2.5")
	if err != nil {
		t.Fatalf("Checksums() unexpected error: %v", err)
	}
	if checksums["easyGit_v0.2.5_linux_amd64.zip"] != "abc123" {
		t.Errorf("unexpected checksum: %s", checksums["easyGit_v0.2.5_linux_amd64.zip"])
	}
}

func TestReleaseClientAssetNameAndURL(t *testing.T) {
	client := NewReleaseClient()
	assetName := client.AssetName("v0.2.5", "linux_amd64")
	if assetName != "easyGit_v0.2.5_linux_amd64.zip" {
		t.Errorf("unexpected asset name: %s", assetName)
	}

	url := client.AssetURL("v0.2.5", assetName)
	expected := "https://github.com/KevinYouu/easyGit/releases/download/v0.2.5/easyGit_v0.2.5_linux_amd64.zip"
	if url != expected {
		t.Errorf("unexpected asset URL: %s", url)
	}
}
