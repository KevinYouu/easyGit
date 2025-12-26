# easyGit 优化计划

**制定日期**: 2025-12-26
**更新日期**: 2025-12-26（移除配置存储迁移，保留 SQLite）
**目标**: 提升代码质量、性能、测试覆盖率和文档完整性

本文档基于《项目分析报告》，提供可执行的详细优化方案。

---

## 📋 总览

| 阶段 | 优先级 | 任务数 | 预估工作量 | 关键收益 |
|------|--------|--------|------------|----------|
| **Phase 0** | 🔴 P0 | 4 个 | 4-7 天 | 代码质量、合规性 |
| **Phase 1** | 🟡 P1 | 3 个 | 3-4 天 | 性能、用户体验 |
| **Phase 2** | 🟢 P2 | 3 个 | 4-6 天 | 代码整洁度 |

**总工作量**: 11-17 天（约 2-3.5 周）

---

## 🔴 Phase 0: 必须修复（4-7 天）

### P0.1 清理调试代码 ⚡

**优先级**: 最高
**工作量**: 5 分钟
**影响**: 用户体验、代码专业度

#### 问题

**位置**:
- `internal/config/option.go:36, 46, 53`
- `internal/config/updateUsage.go:8, 15, 27, 36, 42`

```go
// ❌ 生产代码中的调试输出
fmt.Println("❌ line 31 err ➡️", err)
fmt.Println("❌ line 40 err ➡️", err)
fmt.Println("❌ line 47 err ➡️", err)
```

#### 解决方案

**直接移除**（推荐）

```diff
// internal/config/option.go
func GetOptions() ([]Option, error) {
    // ...
    if err := rows.Scan(&option.Label, &option.Value, &option.Usage); err != nil {
-       fmt.Println("❌ line 40 err ➡️", err)
        return nil, err
    }
}
```

```diff
// internal/config/updateUsage.go
func IncrementUsage(value string) error {
    db, err := openDB()
    if err != nil {
-       fmt.Println("❌ line 8 err ➡️", err)
        return err
    }
    // ...
}
```

**共需移除 8 处调试输出**

#### 验证步骤

```bash
# 1. 搜索确认无残留
grep -r "line.*err ➡️" internal/

# 2. 测试功能正常
go test ./internal/config/...

# 3. 运行程序确认无调试输出
./easyGit version
./easyGit push-all  # 选择提交类型，确认无调试输出
```

#### 完成条件

- ✅ 移除所有 8 处 `fmt.Println("❌ line X err ➡️", err)`
- ✅ 功能正常运行
- ✅ 无调试输出到用户终端

---

### P0.2 修复函数拼写错误 ⚡

**优先级**: 最高
**工作量**: 10 分钟
**影响**: 代码质量、专业度

#### 问题

**位置**: `internal/logs/logs.go:9`

```go
func Waring(text string) {  // ❌ 应该是 Warning
    fmt.Println(colors.RenderColor("yellow", text))
}
```

**影响范围**: 多个文件调用
- `internal/gitcmd/merge.go`
- `internal/gitcmd/cherry_pick.go`
- 其他文件

#### 解决方案

**步骤 1: 重命名函数**

```diff
// internal/logs/logs.go
-func Waring(text string) {
+func Warning(text string) {
     fmt.Println(colors.RenderColor("yellow", text))
 }
```

**步骤 2: 批量替换调用**

```bash
# 查找所有调用
grep -r "logs.Waring" internal/

# 使用 sed 批量替换
find internal/ -name "*.go" -type f -exec sed -i 's/logs\.Waring/logs.Warning/g' {} +

# macOS 使用
find internal/ -name "*.go" -type f -exec sed -i '' 's/logs\.Waring/logs.Warning/g' {} +
```

#### 验证步骤

```bash
# 1. 确认无拼写错误残留
grep -r "Waring" internal/
# 应该返回空

# 2. 编译检查
go build ./...

# 3. 运行测试
go test ./...
```

#### 完成条件

- ✅ 函数改名为 `Warning`
- ✅ 所有调用位置更新
- ✅ 编译无错误

---

### P0.3 添加核心功能测试 🧪

**优先级**: 最高
**工作量**: 3-5 天
**目标**: 测试覆盖率达到 30%

#### 测试优先级

**P0 - 必须测试**:
1. ✅ 版本号递增逻辑 - `internal/gitcmd/tag.go:66-94`
2. ✅ 配置管理 - `internal/config/`
3. ✅ 命令执行 - `internal/command/runCommand.go`

