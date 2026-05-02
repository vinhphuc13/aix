package cmd

import (
	"fmt"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var continueCmd = &cobra.Command{
	Use:   "continue [session-id]",
	Short: "Resume a session and print its context",
	Long:  `Resume a session and print the context block.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aixDir := mustFindAIXDir()

		if len(args) == 1 {
			id := args[0]
			if _, err := session.Load(aixDir, id); err != nil {
				return fmt.Errorf("session %s not found", id)
			}
			if err := session.SetCurrent(aixDir, id); err != nil {
				return err
			}
			fmt.Printf("Switched to session %s\n\n", id)
		}

		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		_ = event.Append(aixDir, s.ID, event.EventSessionResumed, map[string]string{
			"name": s.Name,
		})

		recentEvents, _ := event.ReadLast(aixDir, s.ID, 10)
		context := inject.RenderContext(s, recentEvents)

		fmt.Println(context)
		return nil
	},
}
