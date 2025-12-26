[English](README.md) | 简体中文

# easyGit

🚀 一个现代化的命令行工具，通过交互式终端界面简化 Git 操作。支持 Linux、macOS 和 Windows。

> 这个项目使用自己的功能来提交代码 - 自我测试的最佳实践！

![easyGit 演示](assets/fast-git.gif)

## ✨ 特性

- 🎯 **交互式 Git 操作** - 通过美观的 TUI 界面选择文件、分支和提交
- 🌍 **双语支持** - 支持中英文，自动检测系统语言
- 🎨 **现代化界面** - 美观的主题、动画和进度指示器
- ⚡ **快速高效** - 为常见 Git 操作提供简化的工作流程
- 🔧 **跨平台** - 支持 Linux、macOS 和 Windows

## 📦 安装

### 前置要求

- 系统必须安装 [Git](https://git-scm.com/)

### 快速安装（推荐）

```bash
# Linux/macOS
curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash

# 或者使用 wget
wget -qO- https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
```

```powershell
# Windows
iwr -useb https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1 | iex
```

### 从源码构建

```bash
git clone https://github.com/KevinYouu/easyGit.git
cd easyGit
./build.sh
```

## 🚀 快速开始

### 基础用法

```bash
# 提交工作区所有更改的文件
easyGit pa

# 交互式选择要提交的文件
easyGit ps

# 查看仓库状态
easyGit s
```

### 高级操作

```bash
# 创建并推送标签
easyGit t

# 删除标签
easyGit td

# 将选定分支合并到当前分支
easyGit m

# 拣选提交
easyGit cp

# 重置到选定提交
easyGit rs

# 查看远程仓库
easyGit rv

# 初始化 easyGit 配置
easyGit init

# 更新 easyGit
easyGit update
```

## 📖 命令参考

| 命令 | 别名 | 描述 |
|------|------|------|
| `push-all` | `pa` | 暂存并提交所有更改的文件 |
| `push-selected` | `ps` | 交互式选择文件进行暂存和提交 |
| `status` | `s` | 显示增强 UI 的仓库状态 |
| `tag` | `t` | 创建并推送带语义版本标签 |
| `tag-delete` | `td` | 在本地和远程删除标签 |
| `merge` | `m` | 将选定分支合并到当前分支 |
| `cherry-pick` | `cp` | 从其他分支拣选提交 |
| `reset` | `rs` | 重置仓库到选定提交 |
| `remotes` | `rv` | 列出所有远程仓库 |
| `init` | - | 初始化 easyGit 配置 |
| `update` | - | 更新 easyGit 到最新版本 |
| `version` | - | 显示当前版本信息 |

### 语言支持

easyGit 会自动检测您的系统语言，您也可以手动指定：

```bash
# 强制使用英文
easyGit --language en pa

# 强制使用中文
easyGit --language zh pa
```

## 🛠️ 开发

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/KevinYouu/easyGit.git
cd easyGit

# 安装依赖
go mod tidy

# 构建并注入版本信息
./build.sh

# 或手动构建
go build -o easyGit ./cmd/fastgit
```

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./internal/i18n

# 运行基准测试
go test -bench=. ./internal/i18n
```

### 项目结构

```
easyGit/
├── cmd/fastgit/          # CLI 命令和入口点
│   ├── main.go          # 应用程序入口
│   ├── root.go          # 根命令设置
│   └── commands/        # 各个命令的实现
├── internal/            # 内部包
│   ├── gitcmd/          # Git 操作实现
│   ├── i18n/            # 国际化
│   ├── form/            # TUI 表单组件
│   ├── spinner/         # 加载动画
│   ├── theme/           # 视觉样式
│   └── config/          # 配置管理
└── docs/                # 文档
```

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。对于重大更改，请先开启 issue 讨论您想要更改的内容。

### 开发指南

- 遵循 Go 约定和最佳实践
- 为新功能添加测试
- 根据需要更新文档
- 确保所有命令都支持中英文

## 📄 许可证

本项目采用 [LICENSE](LICENSE) 文件中的许可证。

## 🙏 致谢

基于这些优秀的开源项目构建：

- [Go](https://github.com/golang/go) - Go 编程语言
- [Cobra](https://github.com/spf13/cobra) - 强大的 CLI 应用程序
- [Bubbletea](https://github.com/charmbracelet/bubbletea) - 强大的 TUI 框架
- [Huh](https://github.com/charmbracelet/huh) - 交互式终端表单
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - 终端 UI 样式定义

---

**由 [KevinYouu](https://github.com/KevinYouu) 用 ❤️ 制作**