**P1 - 应该测试**:
4. ✅ Cherry-pick 逻辑 - `internal/gitcmd/cherry_pick.go`
5. ✅ Merge 逻辑 - `internal/gitcmd/merge.go`

#### 实施步骤

#### 步骤 1: 创建测试工具包

**文件**: `internal/testutil/git_helper.go`

```go
package testutil

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

// SetupTestRepo 创建临时 Git 仓库
func SetupTestRepo(t *testing.T) string {
    t.Helper()

    tmpDir := t.TempDir()

    // 初始化仓库
    cmd := exec.Command("git", "init")
    cmd.Dir = tmpDir
    if err := cmd.Run(); err != nil {
        t.Fatalf("git init failed: %v", err)
    }

    // 配置用户
    runGitCmd(t, tmpDir, "config", "user.name", "Test User")
    runGitCmd(t, tmpDir, "config", "user.email", "test@example.com")

    return tmpDir
}

// CreateTestFile 创建测试文件
func CreateTestFile(t *testing.T, repoDir, filename, content string) {
    t.Helper()

    path := filepath.Join(repoDir, filename)
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        t.Fatalf("failed to create file: %v", err)
    }
}

// CommitFile 提交文件
func CommitFile(t *testing.T, repoDir, filename, message string) string {
    t.Helper()

    runGitCmd(t, repoDir, "add", filename)
    runGitCmd(t, repoDir, "commit", "-m", message)

    // 返回 commit hash
    cmd := exec.Command("git", "rev-parse", "HEAD")
    cmd.Dir = repoDir
    output, _ := cmd.Output()
    return strings.TrimSpace(string(output))
}

// CreateTag 创建 tag
func CreateTag(t *testing.T, repoDir, tagName string) {
    t.Helper()
    runGitCmd(t, repoDir, "tag", tagName)
}

// CreateBranch 创建分支
func CreateBranch(t *testing.T, repoDir, branchName string) {
    t.Helper()
    runGitCmd(t, repoDir, "branch", branchName)
}

// CheckoutBranch 切换分支
func CheckoutBranch(t *testing.T, repoDir, branchName string) {
    t.Helper()
    runGitCmd(t, repoDir, "checkout", branchName)
}

// runGitCmd 辅助函数
func runGitCmd(t *testing.T, dir string, args ...string) {
    t.Helper()

    cmd := exec.Command("git", args...)
    cmd.Dir = dir
    if err := cmd.Run(); err != nil {
        t.Fatalf("git %v failed: %v", args, err)
    }
}
```

---

#### 步骤 2: 版本号递增测试

**文件**: `internal/gitcmd/tag_test.go`

```go
package gitcmd

import (
    "testing"
)

func TestGetIncrementedVersion(t *testing.T) {
    tests := []struct {
        name          string
        latestVersion string
        incrementType string
        want          string
    }{
        // 基本测试
        {
            name:          "increment patch",
            latestVersion: "v1.0.0",
            incrementType: "patch",
            want:          "v1.0.1",
        },
        {
            name:          "increment minor",
            latestVersion: "v1.0.5",
            incrementType: "minor",
            want:          "v1.1.0",
        },
        {
            name:          "increment major",
            latestVersion: "v1.5.3",
            incrementType: "major",
            want:          "v2.0.0",
        },

        // 带前缀测试
        {
            name:          "with release prefix",
            latestVersion: "release-1.0.0",
            incrementType: "patch",
            want:          "release-1.0.1",
        },

        // 带后缀测试
        {
            name:          "with suffix",
            latestVersion: "v1.0.0-beta",
            incrementType: "patch",
            want:          "v1.0.1-beta",
        },

        // 边界测试
        {
            name:          "max patch version",
            latestVersion: "v1.0.99",
            incrementType: "patch",
            want:          "v1.0.100",
        },
        {
            name:          "reset patch on minor increment",
            latestVersion: "v1.5.9",
            incrementType: "minor",
            want:          "v1.6.0",
        },
        {
            name:          "reset minor and patch on major increment",
            latestVersion: "v5.9.9",
            incrementType: "major",
            want:          "v6.0.0",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := getIncrementedVersion(tt.latestVersion, tt.incrementType)

            if got != tt.want {
                t.Errorf("getIncrementedVersion(%q, %q) = %q, want %q",
                    tt.latestVersion, tt.incrementType, got, tt.want)
            }
        })
    }
}

// 集成测试 - 需要真实的 Git 环境
func TestTagWorkflow(t *testing.T) {
    // 跳过如果不在 Git 仓库中
    if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
        t.Skip("Not in a git repository")
    }

    // 使用临时仓库测试完整流程
    // 这里需要 testutil 包支持
}
```

