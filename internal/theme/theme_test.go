package theme

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/charmbracelet/x/ansi"
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
	// 验证 Neutral Dark 颜色令牌精确色值(lipgloss.Color 返回 color.Color 接口,
	// 仅断言非 nil 无效——接口零值也是 nil 字符串兜底,故按 RGBA 逐通道断言)
	tests := []struct {
		name string
		got  color.Color
		want color.RGBA
	}{
		{"PrimaryColor", PrimaryColor, color.RGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff}},       // Neutral 50
		{"MutedForeground", MutedForeground, color.RGBA{R: 0xa3, G: 0xa3, B: 0xa3, A: 0xff}}, // Neutral 400
		{"BorderColor", BorderColor, color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff}},         // Neutral 700
		{"SelectionBg", SelectionBg, color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff}},         // Neutral 700
		{"SelectionFg", SelectionFg, color.RGBA{R: 0xfa, G: 0xfa, B: 0xfa, A: 0xff}},         // Neutral 50
		{"SuccessColor", SuccessColor, color.RGBA{R: 0x10, G: 0xb9, B: 0x81, A: 0xff}},       // Emerald 500
		{"ErrorColor", ErrorColor, color.RGBA{R: 0xef, G: 0x44, B: 0x44, A: 0xff}},           // Red 500
		{"WarningColor", WarningColor, color.RGBA{R: 0xf5, G: 0x9e, B: 0x0b, A: 0xff}},       // Amber 500
		{"InfoColor", InfoColor, color.RGBA{R: 0x3b, G: 0x82, B: 0xf6, A: 0xff}},             // Blue 500
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got == nil {
				t.Fatalf("%s 未定义", tc.name)
			}
			r, g, b, a := tc.got.RGBA()
			want := tc.want
			if r != uint32(want.R)*0x101 || g != uint32(want.G)*0x101 ||
				b != uint32(want.B)*0x101 || a != uint32(want.A)*0x101 {
				t.Errorf("%s = #%02x%02x%02x, want #%02x%02x%02x",
					tc.name, r>>8, g>>8, b>>8, want.R, want.G, want.B)
			}
		})
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

