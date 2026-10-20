package gitcmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/KevinYouu/easyGit/internal/command"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// applyCommitTypeDescriptions 提交类型选项附加单行说明:按 commit.type.desc.<value>
// 查 i18n,查不到时保持纯名称(优雅降级,不改库表)。
func applyCommitTypeDescriptions(options []config.Option) []config.Option {
	for i := range options {
		desc := ""
		if key := "commit.type.desc." + options[i].Value; i18n.Has(key) {
			desc = i18n.T(key)
		}
		options[i].Description = desc
	}
	return options
}

// validateCommitMessage 提交消息校验:须含提交类型前缀(如 "fix: ")且主题非空,
// 防止误 Enter 提交 "fix: " 这类空主题消息。
func validateCommitMessage(msg string) error {
	parts := strings.SplitN(msg, ":", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		return fmt.Errorf("%s", i18n.T("form.input.commit.subject.empty"))
	}
	return nil
}

// composeCommitMessage 拼装主题 + 可选正文:正文内部空行与缩进不折叠,
// 仅去掉首尾空白行(否则空正文会留下悬空空行),CRLF 规范为 LF。
func composeCommitMessage(subject, body string) string {
	subject = strings.TrimRight(subject, " \t\r\n")
	body = trimBlankEdgeLines(strings.ReplaceAll(body, "\r\n", "\n"))
	if body == "" {
		return subject
	}
	return subject + "\n\n" + body
}

// trimBlankEdgeLines 去掉文本首尾的空白行,保留中间空行与行内缩进
func trimBlankEdgeLines(text string) string {
	lines := strings.Split(text, "\n")
	start, end := 0, len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return strings.Join(lines[start:end], "\n")
}

// shouldPullBeforePush 判断推送前是否执行 pull:
// 配置 never 时跳过;当前分支未设上游分支时跳过
// (无参数 git pull 必然失败,提前跳过避免终止整个流程)。
func shouldPullBeforePush() bool {
	if v, err := config.GetPullBeforePush(); err == nil && v == config.PullBeforePushNever {
		return false
	}
	if _, err := exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}").Output(); err != nil {
		logs.Info(i18n.T("push.pull.skipped.no.upstream"))
		return false
	}
	return true
}

// appendPushSteps 追加拉取与并行推送,返回命令列表与并行分段起点(串行步数)。
func appendPushSteps(cmds []command.CommandInfo, remotes []string, branch string, pullBefore bool) ([]command.CommandInfo, int) {
	if pullBefore {
		cmds = append(cmds, command.CommandInfo{
			Command:     "git",
			Args:        []string{"pull"},
			Description: i18n.T("git.pull.description"),
			LoadingMsg:  i18n.T("git.pull.loading"),
			SuccessMsg:  i18n.T("git.pull.success"),
		})
	}
	parallelFrom := len(cmds)
	for _, remote := range remotes {
		cmds = append(cmds, command.CommandInfo{
			Command:     "git",
			Args:        []string{"push", remote, branch},
			Description: fmt.Sprintf(i18n.T("git.push.to.remote"), remote),
			LoadingMsg:  fmt.Sprintf(i18n.T("git.push.loading.remote"), remote),
			SuccessMsg:  fmt.Sprintf(i18n.T("git.push.success.remote"), remote),
		})
	}
	return cmds, parallelFrom
}

// buildPushSteps 组装完整步骤序列:暂存 + 提交(仅有改动时)→ pull(串行)
// → 各远程 push(并行),返回并行分段起点。
// 工作区干净时不生成暂存/提交步骤——git commit 此时必然失败并中断 pull/push。
func buildPushSteps(hasChanges bool, message string, remotes []string, branch string, pullBefore bool) ([]command.CommandInfo, int) {
	var steps []command.CommandInfo
	if hasChanges {
		steps = []command.CommandInfo{
			{
				Command:     "git",
				Args:        []string{"add", "-A"},
				Description: i18n.T("git.add.all.description"),
				LoadingMsg:  i18n.T("git.add.all.loading"),
				SuccessMsg:  i18n.T("git.add.all.success"),
			},
			{
				Command:     "git",
				Args:        []string{"commit", "-m", message},
				Description: i18n.T("git.commit.description"),
				LoadingMsg:  i18n.T("git.commit.loading"),
				SuccessMsg:  i18n.T("git.commit.success"),
			},
		}
	}
	return appendPushSteps(steps, remotes, branch, pullBefore)
}

// commitMessageInputs 收集主题(单行,带历史建议)与可选正文(Enter 跳过),
// 返回完整提交信息与主题。
func commitMessageInputs(defaultSubject string) (message, subject string, err error) {
	subject, err = form.InputWithSuggestions(i18n.T("push.input.commit.message"), defaultSubject, validateCommitMessage, config.GetRecentCommitMessages())
	if err != nil {
		return "", "", fmt.Errorf("input: %w", err)
	}

	body, err := form.TextInputOptional(i18n.T("push.input.commit.body"), "")
	if err != nil {
		return "", "", fmt.Errorf("input body: %w", err)
	}
	return composeCommitMessage(subject, body), subject, nil
}

func PushAll() error {
	hasChanges, err := hasPendingChanges()
	if err != nil {
		logs.Error(i18n.T("error.get.file.status"))
		return fmt.Errorf("get file status: %w", err)
	}

	availableRemotes, remotesErr := GetAllRemotes()
	hasRemote := remotesErr == nil && len(availableRemotes) > 0

	if !hasChanges && !hasRemote {
		logs.Info(i18n.T("push.no.changes.no.remote"))
		return nil
	}
	if !hasChanges {
		logs.Info(i18n.T("push.no.changes"))
	} else if !hasRemote {
		logs.Info(i18n.T("push.no.remote.commit.only"))
	}

	// 干净工作区不收集提交信息,直接进入同步与推送
	suffix := ""
	message := ""
	subject := ""
	if hasChanges {
		options, err := config.GetOptions()
		if err != nil {
			logs.Error(i18n.T("error.get.options"))
			return fmt.Errorf("get options: %w", err)
		}
		options = applyCommitTypeDescriptions(options)

		// 类型选项固定且多,开启循环导航
		suffixes, err := form.ListFormWrap(i18n.T("push.select.commit.type"), options, form.ListSingle)
		if err != nil {
			return fmt.Errorf("select form: %w", err)
		}
		suffix = suffixes[0]

		message, subject, err = commitMessageInputs(suffix + ": ")
		if err != nil {
			return err
		}
	}

	var remotes []string
	branch := ""
	pullBefore := false
	if hasRemote {
		branch, err = GetCurrentBranch()
		if err != nil {
			logs.Error(i18n.T("error.get.current.branch"))
			return fmt.Errorf("get current branch: %w", err)
		}

		remotes, err = SelectAndSaveRemotes()
		if err != nil {
			return fmt.Errorf("select remote: %w", err)
		}
		pullBefore = shouldPullBeforePush()
	}

	steps, parallelFrom := buildPushSteps(hasChanges, message, remotes, branch, pullBefore)
	if err := command.RunMultipleCommandsParallel(steps, parallelFrom); err != nil {
		return err
	}

	if hasChanges {
		config.IncrementUsage(suffix)
		// 只记忆主题:建议栏是单行输入,正文入历史会污染候选
		config.AddRecentCommitMessage(subject)
	}
	return nil
}
