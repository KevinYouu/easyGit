package update

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/KevinYouu/easyGit/internal/version"
)

// UpdateSelf 将 easyGit 更新到最新版本。
// Windows 通过本地下载的 PowerShell 脚本完成，Unix 走内置的 下载→校验→解压→原子安装 流程。
func UpdateSelf() error {
	if runtime.GOOS == "windows" {
		return updateWindows()
	}
	return updateUnix()
}

// updateUnix 在 Unix 系统上执行完整更新流程
func updateUnix() error {
	return updateUnixTo(NewReleaseClient(), "")
}

// updateUnixTo 执行 Unix 更新流程并安装到 target；target 为空时自动解析当前二进制路径。
// releaseClient 可注入（测试用）。
func updateUnixTo(releaseClient *ReleaseClient, target string) error {
	fmt.Println(i18n.T("update.checking_latest_version"))

	remoteVersion, err := releaseClient.LatestVersion()
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_get_latest_version"), err)
	}
	fmt.Printf("%s: %s\n", i18n.T("update.latest_version"), remoteVersion)

	// 本地为可比较的发布版本且不低于远程版本时，直接提示已最新
	if cmp, comparable := compareVersions(version.Version, remoteVersion); comparable && cmp >= 0 {
		logs.Success(fmt.Sprintf(i18n.T("update.already_latest"), remoteVersion))
		return nil
	}

	platform, err := getPlatformName()
	if err != nil {
		return err
	}

	assetName := releaseClient.AssetName(remoteVersion, platform)
	zipURL := releaseClient.AssetURL(remoteVersion, assetName)

	tempDir, err := os.MkdirTemp("", "easygit-update-*")
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_create_temp_dir"), err)
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, assetName)

	// 下载安装包（原生 HTTP，带 spinner）
	err = command.RunFuncWithSpinnerOptions(
		i18n.T("update.downloading_latest_release"),
		i18n.T("update.download_completed"),
		func() error {
			return releaseClient.DownloadFile(zipURL, zipPath)
		})
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_download_both"), err)
	}

	// 下载校验文件并验证安装包完整性，防止下载被篡改或截断
	err = command.RunFuncWithSpinnerOptions(
		i18n.T("update.downloading_checksum"),
		i18n.T("update.checksum_verified"),
		func() error {
			checksums, fetchErr := releaseClient.Checksums(remoteVersion)
			if fetchErr != nil {
				return fmt.Errorf("%s: %w", i18n.T("update.failed_get_checksums"), fetchErr)
			}
			// 校验失败的错误包含实际与期望哈希，便于诊断
			return verifyChecksum(checksums, assetName, zipPath)
		})
	if err != nil {
		return err
	}

	// 解压安装包
	fmt.Println(i18n.T("update.extracting_file"))
	if err := ExtractZip(zipPath, tempDir); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_extract_zip"), err)
	}

	// 解析当前运行二进制的真实路径作为安装目标，原子替换
	if target == "" {
		target, err = resolveInstallTarget()
		if err != nil {
			return fmt.Errorf("%s: %w", i18n.T("update.failed_resolve_target"), err)
		}
	}
	extractedBinary := filepath.Join(tempDir, "easyGit")

	fmt.Printf("%s: %s\n", i18n.T("update.installing_to"), target)

	if err := installBinary(extractedBinary, target); err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_install_binary"), err)
	}

	logs.Success(i18n.T("update.completed_successfully"))
	fmt.Println(i18n.T("update.restart_terminal_hint"))

	return nil
}

// updateWindows 下载安装脚本到本地临时目录后执行，避免远程管道执行（iwr | iex）无法审计
func updateWindows() error {
	logs.Info(i18n.T("update.running_windows_script"))

	tempDir, err := os.MkdirTemp("", "easygit-update-*")
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_create_temp_dir"), err)
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "install.ps1")
	err = NewReleaseClient().DownloadFile(installScriptURL, scriptPath)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_download_script"), err)
	}

	// RemoteSigned 允许本地文件执行（Go 下载器不写 MOTW 标记，脚本按本地文件处理），
	// 比 Bypass 限制更严，脚本即使被篡改也无法绕过签名策略
	_, err = command.RunCmdWithSpinnerOptions("powershell",
		[]string{"-NoProfile", "-ExecutionPolicy", "RemoteSigned", "-File", scriptPath},
		i18n.T("update.downloading_running_script"),
		i18n.T("update.script_executed_success"), true)
	if err != nil {
		return fmt.Errorf("%s: %w", i18n.T("update.failed_run_script"), err)
	}

	fmt.Println(i18n.T("update.complete_restart_manual"))
	return nil
}
