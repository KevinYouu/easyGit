package gitcmd

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/colors"
	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

type FileStatus struct {
	Status string
	Path   string
}

func statusColor(status string) string {
	switch status {
	case "M":
		return "yellow"
	case "A":
		return "green"
	case "D":
		return "red"
	case "U":
		return "green"
	case "??":
		return "green"
	default:
		return "white"
	}
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

func Status() error {
	fileStatuss, err := getFileStatuses()
	if err != nil {
		return fmt.Errorf(i18n.T("error.file.status")+" %w", err)
	}

	if len(fileStatuss) == 0 {
		fmt.Println(colors.RenderColor("blue", i18n.T("git.status.no_changes")))
		return nil
	}

	fmt.Println(i18n.T("git.status.title"))
	for _, file := range fileStatuss {
		color := statusColor(file.Status)
		fmt.Println(colors.RenderColor(color, file.Status+" "+file.Path))
	}
	return nil
}
