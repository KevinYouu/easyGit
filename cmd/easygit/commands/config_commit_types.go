package commands

import (
	"errors"
	"fmt"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

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

	selected, err := form.ListFormColumns(i18n.T("config.commit.types.title"), configListSpecs(), options, form.ListSingle)
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

// addCommitType 添加提交类型:输入 → 格式校验 → 保存;
// 格式与重复校验由 config.AddCommitType 统一判定(单一事实来源),
// 此处先经 config.CommitTypePattern 预校验以输出本地化文案,
// 再按哨兵错误映射重复文案,真实错误(如 DB 故障)原样带出。
func addCommitType() {
	label, err := form.Input(i18n.T("config.commit.types.input.title"), "")
	if err != nil {
		// 取消:返回管理页
		return
	}

	// 格式校验:小写字母、数字、连字符(表单层之后的第一道防线,文案本地化)
	if !config.CommitTypePattern.MatchString(label) {
		logs.Error(i18n.T("config.commit.types.invalid"))
		return
	}

	// 保存(内部含重复校验)
	if err := config.AddCommitType(label); err != nil {
		if errors.Is(err, config.ErrCommitTypeExists) {
			logs.Error(fmt.Sprintf(i18n.T("config.commit.types.duplicate"), label))
			return
		}
		logs.Error(i18n.T("error.add.commit.type") + ": " + err.Error())
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

	selected, err := form.ListFormColumns(i18n.T("config.commit.types.delete.select"), configListSpecs(), options, form.ListMulti)
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
		logs.Error(i18n.T("error.delete.commit.type") + ": " + err.Error())
		return
	}
	logs.Success(fmt.Sprintf(i18n.T("config.commit.types.delete.success"), len(selected)))
}
