# easyGit UI 重构计划：Shadcn TUI 风格迁移

本项目将参考 Shadcn/UI 的极简设计语言（Neutral Dark 色系），对当前基于 Bubbletea / Lipgloss / Huh 的 TUI 界面进行全面重构。重构过程分为两个主要阶段。

## 设计目标
- **色系**: 全面转向 Neutral Dark (灰白色系)，取代现有的青蓝/紫色渐变。
- **布局**: 紧凑型布局，减少冗余 Margin/Padding，使用直角边框和左侧单线装饰。
- **交互**: 强化“选中态”的三重强调模式（背景、前景、指示符 `❯`）。
- **一致性**: 所有文本输出必须通过标准排版组件，严禁硬编码颜色。

---

## 第一阶段：组件分析与新版 UI 封装

在不破坏现有业务逻辑的前提下，在底层构建一套全新的 UI 基础库。

### 1.1 设计令牌 (Design Tokens) 重定义
在 `internal/theme/theme.go` 中引入符合 Shadcn 规范的常量：
- `Foreground`: `#fafafa` (Neutral 50) - 主要文字
- `Muted`: `#a3a3a3` (Neutral 400) - 辅助提示文案
- `Border`: `#404040` (Neutral 700) - 默认边框
- `Selection`: `#404040` (Neutral 700) - 选中项背景
- `Primary`: `#fafafa` (Neutral 50) - 关键操作/焦点高亮
- **语义色**: Success (Emerald 500), Error (Red 500), Warning (Amber 500), Info (Blue 500)

### 1.2 封装标准排版组件 (Typography)
创建一套标准文本渲染函数，替代直接使用 `lipgloss.Style`:
- `theme.TextH1(string)`: 粗体 + Primary 色。
- `theme.TextBody(string)`: 默认前景色。
- `theme.TextMuted(string)`: 弱化色 + 倾斜。
- `theme.TextSuccess/Error/Warning/Info(string)`: 语义化着色输出。

### 1.3 封装基础布局组件 (Layout)
- `theme.Panel(content string)`: 提供基础容器，使用 `BorderLeft(true)` 代替全包围边框，作为默认内容块标识。
- `theme.SelectionIndicator()`: 固定宽度为 2 的选中指示符 `❯ `。

### 1.4 重构 Huh 交互主题
实现 `theme.GetShadcnTheme()`：
- **Select/MultiSelect**: 实现 `[背景 Selection] + [前景 Primary] + [指示符 ❯]` 的强调效果。
- **Input**: 极简设计，焦点时仅边框色变化，移除发光和背景填充。
- **Confirm**: 紧凑布局，移除大块颜色块。

### 1.5 简化状态组件
- **Spinner**: 统一使用基础字符动画 (`⠋⠙⠹`)，颜色固定为 `Primary`。
- **Progress**: 移除渐变色，使用 `Primary` 作为进度条前景色。

---

## 第二阶段：全局替换与旧代码清理

在新版 UI 组件准备就绪后，分步骤替换业务层调用。

### 2.1 业务代码样式替换
遍历以下目录，将旧样式（如 `TitleStyle`, `BaseStyle`）替换为新版排版函数：
- `cmd/easygit/`: 主程序入口及子命令说明。
- `internal/gitcmd/`: 各 Git 操作的输出信息渲染。
- `internal/form/`: 确保 `RunSelect`, `RunInput` 等默认加载 `GetShadcnTheme()`。

### 2.2 统一指示符与间距适配
- 检查所有列表选项，确保指示符 `❯` 紧贴文字。
- 走查 `confirm.go` 和 `select.go` 中的水平留白，确保在窄屏终端下依然清晰。

### 2.3 彻底清理旧资产
- 在 `internal/theme/theme.go` 中删除：
    - 旧的渐变色变量 (`GradientStart`, `SecondaryColor` 等)。
    - 旧的复杂样式变量 (`TitleStyle`, `CardStyle`, `FocusedInputStyle` 等)。
    - 多余的动画帧定义 (`Pulse`, `Arrow` 等)。
- 全局搜索并删除所有业务代码中的 `lipgloss.Color("...")` 硬编码。

### 2.4 视觉走查 (Visual Audit)
- 执行交互式提交流程。
- 执行分支/远程列表选择。
- 执行 Cherry-pick 冲突处理流程。
- 确保所有界面在 Neutral Dark 风格下展现出高度的专业感和一致性。

---

## 验收标准
1. [ ] 界面无任何高饱和渐变色。
2. [ ] 选中项焦点清晰，符合“三重强调”模式。
3. [ ] 所有文本颜色符合 Neutral 设计令牌。
4. [ ] 布局紧凑，无非必要空行，使用左侧单线装饰内容块。
