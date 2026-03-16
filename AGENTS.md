# easyGit 开发规则

## 1. 核心规范

- 命名: 文件 `snake_case`, 导出 `PascalCase`, 私有/变量 `camelCase`, 包名小写
- Git 操作: 必须封装在 `internal/gitcmd`，一操作一文件，禁止跨包直接调用
- i18n: 严禁硬编码。文本统一使用 `i18n.T("key")`，键名点分层级
- TUI/UI: 必须使用 `internal/form` (表单)、`internal/spinner` (进度)、`internal/theme` (样式)
- 错误处理: 立即检查 `err`；Git 失败用 `logs.Error`；输入验证在表单层完成

## 2. 质量要求

- 必做检查: 提交前必须执行 `make all`
- 测试与验证: 补齐单元测试；提供最小可复现步骤及预期结果
- 文档同步: 完成功能后更新 `README.md` 和 `README-ZH.md` 并清理 `TODO.md`
- 性能: 启动 < 100ms，避免冗余 Git 命令调用

## 3. 禁止事项

- 硬编码文本 | 使用 `fmt.Println` | 依赖 CGO | 跨包 Git 调用