---

#### 步骤 3: 配置管理测试

**文件**: `internal/config/option_test.go`

```go
package config

import (
    "os"
    "path/filepath"
    "testing"
)

func setupTestDB(t *testing.T) (cleanup func()) {
    t.Helper()

    // 使用临时目录
    tmpDir := t.TempDir()
    oldHome := os.Getenv("HOME")

    os.Setenv("HOME", tmpDir)

    // 返回清理函数
    return func() {
        os.Setenv("HOME", oldHome)
    }
}

func TestGetDefaultOptions(t *testing.T) {
    opts := GetDefaultOptions()

    if len(opts) == 0 {
        t.Error("GetDefaultOptions returned empty slice")
    }

    // 验证包含常见的提交类型
    found := make(map[string]bool)
    for _, opt := range opts {
        found[opt.Value] = true
    }

    required := []string{"fix", "feat", "docs", "test"}
    for _, r := range required {
        if !found[r] {
            t.Errorf("Missing required option: %s", r)
        }
    }
}

func TestSaveAndGetOptions(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    // 保存选项
    options := []Option{
        {Label: "fix", Value: "fix", Usage: 10},
        {Label: "feat", Value: "feat", Usage: 5},
        {Label: "docs", Value: "docs", Usage: 2},
    }

    if err := SaveOptions(options); err != nil {
        t.Fatalf("SaveOptions failed: %v", err)
    }

    // 读取选项
    retrieved, err := GetOptions()
    if err != nil {
        t.Fatalf("GetOptions failed: %v", err)
    }

    if len(retrieved) != len(options) {
        t.Errorf("got %d options, want %d", len(retrieved), len(options))
    }

    // 验证按 usage 降序排序
    if len(retrieved) >= 2 {
        if retrieved[0].Usage < retrieved[1].Usage {
            t.Error("Options not sorted by usage DESC")
        }
    }
}

func TestIncrementUsage(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    // 初始化
    options := []Option{
        {Label: "fix", Value: "fix", Usage: 0},
    }
    SaveOptions(options)

    // 增量更新
    if err := IncrementUsage("fix"); err != nil {
        t.Fatalf("IncrementUsage failed: %v", err)
    }

    // 验证结果
    retrieved, _ := GetOptions()
    for _, opt := range retrieved {
        if opt.Value == "fix" {
            if opt.Usage != 1 {
                t.Errorf("Usage = %d, want 1", opt.Usage)
            }
            return
        }
    }

    t.Error("Option 'fix' not found after increment")
}

func TestConcurrentIncrementUsage(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    // 初始化
    options := []Option{
        {Label: "fix", Value: "fix", Usage: 0},
    }
    SaveOptions(options)

    // 并发增量更新
    const concurrency = 10
    done := make(chan bool, concurrency)

    for i := 0; i < concurrency; i++ {
        go func() {
            IncrementUsage("fix")
            done <- true
        }()
    }

    // 等待完成
    for i := 0; i < concurrency; i++ {
        <-done
    }

    // 验证结果（SQLite 事务应该保证原子性）
    retrieved, _ := GetOptions()
    for _, opt := range retrieved {
        if opt.Value == "fix" {
            if opt.Usage != concurrency {
                t.Errorf("Usage = %d, want %d (possible race condition)", opt.Usage, concurrency)
            }
            return
        }
    }
}
```

---

#### 步骤 4: 命令执行测试

**文件**: `internal/command/runCommand_test.go`

```go
package command

import (
    "testing"
)

func TestRunCommand(t *testing.T) {
    tests := []struct {
        name    string
        info    CommandInfo
        wantErr bool
    }{
        {
            name: "simple echo",
            info: CommandInfo{
                Command: "echo",
                Args:    []string{"hello"},
                Message: "Testing echo",
            },
            wantErr: false,
        },
        {
            name: "git version",
            info: CommandInfo{
                Command: "git",
                Args:    []string{"--version"},
                Message: "Testing git",
            },
            wantErr: false,
        },
        {
            name: "nonexistent command",
            info: CommandInfo{
                Command: "nonexistent-command-xyz",
                Args:    []string{},
                Message: "Testing failure",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := RunCommand(tt.info)

            if (err != nil) != tt.wantErr {
                t.Errorf("RunCommand() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

#### 步骤 5: 设置 CI 自动测试

**文件**: `.github/workflows/test.yml`

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

      - name: Generate coverage report
        run: |
          go tool cover -func=coverage.out
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
          echo "Coverage: $COVERAGE"

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 30" | bc -l) )); then
            echo "⚠️  Coverage $COVERAGE% is below threshold 30%"
            echo "This is a warning, not blocking the build yet"
          else
            echo "✅ Coverage $COVERAGE% meets threshold"
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          flags: unittests
```

