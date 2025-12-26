package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/logs"
)

// RunCmd run a command with args
// command: command
// args: parameters
// successLog: log when success
func RunCmd(command string, args []string, successLog string) (string, error) {
	output, err := exec.Command(command, args...).CombinedOutput()
	if err != nil {
		return string(output), err
	}

	trimmedOutput := strings.TrimSpace(string(output))
	if trimmedOutput != "" {
		fmt.Println(trimmedOutput)
	}
	logs.Success(successLog + "\n")

	return string(output), nil
}
