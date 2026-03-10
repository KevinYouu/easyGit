# UI 重构计划：基于 Shadcn 风格的 TUI 设计语言

## 1. 目标与背景

当前项目的 UI 是基于 bubbletea 构建的终端用户界面（TUI）。为了提升整体界面的美观度、一致性和可用性，我们引入了类似于 Shadcn/UI 的极简主义设计语言。

由于终端环境的限制（无真实阴影、基于字符网格、颜色受限），我们将 Shadcn 的视觉结构和交互模式映射为 TUI 的设计令牌（Design Tokens）和组件规范。

## 2. TUI 设计语言规范 (已实现)

### 2.1 设计令牌 (Design Tokens)

统一定义了全局样式令牌，并严格遵循 Zinc Dark 色系：

核心颜色令牌 (Zinc Dark)：
`foreground`: `#fafafa` (高亮白文字，Zinc 50)
`primary`: `#fafafa` (主操作/高亮文字)
`mutedForeground`: `#a1a1aa` (弱化提示文本，Zinc 400)
`border`: `#3f3f46` (边框/分隔线，Zinc 700，增强可见度)
`input`: `#3f3f46` (输入边界色)
`selection`: `#3f3f46` (选中项/焦点行背景色) \* `selectionForeground`: `#fafafa` (选中项文字前景色)

Diff 颜色令牌：
`diffAddedBg`: `#1a3a1a` (暗绿色背景)
`diffRemovedBg`: `#4a1515` (暗红色背景)

语义别名：
`success`: `#10b981` (Emerald 500)
`warning`: `#f59e0b` (Amber 500)
`destructive`: `#ef4444` (Shadcn Red)
`info`: `#3b82f6` (Blue 500)

### 2.2 基础排版组件 (`<Text />`)

封装于 `ui/src/components/base/Text.tsx`，支持以下变体：
`h1/h2`: 粗体 + Primary 色。
`body`: 默认前景色。
`muted`: 弱化文本颜色 + `dimColor`。
`success/error/warning/info`: 语义化着色。

### 2.3 交互指示器 (`<SelectionIndicator />`)

封装于 `ui/src/components/base/SelectionIndicator.tsx`：
统一管理选中指示符 `❯`。
固定宽度为 `2`，确保指示符紧贴选项内容，避免视觉断裂。

## 3. 核心交互模式

选中态 (Selection)：采用 “背景色 (`selection`) + 前景色 (`selectionForeground`) + 指示符 (`❯`)” 的三重强调模式，确保在任何终端环境下焦点都极其清晰。
紧凑布局 (Compact Layout)：
审批弹窗 (Approval Prompt)：移除大块黄色和冗余边框。使用左侧单像素装饰线标识内容块。
垂直间距：收紧非必要的垂直 Margin/Padding，优先展示核心内容（如代码 Diff 或命令预览）。
边框规范：
默认容器（`Panel`）使用 `colors.border` (#3f3f46)。
获取焦点或处于等待审批状态时，边框颜色可动态切换为 `colors.primary` 或 `colors.warning`。

## 4. 实施成果验证

[x] 基础设施: 主题令牌化，排版与指示器组件化。
[x] 原子组件: `Button`, `ProgressBar`, `Badge`, `CodeBlock` 已全部适配新 Token。
[x] 复杂视图: `ModelPicker`, `ProviderPicker`, `SlashCommandMenu`, `ApprovalPrompt` 等已重构为紧凑、清晰的 Shadcn 风格。
[x] 代码质量: 全面清理硬编码颜色，通过 `make check-ui` 验证。

## 5. 维护指南

1.  禁止硬编码: 任何颜色、间距都必须通过 `THEME` 令牌或 `Text` 变体引入。
2.  组件复用: 优先使用 `Panel`, `Text`, `SelectionIndicator` 等基础组件构建新功能。
3.  视觉平衡: 保持 1-2 字符的水平留白，避免内容直接触碰容器边框。
