package mcpserver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
)

func New(aixDir string) *server.MCPServer {
	s := server.NewMCPServer("aix", "0.1.0")

	s.AddTool(mcp.NewTool("aix_status",
		mcp.WithDescription("Get current aix session status: goal, open tasks, decisions, active files, last checkpoint."),
	), makeHandler(aixDir, handleStatus))

	s.AddTool(mcp.NewTool("aix_add_task",
		mcp.WithDescription("Add a task to the current session."),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("note", mcp.Description("Optional note about the task")),
	), makeHandler(aixDir, handleAddTask))

	s.AddTool(mcp.NewTool("aix_done",
		mcp.WithDescription("Mark a task as done by partial title match."),
		mcp.WithString("task", mcp.Required(), mcp.Description("Task title or partial title to match")),
	), makeHandler(aixDir, handleDone))

	s.AddTool(mcp.NewTool("aix_add_decision",
		mcp.WithDescription("Record an architectural or implementation decision."),
		mcp.WithString("summary", mcp.Required(), mcp.Description("Decision summary")),
		mcp.WithString("rationale", mcp.Description("Optional rationale")),
	), makeHandler(aixDir, handleAddDecision))

	s.AddTool(mcp.NewTool("aix_add_note",
		mcp.WithDescription("Add an engineering note to the current session."),
		mcp.WithString("content", mcp.Required(), mcp.Description("Note content")),
		mcp.WithString("tag", mcp.Description("Optional tag: arch, risk, todo, etc.")),
	), makeHandler(aixDir, handleAddNote))

	s.AddTool(mcp.NewTool("aix_checkpoint",
		mcp.WithDescription("Save the current session state as a named checkpoint."),
		mcp.WithString("message", mcp.Required(), mcp.Description("Checkpoint message describing what was done")),
	), makeHandler(aixDir, handleCheckpoint))

	s.AddTool(mcp.NewTool("aix_focus",
		mcp.WithDescription("Set the current focus for this session (shown at the top of injected context)."),
		mcp.WithString("focus", mcp.Required(), mcp.Description("What you are currently working on")),
	), makeHandler(aixDir, handleFocus))

	s.AddTool(mcp.NewTool("aix_list_sessions",
		mcp.WithDescription("List all aix sessions in this project. Use this when the user wants to switch features or resume a different session."),
	), makeSessionsHandler(aixDir, handleListSessions))

	s.AddTool(mcp.NewTool("aix_switch_session",
		mcp.WithDescription("Switch to a different session by name or ID. Call aix_list_sessions first to see available sessions."),
		mcp.WithString("session", mcp.Required(), mcp.Description("Session name or partial ID to switch to")),
	), makeSessionsHandler(aixDir, handleSwitchSession))

	return s
}

// makeHandler wraps each tool handler: loads session, calls fn, syncs context file.
func makeHandler(aixDir string, fn func(string, *session.Session, mcp.CallToolRequest) (string, error)) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		s, err := session.LoadCurrent(aixDir)
		if err != nil {
			return mcp.NewToolResultError("no active aix session: " + err.Error()), nil
		}
		msg, err := fn(aixDir, s, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
		_ = inject.WriteContextFile(aixDir, s, recentEvts)
		return mcp.NewToolResultText(msg), nil
	}
}

func handleStatus(aixDir string, s *session.Session, _ mcp.CallToolRequest) (string, error) {
	recentEvts, _ := event.ReadLast(aixDir, s.ID, 10)
	return inject.RenderContext(s, recentEvts), nil
}

func handleAddTask(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	title, err := req.RequireString("title")
	if err != nil {
		return "", fmt.Errorf("title is required")
	}
	note := req.GetString("note", "")
	t := session.Task{
		ID:        session.NewID(),
		Title:     title,
		Note:      note,
		Status:    session.TaskOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.Tasks = append(s.Tasks, t)
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventTaskAdded, map[string]string{
		"id": t.ID, "title": t.Title,
	})
	return fmt.Sprintf("[ ] %s", t.Title), nil
}

func handleDone(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	query, err := req.RequireString("task")
	if err != nil {
		return "", fmt.Errorf("task is required")
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
		return "", fmt.Errorf("open task not found matching %q", query)
	}
	s.Tasks[idx].Status = session.TaskDone
	s.Tasks[idx].UpdatedAt = time.Now()
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventTaskDone, map[string]string{
		"id": s.Tasks[idx].ID, "title": s.Tasks[idx].Title,
	})
	return fmt.Sprintf("[x] Done: %s", s.Tasks[idx].Title), nil
}

