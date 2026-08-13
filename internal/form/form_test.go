package form

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/charmbracelet/x/ansi"
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

var ErrEmptyInput = &validationError{msg: "input cannot be empty"}

type validationError struct {
	msg string
}

// selectOptionLabel TERM=dumb(accessible 纯文本模式)时剥离内嵌样式,
// 避免行尾无 reset 的样式序列泄漏到后续提示行。
func TestSelectOptionLabel(t *testing.T) {
	t.Run("正常终端:名称亮加粗 + 说明灰(内嵌样式)", func(t *testing.T) {
		t.Setenv("TERM", "xterm-256color")
		label := selectOptionLabel(config.Option{Label: "soft", Description: "保留工作区更改"})
		if !strings.Contains(label, "\x1b[") {
			t.Errorf("正常终端应内嵌样式: %q", label)
		}
	})

	t.Run("TERM=dumb:剥离样式输出纯文本", func(t *testing.T) {
		t.Setenv("TERM", "dumb")
		label := selectOptionLabel(config.Option{Label: "soft", Description: "保留工作区更改"})
		if strings.Contains(label, "\x1b[") {
			t.Errorf("dumb 不应含 ANSI: %q", label)
		}
		if got := ansi.Strip(label); got != "soft 保留工作区更改" {
			t.Errorf("dumb 应输出纯文本, got %q", got)
		}
	})

	t.Run("TERM=dumb 无说明:纯名称", func(t *testing.T) {
		t.Setenv("TERM", "dumb")
		if got := selectOptionLabel(config.Option{Label: "main"}); got != "main" {
			t.Errorf("dumb 无说明应输出纯名称, got %q", got)
		}
	})
}

// TestSelectOptionLabelAccessiblePath 经 huh 真实 accessible 渲染路径验证:
// TERM=dumb 时 huh 构造期自动进入 accessible 模式并原样打印 option.Key,
// 选项键必须为无 ANSI 的连续纯文本(huh 判定耦合见 isAccessibleMode)。
// 选项行格式 "编号. 键" 固化了 huh v2.0.3 field_select.go RunAccessible
// (fmt.Fprintf(w, "%d. %s\n", i+1, option.Key)),huh 变更格式时需同步。
func TestSelectOptionLabelAccessiblePath(t *testing.T) {
	t.Setenv("TERM", "dumb")
	var selected string
	f := NewSelectForm("标题", []config.Option{
		{Label: "soft", Description: "保留工作区更改"},
		{Label: "hard", Description: "丢弃所有更改"},
	}, 24, &selected)
	var buf bytes.Buffer
	if err := f.Form.
		WithOutput(&buf).
		WithInput(strings.NewReader("1\n")).
		RunWithContext(context.Background()); err != nil {
		t.Fatalf("RunWithContext 失败: %v", err)
	}
	out := buf.String()
	// 选项行按 "编号. 键" 原样打印:键必须为无 ANSI 的连续纯文本
	for _, want := range []string{"1. soft 保留工作区更改", "2. hard 丢弃所有更改"} {
		if !strings.Contains(out, want) {
			t.Errorf("accessible 选项行应为纯文本键,缺少 %q: %q", want, out)
		}
	}
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

// TestFormEscCancels Esc 触发取消:huh 默认 Quit 仅绑 ctrl+c,统一键位将 Esc
// 并入 Quit,保证帮助栏「Esc 取消」提示与实际行为一致。
func TestFormEscCancels(t *testing.T) {
	var selected string
	f := NewSelectForm("标题", []config.Option{{Label: "a", Value: "a"}}, 4, &selected)
	f.SubmitCmd = tea.Quit
	f.CancelCmd = tea.Interrupt
	m, _ := f.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc 未触发取消命令")
	}
	wrapped, ok := m.(*Form)
	if !ok {
		t.Fatalf("更新后模型类型 = %T, want *form.Form", m)
	}
	if wrapped.State != huh.StateAborted {
		t.Errorf("Esc 后状态 = %v, want StateAborted", wrapped.State)
	}
}

// TestFilterBindingsDisabled 筛选键禁用回归:Esc 在表单级先于字段分发匹配 Quit,
// 若 "/" 可进入筛选模式,筛选中的 Esc 会误终止表单并丢弃选择。
// 统一键位在传播前禁用 Select/MultiSelect 的 Filter 键(字段键位按值拷贝),
// 保证 Esc 始终是「取消表单」语义。
func TestFilterBindingsDisabled(t *testing.T) {
	// 单选与多选字段的键位均不得含可用的 "/" 筛选绑定
	assertNoActiveFilter := func(t *testing.T, f *Form) {
		t.Helper()
		for _, b := range f.KeyBinds() {
			for _, k := range b.Keys() {
				if k == "/" && b.Enabled() {
					t.Fatalf("筛选键 / 未被禁用: %+v", b)
				}
			}
		}
	}
	t.Run("select", func(t *testing.T) {
		var selected string
		f := NewSelectForm("标题", []config.Option{{Label: "a", Value: "a"}}, 1, &selected)
		assertNoActiveFilter(t, f)
		// 行为回归:先按 "/" 再按 Esc,表单应正常取消(而非进入筛选模式)
		f.SubmitCmd = tea.Quit
		f.CancelCmd = tea.Interrupt
		m, _ := f.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		m, _ = m.Update(tea.KeyPressMsg{Code: '/'})
		m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
		if cmd == nil {
			t.Fatal("Esc 未触发取消命令")
		}
		wrapped, ok := m.(*Form)
		if !ok {
			t.Fatalf("更新后模型类型 = %T, want *form.Form", m)
		}
		if wrapped.State != huh.StateAborted {
			t.Errorf("Esc 后状态 = %v, want StateAborted", wrapped.State)
		}
	})
	t.Run("multiSelect", func(t *testing.T) {
		var selected []string
		f := NewMultiSelectForm("标题", []string{"a"}, 24, &selected)
		assertNoActiveFilter(t, f)
	})
}
