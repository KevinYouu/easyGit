# 项目重命名方案：easyGit → easyGit

**制定日期**: 2025-12-26
**目标**: 将项目从 easyGit 重命名为 easyGit

---

## 📋 重命名概述

### 为什么要重命名？

1. **语义更准确**: "easy" 比 "fast" 更能体现工具的核心价值
   - easyGit 强调**易用性**和**简化操作**
   - easyGit 强调速度，但工具的核心价值是交互式体验和简化流程

2. **用户定位更清晰**:
   - 目标用户：希望简化 Git 操作的开发者
   - 核心卖点：TUI 交互、自动化流程、降低 Git 学习曲线

3. **避免误导**:
   - "fast" 可能让用户期待性能提升
   - 实际上工具提供的是**流程简化**而非性能优化

### 影响范围

需要修改的内容：

| 类别 | 数量 | 影响 |
|------|------|------|
| **代码文件** | 30+ | 🔴 高 |
| **配置文件** | 5 | 🔴 高 |
| **文档** | 4 | 🟡 中 |
| **构建脚本** | 3 | 🔴 高 |
| **GitHub 仓库** | 1 | 🔴 高 |

---

## 🎯 重命名检查清单

### Phase 1: 代码和配置文件（必须）

#### 1.1 Go Module 路径

- [ ] `go.mod` - module 路径
- [ ] 所有 Go 文件中的 import 路径

#### 1.2 配置文件

- [ ] `.goreleaser.yaml` - 项目名称和二进制名
- [ ] `build.sh` - 构建脚本中的名称
- [ ] `install.sh` - 安装脚本中的名称

#### 1.3 源代码

- [ ] 所有文件中的硬编码字符串 "easyGit"
- [ ] 数据库文件名 `.easyGit.db`
- [ ] 日志/输出中的项目名称
- [ ] 版本注入路径

#### 1.4 文档

- [ ] `README.md` - 项目名称、安装命令、示例
- [ ] `README-CN.md` - 中文版文档
- [ ] `CLAUDE.md` - 项目说明
- [ ] `docs/` 目录下的所有文档

#### 1.5 GitHub 仓库设置

- [ ] 仓库名称
- [ ] 仓库描述
- [ ] 主题标签
- [ ] Release 资产名称

---

## 📝 详细修改步骤

### Step 1: 备份当前状态

```bash
# 1. 确保所有更改已提交
git status

# 2. 创建备份分支
git checkout -b backup/before-rename-$(date +%Y%m%d)
git push origin backup/before-rename-$(date +%Y%m%d)

# 3. 返回主分支
git checkout main

# 4. 创建重命名工作分支
git checkout -b refactor/rename-to-easygit
```

---

### Step 2: 修改 Go Module 路径

#### 2.1 修改 go.mod

**文件**: `go.mod`

```diff
-module github.com/KevinYouu/easyGit
+module github.com/KevinYouu/easyGit

 go 1.23  // 同时修复版本问题
```

#### 2.2 批量替换 import 路径

**查找所有需要替换的 import**:

```bash
# 查找所有 import 语句
grep -r "github.com/KevinYouu/easyGit" --include="*.go" .
```

**批量替换**:

```bash
# 使用 sed 批量替换（Linux/Mac）
find . -name "*.go" -type f -exec sed -i 's|github.com/KevinYouu/easyGit|github.com/KevinYouu/easyGit|g' {} +

# 或使用 Perl（跨平台）
find . -name "*.go" -type f -exec perl -pi -e 's|github.com/KevinYouu/easyGit|github.com/KevinYouu/easyGit|g' {} +
```

**需要修改的文件**（预估）:
- `cmd/fastgit/*.go`
- `internal/*/*.go`

**验证**:

```bash
# 检查是否还有旧路径
grep -r "github.com/KevinYouu/easyGit" --include="*.go" .

# 更新依赖
go mod tidy
```

---

### Step 3: 重命名二进制文件和目录

#### 3.1 重命名 cmd 目录

```bash
# 重命名 cmd/fastgit 为 cmd/easygit
git mv cmd/fastgit cmd/easygit
```

#### 3.2 更新 main.go 中的包名（保持为 main）

**文件**: `cmd/easygit/main.go`

