package version

import (
	"fmt"

	"github.com/KevinYouu/easyGit/internal/colors"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/random"
)

var Version = "untracked"

func GetVersion() {
	funcProbs := []random.FuncProbability{
		{Function: func() { GetLogo() }, Probability: 0.98},
		{Function: func() { GetPenguin() }, Probability: 0.01},
		{Function: func() { GetDivineBeast() }, Probability: 0.01},
	}
	random.ExecuteRandomly(funcProbs)

	fmt.Println(i18n.T("version.version"), colors.RenderColor("blue", Version))
	fmt.Println(i18n.T("version.github"), colors.RenderColor("blue", "https://github.com/KevinYouu/easyGit"))
	fmt.Println(i18n.T("version.about"), colors.RenderColor("blue", "https://www.kevnu.com/about"))
}

func GetLogo() {
	fastGit3D := `
     ___           ___           ___           ___           ___                       ___
    /\  \         /\  \         /\  \         /\  \         /\  \          ___        /\  \
   /::\  \       /::\  \       /::\  \        \:\  \       /::\  \        /\  \       \:\  \
  /:/\:\  \     /:/\:\  \     /:/\ \  \        \:\  \     /:/\:\  \       \:\  \       \:\  \
 /::\~\:\  \   /::\~\:\  \   _\:\~\ \  \       /::\  \   /:/  \:\  \      /::\__\      /::\  \
/:/\:\ \:\__\ /:/\:\ \:\__\ /\ \:\ \ \__\     /:/\:\__\ /:/__/_\:\__\  __/:/\/__/     /:/\:\__\
\/__\:\ \/__/ \/__\:\/:/  / \:\ \:\ \/__/    /:/  \/__/ \:\  /\ \/__/ /\/:/  /       /:/  \/__/
     \:\__\        \::/  /   \:\ \:\__\     /:/  /       \:\ \:\__\   \::/__/       /:/  /
      \/__/        /:/  /     \:\/:/  /     \/__/         \:\/:/  /    \:\__\       \/__/
                  /:/  /       \::/  /                     \::/  /      \/__/
                  \/__/         \/__/                       \/__/
`
	fmt.Println(colors.RenderColor("cyan", fastGit3D))
}

func GetPenguin() {
	Penguin := `
        .--.
       |o_o |
       |:_/ |
      //   \\\\
     (|     | )
    /'\\_   _/'\\
    \\___)=(___/	
`
	fmt.Println(Penguin)
}

func GetDivineBeast() {
	divineBeast := `
	 ┏━┓   ┏━┓+ +
	┏┛ ┗┻━━━┛ ┻┓ + +
	┃         ┃  
	┃   ━     ┃ ++ + + +
	████━████ ┃+
	┃         ┃ +
	┃   ┻     ┃
	┃         ┃ + +
	┗━┓     ┏━┛
	  ┃     ┃           
	  ┃     ┃ + + + +
	  ┃     ┃
	  ┃     ┃ +  Divine beast bless
	  ┃     ┃    Code without bugs  
	  ┃     ┃  +         
	  ┃     ┗━━━┓ + +
	  ┃         ┣┓
	  ┃         ┏┛
	  ┗┓┓┏━━┳┓┏┛  + + + +
	   ┃┫┫  ┃┫┫
	   ┗┻┛  ┗┻┛ + + + +
`
	fmt.Println(divineBeast)
}
