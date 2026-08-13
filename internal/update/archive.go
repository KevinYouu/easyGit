package update

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// 解压时父目录的默认权限
const extractDirMode = 0o755

// 解压安全限制（防 zip bomb），变量形式便于测试注入
var (
	// 单个条目解压大小上限（easyGit 安装包 < 20MB，200MB 已足够宽松）
	maxEntrySize = 200 << 20
	// 全部条目解压总大小上限
	maxTotalSize = 500 << 20
	// 条目数量上限
	maxEntryCount = 1024
)

// ExtractZip 安全解压 zip 文件到目标目录。
// 拒绝绝对路径、Windows 盘符路径与含 .. 的条目，防止 Zip Slip 路径穿越写出目标目录覆盖任意文件；
// 限制条目数量与解压总大小，防止 zip bomb 耗尽磁盘。
func ExtractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 条目数量上限，防止海量小文件拖垮磁盘
	if len(reader.File) > maxEntryCount {
		return fmt.Errorf("%s: %d", i18n.T("update.too_many_entries"), len(reader.File))
	}

	// 累计解压字节数，超上限即中止
	var totalWritten int64

	for _, file := range reader.File {
		// 路径穿越防护：Clean 后仍为绝对路径或以 .. 开头的条目直接拒绝
		cleanName := filepath.Clean(file.Name)
		// Windows 盘符路径（如 C:evil、C:/evil）：filepath.IsAbs 在 Unix 上不识别盘符，
		// 需显式检查，避免跨平台运行时 Join 按卷名规则解析逃逸目标目录
		if filepath.IsAbs(cleanName) || cleanName == ".." ||
			strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) ||
			(len(cleanName) >= 2 && cleanName[1] == ':') {
			return fmt.Errorf("%s: %s", i18n.T("update.unsafe_zip_entry"), file.Name)
		}

		// 单条目声明的解压大小超上限直接拒绝（读取前即可判断）
		if file.UncompressedSize64 > uint64(maxEntrySize) {
			return fmt.Errorf("%s: %s", i18n.T("update.entry_too_large"), file.Name)
		}

		path := filepath.Join(dest, cleanName)

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(path), extractDirMode); err != nil {
			return err
		}

		// 显式关闭句柄，避免循环内 defer 累积导致文件句柄泄漏
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			fileReader.Close()
			return err
		}

		// 哨兵 +1 识别超限：实际写入超过上限即判为 zip bomb
		written, copyErr := io.CopyN(targetFile, fileReader, int64(maxEntrySize)+1)
		closeErr := targetFile.Close()
		fileReader.Close()
		if copyErr != nil && copyErr != io.EOF {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if written > int64(maxEntrySize) {
			return fmt.Errorf("%s: %s", i18n.T("update.entry_too_large"), file.Name)
		}

		totalWritten += written
		if totalWritten > int64(maxTotalSize) {
			return fmt.Errorf("%s", i18n.T("update.archive_too_large"))
		}
	}

	return nil
}
