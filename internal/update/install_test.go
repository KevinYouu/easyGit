package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

func TestResolveInstallTarget(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("resolveInstallTarget 不用于 Windows 更新流程")
	}

	target, err := resolveInstallTarget()
	if err != nil {
		t.Fatalf("resolveInstallTarget() unexpected error: %v", err)
	}

	// 测试运行中的二进制路径应被解析为安装目标
	exePath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}
	if resolved, resolveErr := filepath.EvalSymlinks(exePath); resolveErr == nil {
		exePath = resolved
	}
	if target != exePath {
		t.Errorf("expected install target %s, got: %s", exePath, target)
	}

	// 目标路径应为绝对路径且文件名存在
	if !filepath.IsAbs(target) {
		t.Errorf("expected absolute install target, got: %s", target)
	}
	if _, statErr := os.Stat(target); statErr != nil {
		t.Errorf("install target should exist: %v", statErr)
	}
}

func TestInstallBinary(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("installBinary 只用于 Unix 更新流程")
	}

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	target := filepath.Join(tempDir, "easyGit")

	// 构造新版二进制（模拟解压产物）
	newContent := []byte("new binary version")
	if err := os.WriteFile(source, newContent, 0o755); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	// 构造旧版二进制（模拟已安装的旧版本）
	oldContent := []byte("old binary version")
	if err := os.WriteFile(target, oldContent, 0o644); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	if err := installBinary(source, target); err != nil {
		t.Fatalf("installBinary() unexpected error: %v", err)
	}

	// 目标内容应为新版
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read installed binary: %v", err)
	}
	if string(content) != string(newContent) {
		t.Errorf("expected new content, got: %s", content)
	}

	// 目标应具备可执行权限
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("failed to stat installed binary: %v", err)
	}
	if info.Mode()&0o111 == 0 {
		t.Error("installed binary should be executable")
	}
}

func TestInstallBinaryFailureLeavesOldVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("installBinary 只用于 Unix 更新流程")
	}

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	target := filepath.Join(tempDir, "easyGit")

	// 源文件不存在，安装必然失败
	if err := os.WriteFile(target, []byte("old binary version"), 0o644); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	if err := installBinary(source, target); err == nil {
		t.Fatal("expected error when source does not exist")
	}

	// 失败后旧版本必须仍然可用
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("old binary should still exist: %v", err)
	}
	if string(content) != "old binary version" {
		t.Errorf("expected old content preserved, got: %s", content)
	}

	// 暂存文件不应残留
	matches, err := filepath.Glob(filepath.Join(tempDir, ".easygit-update-*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected no leftover temp files, got: %v", matches)
	}
}

func TestCopyFilePreservesMode(t *testing.T) {
	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	target := filepath.Join(tempDir, "target")

	if err := os.WriteFile(source, []byte("content"), 0o750); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	if err := copyFile(source, target); err != nil {
		t.Fatalf("copyFile() unexpected error: %v", err)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("failed to stat target: %v", err)
	}
	if info.Mode().Perm() != 0o750 {
		t.Errorf("expected mode 0750 preserved, got: %o", info.Mode().Perm())
	}
}

// stubSudoCmd 将 sudo 执行器替换为记录调用参数的桩，返回还原函数
func stubSudoCmd(t *testing.T, failAt int) (*[][]string, func()) {
	t.Helper()

	var calls [][]string
	origRunSudoCmd := runSudoCmd
	runSudoCmd = func(args []string, _, _ string) error {
		calls = append(calls, args)
		if failAt > 0 && len(calls) == failAt {
			return fmt.Errorf("stub sudo failure")
		}
		return nil
	}
	return &calls, func() { runSudoCmd = origRunSudoCmd }
}

