package i18n

import (
	"os"
	"testing"
)

func TestSetLanguage(t *testing.T) {
	// 保存原始语言
	original := GetCurrentLanguage()
	defer SetLanguage(original)

	tests := []struct {
		name string
		lang Language
	}{
		{"set to Chinese", LangZH},
		{"set to English", LangEN},
		{"set to Chinese again", LangZH},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetLanguage(tt.lang)
			got := GetCurrentLanguage()
			if got != tt.lang {
				t.Errorf("after SetLanguage(%v), GetCurrentLanguage() = %v, want %v", tt.lang, got, tt.lang)
			}
		})
	}
}

func TestGetLanguage(t *testing.T) {
	lang := GetCurrentLanguage()
	if lang != LangEN && lang != LangZH {
		t.Errorf("GetCurrentLanguage() = %v, want either %v or %v", lang, LangEN, LangZH)
	}
}

func TestT_NonExistentKey(t *testing.T) {
	key := "non.existent.key.that.should.not.exist"
	result := T(key)

	// Should return the key itself if not found
	if result != key {
		t.Errorf("T(%q) = %q, should return key when translation not found", key, result)
	}
}

func TestT_FallbackToEnglish(t *testing.T) {
	// 设置为中文
	SetLanguage(LangZH)
	defer SetLanguage(LangEN)

	// 使用一个只在英文中存在的键(如果有的话)
	// 这里测试fallback逻辑
	key := "version.short"
	result := T(key)

	if result == key {
		t.Errorf("T(%q) should have translation", key)
	}
}

func TestInitLanguage(t *testing.T) {
	// initLanguage 通过 sync.Once 只执行一次
	// 这里只验证它不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("initLanguage panicked: %v", r)
		}
	}()

	// 调用 T 会触发 initLanguage
	_ = T("test.key")
}

func TestIsChineseLocale(t *testing.T) {
	tests := []struct {
		locale   string
		expected bool
	}{
		{"zh_CN", true},
		{"zh_TW", true},
		{"zh_CN.UTF-8", true},
		{"zh_HK", true},
		{"zh", true},
		{"ZH_CN", true}, // 大写
		{"en_US", false},
		{"en_GB", false},
		{"fr_FR", false},
		{"ja_JP", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			result := isChineseLocale(tt.locale)
			if result != tt.expected {
				t.Errorf("isChineseLocale(%q) = %v, want %v", tt.locale, result, tt.expected)
			}
		})
	}
}

func TestDetectSystemLanguage(t *testing.T) {
	// 保存原始环境变量
	originalLang := os.Getenv("LANG")
	originalLCAll := os.Getenv("LC_ALL")
	defer func() {
		os.Setenv("LANG", originalLang)
		os.Setenv("LC_ALL", originalLCAll)
	}()

	tests := []struct {
		name     string
		lang     string
		lcAll    string
		expected Language
	}{
		{"Chinese LANG", "zh_CN.UTF-8", "", LangZH},
		{"English LANG", "en_US.UTF-8", "", LangEN},
		{"Chinese LC_ALL", "", "zh_CN.UTF-8", LangZH},
		{"LC_ALL overrides LANG", "en_US.UTF-8", "zh_CN.UTF-8", LangZH},
		{"No locale set", "", "", LangEN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("LANG", tt.lang)
			os.Setenv("LC_ALL", tt.lcAll)

			result := detectSystemLanguage()
			if result != tt.expected {
				t.Errorf("detectSystemLanguage() = %v, want %v (LANG=%q, LC_ALL=%q)",
					result, tt.expected, tt.lang, tt.lcAll)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	// 测试并发访问的安全性
	done := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		go func() {
			defer func() { done <- true }()

			// 并发读取
			_ = T("version.short")
			_ = GetCurrentLanguage()

			// 并发写入
			SetLanguage(LangZH)
			SetLanguage(LangEN)
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 100; i++ {
		<-done
	}
}
