package commands

import (
	"fmt"
	"regexp"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// commitTypePattern 提交类型合法格式(与 internal/config 校验一致,错误文案本地化)
var commitTypePattern = regexp.MustCompile(`^[a-z0-9-]+$`)

// configCommitTypes 提交类型管理子流程:子菜单「添加 / 删除 / 返回」。
func configCommitTypes() {
	options := []config.Option{
		{
			Label:       i18n.T("config.commit.types.add"),
			Description: i18n.T("config.commit.types.add.desc"),
			Value:       "add",
		},
		{
			Label:       i18n.T("config.commit.types.delete"),
			Description: i18n.T("config.commit.types.delete.desc"),
			Value:       "delete",
		},
		{
			Label:       i18n.T("config.back"),
			Description: i18n.T("config.back.desc"),
			Value:       "back",
		},
	}

	selected, err := form.ListForm(i18n.T("config.commit.types.title"), options, form.ListSingle)
	if err != nil {
		// 取消:返回配置中心
		return
	}

	switch selected[0] {
	case "add":
		addCommitType()
	case "delete":
		deleteCommitTypes()
	}
}

// addCommitType 添加提交类型:输入 → 格式校验 → 重复校验 → 保存
func addCommitType() {
	label, err := form.Input(i18n.T("config.commit.types.input.title"), "")
	if err != nil {
		// 取消:返回管理页
		return
	}

	// 格式校验:小写字母、数字、连字符
	if !commitTypePattern.MatchString(label) {
		logs.Error(i18n.T("config.commit.types.invalid"))
		return
	}

	// 重复校验
	existing, err := config.GetOptions()
	if err != nil {
		logs.Error(i18n.T("error.get.options"))
		return
	}
	for _, opt := range existing {
		if opt.Value == label {
			logs.Error(fmt.Sprintf(i18n.T("config.commit.types.duplicate"), label))
			return
		}
	}

	// 保存
	if err := config.AddCommitType(label); err != nil {
		logs.Error(i18n.T("error.get.options"))
		return
	}
	logs.Success(fmt.Sprintf(i18n.T("config.commit.types.add.success"), label))
}

// deleteCommitTypes 删除提交类型:多选(显示 usage 使用次数)→ 确认 → 删除
func deleteCommitTypes() {
	allTypes, err := config.GetOptions()
	if err != nil {
		logs.Error(i18n.T("error.get.options"))
		return
	}

	// 每个选项附加 usage 统计单行说明
	options := make([]config.Option, len(allTypes))
	for i, opt := range allTypes {
		options[i] = config.Option{
			Label:       opt.Label,
			Value:       opt.Value,
			Description: fmt.Sprintf(i18n.T("config.commit.types.usage"), opt.Usage),
		}
	}

	selected, err := form.ListForm(i18n.T("config.commit.types.delete.select"), options, form.ListMulti)
	if err != nil {
		// 取消:返回管理页
		return
	}
	if len(selected) == 0 {
		return
	}

	// 不允许删空全部类型
	if len(selected) >= len(allTypes) {
		logs.Error(i18n.T("config.commit.types.min"))
		return
	}

	// 删除前确认
	if !form.Confirm(fmt.Sprintf(i18n.T("config.commit.types.delete.confirm"), len(selected))) {
		return
	}

	if err := config.DeleteCommitTypes(selected); err != nil {
		logs.Error(i18n.T("error.get.options"))
		return
	}
	logs.Success(fmt.Sprintf(i18n.T("config.commit.types.delete.success"), len(selected)))
}
