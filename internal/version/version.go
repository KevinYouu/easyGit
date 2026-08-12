package version

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"github.com/KevinYouu/easyGit/internal/i18n"
	"github.com/KevinYouu/easyGit/internal/random"
	"github.com/KevinYouu/easyGit/internal/theme"
)

var Version = "untracked"

func GetVersion() {
	funcProbs := []random.FuncProbability{
		{Function: func() { GetLogo() }, Probability: 0.98},
		{Function: func() { GetPenguin() }, Probability: 0.01},
		{Function: func() { GetDivineBeast() }, Probability: 0.01},
	}
	random.ExecuteRandomly(funcProbs)

	primaryStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor).Bold(true)

	fmt.Printf("%s %s\n", i18n.T("version.version"), primaryStyle.Render(Version))
	fmt.Printf("%s %s\n", i18n.T("version.github"), primaryStyle.Render("https://github.com/KevinYouu/easyGit"))
	fmt.Printf("%s %s\n", i18n.T("version.about"), primaryStyle.Render("https://www.kevnu.com/about"))
}

func GetLogo() {
	logoStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor)
	easyGit3D := `
 ██████████   █████████    █████████  █████ █████   █████████  █████ ███████████
░░███░░░░░█  ███░░░░░███  ███░░░░░███░░███ ░░███   ███░░░░░███░░███ ░█░░░███░░░█
 ░███  █ ░  ░███    ░███ ░███    ░░░  ░░███ ███   ███     ░░░  ░███ ░   ░███  ░ 
 ░██████    ░███████████ ░░█████████   ░░█████   ░███          ░███     ░███    
 ░███░░█    ░███░░░░░███  ░░░░░░░░███   ░░███    ░███    █████ ░███     ░███    
 ░███ ░   █ ░███    ░███  ███    ░███    ░███    ░░███  ░░███  ░███     ░███    
 ██████████ █████   █████░░█████████     █████    ░░█████████  █████    █████   
░░░░░░░░░░ ░░░░░   ░░░░░  ░░░░░░░░░     ░░░░░      ░░░░░░░░░  ░░░░░    ░░░░░                                                              
`
	fmt.Println(logoStyle.Render(easyGit3D))
}

func GetPenguin() {
	penguinStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor)
	Penguin := `
        .--.
       |o_o |
       |:_/ |
      //   \\\\
     (|     | )
    /'\\_   _/'\\
    \\___)=(___/	
`
	fmt.Println(penguinStyle.Render(Penguin))
}

func GetDivineBeast() {
	beastStyle := lipgloss.NewStyle().Foreground(theme.PrimaryColor)
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
	fmt.Println(beastStyle.Render(divineBeast))
}
