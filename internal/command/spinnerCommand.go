package command

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

// RunCmdWithSpinnerOptions 带加载动画的命令执行（带选项）
func RunCmdWithSpinnerOptions(command string, args []string, loadingMsg, successMsg string, showOutput bool) (string, error) {
	return runWithSpinner(loadingMsg, successMsg, showOutput, true, func() (string, error) {
		output, err := exec.Command(command, args...).CombinedOutput()
		return string(output), err
	})
}

// RunFuncWithSpinnerOptions 带加载动画执行任意任务函数（非 exec 场景，如原生 HTTP 下载）。
// 失败时由调用方负责输出具体错误原因，不重复打印 "Failed: <loadingMsg>"，避免出现无原因的错误行。
func RunFuncWithSpinnerOptions(loadingMsg, successMsg string, task func() error) error {
	_, err := runWithSpinner(loadingMsg, successMsg, false, false, func() (string, error) {
		return "", task()
	})
	return err
}

// runWithSpinner 带加载动画执行任务函数，统一 spinner 生命周期与结果展示。
// printErrorHeader 为 true 时失败打印 "Failed: <loadingMsg>" 头，否则由调用方自行输出错误。
func runWithSpinner(loadingMsg, successMsg string, showOutput, printErrorHeader bool, task func() (string, error)) (string, error) {
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

	// 在goroutine中执行任务；任务 panic 时转为 error 返回，避免主协程在 <-errChan 永久阻塞
	go func() {
		var output string
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%s: %v", i18n.T("ui.task_panicked"), r)
				output = ""
			}
			// 先发送结果，再停止动画
			errChan <- err
			result <- output
			done <- true // 确保动画停止
		}()

		output, err = task()
	}()

	// 等待任务完成
	err := <-errChan
	output := <-result

	// 清除加载动画行
	fmt.Print("\r" + strings.Repeat(" ", len(loadingMsg)+10) + "\r")

	if err != nil {
		// 仅 exec 场景打印失败头；函数场景由调用方输出具体原因，避免重复
		if printErrorHeader {
			fmt.Printf("%s %s\n",
				theme.ErrorIconStyle.Render("✗"),
				theme.ErrorStyle.Render(fmt.Sprintf(i18n.T("progress.failed"), loadingMsg)))
		}

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