#### 完成条件

- ✅ 测试覆盖率达到 30%
- ✅ 所有 P0 模块有测试
- ✅ CI 自动运行测试
- ✅ 测试覆盖率报告可查看

---

### P0.4 创建功能文档 📚

**优先级**: 最高
**工作量**: 1-2 天
**影响**: 项目合规性、可维护性

#### 需要创建的文档

```
docs/features/
├── push.md              - Push 相关命令
├── cherry-pick.md       - Cherry-pick 功能
├── merge.md             - 合并功能
├── reset.md             - Reset 功能
├── tag-management.md    - Tag 管理
└── branch-management.md - 分支管理
```

#### 文档模板

每个文档遵循以下结构（根据 CLAUDE.md 规则）：

```markdown
# [功能名称]

## 功能说明
简要描述功能的作用和使用场景

## 实现位置
- **命令入口**: `cmd/fastgit/commands/xxx.go:行号`
- **核心逻辑**: `internal/gitcmd/xxx.go:行号`

## 运行流程
1. 步骤 1
2. 步骤 2
3. 步骤 3

## 配置 / API
配置项说明（如有）

## 使用示例
\`\`\`bash
easyGit xxx
\`\`\`

**交互流程**:
\`\`\`
步骤 1 输出
步骤 2 输出
...
\`\`\`

## 注意事项
- 注意点 1
- 注意点 2
```

#### 示例：Push 功能文档

**文件**: `docs/features/push.md`

```markdown
# Push 功能

## 功能说明

easyGit 提供两种推送方式：
- **push-all**: 自动添加所有更改并推送
- **push-selected**: 选择性推送特定文件

这两个命令简化了 `git add` + `git commit` + `git push` 的流程。

## 实现位置

**Push All**:
- **命令入口**: `cmd/fastgit/commands/push_all.go:8`
- **核心逻辑**: `internal/gitcmd/pushAll.go:19`

**Push Selected**:
- **命令入口**: `cmd/fastgit/commands/push_selected.go:8`
- **核心逻辑**: `internal/gitcmd/pushSelected.go:16`

## 运行流程

### Push All 流程

1. **检查文件状态**
   - 执行 `git status --porcelain`
   - 如果没有更改，显示提示并退出

2. **添加所有文件**
   - 显示进度: "Adding files..."
   - 执行 `git add .`

3. **获取提交信息**
   - 显示交互式选择器：选择提交类型（fix/feat/refactor...）
   - 显示输入框：输入提交描述

4. **提交更改**
   - 构建提交信息：`[类型]: 描述`
   - 执行 `git commit -m "消息"`

5. **拉取远程更新**
   - 执行 `git pull`

6. **推送到远程**
   - 执行 `git push`

### Push Selected 流程

1. **检查文件状态** - 同 Push All
2. **选择文件**
   - 显示文件列表（支持多选）
   - 用户选择要推送的文件
3. **添加选定文件**
   - 对每个文件执行 `git add [文件]`
4. **提交和推送** - 同 Push All 步骤 3-6

## 配置 / API

**提交类型使用统计**:
- 位置: `~/.easyGit.db` → `options` 表
- 自动更新: 每次选择提交类型后调用 `config.IncrementUsage()`
- 排序: 按使用频率排序选项（`ORDER BY usage DESC`）
- 初始化: 第一次使用时自动创建默认选项

## 使用示例

### Push All

\`\`\`bash
$ easyGit push-all
# 或使用别名
$ easyGit pa
\`\`\`

**交互流程**:

\`\`\`
✓ Adding files...
✓ Committing changes...

📝 Select commit type:
  > fix     (修复 Bug)
    feat    (新功能)
    docs    (文档)
    style   (格式)

📝 Commit message:
> 修复用户登录问题

✓ Pulling from remote...
✓ Pushing to remote...
✓ Successfully pushed all changes!
\`\`\`

### Push Selected

\`\`\`bash
$ easyGit push-selected
# 或使用别名
$ easyGit ps
\`\`\`

**交互流程**:

\`\`\`
📁 Select files to push:
  ☑ src/main.go
  ☑ README.md
  ☐ config.yaml
  ☐ test.txt

✓ Adding 2 files...
✓ Committing changes...
...
\`\`\`

## 注意事项

1. **自动添加所有文件** (Push All)
   - 命令会执行 `git add .`，包括所有未跟踪和已修改的文件
   - 确保 `.gitignore` 正确配置

2. **需要远程仓库**
   - 必须已配置远程仓库（origin）
   - 首次推送可能需要 `git push -u origin [branch]`

3. **冲突处理**
   - 如果 `git pull` 有冲突，需要手动解决
   - 解决后重新运行命令

4. **认证**
   - 可能需要输入 Git 凭据
   - 建议配置 SSH 密钥或凭据缓存

5. **提交信息格式**
   - 格式: `[类型]: 描述`
   - 示例: `fix: 修复登录问题`
   - 遵循 Conventional Commits 规范
```

