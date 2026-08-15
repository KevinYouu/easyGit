package config

import (
	"fmt"
	"strings"

	"github.com/KevinYouu/easyGit/internal/i18n"
)

// 配置中心分派键:BuildConfigOptions 的 Value 与 config 命令跳转子流程共用,
// 单一来源避免字符串漂移。
const (
	ConfigKeyLanguage       = "language"
	ConfigKeyPush           = "push"
	ConfigKeyCommitTypes    = "commit-types"
	ConfigKeyTagPatch       = "tag-patch"
	ConfigKeyTheme          = "theme"
	ConfigKeyConflictEditor = "conflict-editor"
	ConfigKeyPullBeforePush = "pull-before-push"
)

type Option struct {
	Label       string
	Value       string
	Usage       int
	Description string   // 选项单行说明(为空则渲染纯名称)
	Cells       []string // 自适应多列布局的逐列单元格(非空时优先于 Label/Description)
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
		{Label: i18n.T("config.option.theme"), Value: ConfigKeyTheme},
		{Label: i18n.T("config.option.conflict.editor"), Value: ConfigKeyConflictEditor},
		{Label: i18n.T("config.option.pull"), Value: ConfigKeyPullBeforePush},
	}

	// 摘要按分派键定位写入,避免硬编码索引在增项时错位
	summary := make(map[string]*Option, len(options))
	for i := range options {
		summary[options[i].Value] = &options[i]
	}

	// 界面语言摘要(未设置时显示自动检测)
	lang, _ := GetLanguage()
	switch lang {
	case "zh":
		summary[ConfigKeyLanguage].Description = i18n.T("config.summary.language.zh")
	case "en":
		summary[ConfigKeyLanguage].Description = i18n.T("config.summary.language.en")
	default:
		summary[ConfigKeyLanguage].Description = i18n.T("config.summary.language.auto")
	}

	// 推送配置摘要(仓库级:非 git 仓库或读取失败时提示)
	pushCfg, err := GetPushConfig()
	if err != nil {
		summary[ConfigKeyPush].Description = i18n.T("config.not.git.repo")
	} else if pushCfg == nil || len(pushCfg.Remotes) == 0 {
		summary[ConfigKeyPush].Description = i18n.T("config.summary.push.not.set")
	} else {
		summary[ConfigKeyPush].Description = fmt.Sprintf(i18n.T("config.summary.push"), strings.Join(pushCfg.Remotes, ", "))
	}

	// 提交类型摘要(按 usage 排序,完整列出全部类型;
	// 超宽时由渲染层 SafeTruncate 截断,数据层不再预省略)
	commitTypes, _ := GetOptions()
	labels := make([]string, 0, len(commitTypes))
	for _, o := range commitTypes {
		labels = append(labels, o.Label)
	}
	summary[ConfigKeyCommitTypes].Description = fmt.Sprintf(i18n.T("config.summary.commit.types"), strings.Join(labels, ", "))

	// 标签版本上限摘要(读取失败回退默认值)
	patch, _ := GetTagPatch()
	summary[ConfigKeyTagPatch].Description = fmt.Sprintf(i18n.T("config.summary.tag.patch"), FormatPatch(patch))

	// 界面主题摘要(未设置时显示自动检测)
	themeSetting, _ := GetTheme()
	switch themeSetting {
	case ThemeDark:
		summary[ConfigKeyTheme].Description = i18n.T("config.summary.theme.dark")
	case ThemeLight:
		summary[ConfigKeyTheme].Description = i18n.T("config.summary.theme.light")
	default:
		summary[ConfigKeyTheme].Description = i18n.T("config.summary.theme.auto")
	}

	// 冲突编辑器摘要(未设置时显示自动检测)
	editorSetting, _ := GetConflictEditor()
	if editorSetting == "" {
		summary[ConfigKeyConflictEditor].Description = i18n.T("config.summary.conflict.editor.auto")
	} else {
		summary[ConfigKeyConflictEditor].Description = editorSetting
	}

	// 推送前 pull 摘要(默认 always)
	pullSetting, _ := GetPullBeforePush()
	if pullSetting == PullBeforePushNever {
		summary[ConfigKeyPullBeforePush].Description = i18n.T("config.summary.pull.never")
	} else {
		summary[ConfigKeyPullBeforePush].Description = i18n.T("config.summary.pull.always")
	}

	return options
}

type Patch struct {
	Prefix string
	Major  int
	Minor  int
	Patch  int
	Suffix string
}
