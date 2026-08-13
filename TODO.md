# TODO

## 进行中

（无）

## 已完成:统一帮助栏 + 选项单行说明

- [x] helpbar 组件(RenderHelpBar/AppendHelpBar/OptionLabel/HelpBarMinTermHeight) → `internal/form/helpbar.go`
- [x] 表单 4 构造器接线帮助栏,Form 包装模型统一生产与测试路径 → `internal/form/form.go`
- [x] 表格单选/多选帮助栏接线,单选高度预留 1 行 → `internal/form/table_select.go`、`layout.go`
- [x] 进度屏硬编码英文 i18n 化 + 帮助栏(完成态保留 exiting 提示) → `progress_model.go`、`spinnerCommand.go`
- [x] reset 模式表格 → 列表式单选表单,4 项单行说明(default 不传参数) → `internal/gitcmd/reset.go`
- [x] 选项说明统一机制:config.Option.Description + SelectForm 单行组装(merge/cherry-pick/rebase/language/commit type 优雅降级)
- [x] 渲染测试全量更新(高度模型 min(n+2, 终端)/4 行隐藏/两段式单行用例) → `make all` 全绿
- [x] docs/features 功能文档 + 测试用例.md 增补 + README 双语同步

## 已完成:渲染测试覆盖所有组件与命令

- [x] Input 命令级用例(pushAll/pushSelected/squash/tag×2) → `render_input_confirm_test.go`
- [x] Confirm 命令级用例(branch×2/drop/merge/reset/tag,含动态消息) → `render_input_confirm_test.go`
- [x] Select 补 merge target、rebase target → `render_command_test.go`
- [x] MultiSelect 补 remote 远端多选 → `render_command_test.go`
- [x] progress_model_test.go 补 View 渲染测试(进度条自适应/状态图标) → 顺带修复窄屏越界
- [x] docs/features/测试用例.md 同步渲染测试矩阵
- [x] make all 全绿并提交
