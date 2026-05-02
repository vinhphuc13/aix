package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "aix",
	Short:        "Local-first AI coding session runtime",
	Long:         "aix persists engineering working state across AI coding sessions.\nThink: \"Git for AI sessions\" — goals, decisions, files, tasks. Not chat history.",
	SilenceUsage: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(checkpointCmd)
	rootCmd.AddCommand(continueCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(doneCmd)
	rootCmd.AddCommand(focusCmd)
	rootCmd.AddCommand(hookCmd)
	rootCmd.AddCommand(mcpCmd)
}

// findAIXDir walks up from cwd to find .aix/
func findAIXDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, ".aix")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no aix project found (run 'aix start <name>')")
		}
		dir = parent
	}
}

// mustFindAIXDir exits with a clear error if .aix/ is not found.
func mustFindAIXDir() string {
	dir, err := findAIXDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
	return dir
}
