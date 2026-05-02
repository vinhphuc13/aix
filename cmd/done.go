package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <task-id-or-partial-title>",
	Short: "Mark a task as done",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		aixDir := mustFindAIXDir()

		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		idx := -1
		for i, t := range s.Tasks {
			if t.Status == session.TaskDone {
				continue
			}
			if strings.HasPrefix(t.ID, query) || strings.Contains(strings.ToLower(t.Title), strings.ToLower(query)) {
				idx = i
				break
			}
		}

		if idx == -1 {
			return fmt.Errorf("open task not found matching %q", query)
		}

		s.Tasks[idx].Status = session.TaskDone
		s.Tasks[idx].UpdatedAt = time.Now()

		if err := session.Save(aixDir, s); err != nil {
			return err
		}
		_ = event.Append(aixDir, s.ID, event.EventTaskDone, map[string]string{
			"id": s.Tasks[idx].ID, "title": s.Tasks[idx].Title,
		})

		recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
		_ = inject.WriteContextFile(aixDir, s, recentEvts)

		fmt.Printf("[x] Done: %s\n", s.Tasks[idx].Title)
		return nil
	},
}
