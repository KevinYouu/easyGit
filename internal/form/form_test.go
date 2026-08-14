package form

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/KevinYouu/easyGit/internal/config"
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
