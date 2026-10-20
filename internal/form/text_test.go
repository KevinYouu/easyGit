package form

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

// TestTextValidate 必填拒绝空内容与纯空白;可空放行(Enter 直接跳过)
func TestTextValidate(t *testing.T) {
	tests := []struct {
		name       string
		allowEmpty bool
		value      string
		wantErr    bool
	}{
		{name: "必填-正常", value: "feat: title\n\nbody", wantErr: false},
		{name: "必填-空", value: "", wantErr: true},
		{name: "必填-纯空白", value: " \n\t ", wantErr: true},
		{name: "可空-空", allowEmpty: true, value: "", wantErr: false},
		{name: "可空-纯空白", allowEmpty: true, value: "  \n", wantErr: false},
		{name: "可空-有内容", allowEmpty: true, value: "body line", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := textValidate(tt.allowEmpty)(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("textValidate(%v)(%q) err = %v, wantErr = %v", tt.allowEmpty, tt.value, err, tt.wantErr)
			}
		})
	}
}

// TestOptionalTextInputSubmitsEmpty 可空表单空内容可直接提交,必填表单被拦截
func TestOptionalTextInputSubmitsEmpty(t *testing.T) {
	tests := []struct {
		name       string
		allowEmpty bool
		wantState  huh.FormState
	}{
		{name: "可空表单空内容可提交", allowEmpty: true, wantState: huh.StateCompleted},
		{name: "必填表单空内容被拦截", allowEmpty: false, wantState: huh.StateNormal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := ""
			form := newTextInputForm("正文", &value, tt.allowEmpty)
			form = dispatchInitCmd(form)
			m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 20})
			form = m.(*Form)

			m, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
			form = dispatchCmd(m.(*Form), cmd)
			if got := form.State; got != tt.wantState {
				t.Errorf("提交后状态 = %v, want %v", got, tt.wantState)
			}
			if value != "" {
				t.Errorf("未输入时值应保持为空, got %q", value)
			}
		})
	}
}

// TestTextInputDefaultValuePreserved 默认值(含换行)原样带入并回写,不折叠
func TestTextInputDefaultValuePreserved(t *testing.T) {
	defaultValue := "feat: title\n\nbody line 1\nbody line 2"
	value := defaultValue
	form := NewTextInputForm("消息", &value)
	form = dispatchInitCmd(form)
	m, _ := form.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	form = m.(*Form)

	if !strings.Contains(form.View().Content, "body line 2") {
		t.Error("提交前视图应包含正文行")
	}

	// 一行未改直接提交:值必须与默认值完全一致
	m, cmd := form.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	form = dispatchCmd(m.(*Form), cmd)
	if value != defaultValue {
		t.Errorf("默认值被改写: got %q, want %q", value, defaultValue)
	}
}
