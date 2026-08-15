# 测试用例补充指南

## 当前测试覆盖率概览

### 覆盖率统计 (2025-12-26)

| 包名 | 覆盖率 | 测试文件 | 状态 | 目标 |
|------|--------|---------|------|------|
| **i18n** | 93.0% | ✅ 完善 | 🟢 优秀 | 保持 |
| **logs** | 100.0% | ✅ 完善 | 🟢 优秀 | 保持 |
| **random** | 100.0% | ✅ 完善 | 🟢 优秀 | 保持 |
| **theme** | 100.0% | ✅ 完善 | 🟢 优秀 | 保持 |
| **config** | 75.3% | ✅ 完善 | 🟢 良好 | 保持 |
| **version** | 42.9% | ⚠️ 部分 | 🟡 中等 | 60% |
| **gitcmd** | 14.3% | ⚠️ 部分 | 🔴 较低 | 60% |
| **update** | 9.8% | ⚠️ 部分 | 🔴 较低 | 30% |
| **command** | 2.0% | ⚠️ 部分 | 🔴 极低 | 50% |
| **form** | 0.0% | ❌ 缺失 | 🔴 无 | 10% |
| **commands** | 0.0% | ❌ 缺失 | 🔴 无 | 30% |

## 是否需要继续编写测试用例?

### 建议: 是的,但要有策略地进行

#### 优先级评估

**必须补充 (Sprint 1):**
1. ✅ **gitcmd 包** - 核心 Git 操作逻辑
   - 当前仅 14.3%,风险高
   - 包含关键业务逻辑
   - 目标: 60%

2. ✅ **command 包** - 命令执行框架
   - 当前仅 2.0%,几乎无覆盖
   - 负责所有命令的执行
   - 目标: 50%

4. ⭐ **update 包** - 自动更新逻辑
   - 关键功能
   - 目标: 30%

**可选补充 (Sprint 3+):**
5. 🔵 **commands 包** - CLI 命令定义
   - 主要是胶水代码
   - 集成测试更合适
   - 目标: 30%

6. 🔵 **form 包** - TUI 组件
   - 交互式组件难以单元测试
   - 重构后再考虑
   - 目标: 10%

## 测试用例编写指南

### gitcmd 包测试补充方案

#### 1. pushAll.go 测试

```go
func TestPushAll(t *testing.T) {
    // 创建测试仓库
    tempDir, cleanup := testutil.CreateTempGitRepo(t)
    defer cleanup()

    // 设置测试环境
    // ...

    tests := []struct {
        name        string
        setupRepo   func(t *testing.T, dir string)
        mockInput   func() // 模拟用户输入
        wantErr     bool
        checkResult func(t *testing.T, dir string)
    }{
        {
            name: "successful push with changes",
            setupRepo: func(t *testing.T, dir string) {
                // 创建文件并提交
            },
            wantErr: false,
        },
        // 更多测试场景...
    }
}
```

**需要测试的场景:**
- ✅ 成功推送场景
- ✅ 无变更时的行为
- ✅ 推送冲突处理
- ✅ 网络错误处理
- ✅ 用户取消操作

#### 2. pushSelected.go 测试

**测试重点:**
- 文件选择逻辑
- 部分文件提交
- 空选择处理

#### 3. cherry_pick.go 测试

**测试重点:**
- 提交选择
- cherry-pick 冲突
- 多提交 cherry-pick

### command 包测试补充方案

#### 1. RunCmd 测试

```go
func TestRunCmd(t *testing.T) {
    tests := []struct {
        name       string
        command    string
        args       []string
        successMsg string
        wantErr    bool
    }{
        {
            name:       "successful command",
            command:    "echo",
            args:       []string{"hello"},
            successMsg: "Success",
            wantErr:    false,
        },
        {
            name:    "failed command",
            command: "false",
            args:    []string{},
            wantErr: true,
        },
        // 更多场景...
    }
}
```

#### 2. RunMultipleCommands 测试

**测试重点:**
- 命令序列执行
- 中间失败处理
- 进度显示
- 错误传播

## 测试策略建议

### 1. 单元测试优先原则

**适合单元测试:**
- ✅ 纯函数逻辑
- ✅ 数据转换
- ✅ 错误处理
- ✅ 状态管理

**不适合单元测试:**
- ❌ UI 渲染逻辑
- ❌ 用户交互流程
- ❌ 复杂的集成场景

### 2. 测试覆盖率目标

**不要盲目追求 100%:**
- 核心业务逻辑: 60-80%
- 工具函数: 80-100%
- UI 组件: 10-30%
- 胶水代码: 20-40%

**整体目标: 40-50% 即可**

### 3. 测试金字塔

```
        /\
       /  \  E2E 测试 (5%)
      /----\
     / 集成  \  集成测试 (25%)
    /--------\
   /   单元   \  单元测试 (70%)
  /____________\
```

