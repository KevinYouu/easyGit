package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

type FileStatus struct {
	Status string
	Path   string
}

// hasPendingChanges 工作区是否有可提交的改动(pa 据此决定是否跳过提交)。
// -uall:配置 status.showUntrackedFiles=no 时 porcelain 会漏掉未跟踪文件,
// 而 git add -A 仍会暂存它们;--ignore-submodules=dirty:子模块仅内容脏时
// porcelain 报告 M,但 add -A 没有可暂存内容,提交必然失败。
func hasPendingChanges() (bool, error) {
	output, err := exec.Command("git", "status", "--porcelain", "-uall", "--ignore-submodules=dirty").Output()
	if err != nil {
		return false, fmt.Errorf(i18n.T("error.command.execution")+" %w", err)
	}
	return strings.TrimSpace(string(output)) != "", nil
}

// getFileStatuses 获取工作区文件状态(git status --porcelain)。
// 直接执行而非走 spinner:瞬时命令,spinner 只会拖慢响应,
// 保证 ps 启动即显示选择列表。
func getFileStatuses() ([]FileStatus, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(i18n.T("error.command.execution")+" %w", err)
	}

	lines := strings.Split(string(output), "\n")
	var files []FileStatus

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			status := fields[0]
			path := strings.Join(fields[1:], " ")
			files = append(files, FileStatus{Status: status, Path: path})
		}
	}

	return files, nil
}
