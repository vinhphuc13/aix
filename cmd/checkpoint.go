package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
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

		cp := createCheckpoint(aixDir, s, checkpointMessage)

		snapID := ""
		if checkpointSnapshot && len(s.ActiveFiles) > 0 {
			projectRoot := filepath.Dir(aixDir)
			var paths []string
			for _, f := range s.ActiveFiles {
				paths = append(paths, f.Path)
			}
			snap, err := snapshot.Create(aixDir, s.ID, cp.ID, projectRoot, paths)
			if err != nil {
				fmt.Printf("warning: snapshot failed: %v\n", err)
			} else {
				snapID = snap.ID
				s.Checkpoints[len(s.Checkpoints)-1].SnapshotID = snapID
				_ = session.Save(aixDir, s)
			}
		}

		recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
		_ = inject.WriteContextFile(aixDir, s, recentEvts)

		fmt.Printf("Checkpoint: %s\n", cp.Message)
		fmt.Printf("  Tasks: %d open, %d done\n", cp.OpenTasks, cp.DoneTasks)
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