```go
package main  // 包名保持不变

// ... 其他代码
```

---

### Step 4: 修改配置文件

#### 4.1 修改 .goreleaser.yaml

**文件**: `.goreleaser.yaml`

```diff
 version: 2

-project_name: easyGit
+project_name: easyGit

 dist: .builds

 before:
   hooks:
     - go mod tidy

 builds:
   - id: build_noncgo
-    main: ./cmd/fastgit
-    binary: easyGit
+    main: ./cmd/easygit
+    binary: easyGit
     ldflags:
-      - -s -w -X github.com/KevinYouu/easyGit/internal/version.Version={{.Version}}
+      - -s -w -X github.com/KevinYouu/easyGit/internal/version.Version={{.Version}}
     env:
       - CGO_ENABLED=0
     goos:
       - linux
       - windows
       - darwin
     goarch:
       - amd64
       - arm64
```

#### 4.2 修改 build.sh

**文件**: `build.sh`

```diff
 #!/bin/bash
-# Local build script for easyGit
+# Local build script for easyGit
 # Created for testing version injection and local development

 set -e

 # Get version from git tag or use default
 VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev-$(date +%Y%m%d)")
-MAIN_PACKAGE="./cmd/fastgit"
-OUTPUT_NAME="easyGit"
+MAIN_PACKAGE="./cmd/easygit"
+OUTPUT_NAME="easyGit"

-echo "Building easyGit version: $VERSION"
+echo "Building easyGit version: $VERSION"

 # Build with version injection
-go build -ldflags="-s -w -X github.com/KevinYouu/easyGit/internal/version.Version=$VERSION" \
+go build -ldflags="-s -w -X github.com/KevinYouu/easyGit/internal/version.Version=$VERSION" \
     -o "$OUTPUT_NAME" "$MAIN_PACKAGE"

 echo "Build completed: $OUTPUT_NAME"
 echo "Run './$OUTPUT_NAME version' to verify version injection"
```

#### 4.3 修改 install.sh（如果存在）

**文件**: `install.sh`

```diff
 #!/bin/bash
-# Install script for easyGit
+# Install script for easyGit

-BINARY_NAME="easyGit"
+BINARY_NAME="easyGit"
-REPO="KevinYouu/easyGit"
+REPO="KevinYouu/easyGit"

 # ... 其他代码
```

---

### Step 5: 修改数据库文件名

#### 5.1 修改配置路径

**文件**: `internal/config/db.go`

```diff
 func getDBPath() string {
     homeDir, err := os.UserHomeDir()
     if err != nil {
         return ""
     }
-    return filepath.Join(homeDir, ".easyGit.db")
+    return filepath.Join(homeDir, ".easyGit.db")
 }
```

#### 5.2 添加数据迁移逻辑（可选）

**文件**: `internal/config/migrate.go`（新建）

```go
package config

import (
    "os"
    "path/filepath"
)

// MigrateOldDB 迁移旧的数据库文件
func MigrateOldDB() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }

    oldPath := filepath.Join(homeDir, ".easyGit.db")
    newPath := filepath.Join(homeDir, ".easyGit.db")

    // 如果新文件已存在，跳过迁移
    if _, err := os.Stat(newPath); err == nil {
        return nil
    }

    // 如果旧文件存在，重命名为新文件
    if _, err := os.Stat(oldPath); err == nil {
        return os.Rename(oldPath, newPath)
    }

    return nil
}
```

**在 main.go 中调用**:

**文件**: `cmd/easygit/main.go`

```diff
 func main() {
+    // 迁移旧数据库（如果存在）
+    config.MigrateOldDB()
+
     defer config.CloseDB()

     if err := cmd.Execute(); err != nil {
         os.Exit(1)
     }
 }
```

---

### Step 6: 修改代码中的硬编码字符串

#### 6.1 查找所有硬编码的 "easyGit"

```bash
# 查找所有包含 "easyGit" 的代码
grep -r "easyGit" --include="*.go" . | grep -v "easyGit"
```

**可能的位置**:
- 日志输出
- 错误信息
- 帮助文本
- 更新检查 URL

#### 6.2 修改更新检查

**文件**: `internal/update/update.go`

