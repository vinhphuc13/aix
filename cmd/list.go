package cmd

import (
	"fmt"
	"sort"

	"github.com/vinhphuc13/aix/internal/session"
	"github.com/spf13/cobra"
)

var listAll bool

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		aixDir, err := findAIXDir()
		if err != nil {
			fmt.Println("No aix project found.")
			return nil
		}

		sessions, err := session.List(aixDir)
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions yet. Run 'aix start <name>'.")
			return nil
		}

		sort.Slice(sessions, func(i, j int) bool {
			return sessions[i].CreatedAt.After(sessions[j].CreatedAt)
		})

		currentID, _ := session.GetCurrent(aixDir)

		printed := 0
		for _, s := range sessions {
			if !listAll && s.Status == session.StatusCompleted {
				continue
			}
			marker := "  "
			if s.ID == currentID {
				marker = "* "
			}
			open := len(filterTasks(s.Tasks, session.TaskOpen, session.TaskInProgress, session.TaskBlocked))
			fmt.Printf("%s[%s] %-30s %d open tasks  %s\n",
				marker, s.ID, s.Name, open,
				s.CreatedAt.Format("2006-01-02"),
			)
			printed++
		}

		if printed == 0 {
			fmt.Println("No active sessions. Use --all to show completed ones.")
		}
		return nil
	},
}

func init() {
	listCmd.Flags().BoolVar(&listAll, "all", false, "include completed sessions")
}
