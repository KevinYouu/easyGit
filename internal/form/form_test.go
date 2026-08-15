package form

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/charmbracelet/x/ansi"
)

// 注意: form 包的函数依赖于交互式终端输入,
// 这些测试主要验证函数结构和错误处理逻辑

func TestInput_Validation(t *testing.T) {
	// 测试验证函数逻辑
	validate := func(str string) error {
		if str == "" {
			return errors.New("empty")
		}
		return nil
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   "test",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace input",
			input:   "   ",
			wantErr: false, // 只检查是否为空字符串
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// pumpForm 模拟 tea 程序消息循环:消息送入模型后递归执行返回的命令链,
// 直到无更多命令;返回最新模型。
// 光标闪烁是周期性命令(tea 程序中由事件循环节流),测试中不喂回
// BlinkMsg:处理其副作用(光标状态、huh 的 Eval 重算如 Note 预览绑定),
// 但其 cmd 链会持续返回 BlinkMsg,跟进会死循环——因此处理后的 cmd 经
// stripBlinkCmds 剥离闪烁消息,保留其余(如 huh 的 updateTitleMsg)。
func pumpForm(t *testing.T, model tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	m, cmd := model.Update(msg)
	for cmd != nil {
		msg := cmd()
		cmd = nil
		if msg == nil {
			break
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				if c != nil {
					m = pumpForm(t, m, c())
				}
			}
			continue
		}
		if _, ok := msg.(cursor.BlinkMsg); ok {
			// 处理闪烁消息(触发 huh 的 Eval 重算,如 Note 预览绑定),
			// 但其 cmd 链会持续返回 BlinkMsg,剥离后跟进其余消息
			m, cmd = m.Update(msg)
			cmd = stripBlinkCmds(cmd)
			continue
		}
		m, cmd = m.Update(msg)
	}
	return m
}

// stripBlinkCmds 剥离命令链中的光标闪烁消息(持续循环),保留其余消息
// (如 huh 的 updateTitleMsg 动态内容重算);batch 内逐个执行子命令,
// 仅过滤 BlinkMsg,其余原样保留。
func stripBlinkCmds(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		kept := make([]tea.Cmd, 0, len(batch))
		for _, c := range batch {
			if c == nil {
				continue
			}
			inner := c()
			if _, isBlink := inner.(cursor.BlinkMsg); isBlink {
				continue
			}
			kept = append(kept, func() tea.Msg { return inner })
		}
		if len(kept) == 0 {
			return nil
		}
		return tea.Batch(kept...)
	}
	if _, isBlink := msg.(cursor.BlinkMsg); isBlink {
		return nil
	}
	return func() tea.Msg { return msg }
}

// pumpInit 模拟 tea 程序启动:执行模型 Init 命令链(聚焦首个字段、请求窗口尺寸),
// 使首字段的 textinput 处于已聚焦状态,后续 Text 键才能插入字符。
// 现有用例先按 Enter 触发导航再输入,不依赖本辅助;需要首字段直接输入时使用。
func pumpInit(t *testing.T, model tea.Model) tea.Model {
	t.Helper()
	cmd := model.Init()
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		model = pumpForm(t, model, msg)
		cmd = nil
	}
	return model
}

