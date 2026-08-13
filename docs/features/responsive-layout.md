# 响应式布局功能

## 功能说明

easyGit 的 TUI 界面支持响应式布局:根据终端实际宽高自动调整表格列数、表单高度与进度条宽度,解决窄屏溢出、宽屏浪费与中文截断乱码等问题。

## 背景问题

| 编号 | 问题 | 说明 |
| ---- | ---- | ---- |
| P0-1 | 按字节截断字符串 | `formatCompactCommit` / `parseCommitInfo` / `reset.go` 使用 `[:n]` 字节切片,中文(3 字节)/emoji(4 字节)被截出半个字符 → 非法 UTF-8 乱码 |
| P0-2 | 窄屏横向溢出 | 布局判定只看高度(`height < 15` 才紧凑),窄高屏(如 40x40)四列布局直接溢出;极窄宽度下 `maxWidth-3` 可能为负 → 切片 panic 风险 |
| P1-1 | 宽屏列浪费 | 消息列封顶 100 列,超宽屏表格挤在左侧、右侧大片空白,且不居中 |
| P1-2 | 高屏行浪费 | huh `Select` 默认高度 10、`MultiSelect` 硬上限 8 行,50 行高屏仍只显示 8 行 |
| P2 | 一致性与健壮性 | 魔数散落各处;日期列不截断;AltScreen 由初始高度决定,resize 跨阈值不切换;进度条固定 40 宽,窄屏折行 |

## 实现位置

- 布局纯函数: `internal/form/layout.go`(新增,含全部命名常量)
- 表格选择器: `internal/form/table_select.go` / `table_multi_select.go`
- huh 表单高度: `internal/form/select.go` / `multiSelect.go`
- 进度条自适应: `internal/command/progress_model.go`
- 截断修复: `internal/gitcmd/reset.go`
- 单元测试: `internal/form/layout_test.go`(表驱动)

## 布局规则

### 列布局(只由宽度决定)

| 终端宽度 | 模式 | 列 |
| -------- | ---- | --- |
| `< 60` | 紧凑(单列) | `hash message (date)` 合并为一行 |
| `< 80` | 三列 | hash / message / date(隐藏 author) |
| `>= 80` | 四列 | hash / message / date / author |

### 消息列宽度

`message = clamp(width - 固定列宽 - 内边距, 20, 160)`,宽屏上限由 100 提至 160,富余宽度全部吸收给消息列。

### 宽屏居中

当 `终端宽度 - 表总宽 > 16` 时表格水平居中,否则左对齐(`ShouldCenterTable`)。

### 表格高度

普通模式占满 `height`(多选模式预留标题/空行/底部帮助行 3 行),再与内容行数取小:

```
viewport = min(CalculateTableHeight(高度, 多选), 内容行数)
```

内容不足一屏时表格按内容行数显示(底部不渲染空白),内容超出一屏时占满终端高度并滚动。

### 表单高度

huh v2 字段的 `Height` 语义为 **内容高度(标题行 + 选项数)**,紧凑主题(无边框、无帮助、标题为一行)下按内容与终端取小:

```
Height = clamp(min(选项数 + 1, 终端高度), 最小3, 终端高度)   // +1 为标题行
```

`SelectForm` / `MultiSelectForm` 共用此公式(`formFieldHeight`):

- 大屏(内容不足一屏):按内容高度显示,不渲染底部空白行 —— 有多少显示多少;
- 矮屏(内容超出一屏):占满终端高度,huh 内部滚动,不再因预留高度浪费底部行;
- 极小终端(<3 行):退化为实际终端高度,不渲染越界。

### 进度条宽度

`min(40, max(width-8, 10))`,状态行超宽按显示宽度截断。

## 关键实现细节

### 宽度感知截断 `SafeTruncate`

统一封装 `charmbracelet/x/ansi.Truncate`:按**显示宽度**(宽字符/emoji 按 2 列)截断并附加 `...`,不再按字节切片。修复点:

- `table_select.go`:`formatCompactCommit` / `parseCommitInfo`(message、date、author 均截断,hash 不截)
- `gitcmd/reset.go:126`:`shortMsg[:37]` 字节截断 → `form.SafeTruncate(msg, resetMessageMaxWidth)`
- `gitcmd/rebase.go`:`GetRecentCommits` 的标签**不截断**提交消息(完整保留供 squash 等命令提取默认消息),表格展示层按列宽自行截断

### 表格原子重建(防止渲染越界 panic)

`bubbles/table` 的 `renderRow` 遍历行单元格并按下标访问 `m.cols[i]`,当行单元格数与列数不一致时(如 resize 从四列切到单列)会 panic(index out of range)。

因此表格模型提供 `applyLayout()`(单选)/ `updateLayout()`(多选)统一重建路径:**列与行在同一个代码路径中构建**,并保留样式与光标(`SetCursor`)。多选的空格切换只走 `rebuildRows()` 重绘行,不重建表格。

### 统一 AltScreen

`TableSelectForm` / `TableMultiSelectForm` 统一全屏:bubbletea v2 中全屏模式改为**声明式**——在 `View()` 返回的 `tea.View` 上设置 `AltScreen = true`,不再使用 `tea.WithAltScreen()` 选项。紧凑模式只影响列布局,不决定是否全屏(resize 跨阈值也不会模式错位)。

### 依赖升级(2026-08)

全部 TUI 依赖升级至 v2 并适配破坏性变更:

| 依赖 | 旧版 | 新版(模块路径) |
| ---- | ---- | --------------- |
| bubbletea | v1.3.10 | [v2.0.8](https://charm.land/bubbletea) `charm.land/bubbletea/v2` |
| bubbles | v1.0.0 | [v2.1.1](https://charm.land/bubbles) `charm.land/bubbles/v2` |
| huh | v1.0.0 | [v2.0.3](https://charm.land/huh) `charm.land/huh/v2` |
| lipgloss | v1.1.0 | [v2.0.6](https://charm.land/lipgloss) `charm.land/lipgloss/v2` |

适配要点:

- `View()` 返回 `tea.View` 结构体(内容经 `tea.NewView(s)` 构造)
- 全屏/AltScreen 改为 View 声明式字段
- `tea.KeyMsg` 更名为 `tea.KeyPressMsg`
- huh v2 字段直接 `View()` 时 viewport 未初始化(渲染为空),必须经 `Form.Update(WindowSizeMsg)` 驱动
- 表格模型显式 `SetWidth`(v2 viewport 宽度为 0 时渲染为空)
- huh v2 修复了 v1 的滚动 bug(光标跟随可见区),高屏展示更多选项

## 测试

`internal/form/layout_test.go` 表驱动覆盖:

- `LayoutMode` 判定:40x10 / 40x40 / 200x10 / 120x30 / 300x60
- `CalculateMessageWidth` 边界:20 / 35 / 60 / 80 / 120 / 200
- `SafeTruncate` / `parseCommitInfo` / `formatCompactCommit`:中文、emoji、ASCII 混合,断言 `utf8.ValidString` 且 `lipgloss.Width ≤ 列宽`;极窄宽度(20)不 panic

`internal/form/render_test.go` 渲染级测试(真实构造 huh 表单/表格模型,按终端尺寸矩阵渲染断言):

- `Select` / `MultiSelect`:各终端高度下可见选项数符合公式(大屏按内容显示、矮屏占满滚动)
- `TableSelect` / `TableMultiSelect`:窄屏三列、宽屏四列、紧凑单列与居中均无 ANSI 泄漏/越界
- 内置命令全覆盖:按各命令实际表单(单选/多选/表格/输入/确认)与选项规模渲染断言,含 drop/cherry-pick/reset/merge 等
- `parseCommitInfo` 支持单行标签(reset 模式等),避免多列布局下整行空白

验证方式:`make all`。
