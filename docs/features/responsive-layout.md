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
- 截断修复: `internal/gitcmd/reset.go`、`internal/gitcmd/rebase.go`
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

普通模式 `height - 4`,多选模式再预留标题/底部帮助行(`CalculateTableHeight`)。

### 表单高度

- `SelectForm`: `Height(min(max(len(options), 3), height-8))`
- `MultiSelectForm`: 移除 8 行硬上限,改 `min(max(len(options)+1, 3), height-8)`

### 进度条宽度

`min(40, max(width-8, 10))`,状态行超宽按显示宽度截断。

## 关键实现细节

### 宽度感知截断 `SafeTruncate`

统一封装 `charmbracelet/x/ansi.Truncate`:按**显示宽度**(宽字符/emoji 按 2 列)截断并附加 `...`,不再按字节切片。修复点:

- `table_select.go`:`formatCompactCommit` / `parseCommitInfo`(message、date、author 均截断,hash 不截)
- `gitcmd/reset.go:126`:`shortMsg[:37]` 字节截断 → `form.SafeTruncate(msg, resetMessageMaxWidth)`
- `gitcmd/rebase.go:130`:同类字节截断第二处实例 → `form.SafeTruncate(message, rebaseMessageMaxWidth)`

### 表格原子重建(防止渲染越界 panic)

`bubbles/table` 的 `renderRow` 遍历行单元格并按下标访问 `m.cols[i]`,当行单元格数与列数不一致时(如 resize 从四列切到单列)会 panic(index out of range)。

因此表格模型提供 `applyLayout()`(单选)/ `updateLayout()`(多选)统一重建路径:**列与行在同一个代码路径中构建**,并保留样式与光标(`SetCursor`)。多选的空格切换只走 `rebuildRows()` 重绘行,不重建表格。

### 统一 AltScreen

`TableSelectForm` / `TableMultiSelectForm` 统一使用 `tea.WithAltScreen()`,紧凑模式只影响列布局,不再由初始高度决定是否全屏(resize 跨阈值也不会模式错位)。

## 测试

`internal/form/layout_test.go` 表驱动覆盖:

- `LayoutMode` 判定:40x10 / 40x40 / 200x10 / 120x30 / 300x60
- `CalculateMessageWidth` 边界:20 / 35 / 60 / 80 / 120 / 200
- `SafeTruncate` / `parseCommitInfo` / `formatCompactCommit`:中文、emoji、ASCII 混合,断言 `utf8.ValidString` 且 `lipgloss.Width ≤ 列宽`;极窄宽度(20)不 panic

验证方式:`make all`。