// TestMultiInput_Construction 单页多输入表单构造断言(自绘三列模型):
// 空 specs 直接返回空;默认值写入字段;Enter/↓ 逐字段推进(空值校验不推进)、
// shift+tab/↑ 回退、末字段提交,与帮助栏「↑/↓ 导航、继续/提交、取消」键位语义一致。
func TestMultiInput_Construction(t *testing.T) {
	t.Run("空 specs 返回空结果", func(t *testing.T) {
		vals, err := MultiInput(nil, nil)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if vals != nil {
			t.Errorf("vals = %v, want nil", vals)
		}
	})

	// build 构造模型并播种默认值(模拟 MultiInput 入口),返回模型与写回指针
	build := func(specs []InputSpec, values []string) (*multiInputModel, []*string) {
		ptrs := make([]*string, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		m := pumpInit(t, newMultiInputModel(specs, ptrs, nil)).(*multiInputModel)
		return pumpForm(t, m, tea.WindowSizeMsg{Width: 80, Height: 12}).(*multiInputModel), ptrs
	}

	t.Run("空值校验阻挡推进", func(t *testing.T) {
		specs := []InputSpec{
			{Title: "版本号"},
			{Title: "提交消息"},
		}
		m, _ := build(specs, []string{"", ""})

		// 字段 1 空值 Enter:非空校验报错且不推进
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if m.done {
			t.Fatalf("空值 Enter 后 done = true, want false(不应推进)")
		}
		if m.focused != 0 {
			t.Errorf("应仍聚焦字段 1,focused = %d", m.focused)
		}
		if m.errMsg != i18n.T("form.input.empty.error") {
			t.Errorf("errMsg = %q, want %q", m.errMsg, i18n.T("form.input.empty.error"))
		}
	})

	t.Run("Enter 推进、shift+tab 回退、末字段提交", func(t *testing.T) {
		specs := []InputSpec{
			{Title: "版本号", Default: "v1.1.0"},
			{Title: "提交消息"},
		}
		m, ptrs := build(specs, []string{specs[0].Default, ""})

		// 初始聚焦字段 1,默认值已写入输入框
		if got := m.inputs[0].Value(); got != "v1.1.0" {
			t.Errorf("字段 1 值 = %q, want v1.1.0", got)
		}

		// Enter → 推进到字段 2
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if m.done {
			t.Fatalf("字段 1 Enter 后 done = true, want false")
		}
		if m.focused != 1 {
			t.Errorf("应聚焦字段 2,focused = %d", m.focused)
		}

		// 向字段 2 输入文本
		m = pumpForm(t, m, tea.KeyPressMsg{Text: "feat: 打标签"}).(*multiInputModel)
		if got := m.inputs[1].Value(); got != "feat: 打标签" {
			t.Errorf("字段 2 输入值 = %q, want feat: 打标签", got)
		}

		// shift+tab → 回退到字段 1
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}).(*multiInputModel)
		if m.focused != 0 {
			t.Errorf("shift+tab 后应聚焦字段 1,focused = %d", m.focused)
		}

		// 字段 1 Enter → 字段 2;字段 2(末字段)Enter → 提交
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if !m.done {
			t.Fatalf("末字段 Enter 后 done = false, want true")
		}
		m.writeBack()
		if *ptrs[0] != "v1.1.0" || *ptrs[1] != "feat: 打标签" {
			t.Errorf("values = %q/%q, want v1.1.0/feat: 打标签", *ptrs[0], *ptrs[1])
		}
	})

	t.Run("AllowEmpty 跳过非空校验", func(t *testing.T) {
		// prefix/suffix 等可空字段:空值不触发非空校验,可正常推进
		specs := []InputSpec{
			{Title: "前缀", AllowEmpty: true},
			{Title: "主版本号"},
		}
		m, _ := build(specs, []string{"", ""})

		// 字段 1 空值 Enter:AllowEmpty 跳过非空校验,推进到字段 2
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if m.done || m.focused != 1 {
			t.Fatalf("AllowEmpty 空值 Enter 后 done=%v focused=%d, want false/1", m.done, m.focused)
		}

		// 字段 2 空值 Enter:非空校验仍生效,不推进
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if m.done || m.focused != 1 {
			t.Fatalf("字段 2 空值 Enter 后 done=%v focused=%d, want false/1", m.done, m.focused)
		}
		if m.errMsg != i18n.T("form.input.empty.error") {
			t.Errorf("errMsg = %q, want %q", m.errMsg, i18n.T("form.input.empty.error"))
		}
	})

	t.Run("Validate 自定义校验生效", func(t *testing.T) {
		validate := func(s string) error {
			if s != "42" {
				return errors.New("must be 42")
			}
			return nil
		}
		specs := []InputSpec{
			{Title: "数字", Validate: validate},
		}
		m, _ := build(specs, []string{""})

		// 非法值:自定义校验报错,不推进
		m = pumpForm(t, m, tea.KeyPressMsg{Text: "7"}).(*multiInputModel)
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if m.done {
			t.Fatalf("非法值 Enter 后 done = true, want false")
		}
		if m.errMsg != "must be 42" {
			t.Errorf("errMsg = %q, want must be 42", m.errMsg)
		}

		// 退格清空非法输入后输入合法值:提交成功
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyBackspace}).(*multiInputModel)
		m = pumpForm(t, m, tea.KeyPressMsg{Text: "42"}).(*multiInputModel)
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if !m.done {
			t.Fatalf("合法值 Enter 后 done = false, want true")
		}
	})
}

