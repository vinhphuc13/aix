package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
	"github.com/vinhphuc13/aix/internal/snapshot"
	"github.com/spf13/cobra"
)

var (
	checkpointMessage  string
	checkpointSnapshot bool
)

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Save current session state as a checkpoint",
	RunE: func(cmd *cobra.Command, args []string) error {
		if checkpointMessage == "" {
			return fmt.Errorf("message required: aix checkpoint -m 'what you did'")
		}

		aixDir := mustFindAIXDir()

		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return err
		}

		cpID := session.NewID()
		snapID := ""

		if checkpointSnapshot && len(s.ActiveFiles) > 0 {
			projectRoot := filepath.Dir(aixDir)
			var paths []string
			for _, f := range s.ActiveFiles {
				paths = append(paths, f.Path)
			}
			snap, err := snapshot.Create(aixDir, s.ID, cpID, projectRoot, paths)
			if err != nil {
				fmt.Printf("warning: snapshot failed: %v\n", err)
			} else {
				snapID = snap.ID
			}
		}

		open, done := 0, 0
		for _, t := range s.Tasks {
			if t.Status == session.TaskDone {
				done++
			} else {
				open++
			}
		}

		cp := session.Checkpoint{
			ID:         cpID,
			Message:    checkpointMessage,
			SnapshotID: snapID,
			CreatedAt:  time.Now(),
			OpenTasks:  open,
			DoneTasks:  done,
		}
		s.Checkpoints = append(s.Checkpoints, cp)

		if err := session.Save(aixDir, s); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		_ = event.Append(aixDir, s.ID, event.EventCheckpoint, map[string]string{
			"message": cp.Message,
			"id":      cp.ID,
		})

		fmt.Printf("Checkpoint: %s\n", cp.Message)
		fmt.Printf("  Tasks: %d open, %d done\n", open, done)
		if snapID != "" {
			fmt.Printf("  Snapshot: %s (%d files)\n", snapID, len(s.ActiveFiles))
		}
		return nil
	},
}

func init() {
	checkpointCmd.Flags().StringVarP(&checkpointMessage, "message", "m", "", "checkpoint message (required)")
	checkpointCmd.Flags().BoolVar(&checkpointSnapshot, "snapshot", false, "copy active files to snapshot")
}
