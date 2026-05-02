package cmd

import (
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
)

func createCheckpoint(aixDir string, s *session.Session, message string) session.Checkpoint {
	open, done := 0, 0
	for _, t := range s.Tasks {
		if t.Status == session.TaskDone {
			done++
		} else {
			open++
		}
	}
	cp := session.Checkpoint{
		ID:        session.NewID(),
		Message:   message,
		CreatedAt: time.Now(),
		OpenTasks: open,
		DoneTasks: done,
	}
	s.Checkpoints = append(s.Checkpoints, cp)
	_ = session.Save(aixDir, s)
	_ = event.Append(aixDir, s.ID, event.EventCheckpoint, map[string]string{
		"message": cp.Message,
		"id":      cp.ID,
	})
	return cp
}