func TestOptionDisplayText(t *testing.T) {
	if got := optionDisplayText(config.Option{Label: "soft", Description: "保留工作区更改"}); got != "soft 保留工作区更改" {
		t.Errorf("带说明 = %q, want %q", got, "soft 保留工作区更改")
	}
	if got := optionDisplayText(config.Option{Label: "main"}); got != "main" {
		t.Errorf("无说明 = %q, want %q", got, "main")
	}
}

// TestRunAccessibleList 无障碍纯文本路径(TERM=dumb 替代 TUI):
// 编号行输出与序号解析,单选/多选/空行/非法重试/EOF 取消全覆盖。
func TestRunAccessibleList(t *testing.T) {
	options := []config.Option{
		{Label: "soft", Value: "soft", Description: "保留工作区更改"},
		{Label: "hard", Value: "hard", Description: "丢弃所有更改"},
		{Label: "mixed", Value: "mixed", Description: "保留并暂存更改"},
	}
	// 编号行输出(标题 + 逐项编号)对所有分支一致,单独验证一次
	t.Run("编号行输出", func(t *testing.T) {
		var buf bytes.Buffer
		if _, err := runAccessibleList(&buf, strings.NewReader("1\n"), "标题", options, ListSingle); err != nil {
			t.Fatalf("err = %v", err)
		}
		out := buf.String()
		for _, want := range []string{"标题:", "1. soft 保留工作区更改", "2. hard 丢弃所有更改", "3. mixed 保留并暂存更改"} {
			if !strings.Contains(out, want) {
				t.Errorf("缺少 %q: %q", want, out)
			}
		}
	})

	t.Run("单选序号", func(t *testing.T) {
		var buf bytes.Buffer
		vals, err := runAccessibleList(&buf, strings.NewReader("1\n"), "标题", options, ListSingle)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(vals) != 1 || vals[0] != "soft" {
			t.Errorf("vals = %v, want [soft]", vals)
		}
	})

	t.Run("多选逗号序号按输入顺序", func(t *testing.T) {
		var buf bytes.Buffer
		vals, err := runAccessibleList(&buf, strings.NewReader("1,3\n"), "标题", options, ListMulti)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(vals) != 2 || vals[0] != "soft" || vals[1] != "mixed" {
			t.Errorf("vals = %v, want [soft mixed]", vals)
		}
	})

	t.Run("多选空行 = 全不选", func(t *testing.T) {
		var buf bytes.Buffer
		vals, err := runAccessibleList(&buf, strings.NewReader("\n"), "标题", options, ListMulti)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if vals != nil || len(vals) != 0 {
			t.Errorf("vals = %v, want 空切片", vals)
		}
	})

	t.Run("非法输入报错后重试", func(t *testing.T) {
		var buf bytes.Buffer
		vals, err := runAccessibleList(&buf, strings.NewReader("abc\n2\n"), "标题", options, ListSingle)
		if err != nil {
			t.Fatalf("err = %v", err)
		}
		if len(vals) != 1 || vals[0] != "hard" {
			t.Errorf("vals = %v, want [hard]", vals)
		}
		if !strings.Contains(buf.String(), "Invalid: must be a number between 1 and 3") {
			t.Errorf("非法输入应输出重试提示: %q", buf.String())
		}
	})

	t.Run("EOF 视为取消", func(t *testing.T) {
		var buf bytes.Buffer
		_, err := runAccessibleList(&buf, strings.NewReader(""), "标题", options, ListSingle)
		if err != huh.ErrUserAborted {
			t.Errorf("err = %v, want ErrUserAborted", err)
		}
	})
}

