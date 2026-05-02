package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var startGoal string

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a new coding session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Use existing .aix/ if found, otherwise create in cwd
		aixDir, err := findAIXDir()
		if err != nil {
			aixDir = filepath.Join(cwd, ".aix")
		}

		if err := session.InitDir(aixDir); err != nil {
			return fmt.Errorf("failed to init .aix: %w", err)
		}

		goal := startGoal
		if goal == "" {
			goal = name
		}

		s, err := session.NewSession(aixDir, name, goal)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		_ = event.Append(aixDir, s.ID, event.EventSessionStarted, map[string]string{
			"name": s.Name,
			"goal": s.Goal,
		})

		createCheckpoint(aixDir, s, "session started")

		recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
		_ = inject.WriteContextFile(aixDir, s, recentEvts)

		fmt.Printf("Started session %s\n", s.ID)
		fmt.Printf("  Name: %s\n", s.Name)
		fmt.Printf("  Goal: %s\n", s.Goal)
		fmt.Printf("  State: %s\n", aixDir)
		fmt.Println()
		fmt.Println("Next:")
		fmt.Println("  aix hook install          # auto-inject context into Claude")
		fmt.Println("  aix add task 'first task'")
		return nil
	},
}

func init() {
	startCmd.Flags().StringVar(&startGoal, "goal", "", "session goal (defaults to name)")
}
