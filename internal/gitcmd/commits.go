package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

// Commit 一条提交记录:列表展示与内部元数据。
// reset/squash/drop/cherry-pick 共用同一类型,统一数据源。
type Commit struct {
	Hash      string
	Message   string
	Date      string
	Author    string
	Email     string
	IsHead    bool
	Timestamp time.Time
}

// GetCommitsOptions 获取当前分支提交历史(HEAD 在最前):
// limit > 0 时仅返回最近 limit 条,否则全部。
// 返回选项(两行式标签)与提交记录(含邮箱/完整时间戳)。
// reset/squash/drop 共用数据源,避免各命令重复解析 git log。
// 提交消息保持完整不截断:表格展示层按列宽截断,而 squash 等命令
// 从标签提取默认消息,截断会导致获取到残缺文本。
func GetCommitsOptions(limit int) ([]config.Option, []Commit, error) {
	args := []string{"log", "--pretty=format:%h|%s|%ad|%an|%ae|%ai", "--date=format:%m-%d %H:%M"}
	if limit > 0 {
		args = append(args, "-n", fmt.Sprintf("%d", limit))
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("%s %w", i18n.T("error.git.log"), err)
	}

	lines := strings.Split(string(output), "\n")
	var options = make([]config.Option, 0, len(lines))
	var commits = make([]Commit, 0, len(lines))

	for i, line := range lines {
		parts := strings.Split(line, "|")
		if len(parts) < 5 {
			continue
		}
		commit := Commit{
			Hash:    parts[0],
			Message: parts[1],
			Date:    parts[2],
			Author:  parts[3],
			Email:   parts[4],
			IsHead:  i == 0,
		}

		// 解析 ISO 时间戳(%ai)用于排序等场景
		if len(parts) >= 6 {
			if ts, err := time.Parse("2006-01-02 15:04:05 -0700", parts[5]); err == nil {
				commit.Timestamp = ts
			}
		}

		commits = append(commits, commit)
		options = append(options, config.Option{
			Label: fmt.Sprintf("%s %s\n%s • %s", commit.Hash, commit.Message, commit.Date, commit.Author),
			Value: commit.Hash,
		})
	}

	if len(commits) == 0 {
		return nil, nil, fmt.Errorf("%s", i18n.T("rebase.not.enough.commits"))
	}
	return options, commits, nil
}

// GetRecentCommits 获取最近 50 条提交,返回选项与哈希列表
// (squash/drop 使用,保持既有签名)。
func GetRecentCommits() ([]config.Option, []string, error) {
	options, commits, err := GetCommitsOptions(50)
	if err != nil {
		return nil, nil, err
	}

	hashes := make([]string, len(commits))
	for i, c := range commits {
		hashes[i] = c.Hash
	}
	return options, hashes, nil
}