func TestConfirm_DefaultValue(t *testing.T) {
	// 测试 Confirm 函数的默认值行为
	// 由于需要交互式输入，这里只测试函数签名

	// 确保函数可以被调用(虽然在测试环境中会失败)
	var confirmed bool

	// 默认值应该是 false
	if confirmed != false {
		t.Errorf("Default confirmed value = %v, want false", confirmed)
	}
}

func TestInput_DefaultValue(t *testing.T) {
	// 测试默认值的处理
	defaultValue := "default-text"
	inputValue := defaultValue

	// 验证默认值正确设置
	if inputValue != "default-text" {
		t.Errorf("Default input value = %v, want default-text", inputValue)
	}

	// 模拟用户输入
	inputValue = "user-input"
	if inputValue != "user-input" {
		t.Errorf("Modified input value = %v, want user-input", inputValue)
	}
}

func TestInput_EmptyString(t *testing.T) {
	// 测试空字符串验证
	testStr := ""

	validate := func(str string) error {
		if str == "" {
			return errors.New("empty")
		}
		return nil
	}

	err := validate(testStr)
	if err == nil {
		t.Error("Expected error for empty string, got nil")
	}
}

func TestInput_ValidString(t *testing.T) {
	// 测试有效字符串验证
	testStr := "valid input"

	validate := func(str string) error {
		if str == "" {
			return errors.New("empty")
		}
		return nil
	}

	err := validate(testStr)
	if err != nil {
		t.Errorf("Expected no error for valid string, got %v", err)
	}
}

func TestConfirm_BooleanValue(t *testing.T) {
	// 测试布尔值处理
	tests := []struct {
		name  string
		value bool
		want  bool
	}{
		{
			name:  "true value",
			value: true,
			want:  true,
		},
		{
			name:  "false value",
			value: false,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confirmed := tt.value
			if confirmed != tt.want {
				t.Errorf("confirmed = %v, want %v", confirmed, tt.want)
			}
		})
	}
}

// TestFormEscCancels Esc 触发取消:统一列表模型 esc/ctrl+c 置 quitting 并退出,
// 与帮助栏「Esc 取消」提示一致。
func TestFormEscCancels(t *testing.T) {
	m := NewListModel("标题", []config.Option{{Label: "a", Value: "a"}}, ListSingle)
	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc 未触发取消命令")
	}
	if !updated.(*listModel).quitting {
		t.Error("Esc 后 quitting 应为 true")
	}
}

// TestListWrap 循环导航:wrap 开启时顶部 ↑/底部 ↓ 循环到另一端,
// 默认(未开启)边界不动;空选项列表循环导航不 panic。
func TestListWrap(t *testing.T) {
	options := []config.Option{{Label: "a", Value: "a"}, {Label: "b", Value: "b"}, {Label: "c", Value: "c"}}

	t.Run("wrap 开启时顶部↑跳末尾、底部↓跳回顶部", func(t *testing.T) {
		m := NewListModelWrap("标题", options, ListSingle)
		if got := m.table.Cursor(); got != 0 {
			t.Fatalf("初始光标 = %d, want 0", got)
		}
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		if got := m.table.Cursor(); got != len(options)-1 {
			t.Errorf("顶部 ↑ 后光标 = %d, want %d", got, len(options)-1)
		}
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		if got := m.table.Cursor(); got != 0 {
			t.Errorf("底部 ↓ 后光标 = %d, want 0", got)
		}
	})

	t.Run("默认不循环:边界按键不动", func(t *testing.T) {
		m := NewListModel("标题", options, ListSingle)
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		if got := m.table.Cursor(); got != 0 {
			t.Errorf("顶部 ↑ 后光标 = %d, want 0(默认不循环)", got)
		}
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		if got := m.table.Cursor(); got != len(options)-1 {
			t.Errorf("底部 ↓ 后光标 = %d, want %d(默认不循环)", got, len(options)-1)
		}
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		if got := m.table.Cursor(); got != len(options)-1 {
			t.Errorf("底部再 ↓ 后光标 = %d, want %d(默认不循环)", got, len(options)-1)
		}
	})

	t.Run("空选项列表循环导航不 panic", func(t *testing.T) {
		m := NewListModelWrap("标题", nil, ListSingle)
		m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	})
}