func handleAddDecision(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	summary, err := req.RequireString("summary")
	if err != nil {
		return "", fmt.Errorf("summary is required")
	}
	rationale := req.GetString("rationale", "")
	d := session.Decision{
		ID:        session.NewID(),
		Summary:   summary,
		Rationale: rationale,
		CreatedAt: time.Now(),
	}
	s.Decisions = append(s.Decisions, d)
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventDecisionAdded, map[string]string{"summary": d.Summary})
	if rationale != "" {
		return fmt.Sprintf("• %s [%s]", d.Summary, d.Rationale), nil
	}
	return fmt.Sprintf("• %s", d.Summary), nil
}

func handleAddNote(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	content, err := req.RequireString("content")
	if err != nil {
		return "", fmt.Errorf("content is required")
	}
	tag := req.GetString("tag", "")
	n := session.Note{
		ID:        session.NewID(),
		Content:   content,
		Tag:       tag,
		CreatedAt: time.Now(),
	}
	s.Notes = append(s.Notes, n)
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventNoteAdded, map[string]string{
		"content": n.Content, "tag": n.Tag,
	})
	if tag != "" {
		return fmt.Sprintf("Note [%s]: %s", tag, content), nil
	}
	return fmt.Sprintf("Note: %s", content), nil
}

func handleCheckpoint(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	message, err := req.RequireString("message")
	if err != nil {
		return "", fmt.Errorf("message is required")
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
		ID:        session.NewID(),
		Message:   message,
		CreatedAt: time.Now(),
		OpenTasks: open,
		DoneTasks: done,
	}
	s.Checkpoints = append(s.Checkpoints, cp)
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventCheckpoint, map[string]string{
		"message": cp.Message, "id": cp.ID,
	})
	return fmt.Sprintf("Checkpoint: %s (%d open, %d done)", message, open, done), nil
}

func handleFocus(aixDir string, s *session.Session, req mcp.CallToolRequest) (string, error) {
	focus, err := req.RequireString("focus")
	if err != nil {
		return "", fmt.Errorf("focus is required")
	}
	old := s.CurrentFocus
	s.CurrentFocus = focus
	if err := session.Save(aixDir, s); err != nil {
		return "", err
	}
	_ = event.Append(aixDir, s.ID, event.EventFocusChanged, map[string]string{
		"from": old, "to": focus,
	})
	return fmt.Sprintf("Focus: %s", focus), nil
}

// makeSessionsHandler is like makeHandler but does NOT pre-load the current
// session — list and switch operate on all sessions, not just the current one.
func makeSessionsHandler(aixDir string, fn func(string, mcp.CallToolRequest) (string, error)) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		msg, err := fn(aixDir, req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(msg), nil
	}
}

func handleListSessions(aixDir string, _ mcp.CallToolRequest) (string, error) {
	sessions, err := session.List(aixDir)
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "No sessions found.", nil
	}

	currentID := currentSessionID(aixDir)

	var b strings.Builder
	for _, s := range sessions {
		open := 0
		for _, t := range s.Tasks {
			if t.Status != session.TaskDone {
				open++
			}
		}
		marker := "  "
		if s.ID == currentID {
			marker = "* "
		}
		lastCP := ""
		if len(s.Checkpoints) > 0 {
			lastCP = fmt.Sprintf(" [%s]", s.Checkpoints[len(s.Checkpoints)-1].Message)
		}
		fmt.Fprintf(&b, "%s%s  %s  %d open tasks%s\n", marker, s.ID[:8], s.Name, open, lastCP)
	}
	b.WriteString("\n* = current session")
	return b.String(), nil
}

func handleSwitchSession(aixDir string, req mcp.CallToolRequest) (string, error) {
	query, err := req.RequireString("session")
	if err != nil {
		return "", fmt.Errorf("session is required")
	}

	sessions, err := session.List(aixDir)
	if err != nil {
		return "", err
	}

	for _, s := range sessions {
		if strings.HasPrefix(s.ID, query) || strings.EqualFold(s.Name, query) ||
			strings.Contains(strings.ToLower(s.Name), strings.ToLower(query)) {
			if err := session.SetCurrent(aixDir, s.ID); err != nil {
				return "", err
			}
			open := 0
			for _, t := range s.Tasks {
				if t.Status != session.TaskDone {
					open++
				}
			}
			return fmt.Sprintf("Switched to: %s (%s)\nGoal: %s\nOpen tasks: %d",
				s.Name, s.ID[:8], s.Goal, open), nil
		}
	}
	return "", fmt.Errorf("no session found matching %q — call aix_list_sessions to see available sessions", query)
}

func currentSessionID(aixDir string) string {
	data, err := os.ReadFile(filepath.Join(aixDir, "current"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// ProjectRoot returns the project root given the aix directory.
func ProjectRoot(aixDir string) string {
	return filepath.Dir(aixDir)
}