```diff
 const (
     repoOwner = "KevinYouu"
-    repoName  = "easyGit"
+    repoName  = "easyGit"
 )
```

```diff
-_, err := command.RunCmdWithSpinnerOptions("powershell",
-    []string{"-Command", "iwr -useb https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1 | iex"},
+_, err := command.RunCmdWithSpinnerOptions("powershell",
+    []string{"-Command", "iwr -useb https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.ps1 | iex"},
     ...
 )
```

---

### Step 7: 修改文档

#### 7.1 修改 README.md

**文件**: `README.md`

```diff
-# easyGit
+# easyGit

-easyGit is an interactive Git CLI tool that simplifies your Git workflow.
+easyGit is an interactive Git CLI tool that simplifies your Git workflow.

 ## Installation

 ### Using install script

 ```bash
-curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
+curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
 ```

 ### Manual installation

 ```bash
-wget https://github.com/KevinYouu/easyGit/releases/latest/download/easyGit-linux-amd64.zip
-unzip easyGit-linux-amd64.zip
-sudo mv easyGit /usr/local/bin/
+wget https://github.com/KevinYouu/easyGit/releases/latest/download/easyGit-linux-amd64.zip
+unzip easyGit-linux-amd64.zip
+sudo mv easyGit /usr/local/bin/
 ```

 ## Usage

 ```bash
-easyGit push-all
-easyGit merge
-easyGit cherry-pick
+easyGit push-all
+easyGit merge
+easyGit cherry-pick
 ```
```

#### 7.2 修改 README-CN.md

**文件**: `README-CN.md`

```diff
-# easyGit
+# easyGit

-easyGit 是一个交互式 Git CLI 工具，简化你的 Git 工作流程。
+easyGit 是一个交互式 Git CLI 工具，简化你的 Git 工作流程。

 ## 安装

 ### 使用安装脚本

 ```bash
-curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
+curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
 ```
```

#### 7.3 修改 CLAUDE.md

**文件**: `CLAUDE.md`

```diff
 # CLAUDE.md

-This file provides guidance to Claude Code when working with code in the easyGit repository.
+This file provides guidance to Claude Code when working with code in the easyGit repository.

 ## Build and Development Commands

 ### Building the Project
-- **Local Development Build**: `./build.sh` - Builds easyGit with version injection
+- **Local Development Build**: `./build.sh` - Builds easyGit with version injection
-- **Standard Go Build**: `go build -o easyGit ./cmd/fastgit`
+- **Standard Go Build**: `go build -o easyGit ./cmd/easygit`
```

#### 7.4 修改其他文档

**文件**: `docs/UI_ENHANCEMENT.md`, `docs/PROJECT_ANALYSIS.md`, `docs/OPTIMIZATION_PLAN.md`

批量替换所有文档中的 "easyGit" 为 "easyGit":

```bash
find docs/ -name "*.md" -type f -exec sed -i 's/easyGit/easyGit/g' {} +
```

---

### Step 8: 更新 GitHub 仓库

#### 8.1 重命名仓库

**步骤**:

1. 访问 GitHub 仓库页面
2. 点击 **Settings**
3. 在 **Repository name** 中输入 `easyGit`
4. 点击 **Rename**

**注意**: GitHub 会自动设置重定向，旧 URL 仍然可用。

#### 8.2 更新仓库描述

**描述**:

```
easyGit - An interactive Git CLI tool that simplifies your Git workflow with TUI
```

**主题标签**:

```
git, cli, tui, golang, git-tools, interactive-cli, terminal-ui, git-workflow
```

#### 8.3 更新本地远程 URL

```bash
# 查看当前远程 URL
git remote -v

# 更新远程 URL
git remote set-url origin https://github.com/KevinYouu/easyGit.git

# 或使用 SSH
git remote set-url origin git@github.com:KevinYouu/easyGit.git

# 验证
git remote -v
```

---

### Step 9: 测试和验证

#### 9.1 编译测试

```bash
# 1. 清理旧构建
go clean -cache

# 2. 更新依赖
go mod tidy

# 3. 本地构建
./build.sh

# 4. 验证二进制文件
./easyGit version

# 5. 运行测试
go test ./...
```

#### 9.2 功能测试