#### 创建脚本

```bash
#!/bin/bash
# 创建功能文档目录和文件

mkdir -p docs/features

# 创建文档文件
cat > docs/features/.gitkeep << EOF
# 功能文档目录
EOF

touch docs/features/push.md
touch docs/features/cherry-pick.md
touch docs/features/merge.md
touch docs/features/reset.md
touch docs/features/tag-management.md
touch docs/features/branch-management.md

echo "✓ 功能文档结构创建完成"
ls -la docs/features/
```

#### 完成条件

- ✅ 创建 `docs/features/` 目录
- ✅ 所有 6 个功能文档创建完成
- ✅ 每个文档遵循规定结构
- ✅ 代码位置引用准确（包含行号）
- ✅ 运行流程清晰
- ✅ 示例完整

---

## 🟡 Phase 1: 应该优化（3-4 天）

### P1.1 优化 Cherry-pick 性能 🚀

**优先级**: 高
**工作量**: 1 天
**收益**: 性能提升 10-50 倍

#### 当前问题

**位置**: `internal/gitcmd/cherry_pick.go:96-267`

```
时间复杂度: O((N + M) × K)
- N = 远程分支数
- M = 本地分支数
- K = 每个分支提交数（最多 30）

Git 命令调用: (N + M) × (1 + K) 次
示例: 10 分支 × 30 提交 = 310 次进程启动

性能影响:
- 小仓库 (< 5 分支): < 1 秒 ✅
- 中型仓库 (10-20 分支): 2-5 秒 🟡
- 大型仓库 (> 20 分支): > 10 秒 ❌
```

#### 优化方案

**核心思想**: 使用 `git log --cherry-pick` 一次性获取所有可 cherry-pick 的提交

**文件**: `internal/gitcmd/cherry_pick.go`

```go
// 优化后的实现
func getAllCommitsForCherryPickOptimized() ([]Commit, error) {
    currentBranch, err := getCurrentBranch()
    if err != nil {
        return nil, err
    }

    // 使用 git log --cherry-pick 一次性获取所有未合并的提交
    // --cherry-pick: 自动过滤已合并的等效提交
    // --all: 检查所有分支
    // --not HEAD: 排除当前分支
    cmd := exec.Command("git", "log",
        "--all",                              // 所有分支
        "--not", "HEAD",                      // 排除当前分支
        "--cherry-pick",                      // 自动过滤已合并
        "--no-merges",                        // 排除合并提交
        "--format=%H|%s|%an|%ae|%ar|%D",     // 自定义格式
        "--max-count=100",                    // 最多 100 个
    )

    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get commits: %w", err)
    }

    return parseCommits(string(output))
}

// 解析提交列表
func parseCommits(output string) ([]Commit, error) {
    lines := strings.Split(strings.TrimSpace(output), "\n")
    commits := make([]Commit, 0, len(lines))

    for _, line := range lines {
        if line == "" {
            continue
        }

        parts := strings.Split(line, "|")
        if len(parts) < 5 {
            continue
        }

        commit := Commit{
            Hash:        parts[0],
            Subject:     parts[1],
            AuthorName:  parts[2],
            AuthorEmail: parts[3],
            Time:        parts[4],
            Branch:      extractBranch(parts[5]),  // 从 refs 提取分支名
        }

        commits = append(commits, commit)
    }

    return commits, nil
}

// 从 refs 中提取分支名
func extractBranch(refs string) string {
    // refs 格式: "origin/feature-x, feature-x" 或 "HEAD -> main, origin/main"
    if refs == "" {
        return ""
    }

    parts := strings.Split(refs, ",")
    for _, ref := range parts {
        ref = strings.TrimSpace(ref)

        // 跳过 HEAD 和远程分支
        if strings.HasPrefix(ref, "HEAD") || strings.HasPrefix(ref, "origin/") {
            continue
        }

        return ref
    }

    // 如果没有本地分支，返回第一个远程分支（去掉 origin/）
    if len(parts) > 0 {
        ref := strings.TrimSpace(parts[0])
        return strings.TrimPrefix(ref, "origin/")
    }

    return ""
}
```

