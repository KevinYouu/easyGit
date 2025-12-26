package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
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
		} else if strings.HasPrefix(arg, "--language=") {
			lang := strings.ToLower(strings.TrimPrefix(arg, "--language="))
			switch lang {
			case "zh", "chinese", "cn":
				i18n.SetLanguage(i18n.LangZH)
			case "en", "english":
				i18n.SetLanguage(i18n.LangEN)
			}
			break
		}
	}

	// Update command descriptions after language is set
	updateRootCommandDescriptions()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
