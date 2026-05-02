package hook

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/vinhphuc13/aix/internal/event"
	"github.com/vinhphuc13/aix/internal/inject"
	"github.com/vinhphuc13/aix/internal/session"
)

// FindAIXDir walks up from dir looking for .aix/
func FindAIXDir(dir string) (string, error) {
	for {
		candidate := filepath.Join(dir, ".aix")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .aix directory found")
		}
		dir = parent
	}
}

// --- UserPromptSubmit ---

type promptInput struct {
	SessionID      string `json:"session_id"`
	HookEventName  string `json:"hook_event_name"`
	Prompt         string `json:"prompt"`
	CWD            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
}

type promptOutput struct {
	SystemMessage string `json:"systemMessage,omitempty"`
}

func HandlePrompt(stdin io.Reader, stdout io.Writer) error {
	var input promptInput
	_ = json.NewDecoder(stdin).Decode(&input)

	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	aixDir, err := FindAIXDir(cwd)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	s, err := session.LoadCurrent(aixDir)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	recentEvents, _ := event.ReadLast(aixDir, s.ID, 10)
	context := inject.RenderContext(s, recentEvents)

	out := promptOutput{SystemMessage: context}
	data, _ := json.Marshal(out)
	fmt.Fprintf(stdout, "%s\n", data)
	return nil
}

// --- PostToolUse ---

type postToolUseInput struct {
	SessionID      string          `json:"session_id"`
	HookEventName  string          `json:"hook_event_name"`
	ToolName       string          `json:"tool_name"`
	ToolInput      json.RawMessage `json:"tool_input"`
	ToolResponse   json.RawMessage `json:"tool_response"`
	CWD            string          `json:"cwd"`
	TranscriptPath string          `json:"transcript_path"`
}

type fileToolInput struct {
	FilePath string `json:"file_path"`
}

func HandlePostToolUse(stdin io.Reader, stdout io.Writer) error {
	var input postToolUseInput
	if err := json.NewDecoder(stdin).Decode(&input); err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	aixDir, err := FindAIXDir(cwd)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	s, err := session.LoadCurrent(aixDir)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	var filePath string
	switch input.ToolName {
	case "Edit", "Write", "MultiEdit":
		var ti fileToolInput
		if err := json.Unmarshal(input.ToolInput, &ti); err == nil {
			filePath = ti.FilePath
		}
	}

	if filePath == "" {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	// Make relative to project root
	projectRoot := filepath.Dir(aixDir)
	relPath := filePath
	if filepath.IsAbs(filePath) {
		if rel, err := filepath.Rel(projectRoot, filePath); err == nil && !filepath.IsAbs(rel) {
			relPath = rel
		}
	}

	now := time.Now()
	found := false
	for i := range s.ActiveFiles {
		if s.ActiveFiles[i].Path == relPath {
			s.ActiveFiles[i].LastEditAt = now
			found = true
			break
		}
	}

	evType := event.EventFileEdited
	if !found {
		s.ActiveFiles = append(s.ActiveFiles, session.ActiveFile{
			Path:       relPath,
			Role:       "primary",
			AddedAt:    now,
			LastEditAt: now,
		})
		evType = event.EventFileAdded
	}

	_ = session.Save(aixDir, s)
	_ = event.Append(aixDir, s.ID, evType, map[string]string{"path": relPath})

	recentEvents, _ := event.ReadLast(aixDir, s.ID, 10)
	_ = inject.WriteContextFile(aixDir, s, recentEvents)

	fmt.Fprintf(stdout, "{}\n")
	return nil
}

// --- Stop ---

type stopInput struct {
	SessionID      string `json:"session_id"`
	HookEventName  string `json:"hook_event_name"`
	Reason         string `json:"reason"`
	CWD            string `json:"cwd"`
	TranscriptPath string `json:"transcript_path"`
}

func HandleStop(stdin io.Reader, stdout io.Writer) error {
	var input stopInput
	_ = json.NewDecoder(stdin).Decode(&input)

	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	aixDir, err := FindAIXDir(cwd)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	s, err := session.LoadCurrent(aixDir)
	if err != nil {
		fmt.Fprintf(stdout, "{}\n")
		return nil
	}

	open, done := countTasks(s.Tasks)
	cp := session.Checkpoint{
		ID:        session.NewID(),
		Message:   "auto: session stopped",
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
	_ = event.Append(aixDir, s.ID, event.EventSessionStopped, map[string]string{
		"reason": input.Reason,
	})

	recentEvents, _ := event.ReadLast(aixDir, s.ID, 10)
	_ = inject.WriteContextFile(aixDir, s, recentEvents)

	fmt.Fprintf(stdout, "{}\n")
	return nil
}

func countTasks(tasks []session.Task) (open, done int) {
	for _, t := range tasks {
		if t.Status == session.TaskDone {
			done++
		} else {
			open++
		}
	}
	return
}
