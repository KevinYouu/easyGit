package i18n

import (
	"os"
	"strings"
	"sync"
)

// Language represents supported languages
type Language string

const (
	LangEN Language = "en"
	LangZH Language = "zh"
)

var (
	currentLang    Language
	runtimeLang    Language // 运行时临时设置的语言
	hasRuntimeLang bool     // 是否设置了运行时语言
	once           sync.Once
	mu             sync.RWMutex
)

// T translates a given key to the current language
func T(key string) string {
	once.Do(initLanguage)

	mu.RLock()
	defer mu.RUnlock()

	var translations map[string]string
	switch currentLang {
	case LangZH:
		translations = zhTranslations
	default:
		translations = enTranslations
	}

	if text, exists := translations[key]; exists {
		return text
	}

	// Fallback to English if key not found in current language
	if currentLang != LangEN {
		if text, exists := enTranslations[key]; exists {
			return text
		}
	}

	// Return key if translation not found
	return key
}

// SetLanguage manually sets the language (runtime temporary setting - priority 2)
func SetLanguage(lang Language) {
	mu.Lock()
	defer mu.Unlock()
	runtimeLang = lang
	hasRuntimeLang = true
	currentLang = lang
}

// GetCurrentLanguage returns the current language
func GetCurrentLanguage() Language {
	mu.RLock()
	defer mu.RUnlock()
	return currentLang
}

// initLanguage detects and sets the language based on priority:
// 1. Database setting (highest priority)
// 2. Runtime setting (if set via SetLanguage)
// 3. System locale (default)
func initLanguage() {
	lang := determineLanguage()
	mu.Lock()
	currentLang = lang
	mu.Unlock()
}

// determineLanguage determines the language based on priority
func determineLanguage() Language {
	// Priority 1: Database setting (需要导入 config 包)
	// 这里先返回检测到的语言,稍后会在外部调用 LoadLanguageFromDB

	// Priority 2: Runtime setting
	if hasRuntimeLang {
		return runtimeLang
	}

	// Priority 3: System locale
	return detectSystemLanguage()
}

// LoadLanguageFromDB loads language from database (priority 1)
// This should be called during application initialization
func LoadLanguageFromDB(dbLang string) {
	mu.Lock()
	defer mu.Unlock()

	// 只有在没有运行时设置时才使用数据库设置
	if !hasRuntimeLang && dbLang != "" {
		lang := Language(dbLang)
		if lang == LangEN || lang == LangZH {
			currentLang = lang
		}
	}
}

// detectSystemLanguage detects system language from environment variables
func detectSystemLanguage() Language {
	// Check environment variables
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if locale := os.Getenv(env); locale != "" {
			if isChineseLocale(locale) {
				return LangZH
			}
		}
	}

	// Default to English
	return LangEN
}

// isChineseLocale checks if the locale indicates Chinese language
func isChineseLocale(locale string) bool {
	locale = strings.ToLower(locale)
	return strings.HasPrefix(locale, "zh") ||
		strings.Contains(locale, "chinese") ||
		strings.Contains(locale, "china")
}