```bash
# 测试基本命令
./easyGit push-all
./easyGit merge
./easyGit cherry-pick
./easyGit reset

# 测试数据库迁移
rm ~/.easyGit.db  # 删除新数据库
touch ~/.easyGit.db  # 创建旧数据库
./easyGit version  # 应该自动迁移

# 验证迁移结果
ls -la ~/.easyGit.db  # 应该存在
ls -la ~/.easyGit.db  # 应该被删除或重命名
```

#### 9.3 文档验证

```bash
# 检查是否还有 "easyGit" 残留
grep -r "easyGit" . --exclude-dir=.git --exclude-dir=.builds

# 应该只在以下位置出现：
# - CHANGELOG.md（历史记录）
# - 备份文件
```

---

### Step 10: 提交和发布

#### 10.1 提交更改

```bash
# 查看所有更改
git status

# 添加所有更改
git add .

# 提交
git commit -m "refactor: rename project from easyGit to easyGit

Major changes:
- Rename Go module path: github.com/KevinYouu/easyGit → easyGit
- Rename binary: easyGit → easyGit
- Rename cmd directory: cmd/fastgit → cmd/easygit
- Update database path: ~/.easyGit.db → ~/.easyGit.db
- Add migration logic for old database
- Update all documentation
- Update build scripts and configurations

BREAKING CHANGE: Binary name changed from 'easyGit' to 'easyGit'

Co-Authored-By: Claude <noreply@anthropic.com>"
```

#### 10.2 推送到 GitHub

```bash
# 推送重命名分支
git push origin refactor/rename-to-easygit

# 创建 Pull Request（在 GitHub 上操作）
# 合并后推送到 main
git checkout main
git merge refactor/rename-to-easygit
git push origin main
```

#### 10.3 发布新版本

```bash
# 创建新版本标签
git tag -a v1.0.0 -m "Release v1.0.0: Renamed to easyGit

This is the first stable release after renaming the project from easyGit to easyGit.

New Features:
- Database migration from .easyGit.db to .easyGit.db
- Improved project naming and branding

Breaking Changes:
- Binary name changed: easyGit → easyGit
- Database file location changed: ~/.easyGit.db → ~/.easyGit.db
"

# 推送标签
git push origin v1.0.0

# GoReleaser 会自动构建和发布
```

---

## 🔄 数据迁移策略

### 用户升级路径

**场景 1: 全新安装**
- 直接使用 `easyGit`
- 创建 `~/.easyGit.db`

**场景 2: 从 easyGit 升级**
- 首次运行 `easyGit` 时自动检测 `~/.easyGit.db`
- 自动重命名为 `~/.easyGit.db`
- 显示迁移成功消息

**场景 3: 并行使用（不推荐）**
- 用户可能同时安装 `easyGit` 和 `easyGit`
- 迁移逻辑会复制而非移动数据库
- 避免数据丢失

### 迁移逻辑改进

**文件**: `internal/config/migrate.go`

```go
package config

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "github.com/KevinYouu/easyGit/internal/logs"
    "github.com/KevinYouu/easyGit/internal/i18n"
)

// MigrateOldDB 迁移旧的数据库文件
func MigrateOldDB() error {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return err
    }

    oldPath := filepath.Join(homeDir, ".easyGit.db")
    newPath := filepath.Join(homeDir, ".easyGit.db")

    // 如果新文件已存在，跳过迁移
    if _, err := os.Stat(newPath); err == nil {
        return nil
    }

    // 如果旧文件不存在，跳过迁移
    if _, err := os.Stat(oldPath); os.IsNotExist(err) {
        return nil
    }

    // 复制旧数据库到新位置（而非移动，以防用户仍在使用旧版本）
    if err := copyFile(oldPath, newPath); err != nil {
        return fmt.Errorf("failed to migrate database: %v", err)
    }

    logs.Info(i18n.T("config.migrate.success"))
    logs.Info(fmt.Sprintf("Old database: %s", oldPath))
    logs.Info(fmt.Sprintf("New database: %s", newPath))

    return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destFile.Close()

    _, err = io.Copy(destFile, sourceFile)
    return err
}
```

**添加翻译**:

**文件**: `internal/i18n/en.go`

```go
"config.migrate.success": "Database migrated successfully from easyGit to easyGit",
```