#### 性能基准测试

**文件**: `internal/gitcmd/cherry_pick_benchmark_test.go`

```go
package gitcmd

import (
    "os/exec"
    "testing"
)

func BenchmarkGetAllCommitsOld(b *testing.B) {
    // 需要在真实 Git 仓库中运行
    if !isGitRepo() {
        b.Skip("Not in a git repository")
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        getAllCommitsForCherryPick()  // 旧实现
    }
}

func BenchmarkGetAllCommitsOptimized(b *testing.B) {
    if !isGitRepo() {
        b.Skip("Not in a git repository")
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        getAllCommitsForCherryPickOptimized()  // 新实现
    }
}

func isGitRepo() bool {
    cmd := exec.Command("git", "rev-parse", "--git-dir")
    return cmd.Run() == nil
}
```

运行基准测试：

```bash
# 在项目根目录运行
go test -bench=. -benchmem ./internal/gitcmd/

# 预期输出:
# BenchmarkGetAllCommitsOld-8         1    5234567890 ns/op
# BenchmarkGetAllCommitsOptimized-8   10    123456789 ns/op
# 性能提升: ~42 倍
```

#### 完成条件

- ✅ Git 命令调用从 300+ 减少到 1 次
- ✅ 性能基准测试显示 10 倍以上提升
- ✅ 功能正常（能正确获取和过滤提交）
- ✅ 大型仓库测试通过

---

### P1.2 消除重复代码 🔄

**优先级**: 中
**工作量**: 1-2 天
**收益**: 降低维护成本

#### 重复 1: Git 工作流

**位置**: `pushAll.go` 和 `pushSelected.go`

**创建通用工作流**:

**文件**: `internal/gitcmd/workflow.go`

```go
package gitcmd

import (
    "github.com/KevinYouu/easyGit/internal/command"
    "github.com/KevinYouu/easyGit/internal/i18n"
)

// GitWorkflow Git 工作流配置
type GitWorkflow struct {
    AddCommand     []string  // git add 参数
    CommitMessage  string    // 提交信息
    NeedPull       bool      // 是否需要 pull
    NeedPush       bool      // 是否需要 push
}

// Execute 执行 Git 工作流
func (w *GitWorkflow) Execute() error {
    steps := []command.CommandInfo{}

    // 1. Add
    if len(w.AddCommand) > 0 {
        steps = append(steps, command.CommandInfo{
            Command: "git",
            Args:    w.AddCommand,
            Message: i18n.T("push.all.adding.files"),
        })
    }

    // 2. Commit
    if w.CommitMessage != "" {
        steps = append(steps, command.CommandInfo{
            Command: "git",
            Args:    []string{"commit", "-m", w.CommitMessage},
            Message: i18n.T("push.all.committing.changes"),
        })
    }

    // 3. Pull
    if w.NeedPull {
        steps = append(steps, command.CommandInfo{
            Command: "git",
            Args:    []string{"pull"},
            Message: i18n.T("push.all.pulling.remote"),
        })
    }

    // 4. Push
    if w.NeedPush {
        steps = append(steps, command.CommandInfo{
            Command: "git",
            Args:    []string{"push"},
            Message: i18n.T("push.all.pushing.remote"),
        })
    }

    // 执行所有步骤
    return command.RunCommandsWithProgress(steps)
}
```

**重构 pushAll.go**:

```diff
 func PushAll() error {
     // ... 检查文件状态

     // 获取提交信息
     commitType := selectCommitType()
     message := getCommitMessage()
     fullMessage := fmt.Sprintf("%s: %s", commitType, message)

-    // 创建命令序列
-    commands := []command.CommandInfo{
-        {Command: "git", Args: []string{"add", "."}},
-        {Command: "git", Args: []string{"commit", "-m", fullMessage}},
-        {Command: "git", Args: []string{"pull"}},
-        {Command: "git", Args: []string{"push"}},
-    }
-
-    return command.RunCommandsWithProgress(commands)
+    // 使用通用工作流
+    workflow := &GitWorkflow{
+        AddCommand:    []string{"add", "."},
+        CommitMessage: fullMessage,
+        NeedPull:      true,
+        NeedPush:      true,
+    }
+
+    return workflow.Execute()
 }
```

#### 完成条件

- ✅ Git 工作流抽象完成
- ✅ pushAll 和 pushSelected 使用统一工作流
- ✅ 代码行数减少
- ✅ 功能正常运行

---

### P1.3 改进错误处理 💬

**优先级**: 中
**工作量**: 1 天
**收益**: 用户体验提升

#### 问题示例

```go
// ❌ 技术性错误直接暴露
return fmt.Errorf("get latest tag error: %w", err)
return fmt.Errorf("getFileStatuses: %w", err)
```

#### 改进方案

**创建错误包装器**:

**文件**: `internal/errors/errors.go`

```go
package errors

import (
    "fmt"
    "github.com/KevinYouu/easyGit/internal/i18n"
)

// UserError 用户友好的错误
type UserError struct {
    Message string  // 用户可见的错误信息
    Hint    string  // 解决建议
    Err     error   // 原始错误（用于日志）
}

func (e *UserError) Error() string {
    if e.Hint != "" {
        return fmt.Sprintf("%s\n\n💡 %s", e.Message, e.Hint)
    }
    return e.Message
}

func (e *UserError) Unwrap() error {
    return e.Err
}

// New 创建用户友好错误
func New(messageKey string, hint string, err error) *UserError {
    return &UserError{
        Message: i18n.T(messageKey),
        Hint:    hint,
        Err:     err,
    }
}
```

**使用示例**:

```diff
 func getLatestTag() (string, error) {
     cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
     output, err := cmd.Output()
     if err != nil {
-        return "", fmt.Errorf("get latest tag error: %w", err)
+        return "", errors.New(
+            "tag.error.not.found",
+            i18n.T("tag.hint.create.initial"),
+            err,
+        )
     }
     return string(output), nil
 }
```

**翻译键**:

```go
// internal/i18n/zh.go
"tag.error.not.found": "未找到任何 tag",
"tag.hint.create.initial": "提示: 使用 'git tag v0.0.0' 创建初始 tag",

// internal/i18n/en.go
"tag.error.not.found": "No tags found",
"tag.hint.create.initial": "Hint: Use 'git tag v0.0.0' to create initial tag",
```

#### 完成条件

- ✅ 创建错误包装器
- ✅ 主要错误使用友好提示
- ✅ 错误信息国际化
- ✅ 用户测试反馈良好

---

## 🟢 Phase 2: 可以改进（4-6 天）

### P2.1 拆分大文件 📂

**优先级**: 低
**工作量**: 2-3 天

#### 需要拆分的文件

1. **`internal/gitcmd/cherry_pick.go`** (383 行)
   ```
   拆分为:
   cherry_pick/
   ├── query.go      - Git 查询逻辑
   ├── filter.go     - 提交过滤
   ├── ui.go         - 用户交互
   └── execute.go    - 执行 cherry-pick
   ```

2. **`internal/command/progress_model.go`** (437 行)
   - 按功能拆分为多个文件

3. **`internal/theme/theme.go`** (481 行)
   - 按主题类型拆分

**完成条件**:
- ✅ 单个文件不超过 300 行
- ✅ 功能正常运行

---

### P2.2 统一输出格式 🎨

**优先级**: 低
**工作量**: 1 天

#### 问题

73 个位置直接使用 `fmt.Println`

#### 方案

扩展 `logs` 包：

```go
// internal/logs/logs.go
func Success(text string) {
    fmt.Println(colors.RenderColor("green", "✓ " + text))
}

func Error(text string) {
    fmt.Println(colors.RenderColor("red", "✗ " + text))
}

func Info(text string) {
    fmt.Println(colors.RenderColor("blue", "ℹ " + text))
}

func Step(step int, total int, text string) {
    fmt.Printf("[%d/%d] %s\n", step, total, text)
}
```

**完成条件**:
- ✅ 统一使用 logs 包
- ✅ 输出格式一致

