package config

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// 配置中心分派键:BuildConfigOptions 的 Value 与 config 命令跳转子流程共用,
// 单一来源避免字符串漂移。
const (
	ConfigKeyLanguage    = "language"
	ConfigKeyPush        = "push"
	ConfigKeyCommitTypes = "commit-types"
	ConfigKeyTagPatch    = "tag-patch"
)

type Option struct {
	Label       string
	Value       string
	Usage       int
	Description string // 选项单行说明(为空则渲染纯名称)
}

// FormatPatch 将版本上限格式化为可读字符串(prefix + major.minor.patch + suffix)
func FormatPatch(p Patch) string {
	return fmt.Sprintf("%s%d.%d.%d%s", p.Prefix, p.Major, p.Minor, p.Patch, p.Suffix)
}

// BuildConfigOptions 组装配置中心主列表:每项 Label 为配置项名称(本地化),
// Description 为当前值单行摘要;Value 为分派键。摘要读取失败时优雅降级,
// 显示提示而非中断(列表项仍可进入对应子流程)。
func BuildConfigOptions() []Option {
	options := []Option{
		{Label: i18n.T("config.option.language"), Value: ConfigKeyLanguage},
		{Label: i18n.T("config.option.push"), Value: ConfigKeyPush},
		{Label: i18n.T("config.option.commit.types"), Value: ConfigKeyCommitTypes},
		{Label: i18n.T("config.option.tag.patch"), Value: ConfigKeyTagPatch},
	}

	// 界面语言摘要(未设置时显示自动检测)
	lang, _ := GetLanguage()
	switch lang {
	case "zh":
		options[0].Description = i18n.T("config.summary.language.zh")
	case "en":
		options[0].Description = i18n.T("config.summary.language.en")
	default:
		options[0].Description = i18n.T("config.summary.language.auto")
	}

	// 推送配置摘要(仓库级:非 git 仓库或读取失败时提示)
	pushCfg, err := GetPushConfig()
	if err != nil {
		options[1].Description = i18n.T("config.not.git.repo")
	} else if pushCfg == nil || len(pushCfg.Remotes) == 0 {
		options[1].Description = i18n.T("config.summary.push.not.set")
	} else {
		options[1].Description = fmt.Sprintf(i18n.T("config.summary.push"), strings.Join(pushCfg.Remotes, ", "))
	}

	// 提交类型摘要(按 usage 排序,前 3 个,超长省略)
	commitTypes, _ := GetOptions()
	labels := make([]string, 0, min(len(commitTypes), 3))
	for i, o := range commitTypes {
		if i >= 3 {
			break
		}
		labels = append(labels, o.Label)
	}
	summary := strings.Join(labels, ", ")
	if len(commitTypes) > 3 {
		summary += ", …"
	}
	options[2].Description = fmt.Sprintf(i18n.T("config.summary.commit.types"), summary)

	// 标签版本上限摘要(读取失败回退默认值)
	patch, _ := GetTagPatch()
	options[3].Description = fmt.Sprintf(i18n.T("config.summary.tag.patch"), FormatPatch(patch))

	return options
}

type Patch struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
	Suffix string
}
