package commands

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/config"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/logs"
	"github.com/spf13/cobra"
)

var SetLanguageCmd = &cobra.Command{
	Use:   "set-language",
	Short: "Set default language",
	Long:  "Set the default language for easyGit (en/zh)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// 显示当前语言设置
			showCurrentLanguage()
			return
		}

		lang := args[0]
		if err := setLanguage(lang); err != nil {
			logs.Error(i18n.T("language.set.error") + ": " + err.Error())
			return
		}

		logs.Success(i18n.T("language.set.success"))
	},
}

func showCurrentLanguage() {
	currentLang := i18n.GetCurrentLanguage()

	// 从数据库读取
	dbLang, _ := config.GetLanguage()

	fmt.Println(i18n.T("language.current.title"))
	fmt.Printf("  %s: %s\n", i18n.T("language.current.active"), string(currentLang))

	if dbLang != "" {
		fmt.Printf("  %s: %s\n", i18n.T("language.current.database"), dbLang)
	} else {
		fmt.Printf("  %s: %s\n", i18n.T("language.current.database"), i18n.T("language.current.not.set"))
	}

	fmt.Printf("\n%s:\n", i18n.T("language.available"))
	fmt.Println("  - en (English)")
	fmt.Println("  - zh (中文)")
}

func setLanguage(lang string) error {
	// 标准化语言代码
	var langCode string
	switch lang {
	case "zh", "chinese", "cn", "中文":
		langCode = "zh"
	case "en", "english", "英文":
		langCode = "en"
	default:
		return fmt.Errorf("%s", i18n.T("language.invalid"))
	}

	// 保存到数据库
	if err := config.SaveLanguage(langCode); err != nil {
		return err
	}

	// 立即应用
	if langCode == "zh" {
		i18n.SetLanguage(i18n.LangZH)
	} else {
		i18n.SetLanguage(i18n.LangEN)
	}

	return nil
}
