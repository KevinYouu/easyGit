# TODO

## 已完成:TUI 布局优化(分隔线 + ❯ 选中指示符)

- [x] 选中指示符统一 `❯`(显示宽 2,与 huh 默认 `> ` 同宽零位移):huh Select/MultiSelect 主题换字(`SelectSelector`/`MultiSelectSelector` = `"❯ "` + PrimaryColor + Bold)→ `internal/theme/theme.go`
- [x] 表单包装模型:标题下顶部线 + 帮助栏上底部线(≥6 行终端),`dividerAfterTitle` 随 `Update` 重建保留 → `internal/form/form.go`、`select.go`、`multiSelect.go`
- [x] 布局公式更新:指示列 2 宽(单选最左列)、表单高度模型 +2 行(`min(n+4, 终端)`)、表格多选预留 4 行/单选 2 行 → `internal/form/layout.go`
- [x] 表格单选:最左 2 宽指示列,`rebuildRows(cursorRow)` 让 ❯ 跟随光标 → `table_select.go`
- [x] 表格多选:光标行复选框 `❯[x]`/`❯[ ]`(4 列宽恰好容纳),View 结构「标题 + 顶部线 + 行 + 底部线 + 帮助」→ `table_multi_select.go`
- [x] 进度屏:状态行下空行改分隔线,执行中帮助栏前插底部线(完成态保留 exiting 提示)→ `progress_model.go`
- [x] 渲染/布局/主题测试全量更新(高度公式、列数、`❯` 断言)→ `make all` 全绿
- [x] 终端尺寸矩阵 40/60/80/120/300 宽 × 6/10/24 高:分隔线位置正确、❯ 跟随光标、无折行/溢出
- [x] docs/features 功能文档 + 测试用例.md 增补 + README 双语同步,提交

## 已完成:统一帮助栏 + 选项单行说明

- [x] 帮助栏键位改 `[Esc]` 前缀式(主色加粗无背景),主题 Zinc Dark → Neutral Dark(Muted 400 / Border 700) → `internal/form/helpbar.go`、`internal/theme/theme.go`
- [x] 修复选中态背景断裂:OptionLabel 内嵌段剥离段尾 reset,说明段前复位字重(`\x1b[22m`),背景连续贯穿整行且说明保持常规字重 → `internal/form/helpbar.go`
- [x] TERM=dumb(accessible)时 selectOptionLabel 剥离样式,防无 reset 序列泄漏;`isAccessibleMode` 统一判定(与 huh 构造期耦合注释钉住) → `internal/form/select.go`、`form.go`
- [x] 回归测试:选中态背景连续(名称与说明之间无 reset)、说明段生效字重常规(状态机断言)、TERM=dumb 纯文本(含 huh accessible 真实路径) → `helpbar_test.go`、`form_test.go`

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
