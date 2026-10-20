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
// 标签仅供列表展示:表格层按列宽截断,截断不影响任何命令的逻辑;
// squash 的默认合并消息取 GetCommitMessageBodies(完整主题 + 正文)。
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

// GetCommitMessageBodies 按入参顺序返回每条提交的完整信息(主题 + 正文,
// 保留原始换行与空行)。%s 只取主题,合并多条提交时正文会整体丢失,
// squash 需要完整文本拼装默认消息。
// 单次 git show 读取全部提交(逐条调用在 50 条上限下是明显冗余):
// 记录以 NUL 分隔、完整哈希与正文以 0x1f 分隔,正文按哈希前缀回填到入参顺序,
// 不依赖 git 的输出顺序,也不依赖 --no-walk=unsorted(git 2.32+)之类的新参数。
func GetCommitMessageBodies(hashes []string) ([]string, error) {
	if len(hashes) == 0 {
		return nil, nil
	}

	args := append([]string{"show", "-s", "-z", "--format=%H%x1f%B"}, hashes...)
	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return nil, fmt.Errorf("%s %w", i18n.T("error.git.log"), err)
	}

	bodies := make(map[string]string, len(hashes))
	for record := range strings.SplitSeq(string(output), "\x00") {
		fullHash, body, ok := strings.Cut(record, "\x1f")
		if ok {
			bodies[fullHash] = body
		}
	}

	messages := make([]string, 0, len(hashes))
	for _, hash := range hashes {
		body, ok := lookupCommitBody(bodies, hash)
		if !ok {
			return nil, fmt.Errorf("%s %s", i18n.T("error.git.log"), hash)
		}
		messages = append(messages, body)
	}
	return messages, nil
}

// lookupCommitBody 按哈希(完整或前缀)取完整消息
func lookupCommitBody(bodies map[string]string, hash string) (string, bool) {
	if body, ok := bodies[hash]; ok {
		return body, true
	}
	for fullHash, body := range bodies {
		if strings.HasPrefix(fullHash, hash) {
			return body, true
		}
	}
	return "", false
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
