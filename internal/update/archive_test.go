package update

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// createTestZip 创建包含指定条目的 zip 文件，返回文件路径
func createTestZip(t *testing.T, entries map[string]string) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "test.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	for name, content := range entries {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %s: %v", name, err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write zip entry %s: %v", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	return zipPath
}

func TestExtractZip(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"easyGit":        "binary content",
		"LICENSE":        "MIT license",
		"docs/readme.md": "readme content",
	})

	dest := t.TempDir()
	if err := ExtractZip(zipPath, dest); err != nil {
		t.Fatalf("ExtractZip() unexpected error: %v", err)
	}

	// 验证根目录文件
	content, err := os.ReadFile(filepath.Join(dest, "easyGit"))
	if err != nil {
		t.Fatalf("failed to read extracted binary: %v", err)
	}
	if string(content) != "binary content" {
		t.Errorf("unexpected binary content: %s", content)
	}

	// 验证子目录文件
	content, err = os.ReadFile(filepath.Join(dest, "docs", "readme.md"))
	if err != nil {
		t.Fatalf("failed to read extracted nested file: %v", err)
	}
	if string(content) != "readme content" {
		t.Errorf("unexpected nested file content: %s", content)
	}
}

func TestExtractZipRejectPathTraversal(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"../evil.sh": "rm -rf /",
	})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for path traversal entry")
	}

	// 确认恶意文件没有写到 dest 目录内（Clean 后的路径应为 dest/evil.sh）
	if _, statErr := os.Stat(filepath.Join(dest, "evil.sh")); statErr == nil {
		t.Error("path traversal file should not be extracted")
	}
}

func TestExtractZipRejectAbsolutePath(t *testing.T) {
	zipPath := createTestZip(t, map[string]string{
		"/tmp/evil.sh": "malicious",
	})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for absolute path entry")
	}

	if _, statErr := os.Stat("/tmp/evil.sh"); statErr == nil {
		t.Error("absolute path entry should not be extracted")
	}
}

func TestExtractZipRejectNestedTraversal(t *testing.T) {
	// 深层目录逃逸：a/../../evil.sh 解析后落在目标目录外
	zipPath := createTestZip(t, map[string]string{
		"a/../../evil.sh": "malicious",
	})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for nested traversal entry")
	}

	// Clean 后的目标应为 dest 目录外，不会被写出
	if _, statErr := os.Stat(filepath.Join(dest, "evil.sh")); statErr == nil {
		t.Error("nested traversal file should not be extracted")
	}
}

func TestExtractZipRejectWindowsDriveLetter(t *testing.T) {
	// Windows 盘符路径（C:evil）在 Unix 上不是绝对路径，filepath.IsAbs 判不出，
	// 必须显式拒绝，否则跨平台运行时 Join 会按卷名规则解析逃逸目标目录
	zipPath := createTestZip(t, map[string]string{
		"C:evil": "malicious",
	})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for Windows drive letter entry")
	}
	if _, statErr := os.Stat(filepath.Join(dest, "C:evil")); statErr == nil {
		t.Error("drive letter entry should not be extracted")
	}
}

func TestExtractZipRejectTooManyEntries(t *testing.T) {
	// 条目数量超上限应直接拒绝（zip bomb 防护）
	origCount := maxEntryCount
	maxEntryCount = 2
	t.Cleanup(func() { maxEntryCount = origCount })

	zipPath := createTestZip(t, map[string]string{"a": "1", "b": "2", "c": "3"})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for too many entries")
	}
	if !strings.Contains(err.Error(), i18n.T("update.too_many_entries")) {
		t.Errorf("expected too_many_entries error, got: %v", err)
	}
}

func TestExtractZipRejectOversizedEntryPrecheck(t *testing.T) {
	// 单条目声明的解压大小超上限，读取前即拒绝
	origSize := maxEntrySize
	maxEntrySize = 4
	t.Cleanup(func() { maxEntrySize = origSize })

	zipPath := createTestZip(t, map[string]string{"big.bin": "12345"})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error for oversized entry")
	}
	if !strings.Contains(err.Error(), i18n.T("update.entry_too_large")) {
		t.Errorf("expected entry_too_large error, got: %v", err)
	}
}

func TestExtractZipRejectTotalSizeTooLarge(t *testing.T) {
	// 单条目未超限但累计解压总大小超上限应拒绝
	origSize, origTotal := maxEntrySize, maxTotalSize
	maxEntrySize = 16
	maxTotalSize = 10
	t.Cleanup(func() { maxEntrySize, maxTotalSize = origSize, origTotal })

	zipPath := createTestZip(t, map[string]string{"a.bin": "12345678", "b.bin": "12345678"})

	dest := t.TempDir()
	err := ExtractZip(zipPath, dest)
	if err == nil {
		t.Fatal("expected error when total extracted size exceeds limit")
	}
	if !strings.Contains(err.Error(), i18n.T("update.archive_too_large")) {
		t.Errorf("expected archive_too_large error, got: %v", err)
	}
}
