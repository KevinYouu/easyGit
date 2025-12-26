package gitcmd

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

type FileStatus struct {
	Status string
	Path   string
}

func getFileStatuses() ([]FileStatus, error) {
	output, err := command.RunCmdWithSpinnerOptions("git", []string{"status", "--porcelain"}, i18n.T("progress.loading"), i18n.T("success.step.complete"), false)
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