**当前状态:**
- 单元测试: ✅ 已有基础
- 集成测试: ❌ 缺失
- E2E 测试: ❌ 缺失

**建议:**
1. 先完善单元测试 (gitcmd, command)
2. 再添加集成测试 (完整命令流程)
3. 最后考虑 E2E (可选)

### 4. Mock vs 真实环境

**使用 Mock:**
- 外部 API 调用
- 网络请求
- 文件系统操作 (可选)

**使用真实环境:**
- Git 操作 (使用临时仓库)
- 数据库操作 (使用测试数据库)
- 配置读写

**我们的选择:**
- ✅ Git 操作使用真实临时仓库
- ✅ 数据库使用真实临时文件
- ✅ 不使用复杂的 Mock 框架

## 测试工具和辅助函数

### 现有测试工具

1. **testutil.CreateTempGitRepo(t)**
   - 创建临时 Git 仓库
   - 自动清理
   - 配置测试用户

2. **testutil.RunGitCommand(t, dir, args...)**
   - 执行 Git 命令
   - 自动错误处理

3. **config.SetTestDBPath(path)**
   - 设置测试数据库
   - 隔离测试环境

### 需要补充的测试工具

```go
// 建议添加到 testutil/

// MockUserInput 模拟用户输入
func MockUserInput(inputs ...string) func()

// AssertGitStatus 验证 Git 状态
func AssertGitStatus(t *testing.T, dir string, want GitStatus)

// AssertFileExists 验证文件存在
func AssertFileExists(t *testing.T, path string)

// SetupTestBranches 创建测试分支
func SetupTestBranches(t *testing.T, dir string, branches ...string)
```

## 测试执行和 CI/CD

### 本地测试

```bash
# 运行所有测试
go test ./...

# 运行特定包
go test ./internal/gitcmd -v

# 查看覆盖率
go test ./... -cover

# 生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 只运行快速测试
go test ./... -short

# 并行测试
go test ./... -parallel 4
```

### CI 配置建议

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test ./... -cover -race

      - name: Check coverage
        run: |
          go test ./... -coverprofile=coverage.out
          go tool cover -func=coverage.out | grep total
```

## 测试最佳实践

### 1. 测试命名

```go
// ✅ 好的命名
func TestPushAll_WithChanges_Success(t *testing.T)
func TestPushAll_NoChanges_ReturnsError(t *testing.T)
func TestPushAll_UserCancel_ReturnsNil(t *testing.T)

// ❌ 不好的命名
func TestPushAll1(t *testing.T)
func TestPush(t *testing.T)
```

### 2. 测试结构

```go
func TestFunction(t *testing.T) {
    // Arrange (准备)
    setup := prepareTestData()

    // Act (执行)
    result := FunctionUnderTest(setup)

    // Assert (验证)
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### 3. 表驱动测试

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"case 1", "input1", "output1", false},
        {"case 2", "input2", "output2", false},
        {"error case", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 4. 清理资源

```go
func TestWithCleanup(t *testing.T) {
    // 使用 t.Cleanup 确保清理
    tmpDir := t.TempDir() // 自动清理

    t.Cleanup(func() {
        // 额外的清理工作
    })

    // 或使用 defer
    resource := acquireResource()
    defer resource.Close()
}
```

## 测试覆盖率提升计划

### 短期目标 (1-2周)

- [x] 当前整体: ~35%
- [ ] 目标: 45%
- [ ] 重点: gitcmd (14% → 60%), command (2% → 50%)

### 中期目标 (1个月)

- [ ] 目标: 50%
- [ ] 增加集成测试
- [ ] 完善边界测试

### 长期目标 (3个月)

- [ ] 目标: 55-60%
- [ ] E2E 测试覆盖
- [ ] 性能测试
- [ ] 压力测试

## 总结和建议

### 当前状态评估

✅ **已经做得好的:**
- 有基础的测试框架
- 核心工具包覆盖率高
- 测试辅助函数完善

⚠️ **需要改进的:**
- gitcmd 和 command 包覆盖不足
- 缺少集成测试
- 测试文档不完整

### 行动建议

1. **立即执行 (本周):**
   - 补充 gitcmd 包测试
   - 补充 command 包测试
   - 目标: 整体覆盖率 → 45%

2. **近期执行 (2周内):**
   - 补充 update 测试
   - 添加基本集成测试
   - 目标: 整体覆盖率 → 50%

3. **未来考虑:**
   - UI 组件测试 (重构后)
   - E2E 自动化测试
   - 性能基准测试

**结论: 是的,需要继续编写测试用例,但要聚焦在核心业务逻辑 (gitcmd, command),而不是追求所有代码的 100% 覆盖。**
