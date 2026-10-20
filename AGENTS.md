# easyGit 开发规则

easyGit 是 Go 写的交互式 Git CLI（cobra + charm.land bubbletea/huh/lipgloss v2），偏好存 `~/.easyGit.db`（SQLite），支持 Linux/macOS/Windows。
`cmd/easygit` 只放 cobra 命令定义与分发；`internal/gitcmd` 是全部 Git 操作实现；`internal/{form,theme,command,i18n,config,logs}` 是共享 UI 与基础设施。

## 命令

- 提交前执行 `make all`（默认目标，即 vet + fmt + test + build）。`fmt` 会顺带跑 `go fix`，它改的文件要一并提交。
- 单个用例：`go test ./internal/gitcmd -run TestGetCurrentBranch`
- 本地试跑：`make dev <args>`；需要版本号注入的二进制用 `./build.sh`。

## 结构

- 新增子命令：在 `cmd/easygit/commands/` 建文件，注册进 `cmd/easygit/root.go` 的 `rootCmd.AddCommand(...)`，并加入 `cmd/easygit/commands/menu.go` 的 `menuItems`；`menu_test.go` 会校验 `descKey` 在 i18n 中存在以及高频项的排列顺序。
- 一个 Git 操作一个文件，放 `internal/gitcmd`；新增 Git 调用不要写在 `cmd/`、`internal/form`、`internal/command` 里。
- 文件 `snake_case`，导出 `PascalCase`，包名小写；`internal/gitcmd/pushAll.go`、`internal/command/spinnerCommand.go` 等旧文件保持原名，不做重命名清扫。

## 规范

- 文本一律 `i18n.T("a.b.c")`；新键同时加到 `internal/i18n/en.go` 和 `internal/i18n/zh.go`（漏掉 zh 会回退英文）。
- 用户可见输出走 `internal/logs`（`Error`/`Warning`/`Success`）与 `internal/theme` 样式；颜色令牌只在 `internal/theme` 定义，业务包不写颜色字面量。
- 交互一律用 `internal/form`（`ListFormColumns`、`InputWithValidate`、`Confirm`），进度用 `internal/command`（`progress_model.go`、`spinnerCommand.go`）完成，不把 huh/bubbletea 直接接进业务代码；输入校验写在 `InputWithValidate` 的 validate 回调里。
- `internal/gitcmd` 的交互入口写成包级 `var xxxMenu = func(...)`，用例可替换它注入选择结果。
- UI 风格遵循 `docs/ui-shadcn-tui-refactor.md`；非必要不写注释。
- 保持 `CGO_ENABLED=0` 可编译（SQLite 用 `modernc.org/sqlite`），不引入 CGO 依赖；启动保持轻量（<100ms），避免重复调用 `git`。
- 非必要不写注释

## 测试

- 用例依赖本机 `git`：临时仓库用 `testutil.CreateTempGitRepo(t)`；涉及配置的用例先 `config.SetTestDBPath(t.TempDir() + "/test.db")`，否则会写真实 `~/.easyGit.db`。
- `internal/gitcmd` 的用例靠 `os.Chdir` 切到临时仓库，所以不要加 `t.Parallel`，并用 `defer` 切回原目录。

## 文档

- 功能完成后同步 `README.md` 和 `README-ZH.md`，设计说明写进 `docs/features/`，并删掉 `TODO.md` 中已完成的条目（完成项不再在 TODO 累积）。