// stubSudoCmdExec 将 sudo 执行器替换为模拟真实文件操作的桩（cp/chmod/mv/rm 落到真实文件系统），
// 用于验证失败路径的清理行为；第 failAt 次调用返回错误。
func stubSudoCmdExec(t *testing.T, failAt int) (*[][]string, func()) {
	t.Helper()

	var calls [][]string
	origRunSudoCmd := runSudoCmd
	runSudoCmd = func(args []string, _, _ string) error {
		calls = append(calls, args)
		if failAt > 0 && len(calls) == failAt {
			return fmt.Errorf("stub sudo failure")
		}
		switch args[0] {
		case "cp":
			return copyFile(args[1], args[2])
		case "chmod":
			return os.Chmod(args[2], binaryMode)
		case "mv":
			return os.Rename(args[1], args[2])
		case "rm":
			return os.Remove(args[1])
		}
		return nil
	}
	return &calls, func() { runSudoCmd = origRunSudoCmd }
}

func TestInstallBinarySudoFallbackOnNoWritePermission(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("installBinary 只用于 Unix 更新流程")
	}
	if os.Geteuid() == 0 {
		t.Skip("root 不受目录写权限限制，跳过")
	}

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}

	// 目标目录只读，CreateTemp 探测失败 → 自动切 sudo 流程
	binDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}
	target := filepath.Join(binDir, "easyGit")
	if err := os.Chmod(binDir, 0o500); err != nil {
		t.Fatalf("failed to chmod bin dir: %v", err)
	}
	// 先于 TempDir 清理恢复写权限，保证测试目录可被删除
	t.Cleanup(func() { os.Chmod(binDir, 0o700) })

	calls, restore := stubSudoCmd(t, 0)
	defer restore()

	if err := installBinary(source, target); err != nil {
		t.Fatalf("installBinary() unexpected error: %v", err)
	}

	// 应依次执行 sudo cp / chmod / mv
	if len(*calls) != 3 {
		t.Fatalf("expected 3 sudo calls, got %d: %v", len(*calls), *calls)
	}
	cpArgs := (*calls)[0]
	if len(cpArgs) != 3 || cpArgs[0] != "cp" || cpArgs[1] != source {
		t.Errorf("unexpected cp args: %v", cpArgs)
	}
	// 暂存文件必须位于目标目录内，保证 mv 原子性
	if filepath.Dir(cpArgs[2]) != binDir {
		t.Errorf("expected sudo temp file in %s, got: %s", binDir, cpArgs[2])
	}
	if mvArgs := (*calls)[2]; mvArgs[0] != "mv" || mvArgs[2] != target {
		t.Errorf("unexpected mv args: %v", mvArgs)
	}
}

func TestInstallBinarySudoFallbackOnRenamePermissionError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("installBinary 只用于 Unix 更新流程")
	}

	// 目录可写但旧二进制不可替换（粘滞位/属主不同）时 rename 返回权限错误，应回退 sudo
	origRenameFile := renameFile
	renameFile = func(_, _ string) error { return os.ErrPermission }
	t.Cleanup(func() { renameFile = origRenameFile })

	calls, restore := stubSudoCmd(t, 0)
	defer restore()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	target := filepath.Join(tempDir, "easyGit")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	if err := installBinary(source, target); err != nil {
		t.Fatalf("installBinary() unexpected error: %v", err)
	}

	if len(*calls) != 3 {
		t.Errorf("expected sudo fallback with 3 calls, got %d: %v", len(*calls), *calls)
	}
	// 直接安装失败时旧版本必须保持原样（桩未实际替换）
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read target: %v", err)
	}
	if string(content) != "old binary" {
		t.Errorf("expected old content preserved, got: %s", content)
	}
}

