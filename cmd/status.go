package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var statusJSON bool

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current session status",
	RunE: func(cmd *cobra.Command, args []string) error {
		aixDir := mustFindAIXDir()

		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		if statusJSON {
			data, _ := json.MarshalIndent(s, "", "  ")
			fmt.Println(string(data))
			return nil
		}

		recentEvents, _ := event.ReadLast(aixDir, s.ID, 10)

		fmt.Printf("Session: %s (%s)\n", s.Name, s.ID)
		fmt.Printf("Goal:    %s\n", s.Goal)
		fmt.Printf("Status:  %s\n", string(s.Status))
		if s.CurrentFocus != "" {
			fmt.Printf("Focus:   %s\n", s.CurrentFocus)
		}
		if len(s.Checkpoints) > 0 {
			last := s.Checkpoints[len(s.Checkpoints)-1]
			fmt.Printf("Last Checkpoint: %s (%s)\n",
				last.Message, last.CreatedAt.Format("2006-01-02 15:04"))
		}

		open := filterTasks(s.Tasks, session.TaskOpen, session.TaskInProgress, session.TaskBlocked)
		done := filterTasks(s.Tasks, session.TaskDone)
		fmt.Printf("\nTasks: %d open, %d done\n", len(open), len(done))
		for _, t := range open {
			fmt.Printf("  %s %s\n", taskIcon(t.Status), t.Title)
			if t.Note != "" {
				fmt.Printf("       %s\n", t.Note)
			}
		}

		if len(s.Decisions) > 0 {
			fmt.Printf("\nDecisions (%d):\n", len(s.Decisions))
			decisions := s.Decisions
			if len(decisions) > 5 {
				fmt.Printf("  (showing last 5 of %d)\n", len(decisions))
				decisions = decisions[len(decisions)-5:]
			}
			for _, d := range decisions {
				if d.Rationale != "" {
					fmt.Printf("  • %s [%s]\n", d.Summary, d.Rationale)
				} else {
					fmt.Printf("  • %s\n", d.Summary)
				}
			}
		}

		if len(s.ActiveFiles) > 0 {
			fmt.Printf("\nActive Files (%d):\n", len(s.ActiveFiles))
			for _, f := range s.ActiveFiles {
				fmt.Printf("  [%s] %s\n", f.Role, f.Path)
			}
		}

		if len(recentEvents) > 0 {
			fmt.Printf("\nRecent Activity:\n")
			for _, e := range recentEvents {
				summary := eventDataSummary(e.Data)
				fmt.Printf("  %s [%s] %s\n",
					e.Timestamp.Format("15:04"),
					string(e.Type),
					summary,
				)
			}
		}

		return nil
	},
}

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output raw JSON")
}