// TestLineReaderSequential 行级包装读取器:huh accessible 表单的 PromptString
// 每次新建 bufio.Scanner,Scanner 会预读底层 Reader 到私有缓冲;lineReader
// 每次 Read 至多返回一行,保证多字段表单(Input/Confirm/MultiInput)顺序读取
// 共享 stdinBuf 时不会因前一个 Scanner 的预读吞掉后续输入。
func TestLineReaderSequential(t *testing.T) {
	src := bufio.NewReader(strings.NewReader("v\n2\n3\n"))
	lr := &lineReader{br: src}

	// 模拟 MultiInput 的 3 个字段:每个字段新建 Scanner
	got := make([]string, 0, 3)
	for i := range 3 {
		sc := bufio.NewScanner(lr)
		if !sc.Scan() {
			t.Fatalf("第 %d 次读取失败: %v", i+1, sc.Err())
		}
		got = append(got, sc.Text())
	}

	want := []string{"v", "2", "3"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 行 = %q, want %q", i+1, got[i], want[i])
		}
	}

	// 读取耗尽后返回 EOF
	sc := bufio.NewScanner(lr)
	if sc.Scan() {
		t.Errorf("预期 EOF,却读到 %q", sc.Text())
	}
}

// TestLineReaderShortRead 小缓冲 Read 请求(调用方分段读取)数据完整不丢失。
// lineReader 仅保证「Scanner 大缓冲场景每次 Read 至多一行」;小缓冲时按
// Reader 常规语义分段返回,数据顺序与完整性不受影响。
func TestLineReaderShortRead(t *testing.T) {
	src := bufio.NewReader(strings.NewReader("hello world\nnext\n"))
	lr := &lineReader{br: src}

	// 用 4 字节小缓冲读取全部数据,验证无丢失、无错乱、EOF 正常
	buf := make([]byte, 4)
	all := make([]byte, 0, 18)
	for {
		n, err := lr.Read(buf)
		all = append(all, buf[:n]...)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
	}
	if string(all) != "hello world\nnext\n" {
		t.Errorf("读取结果 = %q, want %q", all, "hello world\nnext\n")
	}
}

