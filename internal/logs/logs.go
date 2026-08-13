package logs

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/theme"
)

// Warning should be used to render warning text
func Warning(text string) {
	fmt.Printf("%s %s\n", theme.WarningIconStyle.Render("⚠"), theme.WarningStyle.Render(text))
}

func Info(text string) {
	fmt.Printf("%s %s\n", theme.InfoIconStyle.Render("ℹ"), theme.InfoStyle.Render(text))
}

func Success(text string) {
	fmt.Printf("%s %s\n", theme.SuccessIconStyle.Render("✓"), theme.SuccessStyle.Render(text))
}

func Error(text string) {
	fmt.Printf("%s %s\n", theme.ErrorIconStyle.Render("✗"), theme.ErrorStyle.Render(text))
}
