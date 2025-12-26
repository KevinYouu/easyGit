# TODO 任务清单

## [x] 语言设置持久化 (2025-12-26)

### 已完成任务

- [x] 分析现有语言设置机制
- [x] 在 config 包中添加语言设置存储
- [x] 修改 i18n 包实现三级优先级
- [x] 添加语言设置命令
- [x] 编写测试用例
- [x] 更新文档

### 功能特性

#### 三级语言优先级

1. **运行时参数(最高优先级)**
   - `--language=zh` 或 `-l zh`
   - 临时生效,不持久化

2. **数据库配置**
   - 通过 `set-language` 命令设置
   - 持久化存储在 `~/.easyGit.db`

3. **系统环境变量(最低优先级)**
   - 自动检测 `LC_ALL`, `LC_MESSAGES`, `LANG`
   - 默认为英文

#### 新增命令

```bash
# 查看当前语言设置
./easyGit set-language

# 设置为中文(持久化)
./easyGit set-language zh

# 设置为英文(持久化)
./easyGit set-language en

# 临时使用中文(不影响数据库)
./easyGit version --language=zh
```

### 代码修改

#### 新增文件

- `internal/config/language.go` - 语言数据库存储
- `internal/config/language_test.go` - 配置测试
- `cmd/easygit/commands/set_language.go` - 设置命令
- `internal/i18n/language_priority_test.go` - 优先级测试
- `docs/features/语言设置.md` - 功能文档

#### 修改文件

- `internal/config/db.go` - 添加 settings 表
- `internal/i18n/i18n.go` - 实现优先级逻辑
- `internal/i18n/zh.go` - 添加中文翻译
- `internal/i18n/en.go` - 添加英文翻译
- `cmd/easygit/main.go` - 启动时加载数据库语言
- `cmd/easygit/root.go` - 注册新命令

### 数据库结构

新增 `settings` 表:
```sql
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
```

### 测试覆盖

- ✅ 语言保存和读取
- ✅ 优先级测试
- ✅ 并发安全测试
- ✅ 数据持久化测试
- ✅ 所有测试通过

### 文档更新

- ✅ `docs/features/语言设置.md` - 详细功能文档
- ✅ 包含使用示例和场景说明

## [x] 测试用例开发 (2025-12-26)

### 已完成任务

- [x] 检查现有测试覆盖率
- [x] 为 gitcmd 包补充测试用例
- [x] 为 form 包补充测试用例
- [x] 运行所有测试并修复问题
- [x] 执行代码质量检查
- [x] 创建测试文档

### 测试覆盖率提升

#### gitcmd 包
- **之前**: 4.7%
- **之后**: 14.3%
- **提升**: +9.6%

新增测试文件:
- `internal/gitcmd/branch_test.go` - 分支操作测试
- `internal/gitcmd/merge_test.go` - 合并操作测试
- `internal/gitcmd/reset_test.go` - 重置操作测试

#### form 包
- **之前**: 0.0%
- **之后**: 0.0%
- **说明**: 交互式组件难以单元测试,已添加数据验证逻辑测试

新增测试文件:
- `internal/form/form_test.go` - 表单组件测试

#### testutil 工具包增强

新增辅助函数:
- `CreateTempGitRepo(t)` - 创建临时 Git 仓库并返回清理函数
- `RunGitCommand(t, dir, args...)` - 执行 Git 命令的导出版本

### 代码质量检查

全部通过:
- ✅ `go vet ./...` - 无警告
- ✅ `go build ./...` - 编译成功
- ✅ `go test ./...` - 所有测试通过

### 文档更新

新增文档:
- `docs/features/测试用例.md` - 完整的测试用例文档

文档包含:
- 测试覆盖率统计
- 测试用例详解
- 运行测试指南
- 测试策略说明
- 持续改进方向

## [ ] 待优化项

### 测试覆盖率待提升

1. **command 包** (2.0%)
   - 命令执行框架
   - 交互式进度显示

2. **spinner 包** (12.7%)
   - 加载动画组件
   - 状态管理

3. **update 包** (9.8%)
   - 自动更新功能
   - 版本检查逻辑

4. **gitcmd 包** (14.3%)
   - 可继续补充更多边界测试
   - cherry-pick 功能测试
   - push 功能测试

### 功能待开发

- [ ] 添加集成测试
- [ ] 性能测试和基准测试
- [ ] CI/CD 自动化测试
- [ ] Mock 测试框架引入

## 测试统计

### 当前覆盖率汇总

| 包名 | 覆盖率 | 状态 |
|------|--------|------|
| logs | 100.0% | ✅ 优秀 |
| random | 100.0% | ✅ 优秀 |
| theme | 100.0% | ✅ 优秀 |
| i18n | 90.6% | ✅ 良好 |
| colors | 75.0% | ✅ 良好 |
| config | 73.4% | ✅ 良好 |
| version | 42.9% | ⚠️ 中等 |
| gitcmd | 14.3% | ⚠️ 改进中 |
| spinner | 12.7% | ⚠️ 偏低 |
| update | 9.8% | ⚠️ 偏低 |
| command | 2.0% | ⚠️ 偏低 |
| form | 0.0% | ℹ️ 交互式 |

### 新增测试数量

- gitcmd/branch_test.go: 5 个测试用例
- gitcmd/merge_test.go: 6 个测试用例
- gitcmd/reset_test.go: 4 个测试用例
- form/form_test.go: 6 个测试用例

**总计**: 21 个新增测试用例

## 相关命令

```bash
# 运行所有测试
go test ./...

# 查看覆盖率
go test ./... -cover

# 运行特定包测试
go test ./internal/gitcmd -v

# 代码质量检查
go vet ./...
go build ./...
```