func TestInstallBinaryRenameErrorNoSudoFallback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("installBinary 只用于 Unix 更新流程")
	}

	// 非权限类 rename 错误不应触发 sudo，直接返回失败
	origRenameFile := renameFile
	renameFile = func(_, _ string) error { return fmt.Errorf("target is a directory") }
	t.Cleanup(func() { renameFile = origRenameFile })

	calls, restore := stubSudoCmd(t, 0)
	defer restore()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	target := filepath.Join(tempDir, "easyGit")
	if err := os.WriteFile(target, []byte("old binary"), 0o755); err != nil {
		t.Fatalf("failed to write old binary: %v", err)
	}

	if err := installBinary(source, target); err == nil {
		t.Fatal("expected error for non-permission rename failure")
	}
	if len(*calls) != 0 {
		t.Errorf("expected no sudo calls, got %d: %v", len(*calls), *calls)
	}
}

func TestRunSudoInstallFailureCleansUp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("runSudoInstall 只用于 Unix 更新流程")
	}

	// mv 步骤失败（第 3 次调用）时，sudo cp 已生成 root 属主暂存文件，
	// 失败路径必须追加一次 sudo rm -f 清理，不留残留
	calls, restore := stubSudoCmdExec(t, 3)
	defer restore()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	binDir := filepath.Join(tempDir, "bin")
	target := filepath.Join(binDir, "easyGit")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	err := runSudoInstall(source, target)
	if err == nil {
		t.Fatal("expected error when mv step fails")
	}
	if !strings.Contains(err.Error(), i18n.T("update.failed_rename_binary")) {
		t.Errorf("expected failed_rename_binary in error, got: %v", err)
	}

	// 共 4 次调用：cp / chmod / mv / rm，最后一次为清理
	if len(*calls) != 4 {
		t.Fatalf("expected 4 sudo calls including cleanup, got %d: %v", len(*calls), *calls)
	}
	cleanupArgs := (*calls)[3]
	if len(cleanupArgs) != 3 || cleanupArgs[0] != "rm" || cleanupArgs[1] != "-f" {
		t.Errorf("expected sudo rm -f cleanup call, got: %v", cleanupArgs)
	}
	if cleanupArgs[2] != filepath.Join(binDir, fmt.Sprintf("%s%d", tmpBinaryPrefix, os.Getpid())) {
		t.Errorf("expected cleanup to target sudo temp file, got: %s", cleanupArgs[2])
	}

	// 模拟文件系统下 mv 失败后暂存文件应被 rm 步骤删除
	matches, err := filepath.Glob(filepath.Join(binDir, tmpBinaryPrefix+"*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected no leftover temp files, got: %v", matches)
	}
}

func TestRunSudoInstallCpFailureCleansUp(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("runSudoInstall 只用于 Unix 更新流程")
	}

	// cp 步骤失败（第 1 次调用）时提前返回，同样应触发清理（rm -f 对不存在的文件无副作用）
	calls, restore := stubSudoCmdExec(t, 1)
	defer restore()

	tempDir := t.TempDir()
	source := filepath.Join(tempDir, "source")
	if err := os.WriteFile(source, []byte("new binary"), 0o755); err != nil {
		t.Fatalf("failed to write source: %v", err)
	}
	binDir := filepath.Join(tempDir, "bin")
	target := filepath.Join(binDir, "easyGit")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("failed to create bin dir: %v", err)
	}

	err := runSudoInstall(source, target)
	if err == nil {
		t.Fatal("expected error when cp step fails")
	}
	if !strings.Contains(err.Error(), i18n.T("update.failed_copy_binary")) {
		t.Errorf("expected failed_copy_binary in error, got: %v", err)
	}

	// 共 2 次调用：cp / rm
	if len(*calls) != 2 {
		t.Fatalf("expected 2 sudo calls including cleanup, got %d: %v", len(*calls), *calls)
	}
	if (*calls)[1][0] != "rm" {
		t.Errorf("expected cleanup call after cp failure, got: %v", (*calls)[1])
	}

	// 目标目录不应有残留
	matches, err := filepath.Glob(filepath.Join(binDir, tmpBinaryPrefix+"*"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected no leftover temp files, got: %v", matches)
	}
}
