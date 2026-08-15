package i18n

import (
	"github.com/spf13/cobra"
)

// AddLanguageFlag adds a global language flag to the root command
func AddLanguageFlag(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringP("language", "l", "", "Set language (en/zh)")
}
