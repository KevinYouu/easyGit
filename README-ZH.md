[English](README.md) | 简体中文

# easyGit

🚀 一个现代化的交互式 Git CLI 工具，简化您的日常 Git 工作流。提交代码、管理分支、处理标签等操作，全部通过直观的终端界面完成。

> 这个项目使用自己的功能来提交代码 - 自我测试的最佳实践!

![easyGit 演示](assets/easygit.gif)

## 📖 什么是 easyGit？

**easyGit 是为热爱 Git 但希望工作更高效的开发者设计的。**

无需输入多个 Git 命令或记忆复杂的参数，easyGit 提供：

- ⚡ **一键工作流** - 用单个命令推送所有更改或选择特定文件
- 🎯 **交互式选择** - 通过可视化菜单选择分支、提交、标签和文件
- 🔧 **智能配置** - 保存您的推送偏好设置（远程仓库/分支）和语言设置
- 🌍 **多语言** - 完整的中英文支持，自动检测系统语言
- 🎨 **美观界面** - 现代化 TUI 组件提供清晰的视觉反馈

## 📖 设计哲学

**easyGit 是一个 Git 工作流增强工具，而不是 Git 命令的替代品。**

我们专注于:

- 🎯 **整合多步骤工作流** - 将多个 Git 命令组合为一个交互式流程
- 🧠 **简化复杂命令** - 为难记忆的参数提供友好的界面
- 🚀 **优化重复性工作** - 批量操作和智能默认值
- ✨ **增强用户体验** - 美观的 TUI 界面和清晰的视觉反馈

📚 了解更多: [设计哲学](docs/DESIGN_PHILOSOPHY_ZH.md) | [Design Philosophy (English)](docs/DESIGN_PHILOSOPHY.md)

## ✨ 核心功能

### 提交和推送
- **push-all** - 用一个命令暂存并推送所有更改的文件
- **push-selected** - 交互式选择要提交的文件
- **set-push-config** - 保存默认的远程仓库和分支偏好设置

### 分支和合并管理
- **merge** - 通过交互式选择将任意分支合并到当前分支
- **branch-delete** - 安全地删除本地或远程分支
- **cherry-pick** - 通过可视化选择从其他分支拣选提交

### 标签操作
- **tag-create** - 创建并推送标签，支持语义化版本控制
- **tag-delete** - 在本地和远程删除标签

### 仓库控制
- **reset** - 通过可视化提交历史重置到任意提交
- **init** - 为仓库初始化 easyGit 配置
- **update** - 更新 easyGit 到最新版本

### 自定义设置
- **set-language** - 在中英文之间切换
- **--language 参数** - 按命令覆盖语言设置

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
easyGit push-all

# 交互式选择要提交的文件
easyGit push-selected

# 初始化 easyGit 配置
easyGit init
```

### 高级操作

```bash
# 创建并推送标签
easyGit tag-create

# 删除标签
easyGit tag-delete

# 删除本地或远程分支
easyGit branch-delete

# 将选定分支合并到当前分支
easyGit merge

# 拣选提交
easyGit cherry-pick

# 重置到选定提交
easyGit reset

# 配置推送设置（远程仓库和分支）
easyGit set-push-config

# 设置界面语言
easyGit set-language

# 更新 easyGit
easyGit update
```

## 📖 命令参考

| 命令               | 描述                       |
| ------------------ | -------------------------- |
| `push-all`         | 暂存并提交所有更改的文件   |
| `push-selected`    | 交互式选择文件进行暂存提交 |
| `tag-create`       | 创建并推送带语义版本标签   |
| `tag-delete`       | 在本地和远程删除标签       |
| `branch-delete`    | 删除本地或远程分支         |
| `merge`            | 将选定分支合并到当前分支   |
| `cherry-pick`      | 从其他分支拣选提交         |
| `reset`            | 重置仓库到选定提交         |
| `init`             | 初始化 easyGit 配置        |
| `set-push-config`  | 配置默认推送远程和分支     |
| `set-language`     | 设置默认语言（en/zh）      |
| `update`           | 更新 easyGit 到最新版本    |
| `version`          | 显示当前版本信息           |

### 语言支持

easyGit 会自动检测您的系统语言，您也可以手动指定：

```bash
# 强制使用英文
easyGit --language en push-all

# 强制使用中文
easyGit --language zh push-all
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
go build -o easyGit ./cmd/easygit
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
├── cmd/easygit/          # CLI 命令和入口点
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
