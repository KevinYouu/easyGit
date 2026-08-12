package spinner

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewSpinner(t *testing.T) {
	message := "测试加载消息"
	spinner := NewSpinner(message)

	if spinner.message != message {
		t.Errorf("expected message %q, got %q", message, spinner.message)
	}

	if spinner.done {
		t.Error("new spinner should not be done")
	}

	if spinner.success {
		t.Error("new spinner should not be marked as success")
	}
}

func TestSetMessage(t *testing.T) {
	spinner := NewSpinner("initial message")
	newMessage := "updated message"

	spinner.SetMessage(newMessage)

	if spinner.message != newMessage {
		t.Errorf("expected message %q, got %q", newMessage, spinner.message)
	}
}

func TestSetDone(t *testing.T) {
	tests := []struct {
		name      string
		success   bool
		resultMsg string
		err       error
	}{
		{
			name:      "success without error",
			success:   true,
			resultMsg: "操作成功",
			err:       nil,
		},
		{
			name:      "failure with error",
			success:   false,
			resultMsg: "操作失败",
			err:       fmt.Errorf("test error"),
		},
		{
			name:      "failure without error",
			success:   false,
			resultMsg: "操作失败",
			err:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner("test")
			spinner.SetDone(tt.success, tt.resultMsg, tt.err)

			if !spinner.done {
				t.Error("spinner should be marked as done")
			}

			if spinner.success != tt.success {
				t.Errorf("expected success=%v, got %v", tt.success, spinner.success)
			}

			if spinner.resultMsg != tt.resultMsg {
				t.Errorf("expected resultMsg %q, got %q", tt.resultMsg, spinner.resultMsg)
			}

			if spinner.err != tt.err {
				t.Errorf("expected err %v, got %v", tt.err, spinner.err)
			}
		})
	}
}

func TestView(t *testing.T) {
	t.Run("spinner not done", func(t *testing.T) {
		spinner := NewSpinner("loading...")
		view := spinner.View().Content

		if !strings.Contains(view, "loading...") {
			t.Errorf("view should contain loading message, got: %q", view)
		}
	})

	t.Run("spinner done with success", func(t *testing.T) {
		spinner := NewSpinner("loading...")
		spinner.SetDone(true, "完成", nil)
		view := spinner.View().Content

		if !strings.Contains(view, "完成") {
			t.Errorf("view should contain success message, got: %q", view)
		}

		if !strings.Contains(view, "✓") {
			t.Errorf("view should contain success symbol, got: %q", view)
		}
	})

	t.Run("spinner done with failure", func(t *testing.T) {
		spinner := NewSpinner("loading...")
		spinner.SetDone(false, "失败", fmt.Errorf("test error"))
		view := spinner.View().Content

		if !strings.Contains(view, "失败") {
			t.Errorf("view should contain failure message, got: %q", view)
		}

		if !strings.Contains(view, "✗") {
			t.Errorf("view should contain error symbol, got: %q", view)
		}

		if !strings.Contains(view, "test error") {
			t.Errorf("view should contain error details, got: %q", view)
		}
	})
}
