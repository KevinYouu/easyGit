package i18n

import (
	"sync"
	"testing"
)

func TestLoadLanguageFromDB(t *testing.T) {
	// 重置状态
	mu.Lock()
	hasRuntimeLang = false
	currentLang = LangEN
	mu.Unlock()

	tests := []struct {
		name       string
		dbLang     string
		wantLang   Language
		setRuntime bool
		runtimeVal Language
	}{
		{
			name:       "load chinese from db",
			dbLang:     "zh",
			wantLang:   LangZH,
			setRuntime: false,
		},
		{
			name:       "load english from db",
			dbLang:     "en",
			wantLang:   LangEN,
			setRuntime: false,
		},
		{
			name:       "empty db lang",
			dbLang:     "",
			wantLang:   LangEN, // 保持默认
			setRuntime: false,
		},
		{
			name:       "runtime overrides db",
			dbLang:     "en",
			wantLang:   LangZH, // runtime 设置优先
			setRuntime: true,
			runtimeVal: LangZH,
		},
		{
			name:       "invalid lang code",
			dbLang:     "invalid",
			wantLang:   LangEN, // 应该保持默认
			setRuntime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 重置状态
			mu.Lock()
			hasRuntimeLang = false
			currentLang = LangEN
			mu.Unlock()

			// 如果需要设置 runtime
			if tt.setRuntime {
				SetLanguage(tt.runtimeVal)
			}

			// 加载数据库语言
			LoadLanguageFromDB(tt.dbLang)

			// 验证结果
			got := GetCurrentLanguage()
			if got != tt.wantLang {
				t.Errorf("GetCurrentLanguage() = %v, want %v", got, tt.wantLang)
			}
		})
	}
}

func TestSetLanguage_RuntimePriority(t *testing.T) {
	// 重置状态
	mu.Lock()
	hasRuntimeLang = false
	currentLang = LangEN
	mu.Unlock()

	// 先加载数据库设置
	LoadLanguageFromDB("en")
	if GetCurrentLanguage() != LangEN {
		t.Errorf("After LoadLanguageFromDB(en), got %v, want %v", GetCurrentLanguage(), LangEN)
	}

	// 运行时设置应该覆盖
	SetLanguage(LangZH)
	if GetCurrentLanguage() != LangZH {
		t.Errorf("After SetLanguage(zh), got %v, want %v", GetCurrentLanguage(), LangZH)
	}

	// 再次加载数据库不应该改变(runtime 优先级更高)
	LoadLanguageFromDB("en")
	if GetCurrentLanguage() != LangZH {
		t.Errorf("Runtime setting should not be overridden, got %v, want %v", GetCurrentLanguage(), LangZH)
	}
}

func TestLanguagePriority(t *testing.T) {
	// 测试优先级: Runtime > Database > System

	// 重置状态
	mu.Lock()
	hasRuntimeLang = false
	currentLang = LangEN
	once = sync.Once{} // 重置 once
	mu.Unlock()

	// 场景1: 只有系统语言(最低优先级)
	// 系统语言通过环境变量检测,这里跳过

	// 场景2: 数据库设置存在
	LoadLanguageFromDB("zh")
	if GetCurrentLanguage() != LangZH {
		t.Errorf("Database setting: got %v, want %v", GetCurrentLanguage(), LangZH)
	}

	// 场景3: Runtime 设置覆盖数据库
	SetLanguage(LangEN)
	if GetCurrentLanguage() != LangEN {
		t.Errorf("Runtime override: got %v, want %v", GetCurrentLanguage(), LangEN)
	}
}

func TestGetCurrentLanguage_Concurrent(t *testing.T) {
	// 并发测试
	done := make(chan bool)

	for range 10 {
		go func() {
			_ = GetCurrentLanguage()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}
}

func TestSetLanguage_Concurrent(t *testing.T) {
	// 并发设置测试
	done := make(chan bool)

	for range 5 {
		go func() {
			SetLanguage(LangZH)
			done <- true
		}()

		go func() {
			SetLanguage(LangEN)
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// 验证最终状态是有效的
	lang := GetCurrentLanguage()
	if lang != LangZH && lang != LangEN {
		t.Errorf("Invalid language after concurrent updates: %v", lang)
	}
}
