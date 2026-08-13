package update

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// 安装后二进制的权限：可读写 + 可执行
const binaryMode = 0o755

// renameFile 原子替换的底层入口（测试注入用）
var renameFile = os.Rename

// runSudoCmd 执行单个 sudo 安装步骤（测试注入用）
var runSudoCmd = func(args []string, loadingMsg, successMsg string) error {
	_, err := command.RunCmdWithSpinnerOptions("sudo", args, loadingMsg, successMsg, true)
	return err
}

// 直接安装时暂存文件的命名模式（CreateTemp 要求包含 *）
const tmpBinaryPattern = ".easygit-update-*"

// 直接安装时暂存文件的前缀（sudo 流程拼接唯一后缀用）
const tmpBinaryPrefix = ".easygit-update-"

// getInstallDir 返回默认安装目录（仅当无法从当前进程解析安装路径时使用）
func getInstallDir() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "/opt/homebrew/bin"
		}
		return "/usr/local/bin"
	case "windows":
		// Windows 通过 PowerShell 脚本处理，不返回安装目录
		return ""
	default:
		return "/usr/local/bin"
	}
}

// getPlatformName 返回当前平台对应的发布资产后缀
func getPlatformName() (string, error) {
	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return "windows_amd64", nil
		case "arm64":
			return "windows_arm64", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux_amd64", nil
		case "arm64", "aarch64":
			return "linux_arm64", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "darwin_amd64", nil
		case "arm64":
			return "darwin_arm64", nil
		}
	}
	return "", fmt.Errorf("%s: %s/%s", i18n.T("update.unsupported_platform"), runtime.GOOS, runtime.GOARCH)
}

// resolveInstallTarget 解析当前运行二进制的真实路径作为安装目标。
// 兼容 symlink 与自定义安装位置（如 make install 到 GOPATH/bin），避免新版装到默认目录却不生效。
func resolveInstallTarget() (string, error) {
	exePath, err := os.Executable()
	if err == nil {
		// 解析符号链接，确保覆盖的是实际二进制而非链接本身
		if resolved, resolveErr := filepath.EvalSymlinks(exePath); resolveErr == nil {
			return resolved, nil
		}
		return exePath, nil
	}

	// 无法解析当前进程路径时回退到系统默认目录
	installDir := getInstallDir()
	if installDir != "" {
		return filepath.Join(installDir, "easyGit"), nil
	}
	return "", fmt.Errorf("%s: %w", i18n.T("update.failed_resolve_target"), err)
}

// installBinary 将新二进制原子替换到目标路径。
// 先在目标目录创建暂存文件验证写权限：有权限直接安装，否则自动切换到 sudo 流程。
// 替换通过 rename 原子完成，rename 成功前旧版本始终可用。
func installBinary(source, target string) error {
	dir := filepath.Dir(target)

	// 在目标目录创建暂存文件，同时完成权限探测与原子替换准备
	tmpFile, err := os.CreateTemp(dir, tmpBinaryPattern)
	if err != nil {
		// 无写权限，改用 sudo 安装
		return switchToSudo(source, target)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	// 任一步失败都清理暂存文件，不留残留
	defer os.Remove(tmpPath)

	if err := copyFile(source, tmpPath); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_copy_binary"), err)
	}
	if err := os.Chmod(tmpPath, binaryMode); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_set_permissions"), err)
	}
	// 原子替换：rename 成功前旧版本始终可用
	if err := renameFile(tmpPath, target); err != nil {
		// 目录可写但旧文件不可替换（粘滞位目录、root 拥有的旧二进制等）时回退 sudo，
		// 否则更新在最后一步失败且无法自动补救
		if errors.Is(err, os.ErrPermission) {
			return switchToSudo(source, target)
		}
		return fmt.Errorf("%s: %w", i18n.T("update.failed_rename_binary"), err)
	}

	return nil
}

// switchToSudo 提示需要管理员权限并切换到 sudo 安装流程
func switchToSudo(source, target string) error {
	logs.Warning(i18n.T("update.root_permissions_required"))
	fmt.Println(i18n.T("update.password_prompt_hint"))
	return runSudoInstall(source, target)
}

// copyFile 复制文件内容并保留源文件权限
func copyFile(source, target string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

// runSudoInstall 使用 sudo 在目标目录内完成原子替换（cp → chmod → mv）。
// 任一步失败时暂存文件可能为 root 属主，用户态删除必然失败，须经 sudo 清理。
func runSudoInstall(source, target string) error {
	logs.Info(i18n.T("update.installing_with_sudo"))

	dir := filepath.Dir(target)
	// 目标目录内生成唯一暂存名，保证 mv 在同一文件系统内原子完成
	tmpPath := filepath.Join(dir, fmt.Sprintf("%s%d", tmpBinaryPrefix, os.Getpid()))

	// 清理可能残留的暂存文件
	os.Remove(tmpPath)

	// 安装成功前失败都要清理暂存文件：sudo cp 成功后文件归 root，用户态 Remove 删不掉
	installed := false
	defer func() {
		os.Remove(tmpPath)
		if !installed {
			// 尽力清理；失败时 spinner 已输出错误提示，不阻塞主流程
			runSudoCmd([]string{"rm", "-f", tmpPath},
				i18n.T("update.cleaning_up_temp_file"),
				i18n.T("update.cleaning_up_temp_file"))
		}
	}()

	if err := runSudoCmd([]string{"cp", source, tmpPath},
		i18n.T("update.installing_binary_sudo"),
		i18n.T("update.binary_installed_success")); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_copy_binary"), err)
	}

	if err := runSudoCmd([]string{"chmod", "+x", tmpPath},
		i18n.T("update.setting_executable_permissions"),
		i18n.T("update.permissions_set_success")); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_set_permissions"), err)
	}

	if err := runSudoCmd([]string{"mv", tmpPath, target},
		i18n.T("update.installing_binary_sudo"),
		i18n.T("update.binary_installed_success")); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_rename_binary"), err)
	}

	installed = true
	return nil
}