---

### P2.3 改进代码注释 📝

**优先级**: 低
**工作量**: 1-2 天

**目标**: 为关键函数添加 godoc 注释

**示例**:

```go
// GetIncrementedVersion 根据递增类型计算新版本号
//
// 支持的版本格式:
//   - 标准格式: v1.2.3
//   - 带前缀: release-1.2.3
//   - 带后缀: v1.2.3-beta
//
// 参数:
//   - latestVersion: 当前最新版本号
//   - incrementType: 递增类型 (major/minor/patch)
//
// 返回值:
//   - string: 新版本号
//   - error: 解析错误时返回
//
// 示例:
//   version := getIncrementedVersion("v1.0.0", "patch")
//   // 返回: "v1.0.1"
func getIncrementedVersion(latestVersion string, incrementType string) (string, error) {
    // ...
}
```

**完成条件**:
- ✅ 所有导出函数有 godoc
- ✅ 复杂逻辑有注释

---

## 📊 实施时间表

### Week 1: Phase 0 高优先级

| 日期 | 任务 | 状态 |
|------|------|------|
| Day 1 上午 | P0.1 清理调试代码 (5分钟) | [ ] |
| Day 1 上午 | P0.2 修复拼写错误 (10分钟) | [ ] |
| Day 1 下午 - Day 3 | P0.3 添加核心测试 | [ ] |
| Day 4-5 | P0.4 创建功能文档 | [ ] |

### Week 2: Phase 1 性能优化

| 日期 | 任务 | 状态 |
|------|------|------|
| Day 6 | P1.1 优化 Cherry-pick | [ ] |
| Day 7-8 | P1.2 消除重复代码 | [ ] |
| Day 9 | P1.3 改进错误处理 | [ ] |

### Week 3: Phase 2 代码整洁

| 日期 | 任务 | 状态 |
|------|------|------|
| Day 10-12 | P2.1 拆分大文件 | [ ] |
| Day 13 | P2.2 统一输出格式 | [ ] |
| Day 14-15 | P2.3 改进代码注释 | [ ] |

---

## ✅ 验收标准

### Phase 0 完成标准

- [x] ✅ 无调试代码残留
- [x] ✅ 无拼写错误
- [x] ✅ 测试覆盖率 ≥ 30%
- [x] ✅ 所有功能有文档

### Phase 1 完成标准

- [x] ✅ Cherry-pick 性能提升 10 倍以上
- [x] ✅ 主要重复代码消除
- [x] ✅ 错误信息用户友好

### Phase 2 完成标准

- [x] ✅ 无超过 300 行的文件
- [x] ✅ 输出格式统一
- [x] ✅ 关键函数有 godoc

---

## 🎯 成功指标

### 短期（1 周）
- ✅ 代码质量提升（无调试代码、无拼写错误）
- ✅ 测试覆盖率: 0% → 30%
- ✅ 文档完整性: 0% → 100%

### 中期（2-3 周）
- ✅ Cherry-pick 性能: 10s → 1s
- ✅ 重复代码减少 50%
- ✅ 错误信息用户友好

### 长期（1-2 月）
- ✅ 测试覆盖率: 30% → 60%
- ✅ 代码组织优化
- ✅ 用户体验显著提升

---

## 💡 关于配置存储

### SQLite 是正确的选择 ✅

经过技术分析，当前使用 SQLite (modernc.org/sqlite) 是**最佳方案**：

**你从 JSON → SQLite 的理由完全正确**:
- ✅ JSON 全文件读写性能差
- ✅ JSON 无法原子性增量更新
- ✅ JSON 排序需要加载整个文件

**SQLite 优势（CLI 工具场景）**:
- ✅ 启动极快（< 1ms）- CLI 工具的生命线
- ✅ 单文件存储 - 用户友好
- ✅ SQL 查询 - 代码简洁（ORDER BY usage DESC）
- ✅ ACID 事务 - 原子性保证
- ✅ 无 CGO - 跨平台编译友好

**其他方案不适合**:
- ❌ BadgerDB: 启动慢（10-50ms），多文件存储，API 复杂
- ❌ BoltDB: 已停止维护（2017年归档）
- ❌ JSON/YAML: 性能问题，无原子更新

**结论**: 保留 SQLite，无需迁移配置存储。

---

**优化计划制定完成**

*建议从 Phase 0 的快速修复开始（P0.1 和 P0.2），总共只需 15 分钟，立即见效！*