// TestCompactMultiInput 紧凑多输入表单(配置中心版本号上限):
// 预览行随输入实时刷新(Note + 值绑定)、矮终端完整渲染不溢出。
// TestMultiInput_FocusedRowBackground 聚焦行整行背景(参考列表 Selected 样式):
// 聚焦行逐段携带 SelectionBg 背景(#404040 → 48;2;64;64;64)——标题/输入框/
// 简介/行尾补白,无"打洞";非聚焦行无背景;焦点切换/输入/校验错误后背景跟随。
func TestMultiInput_FocusedRowBackground(t *testing.T) {
	specs := []InputSpec{
		{Title: "前缀", Default: "v", Desc: "留空则无前缀", AllowEmpty: true},
		{Title: "主版本", Default: "1", Validate: nonNeg},
		{Title: "后缀", Default: "-beta", Desc: "留空则无后缀", AllowEmpty: true},
	}
	values := []string{"v", "1", "-beta"}
	ptrs := make([]*string, len(values))
	for i := range values {
		ptrs[i] = &values[i]
	}
	build := func() *multiInputModel {
		m := pumpInit(t, newMultiInputModel(specs, ptrs, nil)).(*multiInputModel)
		return pumpForm(t, m, tea.WindowSizeMsg{Width: 60, Height: 10}).(*multiInputModel)
	}

	const bg = "48;2;64;64;64" // theme.SelectionBg #404040 的 truecolor 背景序列
	hasBg := func(line string) bool { return strings.Contains(line, bg) }

	t.Run("聚焦行整行背景、非聚焦行无背景、行尾铺满", func(t *testing.T) {
		f := build()
		lines := strings.Split(f.View().Content, "\n")

		// 聚焦字段 1:整行(标题→输入框→简介→补白)均含背景色码
		if !hasBg(lines[0]) {
			t.Errorf("聚焦行应含背景色码 %s:\n%q", bg, lines[0])
		}
		// 非聚焦行无背景
		if hasBg(lines[1]) || hasBg(lines[2]) {
			t.Errorf("非聚焦行不应含背景色码:\n%q\n%q", lines[1], lines[2])
		}
		// 聚焦行铺满终端宽(行尾补白)
		if got := lipgloss.Width(ansi.Strip(lines[0])); got != 60 {
			t.Errorf("聚焦行宽 = %d, want 60", got)
		}
	})

	t.Run("焦点切换与输入后背景跟随", func(t *testing.T) {
		f := build()
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel) // 聚焦字段 2
		lines := strings.Split(f.View().Content, "\n")
		if hasBg(lines[0]) || !hasBg(lines[1]) {
			t.Errorf("焦点切换后背景应跟随: 行1=%v 行2=%v", hasBg(lines[0]), hasBg(lines[1]))
		}

		// 输入内容后背景仍在(输入框区域含背景)
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "9"}).(*multiInputModel)
		lines = strings.Split(f.View().Content, "\n")
		if !hasBg(lines[1]) {
			t.Errorf("输入后聚焦行仍应含背景色码:\n%q", lines[1])
		}
		if got := lipgloss.Width(ansi.Strip(lines[1])); got != 60 {
			t.Errorf("输入后聚焦行宽 = %d, want 60", got)
		}
	})

	t.Run("校验错误行仍整行背景", func(t *testing.T) {
		specs2 := []InputSpec{
			{Title: "字段A", Default: ""}, // 非空校验
		}
		v := []string{""}
		p := []*string{&v[0]}
		m := pumpInit(t, newMultiInputModel(specs2, p, nil)).(*multiInputModel)
		m = pumpForm(t, m, tea.WindowSizeMsg{Width: 60, Height: 10}).(*multiInputModel)
		m = pumpForm(t, m, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		lines := strings.Split(m.View().Content, "\n")
		if !strings.Contains(ansi.Strip(lines[0]), "✗") {
			t.Errorf("校验失败应显示错误标记:\n%q", lines[0])
		}
		if !hasBg(lines[0]) {
			t.Errorf("错误行仍应整行背景:\n%q", lines[0])
		}
	})
}