// TestSelectorIndicators 选中指示符 ❯:两个主题的 Focused 单选/多选指示符
// 均为 "❯ "+布局间距(显示宽 2,与 huh 默认 "> " 同宽,零布局位移)
func TestSelectorIndicators(t *testing.T) {
	for name, th := range map[string]huh.Theme{"compact": GetCompactTheme(), "custom": GetCustomTheme()} {
		t.Run(name, func(t *testing.T) {
			styles := th.Theme(true)
			selectSel := styles.Focused.SelectSelector.Render()
			multiSel := styles.Focused.MultiSelectSelector.Render()
			for field, got := range map[string]string{"SelectSelector": selectSel, "MultiSelectSelector": multiSel} {
				if !strings.Contains(ansi.Strip(got), "❯ ") {
					t.Errorf("%s = %q, want 含 %q", field, got, "❯ ")
				}
				if width := lipgloss.Width(got); width != 2 {
					t.Errorf("%s 显示宽 = %d, want 2(与 huh 默认 > 同宽)", field, width)
				}
			}
		})
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

// ─── 主题模式 ────────────────────────────────────────────────────────────────

func TestValidMode(t *testing.T) {
	for _, mode := range []Mode{ModeAuto, ModeDark, ModeLight} {
		if !ValidMode(mode) {
			t.Errorf("ValidMode(%q) = false, want true", mode)
		}
	}
	if ValidMode(Mode("invalid")) {
		t.Error("ValidMode(invalid) = true, want false")
	}
}

// TestResolveMode dark/light 直通;auto/非法值解析为 dark/light
// (终端背景检测结果不可控,仅断言不返回 auto/非法值)
func TestResolveMode(t *testing.T) {
	if got := ResolveMode(ModeDark); got != ModeDark {
		t.Errorf("ResolveMode(dark) = %q, want dark", got)
	}
	if got := ResolveMode(ModeLight); got != ModeLight {
		t.Errorf("ResolveMode(light) = %q, want light", got)
	}
	for _, mode := range []Mode{ModeAuto, Mode("invalid")} {
		got := ResolveMode(mode)
		if got != ModeDark && got != ModeLight {
			t.Errorf("ResolveMode(%q) = %q, want dark or light", mode, got)
		}
	}
}

// TestApplyModeLight 浅色模式切换:令牌切到浅色板,样式重建后渲染新色
func TestApplyModeLight(t *testing.T) {
	// 无论测试顺序如何,结束后恢复深色(其他测试依赖默认深色)
	t.Cleanup(func() { ApplyMode(ModeDark) })

	if got := ApplyMode(ModeLight); got != ModeLight {
		t.Fatalf("ApplyMode(light) = %q, want light", got)
	}
	if CurrentMode() != ModeLight || IsDark() {
		t.Error("CurrentMode/IsDark 未切换到浅色")
	}

	// 浅色板精确色值断言(RGBA 逐通道)
	tests := []struct {
		name string
		got  color.Color
		want color.RGBA
	}{
		{"PrimaryColor", PrimaryColor, color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}},       // Neutral 900
		{"MutedForeground", MutedForeground, color.RGBA{R: 0x73, G: 0x73, B: 0x73, A: 0xff}}, // Neutral 500
		{"BorderColor", BorderColor, color.RGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff}},         // Neutral 200
		{"SelectionBg", SelectionBg, color.RGBA{R: 0xe4, G: 0xe4, B: 0xe7, A: 0xff}},         // Neutral 200
		{"SelectionFg", SelectionFg, color.RGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}},         // Neutral 900
		{"SelectionMuted", SelectionMuted, color.RGBA{R: 0x52, G: 0x52, B: 0x52, A: 0xff}},   // Neutral 600
		{"DiffAddedBg", DiffAddedBg, color.RGBA{R: 0xdc, G: 0xfc, B: 0xe7, A: 0xff}},         // Green 100
		{"DiffRemovedBg", DiffRemovedBg, color.RGBA{R: 0xfe, G: 0xe2, B: 0xe2, A: 0xff}},     // Red 100
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b, a := tc.got.RGBA()
			want := tc.want
			if r != uint32(want.R)*0x101 || g != uint32(want.G)*0x101 ||
				b != uint32(want.B)*0x101 || a != uint32(want.A)*0x101 {
				t.Errorf("%s = #%02x%02x%02x, want #%02x%02x%02x",
					tc.name, r>>8, g>>8, b>>8, want.R, want.G, want.B)
			}
		})
	}

	// 样式已重建:渲染输出含浅色主色 SGR(38;2;24;24;27)
	rendered := SpinnerStyle.Render("x")
	if !strings.Contains(rendered, "38;2;24;24;27") {
		t.Errorf("SpinnerStyle 未重建为浅色主色: %q", rendered)
	}

	// 语义色不随模式变化
	r, g, b, _ := ErrorColor.RGBA()
	if r != 0xef*0x101 || g != 0x44*0x101 || b != 0x44*0x101 {
		t.Errorf("ErrorColor 不应随模式变化: #%02x%02x%02x", r>>8, g>>8, b>>8)
	}
}

// TestApplyModeDark 切回深色后令牌恢复
func TestApplyModeDark(t *testing.T) {
	ApplyMode(ModeLight)
	t.Cleanup(func() { ApplyMode(ModeDark) })

	if got := ApplyMode(ModeDark); got != ModeDark {
		t.Fatalf("ApplyMode(dark) = %q, want dark", got)
	}
	if CurrentMode() != ModeDark || !IsDark() {
		t.Error("CurrentMode/IsDark 未切回深色")
	}
	r, g, b, _ := PrimaryColor.RGBA()
	if r != 0xfa*0x101 || g != 0xfa*0x101 || b != 0xfa*0x101 {
		t.Errorf("PrimaryColor 未恢复深色: #%02x%02x%02x", r>>8, g>>8, b>>8)
	}
}

// TestModesAlignWithConfig 主题模式字符串与 config 常量防漂移
func TestModesAlignWithConfig(t *testing.T) {
	if string(ModeAuto) != config.ThemeAuto ||
		string(ModeDark) != config.ThemeDark ||
		string(ModeLight) != config.ThemeLight {
		t.Error("theme.Mode 与 config.Theme* 常量不一致")
	}
}
