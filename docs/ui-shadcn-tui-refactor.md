# UI 重构计划：基于 Shadcn 风格的 TUI 设计语言

## 1. 目标与背景

当前项目的 UI 是基于 bubbletea 构建的终端用户界面（TUI）。为了提升整体界面的美观度、一致性和可用性，我们引入了类似于 Shadcn/UI 的极简主义设计语言。

由于终端环境的限制（无真实阴影、基于字符网格、颜色受限），我们将 Shadcn 的视觉结构和交互模式映射为 TUI 的设计令牌（Design Tokens）和组件规范。

## 2. TUI 设计语言规范 (已实现)

### 2.1 设计令牌 (Design Tokens)

统一定义了全局样式令牌，并支持 Neutral Dark / Neutral Light 双色板（默认跟随终端背景自动切换，可经 `--theme` 标志或配置中心强制）：

核心颜色令牌 (Neutral Dark)：
`foreground`: `#fafafa` (高亮白文字，Neutral 50)
`primary`: `#fafafa` (主操作/高亮文字)
`mutedForeground`: `#a3a3a3` (弱化提示文本，Neutral 400)
`border`: `#404040` (边框/分隔线，Neutral 700，增强可见度)
`input`: `#404040` (输入边界色)
`selection`: `#404040` (选中项/焦点行背景色) \* `selectionForeground`: `#fafafa` (选中项文字前景色)
`selectionMuted`: `#d4d4d4` (选中/聚焦行上的弱化前景，Neutral 300)

核心颜色令牌 (Neutral Light)：
`foreground`: `#18181b` (主文字，Neutral 900)
`primary`: `#18181b` (主操作/高亮文字)
`mutedForeground`: `#737373` (弱化提示文本，Neutral 500)
`border`: `#e4e4e7` (边框/分隔线，Neutral 200)
`selection`: `#e4e4e7` (选中项/焦点行背景色) \* `selectionForeground`: `#18181b` (选中项文字前景色)
`selectionMuted`: `#525252` (选中/聚焦行上的弱化前景，Neutral 600)

Diff 颜色令牌：
暗色: `diffAddedBg`: `#1a3a1a` (暗绿色背景) / `diffRemovedBg`: `#4a1515` (暗红色背景)
亮色: `diffAddedBg`: `#dcfce7` (Green 100) / `diffRemovedBg`: `#fee2e2` (Red 100)

语义别名(两模式一致)：
`success`: `#10b981` (Emerald 500)
`warning`: `#f59e0b` (Amber 500)
`destructive`: `#ef4444` (Shadcn Red)
`info`: `#3b82f6` (Blue 500)

模式机制：`theme.ApplyMode(auto|dark|light)` 切换色板并重建包级样式；令牌调用时求值自动生效。自动检测使用 lipgloss `HasDarkBackground`(OSC 11 查询，失败回退深色)。优先级：`--theme` 标志 > 配置中心(settings 表) > 自动检测。

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
默认容器（`Panel`）使用 `colors.border` (#404040)。
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
