package commands

import (
	"fmt"
	"strconv"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/form"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
)

// configTagPatch 标签版本上限子流程:单页多输入表单编辑
// prefix/major/minor/patch/suffix,数字字段校验非负整数,prefix/suffix 可留空。
func configTagPatch() {
	current, err := config.GetTagPatch()
	if err != nil {
		logs.Error(i18n.T("error.general") + " " + err.Error())
		return
	}

	specs := []form.InputSpec{
		{
			Title:      i18n.T("config.tag.patch.prefix"),
			Default:    current.Prefix,
			Desc:       i18n.T("config.tag.patch.prefix.desc"),
			AllowEmpty: true,
		},
		{Title: i18n.T("config.tag.patch.major"), Default: strconv.Itoa(current.Major), Validate: nonNegativeInt(i18n.T("config.tag.patch.major"))},
		{Title: i18n.T("config.tag.patch.minor"), Default: strconv.Itoa(current.Minor), Validate: nonNegativeInt(i18n.T("config.tag.patch.minor"))},
		{Title: i18n.T("config.tag.patch.patch"), Default: strconv.Itoa(current.Patch), Validate: nonNegativeInt(i18n.T("config.tag.patch.patch"))},
		{
			Title:      i18n.T("config.tag.patch.suffix"),
			Default:    current.Suffix,
			Desc:       i18n.T("config.tag.patch.suffix.desc"),
			AllowEmpty: true,
		},
	}

	values, err := form.MultiInput(specs)
	if err != nil {
		// 取消:返回配置中心
		return
	}

	// 数字字段在表单层已校验,此处转换必然成功(防御性忽略错误)
	major, _ := strconv.Atoi(values[1])
	minor, _ := strconv.Atoi(values[2])
	patch, _ := strconv.Atoi(values[3])

	newPatch := config.Patch{
		Prefix: values[0],
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		Suffix: values[4],
	}

	if err := config.SavePatches([]config.Patch{newPatch}); err != nil {
		logs.Error(i18n.T("error.general") + " " + err.Error())
		return
	}
	logs.Success(fmt.Sprintf(i18n.T("config.tag.patch.save.success"), config.FormatPatch(newPatch)))
}

// nonNegativeInt 生成非负整数校验器(表单层即时反馈,无需提交后报错)
func nonNegativeInt(name string) func(string) error {
	return func(s string) error {
		n, err := strconv.Atoi(s)
		if err != nil || n < 0 {
			return fmt.Errorf(i18n.T("config.tag.patch.invalid.number"), name)
		}
		return nil
	}
}
