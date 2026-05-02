package cmd

import (
	"fmt"
	"os"

	"github.com/vinhphuc13/aix/internal/hook"
	"github.com/spf13/cobra"
)

var hookGlobal bool

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Manage Claude Code hook integration",
}

var hookInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install aix hooks into .claude/settings.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		settingsPath := hook.FindSettingsPath(cwd, hookGlobal)

		hook.PrintWarningIfNotOnPath()

		if err := hook.Install(settingsPath); err != nil {
			return fmt.Errorf("failed to install hooks: %w", err)
		}

		fmt.Printf("Hooks installed: %s\n", settingsPath)
		fmt.Println("Claude Code will now:")
		fmt.Println("  • Inject session context into every prompt (UserPromptSubmit)")
		fmt.Println("  • Track file edits automatically (PostToolUse)")
		fmt.Println("  • Auto-checkpoint on session end (Stop)")
		return nil
	},
}

var hookUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove aix hooks from .claude/settings.json",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()
		settingsPath := hook.FindSettingsPath(cwd, hookGlobal)

		if err := hook.Uninstall(settingsPath); err != nil {
			return fmt.Errorf("failed to uninstall: %w", err)
		}
		fmt.Printf("Hooks removed from %s\n", settingsPath)
		return nil
	},
}

var hookPromptCmd = &cobra.Command{
	Use:    "prompt",
	Short:  "UserPromptSubmit handler (called by Claude Code)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hook.HandlePrompt(os.Stdin, os.Stdout)
	},
}

var hookPostToolUseCmd = &cobra.Command{
	Use:    "posttooluse",
	Short:  "PostToolUse handler (called by Claude Code)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hook.HandlePostToolUse(os.Stdin, os.Stdout)
	},
}

var hookStopCmd = &cobra.Command{
	Use:    "stop",
	Short:  "Stop handler (called by Claude Code)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return hook.HandleStop(os.Stdin, os.Stdout)
	},
}

func init() {
	hookInstallCmd.Flags().BoolVar(&hookGlobal, "global", false, "install in ~/.claude/settings.json")
	hookUninstallCmd.Flags().BoolVar(&hookGlobal, "global", false, "uninstall from ~/.claude/settings.json")

	hookCmd.AddCommand(hookInstallCmd, hookUninstallCmd, hookPromptCmd, hookPostToolUseCmd, hookStopCmd)
}