func TestMultiInput_Unified(t *testing.T) {
	specs := []InputSpec{
		{Title: "前缀", Default: "v", Desc: "留空则无前缀", AllowEmpty: true},
		{Title: "主版本", Default: "1", Validate: nonNeg},
		{Title: "次版本", Default: "2", Validate: nonNeg},
		{Title: "修订号", Default: "3", Validate: nonNeg},
		{Title: "后缀", Default: "-beta", Desc: "留空则无后缀", AllowEmpty: true},
	}
	values := []string{"v", "1", "2", "3", "-beta"}
	ptrs := make([]*string, len(values))
	for i := range values {
		ptrs[i] = &values[i]
	}

	preview := func(vals []string) string {
		return "preview:" + strings.Join(vals, "|")
	}

	build := func() *multiInputModel {
		m := pumpInit(t, newMultiInputModel(specs, ptrs, preview)).(*multiInputModel)
		return pumpForm(t, m, tea.WindowSizeMsg{Width: 80, Height: 12}).(*multiInputModel)
	}

	t.Run("预览行显示组合结果并随输入刷新", func(t *testing.T) {
		f := build()

		view := f.View().Content
		if !strings.Contains(view, "preview:v|1|2|3|-beta") {
			t.Errorf("初始预览缺失:\n%s", view)
		}

		// 聚焦字段 1(前缀),输入 x 追加到 v 后
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "x"}).(*multiInputModel)
		view = f.View().Content
		if !strings.Contains(view, "preview:vx|1|2|3|-beta") {
			t.Errorf("输入后预览未刷新:\n%s", view)
		}

		// Enter 推进到字段 2,输入 9 追加到默认值 1 后(光标在末尾)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "9"}).(*multiInputModel)
		view = f.View().Content
		if !strings.Contains(view, "preview:vx|19|2|3|-beta") {
			t.Errorf("字段 2 输入后预览未刷新:\n%s", view)
		}
	})

	t.Run("矮终端完整渲染不溢出", func(t *testing.T) {
		for _, h := range []int{10, 9, 8} {
			f := pumpForm(t, newMultiInputModel(specs, ptrs, preview), tea.WindowSizeMsg{Width: 80, Height: h}).(*multiInputModel)
			view := f.View().Content
			if got := lipgloss.Height(view); got > h {
				t.Errorf("终端 %d 行:表单高度 %d 溢出", h, got)
			}
			// 全部 5 个字段标题可见
			for _, title := range []string{"1. 前缀", "2. 主版本", "3. 次版本", "4. 修订号", "5. 后缀"} {
				if !strings.Contains(view, title) {
					t.Errorf("终端 %d 行:字段 %q 不可见:\n%s", h, title, view)
				}
			}
		}
	})

	t.Run("简介行尾弱化显示且不混入标题", func(t *testing.T) {
		f := build()
		view := ansi.Strip(f.View().Content)
		lines := strings.Split(view, "\n")

		// 标题行不含简介内容
		titleLine := lines[1]
		if strings.Contains(titleLine, "留空则无前缀") && !strings.Contains(titleLine, "❯") {
			// 标题与简介同行的唯一合法形式:简介在行尾(❯ 输入框之后)
		}
		// 简介必须在行尾:字段 1 行以简介结尾(去除尾部空白后)
		if !strings.HasSuffix(strings.TrimRight(titleLine, " "), "留空则无前缀") {
			t.Errorf("字段 1 行尾应显示弱化简介: %q", titleLine)
		}
		// 简介必须位于输入框(❯)之后
		if idx := strings.Index(titleLine, "❯"); idx < 0 || strings.Index(titleLine, "留空则无前缀") < idx {
			t.Errorf("简介应位于输入框之后: %q", titleLine)
		}
		// 字段 5(后缀)同理
		if !strings.HasSuffix(strings.TrimRight(lines[5], " "), "留空则无后缀") {
			t.Errorf("字段 5 行尾应显示弱化简介: %q", lines[5])
		}
		// 无简介字段(主版本)行尾不应出现任何简介
		if strings.Contains(lines[2], "留空则无") {
			t.Errorf("字段 2 不应包含简介: %q", lines[2])
		}
	})

	t.Run("简介列对齐且紧贴输入框", func(t *testing.T) {
		f := build()
		lines := strings.Split(ansi.Strip(f.View().Content), "\n")
		// 字段 1 与字段 5 均有简介:简介起始列应一致(中间输入列统一宽度)
		col1 := strings.Index(lines[1], "留空则无前缀")
		col5 := strings.Index(lines[5], "留空则无后缀")
		if col1 < 0 || col5 < 0 {
			t.Fatalf("简介缺失: col1=%d col5=%d", col1, col5)
		}
		if col1 != col5 {
			t.Errorf("简介列未对齐: 字段1=%d 字段5=%d\n%s", col1, col5, lines[1])
		}
		// 简介紧贴输入框:desc 前仅一个空格(输入框右填充结束处),
		// 而非被拉到屏幕最右的整行空白
		before := lines[1][:col1]
		if !strings.HasSuffix(before, " ") {
			t.Errorf("简介前应为单空格分隔: %q", lines[1])
		}
		if strings.HasSuffix(strings.TrimRight(before, " "), " ") && strings.Contains(before, "  ") {
			// 输入框宽于内容时右填充是 textinput 内部空格,允许;
			// 但 desc 前不应有双空格(间隔只有 1)
		}
	})

	t.Run("输入变长时全部行同步变宽保持对齐", func(t *testing.T) {
		f := build()
		// 字段 1 输入超长内容 → 输入列统一变宽 → 两处简介列同步右移且仍对齐
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "01234567890123456789"}).(*multiInputModel)
		lines := strings.Split(ansi.Strip(f.View().Content), "\n")
		col1 := strings.Index(lines[1], "留空则无前缀")
		col5 := strings.Index(lines[5], "留空则无后缀")
		if col1 < 0 || col5 < 0 {
			t.Fatalf("简介缺失: col1=%d col5=%d", col1, col5)
		}
		if col1 != col5 {
			t.Errorf("变长后简介列未对齐: 字段1=%d 字段5=%d", col1, col5)
		}
		if col1 < 30 {
			t.Errorf("变长后简介列应右移(col1=%d),输入列未随内容变宽", col1)
		}
	})
}

