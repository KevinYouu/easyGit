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
