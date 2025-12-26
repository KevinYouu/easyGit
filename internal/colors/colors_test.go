package colors

import (
	"strings"
	"testing"
)

func TestRenderColor(t *testing.T) {
	tests := []struct {
		color string
		text  string
	}{
		{"red", "error message"},
		{"green", "success message"},
		{"yellow", "warning message"},
		{"blue", "info message"},
		{"cyan", "debug message"},
		{"magenta", "special message"},
		{"white", "normal message"},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := RenderColor(tt.color, tt.text)

			// 结果应该包含原始文本
			if !strings.Contains(result, tt.text) {
				t.Errorf("RenderColor(%q, %q) should contain text, got: %s", tt.color, tt.text, result)
			}

			// 结果长度应该大于原始文本(包含颜色代码)
			if len(result) <= len(tt.text) {
				t.Errorf("RenderColor(%q, %q) should add color codes, got length %d vs %d",
					tt.color, tt.text, len(result), len(tt.text))
			}
		})
	}
}

func TestRenderColorEmpty(t *testing.T) {
	result := RenderColor("red", "")
	if result == "" {
		t.Error("RenderColor should handle empty text")
	}
}