**文件**: `internal/i18n/zh.go`

```go
"config.migrate.success": "数据库已成功从 easyGit 迁移到 easyGit",
```

---

## 📢 用户沟通

### 发布公告

**GitHub Release 说明**:

```markdown
# easyGit v1.0.0 - Project Renamed

## 🎉 Major Changes

We've renamed the project from **easyGit** to **easyGit** to better reflect our core mission: **making Git easy for everyone**.

### What Changed?

- **Binary name**: `easyGit` → `easyGit`
- **Database file**: `~/.easyGit.db` → `~/.easyGit.db`
- **Repository**: `KevinYouu/easyGit` → `KevinYouu/easyGit`

### Migration

**Automatic Migration**: Your existing data will be automatically migrated when you first run `easyGit v1.0.0`.

**Manual Migration** (if needed):
```bash
# Backup your old database
cp ~/.easyGit.db ~/.easyGit.db.backup

# The new version will automatically detect and migrate
easyGit version
```

### Installation

**New users**:
```bash
curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
```

**Existing users** (upgrade):
```bash
# Uninstall old version
rm /usr/local/bin/easyGit

# Install new version
curl -sSL https://raw.githubusercontent.com/KevinYouu/easyGit/main/install.sh | bash
```

## 🐛 Bug Fixes

- Fixed Go version declaration (1.24.6 → 1.23)
- Removed debug output from production code
- Improved HTTP response validation

## 📚 Documentation

- Updated all documentation to reflect new name
- Added project analysis report
- Added optimization plan

## ⚠️ Breaking Changes

- **Binary name changed**: If you have scripts or aliases using `easyGit`, please update them to `easyGit`
- **Old GitHub redirect**: While `KevinYouu/easyGit` will redirect to `KevinYouu/easyGit`, please update your git remotes

---

Thank you for your continued support! 🙏
```

### README 警告横幅

**文件**: `README.md`（在顶部添加）

```markdown
# easyGit

> ⚠️ **Important**: This project was previously named `easyGit` and has been renamed to `easyGit` as of v1.0.0.
> All functionality remains the same, but the binary name and database location have changed.
> See [Migration Guide](#migration-guide) for details.
```

---

## ✅ 重命名验证清单

### 构建和运行

- [ ] `go mod tidy` 成功
- [ ] `./build.sh` 成功生成 `easyGit` 二进制
- [ ] `./easyGit version` 显示正确版本
- [ ] 所有测试通过 `go test ./...`

### 功能验证

- [ ] 所有命令可正常运行
- [ ] 数据库迁移正常工作
- [ ] 配置文件读写正常
- [ ] 更新检查指向正确仓库

### 文档验证

- [ ] README.md 中所有链接可访问
- [ ] 安装脚本正确
- [ ] 示例命令正确
- [ ] 无 "easyGit" 残留（除历史记录外）

### 发布验证

- [ ] GoReleaser 配置正确
- [ ] GitHub Actions 构建成功
- [ ] Release 资产名称正确
- [ ] 所有平台二进制可下载

---

## 📊 时间表

| 阶段 | 任务 | 时间 |
|------|------|------|
| **准备** | 备份、创建分支 | 10 分钟 |
| **代码修改** | Go module、import、重命名 | 30 分钟 |
| **配置修改** | .goreleaser、build.sh、install.sh | 20 分钟 |
| **文档修改** | README、CLAUDE.md、docs | 30 分钟 |
| **数据迁移** | 实现迁移逻辑 | 30 分钟 |
| **测试验证** | 编译、功能、文档测试 | 1 小时 |
| **提交发布** | 提交、推送、创建 Release | 20 分钟 |
| **总计** | - | **约 3 小时** |

---

## 🎯 后续行动

### 立即执行

1. ✅ 创建备份分支
2. ✅ 执行重命名步骤 1-9
3. ✅ 完成测试验证
4. ✅ 提交并推送

### 发布后

1. 📢 在社交媒体宣布重命名
2. 📝 更新第三方引用（如博客、论坛帖子）
3. 🔍 监控用户反馈
4. 📊 追踪迁移率（通过日志分析）

---

**重命名方案结束**
*准备好开始执行了吗？让我们开始吧！*
