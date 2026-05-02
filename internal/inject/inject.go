package inject

import (
	"fmt"
	"strings"
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/session"
)

func RenderContext(s *session.Session, recentEvents []event.Event) string {
	var b strings.Builder

	b.WriteString("=== AIX SESSION CONTEXT ===\n")
	fmt.Fprintf(&b, "Session: %s (%s)\n", s.Name, s.ID)
	fmt.Fprintf(&b, "Goal: %s\n", s.Goal)
	if s.CurrentFocus != "" {
		fmt.Fprintf(&b, "Current Focus: %s\n", s.CurrentFocus)
	}
	if len(s.Checkpoints) > 0 {
		last := s.Checkpoints[len(s.Checkpoints)-1]
		fmt.Fprintf(&b, "Last Checkpoint: %s (%s)\n", last.Message, fmtTime(last.CreatedAt))
	}

	openTasks := filterTasks(s.Tasks, session.TaskOpen, session.TaskInProgress, session.TaskBlocked)
	fmt.Fprintf(&b, "\nOPEN TASKS (%d):\n", len(openTasks))
	if len(openTasks) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, t := range openTasks {
			if t.Note != "" {
				fmt.Fprintf(&b, "  %s %s — %s\n", taskIcon(t.Status), t.Title, t.Note)
			} else {
				fmt.Fprintf(&b, "  %s %s\n", taskIcon(t.Status), t.Title)
			}
		}
	}

	decisions := s.Decisions
	if len(decisions) > 5 {
		decisions = decisions[len(decisions)-5:]
	}
	fmt.Fprintf(&b, "\nDECISIONS (%d):\n", len(s.Decisions))
	if len(decisions) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, d := range decisions {
			if d.Rationale != "" {
				fmt.Fprintf(&b, "  • %s [%s]\n", d.Summary, d.Rationale)
			} else {
				fmt.Fprintf(&b, "  • %s\n", d.Summary)
			}
		}
	}

	fmt.Fprintf(&b, "\nACTIVE FILES (%d):\n", len(s.ActiveFiles))
	if len(s.ActiveFiles) == 0 {
		b.WriteString("  (none)\n")
	} else {
		for _, f := range s.ActiveFiles {
			fmt.Fprintf(&b, "  [%s] %s\n", f.Role, f.Path)
		}
	}

	if s.ArchNotes != "" {
		b.WriteString("\nARCHITECTURE NOTES:\n")
		for _, line := range strings.Split(s.ArchNotes, "\n") {
			fmt.Fprintf(&b, "  %s\n", line)
		}
	}

	notes := s.Notes
	if len(notes) > 5 {
		notes = notes[len(notes)-5:]
	}
	if len(notes) > 0 {
		b.WriteString("\nNOTES:\n")
		for _, n := range notes {
			tag := ""
			if n.Tag != "" {
				tag = fmt.Sprintf("[%s] ", n.Tag)
			}
			fmt.Fprintf(&b, "  %s%s\n", tag, n.Content)
		}
	}

	if len(recentEvents) > 0 {
		b.WriteString("\nRECENT ACTIVITY:\n")
		for _, e := range recentEvents {
			fmt.Fprintf(&b, "  %s %s %s\n", fmtTime(e.Timestamp), eventIcon(e.Type), eventSummary(e))
		}
	}

	b.WriteString("=== END AIX CONTEXT ===\n")
	return b.String()
}

func filterTasks(tasks []session.Task, statuses ...session.TaskStatus) []session.Task {
	set := make(map[session.TaskStatus]bool)
	for _, s := range statuses {
		set[s] = true
	}
	var out []session.Task
	for _, t := range tasks {
		if set[t.Status] {
			out = append(out, t)
		}
	}
	return out
}

func taskIcon(status session.TaskStatus) string {
	switch status {
	case session.TaskOpen:
		return "[ ]"
	case session.TaskInProgress:
		return "[~]"
	case session.TaskDone:
		return "[x]"
	case session.TaskBlocked:
		return "[!]"
	default:
		return "[ ]"
	}
}

func eventIcon(t event.EventType) string {
	switch t {
	case event.EventSessionStarted, event.EventSessionResumed:
		return "[S]"
	case event.EventFileEdited, event.EventFileAdded:
		return "[F]"
	case event.EventDecisionAdded:
		return "[D]"
	case event.EventTaskAdded, event.EventTaskDone:
		return "[T]"
	case event.EventCheckpoint:
		return "[C]"
	case event.EventNoteAdded:
		return "[N]"
	case event.EventFocusChanged:
		return "[~]"
	default:
		return "[?]"
	}
}

func eventSummary(e event.Event) string {
	switch e.Type {
	case event.EventFileEdited:
		if p, ok := e.Data["path"]; ok {
			return "edited " + p
		}
	case event.EventFileAdded:
		if p, ok := e.Data["path"]; ok {
			return "tracking " + p
		}
	case event.EventTaskAdded:
		if t, ok := e.Data["title"]; ok {
			return "task: " + t
		}
	case event.EventTaskDone:
		if t, ok := e.Data["title"]; ok {
			return "done: " + t
		}
	case event.EventDecisionAdded:
		if s, ok := e.Data["summary"]; ok {
			return "decided: " + s
		}
	case event.EventCheckpoint:
		if m, ok := e.Data["message"]; ok {
			return "checkpoint: " + m
		}
	case event.EventSessionStarted:
		if n, ok := e.Data["name"]; ok {
			return "started: " + n
		}
	case event.EventSessionResumed:
		if n, ok := e.Data["name"]; ok {
			return "resumed: " + n
		}
	case event.EventNoteAdded:
		if c, ok := e.Data["content"]; ok {
			if len(c) > 50 {
				c = c[:50] + "..."
			}
			return "note: " + c
		}
	case event.EventFocusChanged:
		if t, ok := e.Data["to"]; ok {
			return "focus: " + t
		}
	}
	return string(e.Type)
}

func fmtTime(t time.Time) string {
	return t.Format("2006-01-02 15:04")
}
