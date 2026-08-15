[English](README.md) | 简体中文

# easyGit

[![Release](https://img.shields.io/github/v/release/KevinYouu/easyGit)](https://github.com/KevinYouu/easyGit/releases)
[![License](https://img.shields.io/badge/license-GPL--3.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/KevinYouu/easyGit)](go.mod)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()

🚀 一个现代化的交互式 Git CLI 工具，简化您的日常 Git 工作流。提交代码、管理分支、处理标签等操作，全部通过直观的终端界面完成。支持 Linux/macOS/Windows

> 这个项目使用自己的功能来提交代码 - 自我测试的最佳实践!

![easyGit 演示](assets/easygit.gif)

## 📖 什么是 easyGit？

**easyGit 是为热爱 Git 但希望工作更高效的开发者设计的。**

无需输入多个 Git 命令或记忆复杂的参数，easyGit 提供：

- ⚡ **一键工作流** - 用单个命令推送所有更改或选择特定文件;无参数运行 `easyGit` 直接进入交互式操作菜单
- 🔀 **多远程并行推送** - `pa`/`ps`/`tc`/`td`/`bd` 并发推送到多个远程；单个远程失败不阻塞其余，错误摘要列出全部失败步骤
- 🎯 **交互式选择** - 通过可视化菜单选择分支、提交、标签和文件;任意列表按 `/` 输入即过滤
- 🧠 **记忆你的选择** - 提交类型、reset 模式、merge 策略、cherry-pick 选项跨会话记忆;提交消息支持 ↑/↓ 复用最近 10 条
- 🔧 **智能配置** - 一站式 `config` 配置中心：语言、推送偏好(含推送前拉取)、提交类型、标签版本上限、主题、冲突编辑器
- 🌍 **多语言** - 完整的中英文支持，自动检测系统语言
- 🎨 **美观界面** - 现代化 TUI 组件提供清晰的视觉反馈;每个界面以全宽分隔线框定内容区,`❯` 指示符随光标在所有选择列表间移动
- 📐 **响应式布局** - 表格、表单、进度条随终端尺寸自适应,宽屏充分利用,窄屏不溢出
- ⌨️ **统一帮助栏** - 所有界面底部单行快捷键栏(`[键位]` 前缀 + 动作说明),选项说明单行内嵌(名称亮 + 说明灰),极小终端自动隐藏
- 💻 **跨平台** - 在 Linux、macOS 和 Windows 上无缝运行

## 📖 设计哲学

**easyGit 是一个 Git 工作流增强工具，而不是 Git 命令的替代品。**

我们专注于:

- 🎯 **整合多步骤工作流** - 将多个 Git 命令组合为一个交互式流程
- 🧠 **简化复杂命令** - 为难记忆的参数提供友好的界面
- 🚀 **优化重复性工作** - 批量操作和智能默认值
- ✨ **增强用户体验** - 美观的 TUI 界面和清晰的视觉反馈

📚 了解更多: [设计哲学](docs/DESIGN_PHILOSOPHY_ZH.md) | [Design Philosophy (English)](docs/DESIGN_PHILOSOPHY.md)

---

## ⚡ 为什么选择 easyGit?

### easyGit vs 原生 Git

看看使用 easyGit 能节省多少时间:

| 操作           | 原生 Git                                                       | easyGit                     | 节省时间 |
| -------------- | -------------------------------------------------------------- | --------------------------- | -------- |
| 提交所有更改   | `git add .`<br>`git commit -m "msg"`<br>`git push`             | `easyGit pa`                | **~75%** |
| 删除远程分支   | `git push origin --delete branch-name`                         | `easyGit bd` (交互式)       | **~60%** |
| 创建并推送标签 | `git tag v1.0.0`<br>`git push origin v1.0.0`                   | `easyGit tc` (向导模式)     | **~70%** |
| 合并功能分支   | `git checkout main`<br>`git merge feature`<br>`git push`       | `easyGit m` (交互式)        | **~65%** |
| 变基分支       | `git branch` (查找名称)<br>`git rebase <branch>`               | `easyGit r` (可视化选择)    | **~60%** |
| 拣选提交       | `git log` (查找哈希)<br>`git cherry-pick <hash>`<br>`git push` | `easyGit cp` (可视化选择器) | **~70%** |
| 合并多个提交   | `git rebase -i` (手动编辑)                                     | `easyGit sq` (可视化多选)   | **~70%** |
| 删除提交       | `git rebase -i` (手动删除)                                     | `easyGit d` (可视化多选)    | **~75%** |

**核心优势:**

- 🎯 **交互式选择** - 无需记忆分支名、提交哈希或标签名
- 🔄 **工作流整合** - 多个 Git 命令合并为一个
- 💡 **智能默认** - 记住你的偏好设置(远程仓库、分支、语言)
- ✅ **错误预防** - 破坏性操作前进行可视化确认

---

## 📦 安装

### 前置要求

- 系统必须安装 [Git](https://git-scm.com/)

### 系统要求

- **操作系统:** Linux / macOS / Windows 10+
- **Git 版本:** 2.0+
- **终端:** 支持 ANSI 颜色的现代终端
- **磁盘空间:** < 10MB

### 快速安装（推荐）

**Linux / macOS:**

```bash
# 使用 curl
curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash

# 或者使用 wget
wget -qO- https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
```

**Windows:**

```powershell
# 使用 PowerShell
iwr -useb https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1 | iex
```

### 从源码构建

在所有安装了 Go 的平台上都可以构建：

```bash
git clone https://github.com/KevinYouu/easyGit.git
cd easyGit
./build.sh
```

### 验证安装

安装完成后,验证 easyGit 是否正常工作:

```bash
# 检查 easyGit 是否已安装
easyGit version

# 预期输出:
# easyGit version x.x.x
# Built with Go 1.21+

# 尝试简单命令
easyGit --help
```

如遇到任何问题,请查看[常见问题](#-常见问题)章节。

---

## 🚀 快速开始

### 基础用法

```bash
# 无参数运行:进入交互式操作菜单(Enter 选择命令)
easyGit

# 提交工作区所有更改的文件
easyGit push-all  # 或: easyGit pa

# 交互式选择要提交的文件(列表含 M/A/D 状态列)
easyGit push-selected  # 或: easyGit ps

# 初始化 easyGit 配置, 通常不用手动设置, 会在第一次启动时自动初始化
easyGit init
```

> 💡 **提示**: 任意选择列表(提交/分支/标签/文件)按 `/` 输入即过滤;提交消息输入框按 ↑/↓ 复用最近 10 条历史消息。

### 高级操作

```bash
# 创建并推送标签
easyGit tag-create  # 或: easyGit tc

# 删除标签
easyGit tag-delete  # 或: easyGit td

# 删除本地或远程分支
easyGit branch-delete  # 或: easyGit bd

# 将选定分支合并到当前分支
easyGit merge  # 或: easyGit m

# 变基当前分支到另一分支
# 冲突时进入工具内解决闭环(编辑 → 继续/跳过/中止)
easyGit rebase  # 或: easyGit r

# 拣选提交
easyGit cherry-pick  # 或: easyGit cp

# 重置到选定提交
easyGit reset  # 或: easyGit rs

# 将连续提交合并为一个
easyGit squash  # 或: easyGit sq

# 删除特定提交
easyGit drop  # 或: easyGit d

# 配置 easyGit（语言、主题、推送设置(含推送前拉取)、提交类型、标签版本上限）
easyGit config

# 更新 easyGit
easyGit update
```

### 主题（深色/浅色）

- **自动（默认）**：启动时探测终端背景色（OSC 11 查询）；终端不支持或非 TTY 环境回退深色。
- **手动**：`--theme` 标志临时指定（运行时最高优先级），或通过配置中心持久化偏好。

```bash
# 单次覆盖（运行时优先）
easyGit --theme light push-all

# 持久化偏好（下次启动生效）
easyGit config   # → 界面主题 → 自动 / 深色 / 浅色
```

优先级：`--theme` 标志 > 配置中心设置 > 自动检测。

## 📖 命令参考

### 日常操作

| 命令            | 简写 | 描述                       | 使用场景         |
| --------------- | ---- | -------------------------- | ---------------- |
| `push-all`      | `pa` | 暂存并提交所有更改的文件   | 快速提交所有修改 |
| `push-selected` | `ps` | 交互式选择文件进行暂存提交 | 精细控制提交内容 |

### 分支与标签管理

| 命令            | 简写 | 描述                     | 使用场景                     |
| --------------- | ---- | ------------------------ | ---------------------------- |
| `merge`         | `m`  | 将选定分支合并到当前分支 | 集成功能分支                 |
| `branch-delete` | `bd` | 删除本地或远程分支       | 清理已合并或废弃的分支       |
| `tag-create`    | `tc` | 创建并推送带语义版本标签 | 版本发布 (v1.0.0, v2.1.0 等) |
| `tag-delete`    | `td` | 在本地和远程删除标签     | 删除错误或废弃的标签         |

### 高级 Git 操作

| 命令          | 简写 | 描述               | 使用场景               |
| ------------- | ---- | ------------------ | ---------------------- |
| `rebase`      | `r`  | 变基当前分支到另一分支 | 保持线性的提交历史     |
| `cherry-pick` | `cp` | 从其他分支拣选提交 | 应用特定提交到当前分支 |
| `reset`       | `rs` | 重置仓库到选定提交 | 撤销提交或回到之前状态 |
| `squash`      | `sq` | 将连续提交合并为一个 | 清理本地提交历史 (可视化) |
| `drop`        | `d`  | 删除特定历史提交   | 从历史中移除不需要的提交 |

### 配置与工具

| 命令              | 简写 | 描述                                      | 使用场景                         |
| ----------------- | ---- | ----------------------------------------- | -------------------------------- |
| `init`            | -    | 初始化 easyGit 配置                       | 首次设置（通常自动执行）         |
| `config`          | -    | 打开交互式配置中心                        | 修改语言、主题、推送配置(含推送前拉取)、提交类型、标签版本上限、冲突编辑器 |
| `update`          | -    | 更新 easyGit 到最新版本                   | 获取新功能和错误修复             |
| `version`         | `v`  | 显示当前版本信息                          | 查看已安装版本                   |

---

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

---

## 🔧 兼容性

✅ 完全兼容原生 Git 工作流
✅ 可与 Git 命令混合使用
✅ 不修改 `.git` 目录结构
✅ 支持所有 Git hooks

---

## ❓ 常见问题

### easyGit 与其他 Git 工具有什么不同?

easyGit 专注于**工作流整合**而非替换单个 Git 命令。它将多个 Git 操作组合为交互式流程，让复杂任务变得简单。

### easyGit 会修改我的 Git 仓库吗?

不会。easyGit 底层使用标准 Git 命令，不会修改你的 `.git` 目录结构。你可以安全地与常规 Git 命令混用。

### 可以在 CI/CD 流水线中使用 easyGit 吗?

easyGit 是为交互式终端使用而设计的。对于 CI/CD，我们建议使用标准 Git 命令。

### 遇到 Bug 怎么办?

请[提交 issue](https://github.com/KevinYouu/easyGit/issues)，包含:

- 你的操作系统和 Git 版本
- 复现步骤
- 期望行为 vs 实际行为

### 如何验证安装?

```bash
# 查看版本
easyGit version

# 预期输出
easyGit version x.x.x
Built with Go 1.21+
```

---

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。对于重大更改，请先开启 issue 讨论您想要更改的内容。

### 开发指南

- 遵循 Go 约定和最佳实践
- 为新功能添加测试
- 根据需要更新文档
- 确保所有命令都支持中英文

---

## 📄 许可证

本项目采用 GPL-3.0 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

基于这些优秀的开源项目构建：

### 核心依赖

- [Go](https://github.com/golang/go) - Go 编程语言
- [Cobra](https://github.com/spf13/cobra) - 强大的 CLI 框架
- [Bubbletea v2](https://charm.land/bubbletea) - 强大的 TUI 框架
- [Bubbles v2](https://charm.land/bubbles) - Bubbletea 的 TUI 组件
- [Huh v2](https://charm.land/huh) - 交互式终端表单
- [Lipgloss v2](https://charm.land/lipgloss) - 终端 UI 样式定义
- [SQLite](https://gitlab.com/cznic/sqlite) - 纯 Go 实现的 SQLite，用于配置存储
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) - 终端工具
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) - 文本处理和国际化支持

---

**由 [KevinYouu](https://github.com/KevinYouu) 用 ❤️ 制作**