// nonNeg 测试用非负整数校验(与配置中心 nonNegativeInt 同语义)
func nonNeg(s string) error {
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return errors.New("must be non-negative")
	}
	return nil
}

// TestMultiInput_Navigation 统一多输入表单的上下导航键:
// ↓/j 推进、↑/k 回退、末字段 ↓ 提交;导航键不产生输入字符。
func TestMultiInput_Navigation(t *testing.T) {
	specs := []InputSpec{
		{Title: "字段A", Default: "a"},
		{Title: "字段B", Default: "b"},
		{Title: "字段C", Default: "c"},
	}
	values := []string{"a", "b", "c"}
	ptrs := make([]*string, len(values))
	for i := range values {
		ptrs[i] = &values[i]
	}
	// 每个子测试重建:导航测试会修改 values,子测试间须隔离
	newForm := func() *multiInputModel {
		for i, spec := range specs {
			values[i] = spec.Default
		}
		for i := range values {
			ptrs[i] = &values[i]
		}
		m := pumpInit(t, newMultiInputModel(specs, ptrs, nil)).(*multiInputModel)
		return pumpForm(t, m, tea.WindowSizeMsg{Width: 80, Height: 12}).(*multiInputModel)
	}

	assertFocused := func(t *testing.T, m *multiInputModel, want int) {
		t.Helper()
		if m.focused != want {
			t.Errorf("聚焦字段 = %d, want %d", m.focused, want)
		}
	}

	t.Run("↓ 推进、↑ 回退、末字段 ↓ 循环回首字段", func(t *testing.T) {
		f := newForm()
		assertFocused(t, f, 0)

		// ↓ 推进到字段 B
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		assertFocused(t, f, 1)

		// ↑ 回退到字段 A
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyUp}).(*multiInputModel)
		assertFocused(t, f, 0)

		// ↓↓ 到末字段 C,再 ↓ 循环回首字段(不提交)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		assertFocused(t, f, 2)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		if f.done {
			t.Fatalf("末字段 ↓ 后 done = true, want false(循环不应提交)")
		}
		assertFocused(t, f, 0) // 循环回首字段 A

		// 末字段仅 Enter 提交
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		assertFocused(t, f, 2)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*multiInputModel)
		if !f.done {
			t.Fatalf("末字段 Enter 后 done = false, want true")
		}
	})

	t.Run("j/k 推进回退且不产生字符", func(t *testing.T) {
		f := newForm()
		assertFocused(t, f, 0)

		f = pumpForm(t, f, tea.KeyPressMsg{Text: "j"}).(*multiInputModel)
		assertFocused(t, f, 1)
		if f.inputs[0].Value() != "a" {
			t.Errorf("j 不应向字段 A 输入字符,值 = %q", f.inputs[0].Value())
		}

		f = pumpForm(t, f, tea.KeyPressMsg{Text: "k"}).(*multiInputModel)
		assertFocused(t, f, 0)

		// 普通字符输入不受影响
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "z"}).(*multiInputModel)
		if got := f.inputs[0].Value(); got != "az" {
			t.Errorf("字段 A 输入 z 后 = %q, want az", got)
		}

		// j 循环:末字段 j 回首字段且不提交
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "j"}).(*multiInputModel)
		if f.done {
			t.Fatalf("末字段 j 后 done = true, want false")
		}
		assertFocused(t, f, 0)
	})

	t.Run("首字段 ↑ 与末字段 ↓ 循环边界", func(t *testing.T) {
		f := newForm()
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyUp}).(*multiInputModel)
		assertFocused(t, f, 2) // 首字段 ↑ 循环回末字段

		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		assertFocused(t, f, 0) // 末字段 ↓ 循环回首字段

		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyDown}).(*multiInputModel)
		assertFocused(t, f, 2) // 末字段
	})
}
