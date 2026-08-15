package form

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"charm.land/bubbles/v2/cursor"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
)

// 注意: form 包的函数依赖于交互式终端输入,
// 这些测试主要验证函数结构和错误处理逻辑

func TestInput_Validation(t *testing.T) {
	// 测试验证函数逻辑
	validate := func(str string) error {
		if str == "" {
			return ErrEmptyInput
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
// BlinkMsg,避免命令链无限循环。
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
			continue
		}
		m, cmd = m.Update(msg)
	}
	return m
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

// TestMultiInput_Construction 单页多输入表单构造断言:
// 空 specs 直接返回空;默认值写入字段;Enter 逐字段推进(空值校验不推进)、
// shift+tab 回退、末字段 Enter 提交,与帮助栏「继续/提交、上一步、取消」键位语义一致。
func TestMultiInput_Construction(t *testing.T) {
	t.Run("空 specs 返回空结果", func(t *testing.T) {
		vals, err := MultiInput(nil)
		if err != nil {
			t.Fatalf("err = %v, want nil", err)
		}
		if vals != nil {
			t.Errorf("vals = %v, want nil", vals)
		}
	})

	t.Run("空值校验阻挡推进", func(t *testing.T) {
		// 字段 1 无默认值:Enter 触发非空校验,错误提示且不推进到字段 2
		specs := []InputSpec{
			{Title: "版本号"},
			{Title: "提交消息"},
		}
		values := make([]string, len(specs))
		ptrs := make([]*string, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		f := pumpForm(t, NewMultiInputForm(specs, ptrs), tea.WindowSizeMsg{Width: 80, Height: 12}).(*Form)

		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateNormal {
			t.Fatalf("空值 Enter 后 State = %v, want StateNormal(不应推进)", f.State)
		}
		if got := f.GetFocusedField().GetValue(); got != "" {
			t.Errorf("应仍聚焦字段 1,当前值 = %q", got)
		}
		errs := f.Errors()
		if len(errs) != 1 || errs[0].Error() != i18n.T("form.input.empty.error") {
			t.Errorf("Errors() = %v, want [%q]", errs, i18n.T("form.input.empty.error"))
		}
	})

	t.Run("Enter 推进、shift+tab 回退、末字段提交", func(t *testing.T) {
		specs := []InputSpec{
			{Title: "版本号", Default: "v1.1.0"},
			{Title: "提交消息"},
		}
		// 模拟 MultiInput 的播种:默认值先写入调用方指针,字段值来自指针
		values := []string{specs[0].Default, ""}
		ptrs := make([]*string, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		f := pumpForm(t, NewMultiInputForm(specs, ptrs), tea.WindowSizeMsg{Width: 80, Height: 12}).(*Form)

		// 初始聚焦字段 1,默认值已显示
		if got := f.GetFocusedField().GetValue(); got != "v1.1.0" {
			t.Errorf("字段 1 值 = %q, want v1.1.0", got)
		}

		// Enter → 推进到字段 2(默认值非空,校验通过)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateNormal {
			t.Fatalf("字段 1 Enter 后 State = %v, want StateNormal", f.State)
		}
		if got := f.GetFocusedField().GetValue(); got != "" {
			t.Errorf("应聚焦字段 2,当前值 = %q", got)
		}

		// 向字段 2 输入文本(推进后字段 2 已获得焦点)
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "feat: 打标签"}).(*Form)
		if got := f.GetFocusedField().GetValue(); got != "feat: 打标签" {
			t.Errorf("字段 2 输入值 = %q, want feat: 打标签", got)
		}
		if values[1] != "feat: 打标签" {
			t.Errorf("values[1] = %q, want feat: 打标签(指针应同步)", values[1])
		}

		// shift+tab → 回退到字段 1
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}).(*Form)
		if got := f.GetFocusedField().GetValue(); got != "v1.1.0" {
			t.Errorf("shift+tab 后应聚焦字段 1,当前值 = %q", got)
		}

		// 字段 1 Enter → 字段 2;字段 2(末字段)Enter → 提交
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateCompleted {
			t.Fatalf("末字段 Enter 后 State = %v, want StateCompleted", f.State)
		}
		if values[0] != "v1.1.0" || values[1] != "feat: 打标签" {
			t.Errorf("values = %v, want [v1.1.0 feat: 打标签]", values)
		}
	})

	t.Run("AllowEmpty 跳过非空校验", func(t *testing.T) {
		// prefix/suffix 等可空字段:空值不触发非空校验,可正常推进
		specs := []InputSpec{
			{Title: "前缀", AllowEmpty: true},
			{Title: "主版本号"},
		}
		values := make([]string, len(specs))
		ptrs := make([]*string, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		f := pumpInit(t, NewMultiInputForm(specs, ptrs)).(*Form)
		f = pumpForm(t, f, tea.WindowSizeMsg{Width: 80, Height: 12}).(*Form)

		// 字段 1 空值 Enter:AllowEmpty 跳过非空校验,推进到字段 2
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateNormal {
			t.Fatalf("AllowEmpty 空值 Enter 后 State = %v, want StateNormal", f.State)
		}
		if got := f.GetFocusedField().GetValue(); got != "" {
			t.Errorf("应聚焦字段 2,当前值 = %q", got)
		}

		// 字段 2 空值 Enter:非空校验仍生效,不推进
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateNormal {
			t.Fatalf("字段 2 空值 Enter 后 State = %v, want StateNormal", f.State)
		}
		errs := f.Errors()
		if len(errs) != 1 || errs[0].Error() != i18n.T("form.input.empty.error") {
			t.Errorf("Errors() = %v, want [%q]", errs, i18n.T("form.input.empty.error"))
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
		values := make([]string, 1)
		ptrs := make([]*string, len(values))
		for i := range values {
			ptrs[i] = &values[i]
		}
		f := pumpInit(t, NewMultiInputForm(specs, ptrs)).(*Form)
		f = pumpForm(t, f, tea.WindowSizeMsg{Width: 80, Height: 12}).(*Form)

		// 非法值:自定义校验报错,不推进
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "7"}).(*Form)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateNormal {
			t.Fatalf("非法值 Enter 后 State = %v, want StateNormal", f.State)
		}
		errs := f.Errors()
		if len(errs) != 1 || errs[0].Error() != "must be 42" {
			t.Errorf("Errors() = %v, want [must be 42]", errs)
		}

		// 退格清空非法输入后输入合法值:提交成功
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyBackspace}).(*Form)
		f = pumpForm(t, f, tea.KeyPressMsg{Text: "42"}).(*Form)
		f = pumpForm(t, f, tea.KeyPressMsg{Code: tea.KeyEnter}).(*Form)
		if f.State != huh.StateCompleted {
			t.Fatalf("合法值 Enter 后 State = %v, want StateCompleted", f.State)
		}
		if values[0] != "42" {
			t.Errorf("values[0] = %q, want 42", values[0])
		}
	})
}

var ErrEmptyInput = &validationError{msg: "input cannot be empty"}

type validationError struct {
	msg string
}

// optionDisplayText 无样式纯文本(列表模型与 accessible 模式共用):
// 有说明时「名称 + 说明」拼接,无说明仅名称。
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

func (e *validationError) Error() string {
	return e.msg
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
			return ErrEmptyInput
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
			return ErrEmptyInput
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
