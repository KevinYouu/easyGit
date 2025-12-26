package update

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

const (
	repoOwner = "KevinYouu"
	repoName  = "fastGit"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return release.TagName, nil
}

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
	return "", fmt.Errorf(i18n.T("update.unsupported_platform")+": %s/%s", runtime.GOOS, runtime.GOARCH)
}

func getInstallDir() string {
	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return "/opt/homebrew/bin"
		}
		return "/usr/local/bin"
	case "linux":
		return "/usr/local/bin"
	case "windows":
		// Windows 通过 PowerShell 脚本处理
		return ""
	}
	return "/usr/local/bin"
}

func UpdateSelf() error {
	// Windows 使用 PowerShell 脚本
	if runtime.GOOS == "windows" {
		fmt.Println(i18n.T("update.running_windows_script"))
		_, err := command.RunCmdWithSpinnerOptions("powershell",
			[]string{"-Command", "iwr -useb https://raw.githubusercontent.com/KevinYouu/fastGit/main/install.ps1 | iex"},
			i18n.T("update.downloading_running_script"),
			i18n.T("update.script_executed_success"), true)
		if err != nil {
			return fmt.Errorf(i18n.T("update.failed_run_script")+": %w", err)
		}
		fmt.Println(i18n.T("update.complete_restart_manual"))
		return nil
	}

	// Unix 系统（Linux, macOS）使用改进的更新流程
	return updateUnix()
}

func updateUnix() error {
	fmt.Println(i18n.T("update.checking_latest_version"))

	version, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf(i18n.T("update.failed_get_latest_version")+": %w", err)
	}
	fmt.Printf(i18n.T("update.latest_version")+": %s\n", version)

	platform, err := getPlatformName()
	if err != nil {
		return err
	}

	assetName := fmt.Sprintf("fastGit_%s_%s.zip", version, platform)
	url := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", repoOwner, repoName, version, assetName)

	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fastgit-update-*")
	if err != nil {
		return fmt.Errorf(i18n.T("update.failed_create_temp_dir")+": %w", err)
	}
	defer os.RemoveAll(tempDir)

	zipPath := filepath.Join(tempDir, assetName)

	// 使用命令执行器下载文件
	commands := []command.CommandInfo{
		{
			Command:     "curl",
			Args:        []string{"-L", "-o", zipPath, url},
			Description: i18n.T("update.downloading_latest_release"),
			LoadingMsg:  fmt.Sprintf(i18n.T("update.downloading_asset"), assetName),
			SuccessMsg:  i18n.T("update.download_completed"),
		},
	}

	err = command.RunMultipleCommands(commands)
	if err != nil {
		// 如果 curl 失败，尝试使用 wget
		fmt.Println(i18n.T("update.curl_failed_try_wget"))
		commands[0].Command = "wget"
		commands[0].Args = []string{"-O", zipPath, url}
		commands[0].LoadingMsg = fmt.Sprintf(i18n.T("update.downloading_with_wget"), assetName)

		err = command.RunMultipleCommands(commands)
		if err != nil {
			return fmt.Errorf(i18n.T("update.failed_download_both")+": %w", err)
		}
	}

	// 解压文件
	fmt.Println(i18n.T("update.extracting_file"))
	if err := extractZip(zipPath, tempDir); err != nil {
		return fmt.Errorf(i18n.T("update.failed_extract_zip")+": %w", err)
	}

	// 安装文件
	installDir := getInstallDir()
	extractedBinary := filepath.Join(tempDir, "fastGit")
	targetPath := filepath.Join(installDir, "fastGit")

	fmt.Printf(i18n.T("update.installing_to")+": %s...\n", installDir)

	// 检查目标文件是否存在，如果存在先尝试删除以测试权限
	if _, err := os.Stat(targetPath); err == nil {
		// 文件存在，尝试移除以测试权限
		if err := os.Remove(targetPath); err != nil {
			fmt.Println(i18n.T("update.root_permissions_required"))
			fmt.Println(i18n.T("update.password_prompt_hint"))
			err = runSudoInstall(extractedBinary, targetPath)
		} else {
			// 能够删除，说明有权限，直接安装
			fmt.Println(i18n.T("update.direct_install_sufficient_permissions"))
			err = runDirectInstall(extractedBinary, targetPath)
		}
	} else {
		// 文件不存在，尝试创建测试文件来检查权限
		if hasWritePermission(installDir) {
			fmt.Println(i18n.T("update.direct_install_no_sudo"))
			err = runDirectInstall(extractedBinary, targetPath)
		} else {
			fmt.Println(i18n.T("update.root_permissions_required"))
			fmt.Println(i18n.T("update.password_prompt_hint"))
			err = runSudoInstall(extractedBinary, targetPath)
		}
	}

	if err != nil {
		return fmt.Errorf(i18n.T("update.failed_install_binary")+": %w", err)
	}

	fmt.Println(i18n.T("update.completed_successfully"))
	fmt.Println(i18n.T("update.restart_terminal_hint"))

	return nil
}

func extractZip(src, dest string) error {
	reader, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(dest, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.FileInfo().Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

// hasWritePermission 检查是否对目录有写权限
func hasWritePermission(dir string) bool {
	testFile := filepath.Join(dir, ".fastgit-write-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// runDirectInstall 直接安装（无需 sudo）
func runDirectInstall(source, target string) error {
	installCommands := []command.CommandInfo{
		{
			Command:     "cp",
			Args:        []string{source, target},
			Description: i18n.T("update.installing_binary_system"),
			LoadingMsg:  i18n.T("update.installing_binary"),
			SuccessMsg:  i18n.T("update.binary_installed_success"),
		},
		{
			Command:     "chmod",
			Args:        []string{"+x", target},
			Description: i18n.T("update.setting_executable_permissions"),
			LoadingMsg:  i18n.T("update.setting_permissions"),
			SuccessMsg:  i18n.T("update.permissions_set_success"),
		},
	}
	return command.RunMultipleCommands(installCommands)
}

// runSudoInstall 使用交互式 sudo 安装
func runSudoInstall(source, target string) error {
	fmt.Println(i18n.T("update.installing_with_sudo"))

	// 复制文件
	_, err := command.RunCmdWithSpinnerOptions("sudo",
		[]string{"cp", source, target},
		i18n.T("update.installing_binary_sudo"),
		i18n.T("update.binary_installed_success"), true)
	if err != nil {
		return fmt.Errorf(i18n.T("update.failed_copy_binary")+": %w", err)
	}

	// 设置权限
	_, err = command.RunCmdWithSpinnerOptions("sudo",
		[]string{"chmod", "+x", target},
		i18n.T("update.setting_executable_permissions"),
		i18n.T("update.permissions_set_success"), true)
	if err != nil {
		return fmt.Errorf(i18n.T("update.failed_set_permissions")+": %w", err)
	}

	return nil
}
