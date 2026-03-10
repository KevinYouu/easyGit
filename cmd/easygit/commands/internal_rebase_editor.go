package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func InternalRebaseEditorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "_internal_rebase_editor <file>",
		Short:  "Internal command used as GIT_SEQUENCE_EDITOR",
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				fmt.Println("No file provided")
				os.Exit(1)
			}

			filePath := args[0]
			content, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Println("Error reading file:", err)
				os.Exit(1)
			}

			configStr := os.Getenv("EASYGIT_REBASE_CONFIG")
			if configStr == "" {
				// No config, just pass through (shouldn't happen in our new flow)
				os.Exit(0)
			}

			var config map[string]interface{}
			err = json.Unmarshal([]byte(configStr), &config)
			if err != nil {
				fmt.Println("Error parsing config:", err)
				os.Exit(1)
			}

			mode := config["mode"].(string)
			targetsIf := config["targets"].([]interface{})
			targets := make([]string, len(targetsIf))
			for i, v := range targetsIf {
				targets[i] = v.(string)
			}
			newMessage := config["message"].(string)

			lines := strings.Split(string(content), "\n")
			var newLines []string

			// Helper to check if a hash is in the target list
			isTarget := func(hash string) bool {
				for _, t := range targets {
					if strings.HasPrefix(t, hash) {
						return true
					}
				}
				return false
			}

			if mode == "drop" {
				for _, line := range lines {
					if strings.HasPrefix(line, "pick ") {
						parts := strings.SplitN(line, " ", 3)
						if len(parts) >= 2 {
							hash := parts[1]
							if isTarget(hash) {
								// Drop it by not adding it to newLines
								continue
							}
						}
					}
					newLines = append(newLines, line)
				}
			} else if mode == "squash" {
				firstTargetSeen := false
				for _, line := range lines {
					if strings.HasPrefix(line, "pick ") {
						parts := strings.SplitN(line, " ", 3)
						if len(parts) >= 2 {
							hash := parts[1]
							if isTarget(hash) {
								if !firstTargetSeen {
									// First target: change to reword to apply the new message
									firstTargetSeen = true
									if newMessage != "" {
										newLines = append(newLines, "reword "+hash+" "+parts[2])
									} else {
										newLines = append(newLines, line) // keep pick if no new message
									}
								} else {
									// Subsequent targets: change to fixup to melt into previous without keeping its message
									newLines = append(newLines, "fixup "+hash+" "+parts[2])
								}
								continue
							}
						}
					}
					newLines = append(newLines, line)
				}
			}

			newContent := strings.Join(newLines, "\n")
			err = os.WriteFile(filePath, []byte(newContent), 0644)
			if err != nil {
				fmt.Println("Failed to save rebase plan:", err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
