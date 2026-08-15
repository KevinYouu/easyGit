package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/theme"
)

func main() {
	// Initialize database and load language setting (Priority 1: Database)
	if err := config.Initialize(); err == nil {
		if dbLang, err := config.GetLanguage(); err == nil && dbLang != "" {
			i18n.LoadLanguageFromDB(dbLang)
		}
	}

	// Pre-parse language flag before executing the command (Priority 2: Runtime)
	for i, arg := range os.Args {
		if (arg == "-l" || arg == "--language") && i+1 < len(os.Args) {
			lang := strings.ToLower(os.Args[i+1])
			switch lang {
			case "zh", "chinese", "cn":
				i18n.SetLanguage(i18n.LangZH)
			case "en", "english":
				i18n.SetLanguage(i18n.LangEN)
			}
			break
		} else if after, ok := strings.CutPrefix(arg, "--language="); ok {
			lang := strings.ToLower(after)
			switch lang {
			case "zh", "chinese", "cn":
				i18n.SetLanguage(i18n.LangZH)
			case "en", "english":
				i18n.SetLanguage(i18n.LangEN)
			}
			break
		}
	}

	// Pre-parse theme flag before executing the command (Priority 1: Runtime)
	themeMode := ""
	for i, arg := range os.Args {
		if arg == "--theme" && i+1 < len(os.Args) {
			themeMode = strings.ToLower(os.Args[i+1])
			break
		} else if after, ok := strings.CutPrefix(arg, "--theme="); ok {
			themeMode = strings.ToLower(after)
			break
		}
	}

	// Initialize theme (Priority 1: --theme flag; Priority 2: Database; Priority 3: auto-detect)
	if themeMode != "" {
		theme.ApplyMode(theme.Mode(themeMode))
	} else if dbTheme, err := config.GetTheme(); err == nil && dbTheme != "" {
		theme.ApplyMode(theme.Mode(dbTheme))
	} else {
		theme.ApplyMode(theme.ModeAuto)
	}

	// Update command descriptions after language is set
	updateRootCommandDescriptions()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
