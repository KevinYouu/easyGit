# easyGit 编码规范

## 命名规范

- 文件名: snake_case (push_all.go, cherry_pick.go)
- 包名: 小写单词 (gitcmd, i18n, config)
- 函数/方法: PascalCase 导出, camelCase 私有
- 常量: PascalCase 或 UPPER_SNAKE_CASE
- 变量: camelCase

## 代码质量检查

编写完成代码后必须执行以下检查:

```bash
go vet ./...
go build ./...  # 编译检查
go test ./...   # 运行测试
```

## 国际化

- 所有用户可见文本必须使用 `i18n.T("key")`
- 翻译键添加到 `internal/i18n/en.go` 和 `zh.go`
- 键名使用点分层级: `"command.push.description"`

## Git 操作

- 所有 Git 命令封装在 `internal/gitcmd/` 包
- 每个操作独立文件 (merge.go, reset.go 等)
- 使用 `exec.Command("git", args...)` 执行命令
- 错误信息通过 `internal/logs` 统一输出

## TUI 组件

- 表单使用 `internal/form` 包 (Input, Select, MultiSelect)
- 长时操作使用 `internal/spinner` 显示进度
- 样式统一使用 `internal/theme` 主题

## 错误处理

- 立即检查错误: `if err != nil { return err }`
- Git 命令失败使用 `logs.Error()` 输出
- 用户输入验证在表单层完成

## 性能要求

- CLI 启动 < 100ms
- 单次 Git 操作避免多次命令调用
- 配置读取使用 SQLite (已优化)

## 禁止事项

- ❌ 硬编码文本 (必须用 i18n)
- ❌ 直接 fmt.Println 调试代码
- ❌ 跨包直接调用 Git 命令
- ❌ 使用 CGO 依赖
