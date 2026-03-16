package theme

import (
	"strings"
	"testing"
)

func TestGetSpinnerFrames(t *testing.T) {
	frames := GetSpinnerFrames()

	if len(frames) == 0 {
		t.Error("GetSpinnerFrames should return non-empty slice")
	}

	for i, frame := range frames {
		if frame == "" {
			t.Errorf("frame %d should not be empty", i)
		}
	}
}

func TestGetPulseSpinnerFrames(t *testing.T) {
	frames := GetPulseSpinnerFrames()

	if len(frames) == 0 {
		t.Error("GetPulseSpinnerFrames should return non-empty slice")
	}

	for i, frame := range frames {
		if frame == "" {
			t.Errorf("frame %d should not be empty", i)
		}
	}
}

func TestGetDotsSpinnerFrames(t *testing.T) {
	frames := GetDotsSpinnerFrames()

	if len(frames) == 0 {
		t.Error("GetDotsSpinnerFrames should return non-empty slice")
	}

	for i, frame := range frames {
		if frame == "" {
			t.Errorf("frame %d should not be empty", i)
		}
	}
}

func TestGetArrowSpinnerFrames(t *testing.T) {
	frames := GetArrowSpinnerFrames()

	if len(frames) == 0 {
		t.Error("GetArrowSpinnerFrames should return non-empty slice")
	}

	for i, frame := range frames {
		if frame == "" {
			t.Errorf("frame %d should not be empty", i)
		}
	}
}

func TestColors(t *testing.T) {
	// 验证 Zinc Dark 颜色令牌已定义
	colors := []struct {
		name string
		val  any
	}{
		{"PrimaryColor", PrimaryColor},
		{"MutedForeground", MutedForeground},
		{"BorderColor", BorderColor},
		{"SelectionBg", SelectionBg},
		{"SelectionFg", SelectionFg},
		{"SuccessColor", SuccessColor},
		{"ErrorColor", ErrorColor},
		{"WarningColor", WarningColor},
		{"InfoColor", InfoColor},
	}

	for _, c := range colors {
		if c.val == nil {
			t.Errorf("%s should not be nil", c.name)
		}
	}
}

func TestGetSpinnerStyle(t *testing.T) {
	style := GetSpinnerStyle()

	// 验证调用不会 panic
	_ = style.Render("test")
}

func TestGetStatusIcon(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "success status",
			status:   "success",
			expected: "✓",
		},
		{
			name:     "error status",
			status:   "error",
			expected: "✗",
		},
		{
			name:     "warning status",
			status:   "warning",
			expected: "⚠",
		},
		{
			name:     "info status",
			status:   "info",
			expected: "ℹ",
		},
		{
			name:     "loading status",
			status:   "loading",
			expected: "⏳",
		},
		{
			name:     "pending status",
			status:   "pending",
			expected: "○",
		},
		{
			name:     "complete status",
			status:   "complete",
			expected: "✓",
		},
		{
			name:     "running status",
			status:   "running",
			expected: "▶",
		},
		{
			name:     "unknown status",
			status:   "unknown",
			expected: "•",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icon := GetStatusIcon(tt.status)
			if icon != tt.expected {
				t.Errorf("GetStatusIcon(%q) = %q, want %q", tt.status, icon, tt.expected)
			}
		})
	}
}

func TestGetHorizontalRule(t *testing.T) {
	tests := []struct {
		name  string
		width int
	}{
		{"small width", 10},
		{"medium width", 50},
		{"large width", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := GetHorizontalRule(tt.width)
			if rule == "" {
				t.Error("GetHorizontalRule should not return empty string")
			}
			// 验证包含分隔符字符
			if !strings.Contains(rule, "─") {
				t.Error("GetHorizontalRule should contain separator character")
			}
		})
	}
}

func TestGetCustomTheme(t *testing.T) {
	theme := GetCustomTheme()
	if theme == nil {
		t.Error("GetCustomTheme should not return nil")
	}
}

func TestGetCompactTheme(t *testing.T) {
	theme := GetCompactTheme()
	if theme == nil {
		t.Error("GetCompactTheme should not return nil")
	}
}

func TestGetProgressBarStyle(t *testing.T) {
	style := GetProgressBarStyle()

	// 验证调用不会 panic
	_ = style.Render("test")
}

func TestRenderSelection(t *testing.T) {
	result := RenderSelection("test item")
	if result == "" {
		t.Error("RenderSelection should not return empty string")
	}
	if !strings.Contains(result, "test item") {
		t.Error("RenderSelection should contain the input text")
	}
}

func TestRenderMuted(t *testing.T) {
	result := RenderMuted("muted text")
	if result == "" {
		t.Error("RenderMuted should not return empty string")
	}
}

func TestRenderBadge(t *testing.T) {
	variants := []string{"success", "error", "warning", "info", "unknown"}
	for _, v := range variants {
		result := RenderBadge("label", v)
		if result == "" {
			t.Errorf("RenderBadge(%q) should not return empty string", v)
		}
	}
}
