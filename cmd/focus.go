package cmd

import (
	"fmt"
	"strings"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var focusCmd = &cobra.Command{
	Use:   "focus <text>",
	Short: "Set the current focus for this session",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		aixDir := mustFindAIXDir()

		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		old := s.CurrentFocus
		s.CurrentFocus = text

		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventFocusChanged, map[string]string{
			"from": old, "to": text,
		})

		recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
		_ = inject.WriteContextFile(aixDir, s, recentEvts)

		fmt.Printf("Focus: %s\n", text)
		return nil
	},
}
