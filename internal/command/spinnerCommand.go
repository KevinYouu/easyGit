package command

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
	"github.com/charmbracelet/lipgloss"
)

// RunCmdWithSpinnerOptions 带加载动画的命令执行（带选项）
func RunCmdWithSpinnerOptions(command string, args []string, loadingMsg, successMsg string, showOutput bool) (string, error) {
	// 创建加载动画的channel
	done := make(chan bool)
	result := make(chan string, 1) // 添加缓冲避免阻塞
	errChan := make(chan error, 1) // 添加缓冲避免阻塞

	// 启动加载动画
	go func() {
		frames := theme.GetSpinnerFrames()
		style := theme.GetSpinnerStyle()
		frameIndex := 0
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				fmt.Printf("\r%s %s",
					style.Render(frames[frameIndex]),
					theme.InfoStyle.Render(loadingMsg))
				frameIndex = (frameIndex + 1) % len(frames)
			}
		}
	}()

	// 在goroutine中执行命令
	go func() {
		defer func() {
			done <- true // 确保动画停止
		}()

		output, err := exec.Command(command, args...).CombinedOutput()

		// 先发送结果，再停止动画
		errChan <- err
		result <- string(output)
	}()

	// 等待命令完成
	err := <-errChan
	output := <-result

	// 清除加载动画行
	fmt.Print("\r" + strings.Repeat(" ", len(loadingMsg)+10) + "\r")

	if err != nil {
		fmt.Printf("%s %s\n",
			theme.ErrorIconStyle.Render("✗"),
			theme.ErrorStyle.Render(fmt.Sprintf("Failed: %s", loadingMsg)))

		// 显示详细的错误输出
		trimmedOutput := strings.TrimSpace(output)
		if trimmedOutput != "" {
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.PrimaryColor).
					Bold(true).
					Render(i18n.T("ui.error.details")))
			fmt.Printf("%s\n",
				lipgloss.NewStyle().
					Foreground(theme.MutedForeground).
					Render(trimmedOutput))
		}

		return output, err
	}

	// 显示成功信息
	fmt.Printf("%s %s\n",
		theme.SuccessIconStyle.Render("✓"),
		theme.SuccessStyle.Render(successMsg))

	// 如果有输出内容且需要显示，显示它
	trimmedOutput := strings.TrimSpace(output)
	if showOutput && trimmedOutput != "" {
		fmt.Println(lipgloss.NewStyle().Foreground(theme.MutedForeground).Render(trimmedOutput))
	}

	return output, nil
}

// CommandInfo 命令信息结构
type CommandInfo struct {
	Command     string
	Args        []string
	Description string
	LoadingMsg  string
	SuccessMsg  string
}
