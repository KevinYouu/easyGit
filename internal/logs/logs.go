package logs

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/colors"
)

// Warning should be used to render warning text
func Warning(text string) {
	fmt.Println(colors.RenderColor("yellow", text))
}

func Info(text string) {
	fmt.Println(colors.RenderColor("cyan", text))
}

func Success(text string) {
	fmt.Println(colors.RenderColor("green", text))
}

func Error(text string) {
	fmt.Println(colors.RenderColor("red", text))
}
