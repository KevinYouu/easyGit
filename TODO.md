# TODO

## 进行中

（无）

## 进行中:渲染测试覆盖所有组件与命令

- [x] Input 命令级用例(pushAll/pushSelected/squash/tag×2) → `render_input_confirm_test.go`
- [x] Confirm 命令级用例(branch×2/drop/merge/reset/tag,含动态消息) → `render_input_confirm_test.go`
- [x] Select 补 merge target、rebase target → `render_command_test.go`
- [x] MultiSelect 补 remote 远端多选 → `render_command_test.go`
- [x] progress_model_test.go 补 View 渲染测试(进度条自适应/状态图标) → 顺带修复窄屏越界
- [x] docs/features/测试用例.md 同步渲染测试矩阵
- [x] make all 全绿并提交
