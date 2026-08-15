package gitcmd

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// GetCurrentRemote 获取当前分支跟踪的远程仓库
func GetCurrentRemote() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	output, err := cmd.Output()
	if err != nil {
		// 如果当前分支没有跟踪远程,返回默认值 "origin"
		return "origin", nil
	}

	// 输出格式: origin/main
	parts := strings.Split(strings.TrimSpace(string(output)), "/")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "origin", nil
}

// GetAllRemotes 获取所有远程仓库列表
func GetAllRemotes() ([]string, error) {
	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("get remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var remotes []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			remotes = append(remotes, line)
		}
	}

	if len(remotes) == 0 {
		return nil, fmt.Errorf("no remotes configured")
	}

	return remotes, nil
}

// SelectRemoteWithConfig 智能选择远程仓库(支持配置持久化和多选)
func SelectRemoteWithConfig() ([]string, bool, error) {
	remotes, err := GetAllRemotes()
	if err != nil {
		logs.Error(i18n.T("error.get.remotes"))
		return nil, false, err
	}

	// 只有一个远程时直接返回,无需保存配置
	if len(remotes) == 1 {
		return remotes, false, nil
	}

	// 检查是否已有配置
	pushConfig, err := config.GetPushConfig()
	if err != nil {
		logs.Error(i18n.T("error.get.push.config"))
	}

	// 如果有配置且远程仍然存在,使用配置
	if pushConfig != nil && len(pushConfig.Remotes) > 0 {
		// 验证配置的远程是否还存在
		validRemotes := []string{}
		for _, configRemote := range pushConfig.Remotes {
			if slices.Contains(remotes, configRemote) {
				validRemotes = append(validRemotes, configRemote)
			}
		}

		if len(validRemotes) > 0 {
			return validRemotes, false, nil
		}

		// 配置的远程都不存在了,清除配置
		config.ClearPushConfig()
	}

	// 获取当前默认远程
	currentRemote, _ := GetCurrentRemote()

	// 将当前远程移到列表最前面,并转换为 Option 数组
	var options []string
	for _, remote := range remotes {
		if remote == currentRemote {
			// 当前远程放在最前面
			options = append([]string{remote}, options...)
		} else {
			options = append(options, remote)
		}
	}

	// 使用多选表单
	selectedRemotes, err := form.ListForm(i18n.T("git.select.remotes.first"), form.StringOptions(options), form.ListMulti)
	if err != nil {
		return nil, false, err
	}

	if len(selectedRemotes) == 0 {
		return nil, false, fmt.Errorf("no remote selected")
	}

	return selectedRemotes, true, nil
}

// GetRemoteBranches 获取指定远程的分支列表
func GetRemoteBranches(remote string) ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("get remote branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	prefix := remote + "/"

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, prefix); ok {
			// 移除 "origin/" 前缀
			branch := after
			// 跳过 HEAD
			if !strings.Contains(branch, "HEAD") {
				branches = append(branches, branch)
			}
		}
	}

	return branches, nil
}
